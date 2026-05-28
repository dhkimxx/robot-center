package domain

import (
	"encoding/json"
	"time"
)

type SensorDescriptor struct {
	ID          string          `json:"id"`
	MissionID   string          `json:"missionId"`
	RobotCode   string          `json:"robotCode"`
	SensorID    string          `json:"sensorId"`
	ChannelRole string          `json:"channelRole"`
	DisplayName string          `json:"displayName"`
	SensorType  string          `json:"sensorType"`
	Unit        string          `json:"unit,omitempty"`
	Enabled     bool            `json:"enabled"`
	Metadata    json.RawMessage `json:"metadata"`
	FirstSeenAt time.Time       `json:"firstSeenAt"`
	LastSeenAt  time.Time       `json:"lastSeenAt"`
}

type SensorSample struct {
	ID           string          `json:"id"`
	DescriptorID string          `json:"descriptorId,omitempty"`
	MissionID    string          `json:"missionId"`
	RobotCode    string          `json:"robotCode"`
	SensorID     string          `json:"sensorId"`
	ChannelRole  string          `json:"channelRole"`
	MessageID    string          `json:"messageId,omitempty"`
	Timestamp    *time.Time      `json:"timestamp,omitempty"`
	ReceivedAt   time.Time       `json:"receivedAt"`
	Values       json.RawMessage `json:"values,omitempty"`
	ObjectKey    string          `json:"objectKey,omitempty"`
	RawPayload   json.RawMessage `json:"rawPayload"`
}

type SensorLatest struct {
	Descriptor   SensorDescriptor `json:"descriptor"`
	LatestSample *SensorSample    `json:"latestSample,omitempty"`
}

type SensorEnvelope struct {
	MessageID   string             `json:"messageId,omitempty"`
	MessageType string             `json:"messageType,omitempty"`
	RobotCode   string             `json:"robotCode"`
	MissionID   string             `json:"missionId"`
	ChannelRole string             `json:"channelRole"`
	ReceivedAt  time.Time          `json:"receivedAt"`
	Descriptors []SensorDescriptor `json:"descriptors,omitempty"`
	Samples     []SensorSample     `json:"samples,omitempty"`
	RawPayload  json.RawMessage    `json:"rawPayload"`
}
