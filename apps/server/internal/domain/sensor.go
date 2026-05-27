package domain

import (
	"encoding/json"
	"time"
)

type SensorDescriptor struct {
	ID           string          `json:"id"`
	MissionID    string          `json:"missionId"`
	RobotCode    string          `json:"robotCode"`
	SensorID     string          `json:"sensorId"`
	ChannelRole  string          `json:"channelRole"`
	DisplayName  string          `json:"displayName"`
	SensorType   string          `json:"sensorType"`
	ValueType    string          `json:"valueType"`
	Unit         string          `json:"unit,omitempty"`
	SampleRateHz *float64        `json:"sampleRateHz,omitempty"`
	Enabled      bool            `json:"enabled"`
	Metadata     json.RawMessage `json:"metadata"`
	FirstSeenAt  time.Time       `json:"firstSeenAt"`
	LastSeenAt   time.Time       `json:"lastSeenAt"`
}

type SensorSample struct {
	ID           string          `json:"id"`
	DescriptorID string          `json:"descriptorId,omitempty"`
	MissionID    string          `json:"missionId"`
	RobotCode    string          `json:"robotCode"`
	SensorID     string          `json:"sensorId"`
	ChannelRole  string          `json:"channelRole"`
	MessageID    string          `json:"messageId,omitempty"`
	Sequence     int64           `json:"sequence,omitempty"`
	SentAt       *time.Time      `json:"sentAt,omitempty"`
	ReceivedAt   time.Time       `json:"receivedAt"`
	NumericValue *float64        `json:"numericValue,omitempty"`
	TextValue    string          `json:"textValue,omitempty"`
	BoolValue    *bool           `json:"boolValue,omitempty"`
	VectorValue  json.RawMessage `json:"vectorValue,omitempty"`
	ObjectValue  json.RawMessage `json:"objectValue,omitempty"`
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
	Sequence    int64              `json:"sequence,omitempty"`
	SentAt      *time.Time         `json:"sentAt,omitempty"`
	ReceivedAt  time.Time          `json:"receivedAt"`
	Descriptors []SensorDescriptor `json:"descriptors,omitempty"`
	Samples     []SensorSample     `json:"samples,omitempty"`
	RawPayload  json.RawMessage    `json:"rawPayload"`
}
