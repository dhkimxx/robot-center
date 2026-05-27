package model

import (
	"encoding/json"
	"time"
)

type StorageObjectModel struct {
	ID               string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID        string          `gorm:"column:mission_id;type:uuid;not null;index"`
	RobotID          *string         `gorm:"column:robot_id;type:uuid;index"`
	RecordingChunkID *string         `gorm:"column:recording_chunk_id;type:uuid;index"`
	ObjectType       string          `gorm:"column:object_type;not null"`
	Bucket           string          `gorm:"column:bucket;not null"`
	ObjectKey        string          `gorm:"column:object_key;not null;uniqueIndex"`
	ContentType      *string         `gorm:"column:content_type"`
	SizeBytes        *int64          `gorm:"column:size_bytes"`
	Checksum         *string         `gorm:"column:checksum"`
	StartedAt        *time.Time      `gorm:"column:started_at"`
	EndedAt          *time.Time      `gorm:"column:ended_at"`
	Metadata         json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
	CreatedAt        time.Time       `gorm:"column:created_at;not null;default:now()"`
}

func (StorageObjectModel) TableName() string {
	return "storage_objects"
}
