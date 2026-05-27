package model

import (
	"encoding/json"
	"time"
)

type EventModel struct {
	ID                     string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID              string          `gorm:"column:mission_id;type:uuid;not null;index:events_mission_occurred_idx,sort:desc,priority:1"`
	RobotID                *string         `gorm:"column:robot_id;type:uuid;index"`
	EventType              string          `gorm:"column:event_type;not null"`
	Severity               string          `gorm:"column:severity;not null"`
	Title                  string          `gorm:"column:title;not null"`
	Description            *string         `gorm:"column:description"`
	OccurredAt             time.Time       `gorm:"column:occurred_at;not null;index:events_mission_occurred_idx,sort:desc,priority:2"`
	Geom                   []byte          `gorm:"column:geom;type:geometry(Point,4326)"`
	RelatedStorageObjectID *string         `gorm:"column:related_storage_object_id;type:uuid;index"`
	RawPayload             json.RawMessage `gorm:"column:raw_payload;type:jsonb;not null;default:'{}'::jsonb"`
	CreatedAt              time.Time       `gorm:"column:created_at;not null;default:now()"`
}

func (EventModel) TableName() string {
	return "events"
}
