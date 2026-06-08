package api

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
	"strings"
)

// @Summary 녹화 chunk 조회
// @Description 관제 UI가 조회하는 recording chunk와 파일 상태 목록을 최신순으로 최대 300개 반환합니다. missionCode를 지정하면 해당 임무 안에서 최신 300개를 반환합니다.
// @Tags Operator API
// @Produce json
// @Param missionCode query string false "조회할 임무 코드"
// @Success 200 {object} dto.OperatorRecordingsResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/operator/recordings [get]
func (s *Server) handleListRecordings(w http.ResponseWriter, r *http.Request) {
	missionCode := strings.TrimSpace(r.URL.Query().Get("missionCode"))
	if missionCode != "" {
		page, err := s.services.Recording.ListMissionRecordingChunks(r.Context(), store.MissionRecordingChunkQuery{
			MissionCode: missionCode,
			Limit:       300,
			Offset:      0,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, dto.OperatorRecordingsPayload(s.createOperatorRecordingResponses(page.Chunks)))
		return
	}

	recordings, err := s.services.Recording.ListRecordingChunks(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.OperatorRecordingsPayload(s.createOperatorRecordingResponses(recordings)))
}

func (s *Server) createOperatorRecordingResponses(recordings []domain.RecordingChunk) []dto.OperatorRecordingChunkResponse {
	response := make([]dto.OperatorRecordingChunkResponse, 0, len(recordings))
	for _, recording := range recordings {
		response = append(response, s.createOperatorRecordingResponse(recording))
	}
	return response
}

// @Summary 임무 녹화 요약 조회
// @Description 관제 UI 리플레이 화면이 특정 임무의 로봇별 녹화 청크 수와 파일 저장 상태를 요약 조회합니다.
// @Tags Operator API
// @Produce json
// @Param missionCode path string true "조회할 임무 코드"
// @Success 200 {object} dto.OperatorMissionRecordingSummaryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/operator/missions/{missionCode}/recordings/summary [get]
func (s *Server) handleMissionRecordingSummary(w http.ResponseWriter, r *http.Request) {
	missionCode := strings.TrimSpace(r.PathValue("missionCode"))
	if missionCode == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("missionCode is required"))
		return
	}
	summary, err := s.services.Recording.SummarizeMissionRecordings(r.Context(), missionCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, createMissionRecordingSummaryResponse(summary))
}

// @Summary 임무 녹화 chunk 페이지 조회
// @Description 관제 UI 리플레이 화면이 특정 임무의 녹화 chunk를 최신순으로 페이지 조회합니다. robotCode를 지정하면 해당 로봇의 chunk만 반환합니다.
// @Tags Operator API
// @Produce json
// @Param missionCode path string true "조회할 임무 코드"
// @Param robotCode query string false "조회할 로봇 코드"
// @Param limit query int false "페이지 크기. 기본 100, 최대 200"
// @Param offset query int false "조회 시작 위치. 기본 0"
// @Success 200 {object} dto.OperatorMissionRecordingChunksResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/operator/missions/{missionCode}/recordings/chunks [get]
func (s *Server) handleMissionRecordingChunks(w http.ResponseWriter, r *http.Request) {
	missionCode := strings.TrimSpace(r.PathValue("missionCode"))
	if missionCode == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("missionCode is required"))
		return
	}
	limit := clampIntValue(intQueryValue(r, "limit", 100), 1, 200)
	offset := nonNegativeIntQueryValue(r, "offset", 0)
	page, err := s.services.Recording.ListMissionRecordingChunks(r.Context(), store.MissionRecordingChunkQuery{
		MissionCode: missionCode,
		RobotCode:   strings.TrimSpace(r.URL.Query().Get("robotCode")),
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, s.createMissionRecordingChunksResponse(page))
}

func createMissionRecordingSummaryResponse(summary store.MissionRecordingSummary) dto.OperatorMissionRecordingSummaryResponse {
	robots := make([]dto.OperatorMissionRecordingRobotSummaryResponse, 0, len(summary.Robots))
	for _, robot := range summary.Robots {
		robots = append(robots, dto.OperatorMissionRecordingRobotSummaryResponse{
			RobotCode:            robot.RobotCode,
			ChunkCount:           robot.ChunkCount,
			UploadedChunkCount:   robot.UploadedChunkCount,
			RecordingChunkCount:  robot.RecordingChunkCount,
			FinalizingChunkCount: robot.FinalizingChunkCount,
			PartialChunkCount:    robot.PartialChunkCount,
			FirstStartedAt:       robot.FirstStartedAt,
			LastEndedAt:          robot.LastEndedAt,
			AvailableFileCounts:  copyIntMap(robot.AvailableFileCounts),
			MissingFileCounts:    copyIntMap(robot.MissingFileCounts),
		})
	}
	return dto.OperatorMissionRecordingSummaryResponse{
		MissionCode: summary.MissionCode,
		TotalChunks: summary.TotalChunks,
		Robots:      robots,
	}
}

func (s *Server) createMissionRecordingChunksResponse(page store.MissionRecordingChunkPage) dto.OperatorMissionRecordingChunksResponse {
	recordings := s.createOperatorRecordingResponses(page.Chunks)
	nextOffset := page.Offset + len(recordings)
	return dto.OperatorMissionRecordingChunksResponse{
		Recordings: recordings,
		Page: dto.OperatorRecordingPageResponse{
			Limit:      page.Limit,
			Offset:     page.Offset,
			Total:      page.Total,
			HasMore:    nextOffset < page.Total,
			NextOffset: nextOffset,
		},
	}
}

func copyIntMap(input map[string]int) map[string]int {
	if input == nil {
		return map[string]int{}
	}
	output := make(map[string]int, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
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
