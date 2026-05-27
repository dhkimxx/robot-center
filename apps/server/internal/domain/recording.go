package domain

import "time"

type RecordingChunk struct {
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

type RecordingFinalizationJob struct {
	ID                 string         `json:"id"`
	RecordingChunkID   string         `json:"recordingChunkId"`
	RecordingSessionID string         `json:"recordingSessionId"`
	MissionID          string         `json:"missionId"`
	RobotID            string         `json:"robotId"`
	Status             string         `json:"status"`
	Reason             string         `json:"reason,omitempty"`
	Attempts           int            `json:"attempts"`
	LockedBy           string         `json:"lockedBy,omitempty"`
	LockedUntil        *time.Time     `json:"lockedUntil,omitempty"`
	LastError          string         `json:"lastError,omitempty"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
	CompletedAt        *time.Time     `json:"completedAt,omitempty"`
	Chunk              RecordingChunk `json:"chunk"`
}

type RecordingTickResult struct {
	Chunk    RecordingChunk `json:"chunk"`
	Manifest map[string]any `json:"manifest"`
}
