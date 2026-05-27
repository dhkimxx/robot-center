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

func RecordingTick(result domain.RecordingTickResult) RecordingTickResponse {
	return RecordingTickResponse{
		Chunk:    RecordingChunk(result.Chunk),
		Manifest: result.Manifest,
	}
}
