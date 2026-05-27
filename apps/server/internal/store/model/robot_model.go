package model

import (
	"encoding/json"
	"time"

	"robot-center/apps/server/internal/domain"
)

type RobotModel struct {
	ID          string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RobotCode   string          `gorm:"column:robot_code;not null;uniqueIndex"`
	DisplayName string          `gorm:"column:display_name;not null"`
	ModelName   *string         `gorm:"column:model_name"`
	DeviceState string          `gorm:"column:device_state;not null;default:offline"`
	LastSeenAt  *time.Time      `gorm:"column:last_seen_at"`
	ArchivedAt  *time.Time      `gorm:"column:archived_at"`
	Metadata    json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
	CreatedAt   time.Time       `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt   time.Time       `gorm:"column:updated_at;not null;default:now()"`
}

func (RobotModel) TableName() string {
	return "robots"
}

func (record RobotModel) ToDomainRobot() domain.Robot {
	return domain.Robot{
		ID:          record.ID,
		RobotCode:   record.RobotCode,
		DisplayName: record.DisplayName,
		ModelName:   stringFromPointer(record.ModelName),
		DeviceState: domain.NormalizeRobotDeviceState(record.DeviceState),
		LastSeenAt:  record.LastSeenAt,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
}

type RobotConnectionTokenModel struct {
	ID             string     `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RobotID        string     `gorm:"column:robot_id;type:uuid;not null;index"`
	TokenHash      string     `gorm:"column:token_hash;not null"`
	TokenPlaintext *string    `gorm:"column:token_plaintext"`
	Name           string     `gorm:"column:name;not null"`
	IsActive       bool       `gorm:"column:is_active;not null;default:true"`
	LastUsedAt     *time.Time `gorm:"column:last_used_at"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null;default:now()"`
}

func (RobotConnectionTokenModel) TableName() string {
	return "robot_tokens"
}
