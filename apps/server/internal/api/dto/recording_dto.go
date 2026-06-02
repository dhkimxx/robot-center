package dto

import (
	"time"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/utils"
)

type RecordingChunkResponse struct {
	ID                 string                  `json:"id"`
	RecordingSessionID string                  `json:"recordingSessionId"`
	MissionID          string                  `json:"missionId"`
	MissionCode        string                  `json:"missionCode"`
	RobotCode          string                  `json:"robotCode"`
	ChunkIndex         int                     `json:"chunkIndex"`
	Status             string                  `json:"status"`
	StartedAt          time.Time               `json:"startedAt"`
	EndedAt            time.Time               `json:"endedAt"`
	DurationSeconds    int                     `json:"durationSeconds"`
	ManifestObjectKey  string                  `json:"manifestObjectKey"`
	MediaObjectKeys    map[string]string       `json:"mediaObjectKeys"`
	AvailableFileTypes map[string]bool         `json:"availableFileTypes,omitempty"`
	CreatedAt          time.Time               `json:"createdAt"`
	UpdatedAt          time.Time               `json:"updatedAt"`
	Files              []RecordingFileResponse `json:"files,omitempty"`
}

type RecordingFileResponse struct {
	Type        string `json:"type"`
	Label       string `json:"label"`
	Status      string `json:"status"`
	ContentType string `json:"contentType"`
	ObjectKey   string `json:"objectKey,omitempty"`
	URL         string `json:"url,omitempty"`
}

type RecordingTickResponse struct {
	Chunk    RecordingChunkResponse `json:"chunk"`
	Manifest map[string]any         `json:"manifest"`
}

type RecordingTargetsResponse struct {
	Targets []MissionResponse `json:"targets"`
}

type RecordingsResponse struct {
	Recordings []RecordingChunkResponse `json:"recordings"`
}

type RecordingChunkEnvelopeResponse struct {
	Chunk RecordingChunkResponse `json:"chunk"`
}

type RecordingFinalizationJobResponse struct {
	ID                 string                 `json:"id"`
	RecordingChunkID   string                 `json:"recordingChunkId"`
	RecordingSessionID string                 `json:"recordingSessionId"`
	MissionID          string                 `json:"missionId"`
	RobotID            string                 `json:"robotId"`
	Status             string                 `json:"status"`
	Reason             string                 `json:"reason,omitempty"`
	Attempts           int                    `json:"attempts"`
	LockedBy           string                 `json:"lockedBy,omitempty"`
	LockedUntil        *time.Time             `json:"lockedUntil,omitempty"`
	LastError          string                 `json:"lastError,omitempty"`
	CreatedAt          time.Time              `json:"createdAt"`
	UpdatedAt          time.Time              `json:"updatedAt"`
	CompletedAt        *time.Time             `json:"completedAt,omitempty"`
	Chunk              RecordingChunkResponse `json:"chunk"`
}

type RecorderFinalizationJobsResponse struct {
	Jobs []RecordingFinalizationJobResponse `json:"jobs"`
}

type RecorderTickRequest struct {
	MissionCode          string    `json:"missionCode"`
	RobotCode            string    `json:"robotCode"`
	ChunkDurationSeconds int       `json:"chunkDurationSeconds"`
	TickAt               time.Time `json:"tickAt"`
}

type RecorderUploadRequest struct {
	SizeBytes *int64 `json:"sizeBytes,omitempty"`
	Checksum  string `json:"checksum,omitempty"`
	WorkerID  string `json:"workerId,omitempty"`
	Attempt   int    `json:"attempt,omitempty"`
}

type RecorderFinalizationClaimRequest struct {
	WorkerID            string `json:"workerId"`
	Limit               int    `json:"limit"`
	LockDurationSeconds int    `json:"lockDurationSeconds"`
}

type RecorderFinalizationStatusRequest struct {
	WorkerID string `json:"workerId,omitempty"`
	Attempt  int    `json:"attempt,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

func RecordingChunk(chunk domain.RecordingChunk) RecordingChunkResponse {
	return RecordingChunkResponse{
		ID:                 chunk.ID,
		RecordingSessionID: chunk.RecordingSessionID,
		MissionID:          chunk.MissionID,
		MissionCode:        chunk.MissionCode,
		RobotCode:          chunk.RobotCode,
		ChunkIndex:         chunk.ChunkIndex,
		Status:             chunk.Status,
		StartedAt:          chunk.StartedAt,
		EndedAt:            chunk.EndedAt,
		DurationSeconds:    chunk.DurationSeconds,
		ManifestObjectKey:  chunk.ManifestObjectKey,
		MediaObjectKeys:    utils.CopyStringMap(chunk.MediaObjectKeys),
		AvailableFileTypes: utils.CopyBoolMap(chunk.AvailableFileTypes),
		CreatedAt:          chunk.CreatedAt,
		UpdatedAt:          chunk.UpdatedAt,
	}
}

func RecordingTargetsPayload(targets []domain.Mission) RecordingTargetsResponse {
	return RecordingTargetsResponse{
		Targets: Missions(targets),
	}
}

func RecordingsPayload(recordings []RecordingChunkResponse) RecordingsResponse {
	return RecordingsResponse{
		Recordings: recordings,
	}
}

func RecordingChunkPayload(chunk RecordingChunkResponse) RecordingChunkEnvelopeResponse {
	return RecordingChunkEnvelopeResponse{
		Chunk: chunk,
	}
}

func RecordingFinalizationJob(job domain.RecordingFinalizationJob) RecordingFinalizationJobResponse {
	return RecordingFinalizationJobResponse{
		ID:                 job.ID,
		RecordingChunkID:   job.RecordingChunkID,
		RecordingSessionID: job.RecordingSessionID,
		MissionID:          job.MissionID,
		RobotID:            job.RobotID,
		Status:             job.Status,
		Reason:             job.Reason,
		Attempts:           job.Attempts,
		LockedBy:           job.LockedBy,
		LockedUntil:        job.LockedUntil,
		LastError:          job.LastError,
		CreatedAt:          job.CreatedAt,
		UpdatedAt:          job.UpdatedAt,
		CompletedAt:        job.CompletedAt,
		Chunk:              RecordingChunk(job.Chunk),
	}
}

func RecorderFinalizationJobsPayload(jobs []domain.RecordingFinalizationJob) RecorderFinalizationJobsResponse {
	response := make([]RecordingFinalizationJobResponse, 0, len(jobs))
	for _, job := range jobs {
		response = append(response, RecordingFinalizationJob(job))
	}
	return RecorderFinalizationJobsResponse{
		Jobs: response,
	}
}

func RecordingTick(result domain.RecordingTickResult) RecordingTickResponse {
	return RecordingTickResponse{
		Chunk:    RecordingChunk(result.Chunk),
		Manifest: result.Manifest,
	}
}
