package dto

import (
	"time"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/utils"
)

type OperatorRecordingChunkResponse struct {
	ID                 string                          `json:"id"`
	RecordingSessionID string                          `json:"recordingSessionId"`
	MissionID          string                          `json:"missionId"`
	MissionCode        string                          `json:"missionCode"`
	RobotCode          string                          `json:"robotCode"`
	ChunkIndex         int                             `json:"chunkIndex"`
	Status             string                          `json:"status"`
	StartedAt          time.Time                       `json:"startedAt"`
	EndedAt            time.Time                       `json:"endedAt"`
	DurationSeconds    int                             `json:"durationSeconds"`
	ManifestObjectKey  string                          `json:"manifestObjectKey"`
	MediaObjectKeys    map[string]string               `json:"mediaObjectKeys"`
	AvailableFileTypes map[string]bool                 `json:"availableFileTypes,omitempty"`
	CreatedAt          time.Time                       `json:"createdAt"`
	UpdatedAt          time.Time                       `json:"updatedAt"`
	Files              []OperatorRecordingFileResponse `json:"files,omitempty"`
}

type OperatorRecordingFileResponse struct {
	Type        string `json:"type"`
	Label       string `json:"label"`
	Status      string `json:"status"`
	ContentType string `json:"contentType"`
	ObjectKey   string `json:"objectKey,omitempty"`
	URL         string `json:"url,omitempty"`
}

type OperatorRecordingsResponse struct {
	Recordings []OperatorRecordingChunkResponse `json:"recordings"`
}

type OperatorMissionRecordingSummaryResponse struct {
	MissionCode string                                         `json:"missionCode"`
	TotalChunks int                                            `json:"totalChunks"`
	Robots      []OperatorMissionRecordingRobotSummaryResponse `json:"robots"`
}

type OperatorMissionRecordingRobotSummaryResponse struct {
	RobotCode            string         `json:"robotCode"`
	ChunkCount           int            `json:"chunkCount"`
	UploadedChunkCount   int            `json:"uploadedChunkCount"`
	RecordingChunkCount  int            `json:"recordingChunkCount"`
	FinalizingChunkCount int            `json:"finalizingChunkCount"`
	PartialChunkCount    int            `json:"partialChunkCount"`
	FirstStartedAt       *time.Time     `json:"firstStartedAt,omitempty"`
	LastEndedAt          *time.Time     `json:"lastEndedAt,omitempty"`
	AvailableFileCounts  map[string]int `json:"availableFileCounts"`
	MissingFileCounts    map[string]int `json:"missingFileCounts"`
}

type OperatorMissionRecordingChunksResponse struct {
	Recordings []OperatorRecordingChunkResponse `json:"recordings"`
	Page       OperatorRecordingPageResponse    `json:"page"`
}

type OperatorRecordingPageResponse struct {
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
	Total      int  `json:"total"`
	HasMore    bool `json:"hasMore"`
	NextOffset int  `json:"nextOffset"`
}

type RecorderRecordingTargetResponse struct {
	ID          string     `json:"id"`
	MissionCode string     `json:"missionCode"`
	Name        string     `json:"name"`
	MissionType string     `json:"missionType"`
	Status      string     `json:"status"`
	SiteNote    string     `json:"siteNote,omitempty"`
	RobotCode   string     `json:"robotCode,omitempty"`
	RobotCodes  []string   `json:"robotCodes,omitempty"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	EndedAt     *time.Time `json:"endedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type RecorderRecordingTargetsResponse struct {
	Targets []RecorderRecordingTargetResponse `json:"targets"`
}

type RecorderRecordingChunkResponse struct {
	ID                 string            `json:"id"`
	RecordingSessionID string            `json:"recordingSessionId"`
	MissionID          string            `json:"missionId"`
	MissionCode        string            `json:"missionCode"`
	RobotCode          string            `json:"robotCode"`
	ChunkIndex         int               `json:"chunkIndex"`
	Status             string            `json:"status"`
	StartedAt          time.Time         `json:"startedAt"`
	EndedAt            time.Time         `json:"endedAt"`
	DurationSeconds    int               `json:"durationSeconds"`
	ManifestObjectKey  string            `json:"manifestObjectKey"`
	MediaObjectKeys    map[string]string `json:"mediaObjectKeys"`
	AvailableFileTypes map[string]bool   `json:"availableFileTypes,omitempty"`
	CreatedAt          time.Time         `json:"createdAt"`
	UpdatedAt          time.Time         `json:"updatedAt"`
}

