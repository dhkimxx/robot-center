package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
	"strings"
	"time"
)

// @Summary 녹화 대상 임무 조회
// @Description recorder-worker가 구독해야 하는 active mission 목록을 반환합니다.
// @Tags Recorder API
// @Produce json
// @Success 200 {object} dto.RecorderRecordingTargetsResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/recorder/recording-targets [get]
func (s *Server) handleRecordingTargets(w http.ResponseWriter, r *http.Request) {
	targets, err := s.services.Missions.RecordingTargets(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.RecorderRecordingTargetsPayload(targets))
}

// @Summary 녹화 chunk 조회
// @Description 관제 UI가 조회하는 recording chunk와 파일 상태 목록을 반환합니다. missionCode를 지정하면 해당 임무의 chunk만 반환합니다.
// @Tags Operator API
// @Produce json
// @Param missionCode query string false "조회할 임무 코드"
// @Success 200 {object} dto.OperatorRecordingsResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/operator/recordings [get]
func (s *Server) handleListRecordings(w http.ResponseWriter, r *http.Request) {
	recordings, err := s.services.Recording.ListRecordingChunks(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	missionCode := strings.TrimSpace(r.URL.Query().Get("missionCode"))
	response := make([]dto.OperatorRecordingChunkResponse, 0, len(recordings))
	for _, recording := range recordings {
		if missionCode != "" && recording.MissionCode != missionCode {
			continue
		}
		response = append(response, s.createOperatorRecordingResponse(recording))
	}
	writeJSON(w, http.StatusOK, dto.OperatorRecordingsPayload(response))
}

func (s *Server) createOperatorRecordingResponse(recording domain.RecordingChunk) dto.OperatorRecordingChunkResponse {
	response := dto.OperatorRecordingChunk(recording)
	response.Status = normalizeRecordingResponseStatus(recording)
	response.Files = []dto.OperatorRecordingFileResponse{
		s.createOperatorRecordingFileResponse(recording, "rgb_audio_mp4", "RGB MP4", "video/mp4", recording.MediaObjectKeys["rgbMp4"], recording.AvailableFileTypes["rgb_audio_mp4"]),
		s.createOperatorRecordingFileResponse(recording, "thermal_mp4", "Thermal MP4", "video/mp4", recording.MediaObjectKeys["thermal"], recording.AvailableFileTypes["thermal_mp4"]),
		s.createOperatorRecordingFileResponse(recording, "sensor_jsonl", "Sensor JSONL", "application/x-ndjson", recording.MediaObjectKeys["sensor"], recording.AvailableFileTypes["sensor_jsonl"]),
		s.createOperatorRecordingFileResponse(recording, "telemetry_jsonl", "Telemetry/GPS JSONL", "application/x-ndjson", recording.MediaObjectKeys["telemetry"], recording.AvailableFileTypes["telemetry_jsonl"]),
		s.createOperatorRecordingFileResponse(recording, "manifest", "저장 메타데이터", "application/json", recording.ManifestObjectKey, recording.AvailableFileTypes["manifest"] || recording.Status == "uploaded"),
	}
	return response
}

func normalizeRecordingResponseStatus(recording domain.RecordingChunk) string {
	if recording.Status == "uploaded" && !hasAvailableRecordingVideo(recording) {
		return "partial"
	}
	return recording.Status
}

func hasAvailableRecordingVideo(recording domain.RecordingChunk) bool {
	return recording.AvailableFileTypes["rgb_audio_mp4"] || recording.AvailableFileTypes["thermal_mp4"]
}

func (s *Server) createOperatorRecordingFileResponse(recording domain.RecordingChunk, fileType string, label string, contentType string, objectKey string, available bool) dto.OperatorRecordingFileResponse {
	status := "planned"
	fileURL := ""
	if available {
		status = "available"
		fileURL = s.createStorageObjectURL(objectKey)
	} else if recording.Status == "recording" || recording.Status == "pending" {
		status = "recording"
	} else if recording.Status == "finalizing" {
		status = "finalizing"
	} else if recording.Status == "partial" || recording.Status == "stopped" {
		status = "partial"
	} else if recording.Status == "failed" {
		status = "failed"
	}
	return dto.OperatorRecordingFileResponse{
		Type:        fileType,
		Label:       label,
		Status:      status,
		ContentType: contentType,
		ObjectKey:   objectKey,
		URL:         fileURL,
	}
}

func (s *Server) createStorageObjectURL(objectKey string) string {
	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" {
		return ""
	}

	publicMinIOURL := strings.TrimSpace(s.config.MinIOPublicURL)
	if publicMinIOURL != "" {
		fileURL := createObjectStorageURL(publicMinIOURL, s.config.MinIOBucket, objectKey)
		if fileURL != "" {
			return fileURL
		}
	}
	return createObjectStorageURL(s.createLegacyStoragePublicBaseURL(), s.config.MinIOBucket, objectKey)
}

