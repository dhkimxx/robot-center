package model

import (
	"encoding/json"
	"time"
)

type EventModel struct {
	BaseModel
	MissionID              string          `gorm:"column:mission_id;type:uuid;not null;index:events_mission_occurred_idx,sort:desc,priority:1"`
	RobotID                *string         `gorm:"column:robot_id;type:uuid;index"`
	EventID                *string         `gorm:"column:event_id"`
	EventType              string          `gorm:"column:event_type;not null"`
	EventCategory          string          `gorm:"column:event_category;not null;default:'mission'"`
	TrackID                *string         `gorm:"column:track_id;index"`
	Severity               string          `gorm:"column:severity;not null"`
	Title                  string          `gorm:"column:title;not null"`
	Description            *string         `gorm:"column:description"`
	OccurredAt             time.Time       `gorm:"column:occurred_at;not null;index:events_mission_occurred_idx,sort:desc,priority:2"`
	ReceivedAt             time.Time       `gorm:"column:received_at;not null;default:now()"`
	DetectionCount         *int            `gorm:"column:detection_count"`
	Geom                   []byte          `gorm:"column:geom;type:geometry(Point,4326)"`
	RelatedStorageObjectID *string         `gorm:"column:related_storage_object_id;type:uuid;index"`
	Values                 json.RawMessage `gorm:"column:values;type:jsonb;not null;default:'{}'::jsonb"`
	RawMessage             json.RawMessage `gorm:"column:raw_message;type:jsonb;not null;default:'{}'::jsonb"`
	RawPayload             json.RawMessage `gorm:"column:raw_payload;type:jsonb;not null;default:'{}'::jsonb"`
}

func (EventModel) TableName() string {
	return "events"
}
