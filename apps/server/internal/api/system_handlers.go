package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/service"
)

// @Summary 시스템 상태 조회
// @Description 관제 서비스, 녹화 서비스, 저장소, 실시간 연결 상태 요약을 반환합니다. scope=overview는 공통 상태바용 경량 응답이며, scope=full 또는 생략 시 시스템 관리 화면용 상세 응답을 반환합니다.
// @Tags 시스템 API
// @Produce json
// @Param scope query string false "응답 범위" Enums(full, overview) default(full)
// @Success 200 {object} dto.SystemStatusResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/system/status [get]
func (s *Server) handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	requestContext := r.Context()
	scope := normalizeSystemStatusScope(r.URL.Query().Get("scope"))
	robotCount, missionCount, recordingCount, err := s.readSystemSummaryCounts(requestContext, scope)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	sfuRooms := s.sfuHub.Summaries()
	writeJSON(w, http.StatusOK, dto.SystemStatus(dto.SystemStatusInput{
		Environment:               s.config.Environment,
		AppServerPublicURL:        s.config.AppServerPublicURL,
		RecorderWorkerInternalURL: s.config.RecorderWorkerInternalURL,
		MinIOInternalURL:          s.config.MinIOInternalURL,
		MinIOPublicURL:            s.config.MinIOPublicURL,
		MinIOBucket:               s.config.MinIOBucket,
		RecorderWorkerStatus:      s.componentHTTPStatus(requestContext, s.config.RecorderWorkerInternalURL+"/healthz"),
		ObjectStorage:             s.readObjectStorageStatus(requestContext, scope),
		Database:                  s.readDatabaseStatus(requestContext, scope),
		RecorderRuntime:           s.readRecorderRuntimeStatus(requestContext, scope),
		RobotCount:                robotCount,
		MissionCount:              missionCount,
		RecordingCount:            recordingCount,
		SFURooms:                  sfuRooms,
	}))
}

const (
	systemStatusScopeFull     = "full"
	systemStatusScopeOverview = "overview"
)

func normalizeSystemStatusScope(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case systemStatusScopeOverview:
		return systemStatusScopeOverview
	default:
		return systemStatusScopeFull
	}
}

func (s *Server) readSystemSummaryCounts(ctx context.Context, scope string) (int, int, int, error) {
	if scope == systemStatusScopeOverview {
		return 0, 0, 0, nil
	}
	robots, err := s.services.Robots.ListRobots(ctx)
	if err != nil {
		return 0, 0, 0, err
	}
	missions, err := s.services.Missions.ListMissions(ctx)
	if err != nil {
		return 0, 0, 0, err
	}
	recordings, err := s.services.Recording.ListRecordingChunks(ctx)
	if err != nil {
		return 0, 0, 0, err
	}
	return len(robots), len(missions), len(recordings), nil
}

func (s *Server) readObjectStorageStatus(ctx context.Context, scope string) dto.ObjectStorageStatusResponse {
	if scope == systemStatusScopeOverview {
		return dto.ObjectStorageSkipped(s.config.MinIOBucket)
	}
	if s.services.Storage == nil {
		return dto.ObjectStorageUnavailable(s.config.MinIOBucket, nil)
	}
	usage, err := s.services.Storage.GetObjectStorageUsage(ctx)
	if err != nil {
		return dto.ObjectStorageUnavailable(s.config.MinIOBucket, err)
	}
	return dto.ObjectStorageStatus(usage)
}

func (s *Server) readDatabaseStatus(ctx context.Context, scope string) dto.DatabaseStatusResponse {
	if scope == systemStatusScopeOverview {
		return dto.DatabaseSkipped()
	}
	usage, err := s.services.System.GetDatabaseUsage(ctx)
	if err != nil {
		return dto.DatabaseUnavailable(err)
	}
	return dto.DatabaseStatus(usage)
}

func (s *Server) readRecorderRuntimeStatus(ctx context.Context, scope string) domain.RecorderRuntimeStatus {
	if scope == systemStatusScopeOverview {
		return domain.RecorderRuntimeStatus{Status: "skipped"}
	}
	if s.services.RecorderRuntime == nil {
		return domain.RecorderRuntimeStatus{Status: "unavailable", Error: "recorder runtime admin service is not configured"}
	}
	status, err := s.services.RecorderRuntime.GetRecorderRuntimeStatus(ctx)
	if err != nil {
		return domain.RecorderRuntimeStatus{Status: "unavailable", Error: err.Error()}
	}
	return status
}

// @Summary 객체 스토리지 초기화
// @Description 확인 문자열을 받은 뒤 테스트용 객체 스토리지 데이터를 정리합니다. production 환경이거나 녹화 런타임 상태가 정리 가능하지 않으면 실행되지 않습니다.
// @Tags 시스템 API
// @Accept json
// @Produce json
// @Param request body dto.ClearObjectStorageRequest true "객체 스토리지 초기화 요청"
// @Success 200 {object} dto.ClearObjectStorageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/system/object-storage/clear [post]
func (s *Server) handleClearObjectStorage(w http.ResponseWriter, r *http.Request) {
	var request dto.ClearObjectStorageRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := s.services.Storage.ValidateClearObjectStorageRequest(request.Confirmation); err != nil {
		writeSystemActionError(w, err)
		return
	}
	if err := s.ensureObjectStorageClearable(r.Context()); err != nil {
		writeSystemActionError(w, err)
		return
	}

	result, err := s.services.Storage.ClearObjectStorage(r.Context(), request.Confirmation)
	if err != nil {
		writeSystemActionError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ClearObjectStorageResponse{
		ObjectStorage: result,
	})
}

