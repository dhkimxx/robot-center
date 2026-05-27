package model

import (
	"encoding/json"
	"time"
)

type RecordingSessionModel struct {
	ID                   string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID            string          `gorm:"column:mission_id;type:uuid;not null;index"`
	RobotID              string          `gorm:"column:robot_id;type:uuid;not null;index"`
	RecorderSessionID    *string         `gorm:"column:recorder_session_id;type:uuid;index"`
	Status               string          `gorm:"column:status;not null;default:pending"`
	ChunkDurationSeconds int             `gorm:"column:chunk_duration_seconds;not null;default:600"`
	StartedAt            time.Time       `gorm:"column:started_at;not null;default:now()"`
	EndedAt              *time.Time      `gorm:"column:ended_at"`
	LastError            *string         `gorm:"column:last_error"`
	Metadata             json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
}

func (RecordingSessionModel) TableName() string {
	return "recording_sessions"
}

type RecordingChunkModel struct {
	ID                 string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RecordingSessionID string          `gorm:"column:recording_session_id;type:uuid;not null;uniqueIndex:recording_chunks_session_index_unique,priority:1"`
	MissionID          string          `gorm:"column:mission_id;type:uuid;not null;index"`
	RobotID            string          `gorm:"column:robot_id;type:uuid;not null;index"`
	ChunkIndex         int             `gorm:"column:chunk_index;not null;uniqueIndex:recording_chunks_session_index_unique,priority:2"`
	Status             string          `gorm:"column:status;not null;default:pending"`
	StartedAt          time.Time       `gorm:"column:started_at;not null"`
	EndedAt            *time.Time      `gorm:"column:ended_at"`
	DurationSeconds    *float64        `gorm:"column:duration_seconds"`
	ManifestObjectID   *string         `gorm:"column:manifest_object_id;type:uuid;index"`
	Metadata           json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
	CreatedAt          time.Time       `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt          time.Time       `gorm:"column:updated_at;not null;default:now()"`
}

func (RecordingChunkModel) TableName() string {
	return "recording_chunks"
}

type RecordingFinalizationJobModel struct {
	ID                 string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RecordingChunkID   string          `gorm:"column:recording_chunk_id;type:uuid;not null;uniqueIndex"`
	RecordingSessionID string          `gorm:"column:recording_session_id;type:uuid;not null;index"`
	MissionID          string          `gorm:"column:mission_id;type:uuid;not null;index"`
	RobotID            string          `gorm:"column:robot_id;type:uuid;not null;index"`
	Status             string          `gorm:"column:status;not null;default:queued;index"`
	Reason             *string         `gorm:"column:reason"`
	Attempts           int             `gorm:"column:attempts;not null;default:0"`
	LockedBy           *string         `gorm:"column:locked_by"`
	LockedUntil        *time.Time      `gorm:"column:locked_until;index"`
	LastError          *string         `gorm:"column:last_error"`
	Metadata           json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
	CompletedAt        *time.Time      `gorm:"column:completed_at"`
	CreatedAt          time.Time       `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt          time.Time       `gorm:"column:updated_at;not null;default:now()"`
}

func (RecordingFinalizationJobModel) TableName() string {
	return "recording_finalization_jobs"
}