func (s *Server) createLegacyStoragePublicBaseURL() string {
	publicURL, publicErr := url.Parse(s.config.AppServerPublicURL)
	minioURL, minioErr := url.Parse(s.config.MinIOInternalURL)

	scheme := "http"
	if publicErr == nil && publicURL.Scheme != "" {
		scheme = publicURL.Scheme
	} else if minioErr == nil && minioURL.Scheme != "" {
		scheme = minioURL.Scheme
	}

	host := ""
	if publicErr == nil {
		host = publicURL.Hostname()
	}
	if host == "" && minioErr == nil {
		host = minioURL.Hostname()
	}
	if host == "" {
		host = "localhost"
	}

	port := "9000"
	if minioErr == nil && minioURL.Port() != "" {
		port = minioURL.Port()
	}
	return fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(host, port))
}

func createObjectStorageURL(baseURL string, bucket string, objectKey string) string {
	parsedURL, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return ""
	}
	bucket = strings.TrimSpace(bucket)
	if bucket == "" {
		bucket = "robot-center"
	}
	pathSegments := []string{}
	if basePath := strings.Trim(parsedURL.Path, "/"); basePath != "" {
		pathSegments = append(pathSegments, strings.Split(basePath, "/")...)
	}
	pathSegments = append(pathSegments, bucket)
	if trimmedObjectKey := strings.Trim(objectKey, "/"); trimmedObjectKey != "" {
		pathSegments = append(pathSegments, strings.Split(trimmedObjectKey, "/")...)
	}
	parsedURL.Path = "/" + strings.Join(pathSegments, "/")
	parsedURL.RawPath = ""
	parsedURL.RawQuery = ""
	parsedURL.Fragment = ""
	return parsedURL.String()
}

// @Summary 녹화 tick 반영
// @Description recorder-worker가 mission/robot 기준 chunk 생성을 요청합니다.
// @Tags Recorder API
// @Accept json
// @Produce json
// @Param request body dto.RecorderTickRequest true "녹화 tick 요청"
// @Success 200 {object} dto.RecorderRecordingTickResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/v1/recorder/tick [post]
func (s *Server) handleRecorderTick(w http.ResponseWriter, r *http.Request) {
	var request dto.RecorderTickRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result, err := s.services.Recording.ApplyRecordingTick(r.Context(), store.RecordingTickInput{
		MissionCode:          strings.TrimSpace(request.MissionCode),
		RobotCode:            strings.TrimSpace(request.RobotCode),
		ChunkDurationSeconds: request.ChunkDurationSeconds,
		TickAt:               request.TickAt,
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.RecorderRecordingTick(result))
}

// @Summary 녹화 finalization job claim
// @Description recorder-worker가 처리할 finalization job을 claim합니다.
// @Tags Recorder API
// @Accept json
// @Produce json
// @Param request body dto.RecorderFinalizationClaimRequest true "finalization job claim 요청"
// @Success 200 {object} dto.RecorderFinalizationJobsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/v1/recorder/finalization-jobs/claim [post]
func (s *Server) handleRecorderFinalizationJobsClaim(w http.ResponseWriter, r *http.Request) {
	var request dto.RecorderFinalizationClaimRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	jobs, err := s.services.Recording.ClaimFinalizationJobs(
		r.Context(),
		request.WorkerID,
		request.Limit,
		time.Duration(request.LockDurationSeconds)*time.Second,
	)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.RecorderFinalizationJobsPayload(jobs))
}

// @Summary 녹화 finalization job 완료
// @Description recorder-worker가 finalization job 완료를 보고합니다.
// @Tags Recorder API
// @Accept json
// @Produce json
// @Param jobID path string true "finalization job ID"
// @Param request body dto.RecorderFinalizationStatusRequest true "finalization job 상태 요청"
// @Success 200 {object} dto.OKResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/v1/recorder/finalization-jobs/{jobID}/completed [post]
func (s *Server) handleRecorderFinalizationJobCompleted(w http.ResponseWriter, r *http.Request) {
	request, err := decodeRecorderFinalizationStatus(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.services.Recording.MarkFinalizationJobCompleted(r.Context(), r.PathValue("jobID"), request.WorkerID, request.Attempt); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.OKPayload())
}

