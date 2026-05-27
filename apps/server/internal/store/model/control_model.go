package model

import (
	"encoding/json"
	"time"
)

type ControlCommandModel struct {
	ID            string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID     string          `gorm:"column:mission_id;type:uuid;not null;index"`
	RobotID       string          `gorm:"column:robot_id;type:uuid;not null;index"`
	RequestedBy   *string         `gorm:"column:requested_by;type:uuid;index"`
	CommandType   string          `gorm:"column:command_type;not null"`
	Status        string          `gorm:"column:status;not null;default:requested"`
	Payload       json.RawMessage `gorm:"column:payload;type:jsonb;not null;default:'{}'::jsonb"`
	RequestedAt   time.Time       `gorm:"column:requested_at;not null;default:now()"`
	SentAt        *time.Time      `gorm:"column:sent_at"`
	CompletedAt   *time.Time      `gorm:"column:completed_at"`
	FailureReason *string         `gorm:"column:failure_reason"`
}

func (ControlCommandModel) TableName() string {
	return "control_commands"
}

type ControlAckModel struct {
	ID               string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	ControlCommandID string          `gorm:"column:control_command_id;type:uuid;not null;index"`
	RobotID          string          `gorm:"column:robot_id;type:uuid;not null;index"`
	AckStatus        string          `gorm:"column:ack_status;not null"`
	Message          *string         `gorm:"column:message"`
	ReceivedAt       time.Time       `gorm:"column:received_at;not null;default:now()"`
	RawPayload       json.RawMessage `gorm:"column:raw_payload;type:jsonb;not null;default:'{}'::jsonb"`
}

func (ControlAckModel) TableName() string {
	return "control_acks"
}
