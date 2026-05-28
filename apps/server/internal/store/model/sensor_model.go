package model

import (
	"encoding/json"
	"time"
)

type SensorDescriptorModel struct {
	BaseModel
	MissionID   string          `gorm:"column:mission_id;type:uuid;not null;uniqueIndex:sensor_descriptors_mission_robot_sensor_unique,priority:1;index"`
	RobotID     string          `gorm:"column:robot_id;type:uuid;not null;uniqueIndex:sensor_descriptors_mission_robot_sensor_unique,priority:2;index"`
	SensorID    string          `gorm:"column:sensor_id;not null;uniqueIndex:sensor_descriptors_mission_robot_sensor_unique,priority:3"`
	ChannelRole string          `gorm:"column:channel_role;not null"`
	DisplayName string          `gorm:"column:display_name;not null"`
	SensorType  string          `gorm:"column:sensor_type;not null;index"`
	Unit        *string         `gorm:"column:unit"`
	Enabled     bool            `gorm:"column:enabled;not null"`
	Metadata    json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
	FirstSeenAt time.Time       `gorm:"column:first_seen_at;not null;default:now()"`
	LastSeenAt  time.Time       `gorm:"column:last_seen_at;not null;default:now();index"`
}

func (SensorDescriptorModel) TableName() string {
	return "sensor_descriptors"
}

type SensorSampleModel struct {
	BaseModel
	DescriptorID string          `gorm:"column:descriptor_id;type:uuid;not null;index;index:sensor_samples_descriptor_received_idx,priority:1"`
	MissionID    string          `gorm:"column:mission_id;type:uuid;not null;index:sensor_samples_latest_idx,priority:1"`
	RobotID      string          `gorm:"column:robot_id;type:uuid;not null;index:sensor_samples_latest_idx,priority:2"`
	SensorID     string          `gorm:"column:sensor_id;not null;index:sensor_samples_latest_idx,priority:3"`
	ChannelRole  string          `gorm:"column:channel_role;not null;index"`
	MessageID    *string         `gorm:"column:message_id;index"`
	Timestamp    *time.Time      `gorm:"column:sample_timestamp"`
	ReceivedAt   time.Time       `gorm:"column:received_at;not null;default:now();index:sensor_samples_latest_idx,sort:desc,priority:4;index:sensor_samples_descriptor_received_idx,sort:desc,priority:2"`
	Values       json.RawMessage `gorm:"column:values;type:jsonb"`
	ObjectKey    *string         `gorm:"column:object_key"`
	RawPayload   json.RawMessage `gorm:"column:raw_payload;type:jsonb;not null;default:'{}'::jsonb"`
}

func (SensorSampleModel) TableName() string {
	return "sensor_samples"
}
