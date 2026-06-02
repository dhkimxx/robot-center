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

func (s *Server) handleRecordingTargets(w http.ResponseWriter, r *http.Request) {
	targets, err := s.services.Missions.RecordingTargets(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.RecordingTargetsPayload(targets))
}

func (s *Server) handleListRecordings(w http.ResponseWriter, r *http.Request) {
	recordings, err := s.services.Recording.ListRecordingChunks(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	response := make([]dto.RecordingChunkResponse, 0, len(recordings))
	for _, recording := range recordings {
		response = append(response, s.createRecordingResponse(recording))
	}
	writeJSON(w, http.StatusOK, dto.RecordingsPayload(response))
}

func (s *Server) createRecordingResponse(recording domain.RecordingChunk) dto.RecordingChunkResponse {
	response := dto.RecordingChunk(recording)
	response.Files = []dto.RecordingFileResponse{
		s.createRecordingFileResponse(recording, "rgb_audio_mp4", "RGB MP4", "video/mp4", recording.MediaObjectKeys["rgbMp4"], recording.AvailableFileTypes["rgb_audio_mp4"]),
		s.createRecordingFileResponse(recording, "thermal_mp4", "Thermal MP4", "video/mp4", recording.MediaObjectKeys["thermal"], recording.AvailableFileTypes["thermal_mp4"]),
		s.createRecordingFileResponse(recording, "sensor_jsonl", "Sensor JSONL", "application/x-ndjson", recording.MediaObjectKeys["sensor"], recording.AvailableFileTypes["sensor_jsonl"]),
		s.createRecordingFileResponse(recording, "telemetry_jsonl", "Telemetry/GPS JSONL", "application/x-ndjson", recording.MediaObjectKeys["telemetry"], recording.AvailableFileTypes["telemetry_jsonl"]),
		s.createRecordingFileResponse(recording, "manifest", "저장 메타데이터", "application/json", recording.ManifestObjectKey, recording.AvailableFileTypes["manifest"] || recording.Status == "uploaded"),
	}
	return response
}

func (s *Server) createRecordingFileResponse(recording domain.RecordingChunk, fileType string, label string, contentType string, objectKey string, available bool) dto.RecordingFileResponse {
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
	return dto.RecordingFileResponse{
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

	publicURL, publicErr := url.Parse(s.config.PublicURL)
	minioURL, minioErr := url.Parse(s.config.MinIOEndpoint)

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
	bucket := strings.TrimSpace(s.config.MinIOBucket)
	if bucket == "" {
		bucket = "robot-center"
	}
	encodedPath := encodeObjectPath(bucket + "/" + objectKey)
	return fmt.Sprintf("%s://%s/%s", scheme, net.JoinHostPort(host, port), encodedPath)
}

func encodeObjectPath(path string) string {
	segments := strings.Split(path, "/")
	for index, segment := range segments {
		segments[index] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}

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

	writeJSON(w, http.StatusOK, dto.RecordingTick(result))
}

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
	writeJSON(w, http.StatusOK, dto.RecordingChunkPayload(s.createRecordingResponse(chunk)))
}

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
	writeJSON(w, http.StatusOK, dto.RecordingChunkPayload(s.createRecordingResponse(chunk)))
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
