package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/service"
)

// @Summary 시스템 상태 조회
// @Description 관제 서비스, 녹화 서비스, 저장소, 실시간 연결 상태 요약을 반환합니다.
// @Tags 시스템 API
// @Produce json
// @Success 200 {object} dto.SystemStatusResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/system/status [get]
func (s *Server) handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	requestContext := r.Context()
	robots, err := s.services.Robots.ListRobots(requestContext)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	missions, err := s.services.Missions.ListMissions(requestContext)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	recordings, err := s.services.Recording.ListRecordingChunks(requestContext)
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
		ObjectStorage:             s.readObjectStorageStatus(requestContext),
		Database:                  s.readDatabaseStatus(requestContext),
		RecorderRuntime:           s.readRecorderRuntimeStatus(requestContext),
		RobotCount:                len(robots),
		MissionCount:              len(missions),
		RecordingCount:            len(recordings),
		SFURooms:                  sfuRooms,
	}))
}

func (s *Server) readObjectStorageStatus(ctx context.Context) dto.ObjectStorageStatusResponse {
	if s.services.Storage == nil {
		return dto.ObjectStorageUnavailable(s.config.MinIOBucket, nil)
	}
	usage, err := s.services.Storage.GetObjectStorageUsage(ctx)
	if err != nil {
		return dto.ObjectStorageUnavailable(s.config.MinIOBucket, err)
	}
	return dto.ObjectStorageStatus(usage)
}

func (s *Server) readDatabaseStatus(ctx context.Context) dto.DatabaseStatusResponse {
	usage, err := s.services.System.GetDatabaseUsage(ctx)
	if err != nil {
		return dto.DatabaseUnavailable(err)
	}
	return dto.DatabaseStatus(usage)
}

func (s *Server) readRecorderRuntimeStatus(ctx context.Context) domain.RecorderRuntimeStatus {
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
// @Description 확인 문자열을 받은 뒤 테스트용 객체 스토리지 데이터를 정리합니다. production 환경에서는 실행되지 않습니다.
// @Tags 시스템 API
// @Accept json
// @Produce json
// @Param request body dto.ClearObjectStorageRequest true "객체 스토리지 초기화 요청"
// @Success 200 {object} dto.ClearObjectStorageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/system/object-storage/clear [post]
func (s *Server) handleClearObjectStorage(w http.ResponseWriter, r *http.Request) {
	var request dto.ClearObjectStorageRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
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