// @Summary 객체 스토리지 운영 중 정리
// @Description 진행 중인 녹화와 마무리 작업에 연결된 파일은 제외하고, 완료/실패 상태의 녹화 파일과 파일 상태 메타데이터만 정리합니다. production 환경에서는 실행되지 않습니다.
// @Tags 시스템 API
// @Accept json
// @Produce json
// @Param request body dto.PruneObjectStorageRequest true "객체 스토리지 운영 중 정리 요청"
// @Success 200 {object} dto.PruneObjectStorageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/system/object-storage/prune [post]
func (s *Server) handlePruneObjectStorage(w http.ResponseWriter, r *http.Request) {
	var request dto.PruneObjectStorageRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result, err := s.services.Storage.PruneObjectStorage(r.Context(), request.Confirmation)
	if err != nil {
		writeSystemActionError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.PruneObjectStorageResponse{
		ObjectStorage: result,
	})
}

// @Summary 센서 데이터 초기화
// @Description 확인 문자열을 받은 뒤 테스트용 센서 정의와 센서값 데이터를 정리합니다. production 환경에서는 실행되지 않습니다.
// @Tags 시스템 API
// @Accept json
// @Produce json
// @Param request body dto.ClearSensorDataRequest true "센서 데이터 초기화 요청"
// @Success 200 {object} dto.ClearSensorDataResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/system/sensors/clear [post]
func (s *Server) handleClearSensorData(w http.ResponseWriter, r *http.Request) {
	var request dto.ClearSensorDataRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result, err := s.services.Sensors.ClearSensorData(r.Context(), s.config.Environment, request.Confirmation)
	if err != nil {
		writeSystemActionError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ClearSensorDataResponse{
		SensorData: result,
	})
}

// @Summary 이벤트 데이터 초기화
// @Description 확인 문자열을 받은 뒤 테스트용 임무 이벤트와 객체 탐지 데이터를 정리합니다. production 환경에서는 실행되지 않습니다.
// @Tags 시스템 API
// @Accept json
// @Produce json
// @Param request body dto.ClearEventDataRequest true "이벤트 데이터 초기화 요청"
// @Success 200 {object} dto.ClearEventDataResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/system/events/clear [post]
func (s *Server) handleClearEventData(w http.ResponseWriter, r *http.Request) {
	var request dto.ClearEventDataRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result, err := s.services.Events.ClearEventData(r.Context(), s.config.Environment, request.Confirmation)
	if err != nil {
		writeSystemActionError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ClearEventDataResponse{
		EventData: result,
	})
}

// @Summary 녹화 런타임 초기화
// @Description 확인 문자열을 받은 뒤 녹화 서비스의 로컬 임시 파일을 정리합니다. 진행 중인 녹화 상태가 있거나 production 환경이면 실행되지 않습니다.
// @Tags 시스템 API
// @Accept json
// @Produce json
// @Param request body dto.ClearRecorderRuntimeRequest true "녹화 런타임 초기화 요청"
// @Success 200 {object} dto.ClearRecorderRuntimeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/system/recorder-runtime/clear [post]
func (s *Server) handleClearRecorderRuntime(w http.ResponseWriter, r *http.Request) {
	var request dto.ClearRecorderRuntimeRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result, err := s.services.RecorderRuntime.ClearRecorderRuntime(r.Context(), request.Confirmation)
	if err != nil {
		writeSystemActionError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ClearRecorderRuntimeResponse{
		RecorderRuntime: result,
	})
}

// @Summary 녹화 런타임 운영 중 정리
// @Description 녹화 서비스 로컬 임시 파일 중 현재 작성 중인 chunk와 마무리 대기 chunk는 제외하고 오래된 런타임 파일만 정리합니다. production 환경에서는 실행되지 않습니다.
// @Tags 시스템 API
// @Accept json
// @Produce json
// @Param request body dto.PruneRecorderRuntimeRequest true "녹화 런타임 운영 중 정리 요청"
// @Success 200 {object} dto.PruneRecorderRuntimeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/system/recorder-runtime/prune [post]
func (s *Server) handlePruneRecorderRuntime(w http.ResponseWriter, r *http.Request) {
	var request dto.PruneRecorderRuntimeRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result, err := s.services.RecorderRuntime.PruneRecorderRuntime(r.Context(), request.Confirmation)
	if err != nil {
		writeSystemActionError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.PruneRecorderRuntimeResponse{
		RecorderRuntime: result,
	})
}

func (s *Server) ensureObjectStorageClearable(ctx context.Context) error {
	if s.services.RecorderRuntime == nil {
		return fmt.Errorf("%w: recorder runtime status is unavailable", service.ErrSystemActionConflict)
	}
	status, err := s.services.RecorderRuntime.GetRecorderRuntimeStatus(ctx)
	if err != nil {
		return fmt.Errorf("%w: recorder runtime status is unavailable", service.ErrSystemActionConflict)
	}
	if status.Status != "ok" || !status.Clearable {
		return fmt.Errorf("%w: recorder runtime is not clearable", service.ErrSystemActionConflict)
	}
	return nil
}

func writeSystemActionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrSystemActionForbidden):
		writeError(w, http.StatusForbidden, err)
	case errors.Is(err, service.ErrSystemActionConfirmationRequired):
		writeError(w, http.StatusBadRequest, err)
	case errors.Is(err, service.ErrSystemActionConflict):
		writeError(w, http.StatusConflict, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func (s *Server) componentHTTPStatus(ctx context.Context, targetURL string) string {
	if strings.TrimSpace(targetURL) == "" {
		return "unknown"
	}
	client := http.Client{Timeout: 500 * time.Millisecond}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "unreachable"
	}
	response, err := client.Do(request)
	if err != nil {
		return "unreachable"
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		return "ok"
	}
	return "degraded"
}
