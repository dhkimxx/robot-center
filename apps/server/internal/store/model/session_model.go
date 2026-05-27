package model

import (
	"encoding/json"
	"time"
)

type RobotSessionModel struct {
	ID              string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RobotID         string          `gorm:"column:robot_id;type:uuid;not null;index"`
	MissionID       *string         `gorm:"column:mission_id;type:uuid;index"`
	State           string          `gorm:"column:state;not null"`
	ClientIP        *string         `gorm:"column:client_ip;type:inet"`
	UserAgent       *string         `gorm:"column:user_agent"`
	ConnectedAt     time.Time       `gorm:"column:connected_at;not null;default:now()"`
	LastHeartbeatAt time.Time       `gorm:"column:last_heartbeat_at;not null;default:now()"`
	DisconnectedAt  *time.Time      `gorm:"column:disconnected_at"`
	RawPayload      json.RawMessage `gorm:"column:raw_payload;type:jsonb;not null;default:'{}'::jsonb"`
}

func (RobotSessionModel) TableName() string {
	return "robot_sessions"
}

type BrowserSessionModel struct {
	ID             string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID      string          `gorm:"column:mission_id;type:uuid;not null;index"`
	UserID         *string         `gorm:"column:user_id;type:uuid;index"`
	State          string          `gorm:"column:state;not null"`
	ConnectedAt    time.Time       `gorm:"column:connected_at;not null;default:now()"`
	DisconnectedAt *time.Time      `gorm:"column:disconnected_at"`
	Metadata       json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
}

func (BrowserSessionModel) TableName() string {
	return "browser_sessions"
}

type RecorderSessionModel struct {
	ID        string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID string          `gorm:"column:mission_id;type:uuid;not null;index"`
	State     string          `gorm:"column:state;not null"`
	StartedAt time.Time       `gorm:"column:started_at;not null;default:now()"`
	StoppedAt *time.Time      `gorm:"column:stopped_at"`
	LastError *string         `gorm:"column:last_error"`
	Metadata  json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
}

func (RecorderSessionModel) TableName() string {
	return "recorder_sessions"
}