// @Summary 녹화 finalization job 부분 완료
// @Description recorder-worker가 finalization job 부분 완료를 보고합니다.
// @Tags Recorder API
// @Accept json
// @Produce json
// @Param jobID path string true "finalization job ID"
// @Param request body dto.RecorderFinalizationStatusRequest true "finalization job 상태 요청"
// @Success 200 {object} dto.OKResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/v1/recorder/finalization-jobs/{jobID}/partial [post]
func (s *Server) handleRecorderFinalizationJobPartial(w http.ResponseWriter, r *http.Request) {
	request, err := decodeRecorderFinalizationStatus(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.services.Recording.MarkFinalizationJobPartial(r.Context(), r.PathValue("jobID"), request.WorkerID, request.Attempt, request.Reason); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.OKPayload())
}

// @Summary 녹화 finalization job 실패
// @Description recorder-worker가 finalization job 실패를 보고합니다.
// @Tags Recorder API
// @Accept json
// @Produce json
// @Param jobID path string true "finalization job ID"
// @Param request body dto.RecorderFinalizationStatusRequest true "finalization job 상태 요청"
// @Success 200 {object} dto.OKResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/v1/recorder/finalization-jobs/{jobID}/failed [post]
func (s *Server) handleRecorderFinalizationJobFailed(w http.ResponseWriter, r *http.Request) {
	request, err := decodeRecorderFinalizationStatus(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.services.Recording.MarkFinalizationJobFailed(r.Context(), r.PathValue("jobID"), request.WorkerID, request.Attempt, request.Reason); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.OKPayload())
}

// @Summary 녹화 chunk 업로드 완료
// @Description recorder-worker가 chunk manifest 업로드 완료를 보고합니다.
// @Tags Recorder API
// @Accept json
// @Produce json
// @Param chunkID path string true "recording chunk ID"
// @Param request body dto.RecorderUploadRequest false "업로드 메타데이터"
// @Success 200 {object} dto.RecorderRecordingChunkEnvelopeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/v1/recorder/chunks/{chunkID}/uploaded [post]
func (s *Server) handleRecorderChunkUploaded(w http.ResponseWriter, r *http.Request) {
	uploadMetadata, err := decodeRecorderUploadMetadata(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	chunk, err := s.services.Recording.MarkRecordingChunkUploaded(r.Context(), r.PathValue("chunkID"), uploadMetadata)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.RecorderRecordingChunkPayload(dto.RecorderRecordingChunk(chunk)))
}

// @Summary 녹화 파일 업로드 완료
// @Description recorder-worker가 chunk의 개별 파일 업로드 완료를 보고합니다.
// @Tags Recorder API
// @Accept json
// @Produce json
// @Param chunkID path string true "recording chunk ID"
// @Param fileType path string true "업로드된 파일 타입"
// @Param request body dto.RecorderUploadRequest false "업로드 메타데이터"
// @Success 200 {object} dto.RecorderRecordingChunkEnvelopeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/v1/recorder/chunks/{chunkID}/files/{fileType}/uploaded [post]
func (s *Server) handleRecorderFileUploaded(w http.ResponseWriter, r *http.Request) {
	uploadMetadata, err := decodeRecorderUploadMetadata(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	chunk, err := s.services.Recording.MarkRecordingFileUploaded(r.Context(), r.PathValue("chunkID"), r.PathValue("fileType"), uploadMetadata)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.RecorderRecordingChunkPayload(dto.RecorderRecordingChunk(chunk)))
}

func decodeRecorderUploadMetadata(r *http.Request) (store.RecordingUploadMetadata, error) {
	defer r.Body.Close()
	rawPayload, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return store.RecordingUploadMetadata{}, err
	}
	if len(strings.TrimSpace(string(rawPayload))) == 0 {
		return store.RecordingUploadMetadata{}, nil
	}

	var request dto.RecorderUploadRequest
	if err := json.Unmarshal(rawPayload, &request); err != nil {
		return store.RecordingUploadMetadata{}, err
	}
	return store.RecordingUploadMetadata{
		SizeBytes: request.SizeBytes,
		Checksum:  strings.TrimSpace(request.Checksum),
		WorkerID:  strings.TrimSpace(request.WorkerID),
		Attempt:   request.Attempt,
	}, nil
}

func decodeRecorderFinalizationStatus(r *http.Request) (dto.RecorderFinalizationStatusRequest, error) {
	defer r.Body.Close()
	rawPayload, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return dto.RecorderFinalizationStatusRequest{}, err
	}
	if len(strings.TrimSpace(string(rawPayload))) == 0 {
		return dto.RecorderFinalizationStatusRequest{}, nil
	}

	var request dto.RecorderFinalizationStatusRequest
	if err := json.Unmarshal(rawPayload, &request); err != nil {
		return dto.RecorderFinalizationStatusRequest{}, err
	}
	request.WorkerID = strings.TrimSpace(request.WorkerID)
	request.Reason = strings.TrimSpace(request.Reason)
	return request, nil
}
