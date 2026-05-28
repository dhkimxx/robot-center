package model

import (
	"encoding/json"
	"time"
)

type StorageObjectModel struct {
	BaseModel
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
}

func (StorageObjectModel) TableName() string {
	return "storage_objects"
}