type RecorderRecordingTickResponse struct {
	Chunk    RecorderRecordingChunkResponse `json:"chunk"`
	Manifest map[string]any                 `json:"manifest"`
}

type RecorderRecordingChunkEnvelopeResponse struct {
	Chunk RecorderRecordingChunkResponse `json:"chunk"`
}

type RecorderFinalizationJobResponse struct {
	ID                 string                         `json:"id"`
	RecordingChunkID   string                         `json:"recordingChunkId"`
	RecordingSessionID string                         `json:"recordingSessionId"`
	MissionID          string                         `json:"missionId"`
	RobotID            string                         `json:"robotId"`
	Status             string                         `json:"status"`
	Reason             string                         `json:"reason,omitempty"`
	Attempts           int                            `json:"attempts"`
	LockedBy           string                         `json:"lockedBy,omitempty"`
	LockedUntil        *time.Time                     `json:"lockedUntil,omitempty"`
	LastError          string                         `json:"lastError,omitempty"`
	CreatedAt          time.Time                      `json:"createdAt"`
	UpdatedAt          time.Time                      `json:"updatedAt"`
	CompletedAt        *time.Time                     `json:"completedAt,omitempty"`
	Chunk              RecorderRecordingChunkResponse `json:"chunk"`
}

type RecorderFinalizationJobsResponse struct {
	Jobs []RecorderFinalizationJobResponse `json:"jobs"`
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

func OperatorRecordingChunk(chunk domain.RecordingChunk) OperatorRecordingChunkResponse {
	return OperatorRecordingChunkResponse{
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

func OperatorRecordingsPayload(recordings []OperatorRecordingChunkResponse) OperatorRecordingsResponse {
	return OperatorRecordingsResponse{
		Recordings: recordings,
	}
}

func RecorderRecordingTarget(mission domain.Mission) RecorderRecordingTargetResponse {
	return RecorderRecordingTargetResponse{
		ID:          mission.ID,
		MissionCode: mission.MissionCode,
		Name:        mission.Name,
		MissionType: mission.MissionType,
		Status:      mission.Status,
		SiteNote:    mission.SiteNote,
		RobotCode:   mission.RobotCode,
		RobotCodes:  append([]string(nil), mission.RobotCodes...),
		StartedAt:   mission.StartedAt,
		EndedAt:     mission.EndedAt,
		CreatedAt:   mission.CreatedAt,
		UpdatedAt:   mission.UpdatedAt,
	}
}

func RecorderRecordingTargetsPayload(targets []domain.Mission) RecorderRecordingTargetsResponse {
	response := make([]RecorderRecordingTargetResponse, 0, len(targets))
	for _, target := range targets {
		response = append(response, RecorderRecordingTarget(target))
	}
	return RecorderRecordingTargetsResponse{
		Targets: response,
	}
}

func RecorderRecordingChunk(chunk domain.RecordingChunk) RecorderRecordingChunkResponse {
	return RecorderRecordingChunkResponse{
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

func RecorderRecordingChunkPayload(chunk RecorderRecordingChunkResponse) RecorderRecordingChunkEnvelopeResponse {
	return RecorderRecordingChunkEnvelopeResponse{
		Chunk: chunk,
	}
}

func RecorderFinalizationJob(job domain.RecordingFinalizationJob) RecorderFinalizationJobResponse {
	return RecorderFinalizationJobResponse{
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
		Chunk:              RecorderRecordingChunk(job.Chunk),
	}
}

func RecorderFinalizationJobsPayload(jobs []domain.RecordingFinalizationJob) RecorderFinalizationJobsResponse {
	response := make([]RecorderFinalizationJobResponse, 0, len(jobs))
	for _, job := range jobs {
		response = append(response, RecorderFinalizationJob(job))
	}
	return RecorderFinalizationJobsResponse{
		Jobs: response,
	}
}

func RecorderRecordingTick(result domain.RecordingTickResult) RecorderRecordingTickResponse {
	return RecorderRecordingTickResponse{
		Chunk:    RecorderRecordingChunk(result.Chunk),
		Manifest: result.Manifest,
	}
}
