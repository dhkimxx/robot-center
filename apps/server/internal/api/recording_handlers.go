package api

import (
	"encoding/json"
	"io"
	"net/http"
	"robot-center/apps/server/internal/api/dto"
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
