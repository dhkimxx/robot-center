package domain

import (
	"encoding/json"
	"strings"
	"time"
)

type Robot struct {
	ID          string     `json:"id"`
	RobotCode   string     `json:"robotCode"`
	DisplayName string     `json:"displayName"`
	ModelName   string     `json:"modelName,omitempty"`
	Status      string     `json:"status"`
	LastSeenAt  *time.Time `json:"lastSeenAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type RobotConnectionInfo struct {
	ServerURL  string `json:"serverUrl"`
	RobotCode  string `json:"robotCode"`
	RobotToken string `json:"robotToken"`
}

type Mission struct {
	ID          string     `json:"id"`
	MissionCode string     `json:"missionCode"`
	Name        string     `json:"name"`
	MissionType string     `json:"missionType"`
	Status      string     `json:"status"`
	SiteNote    string     `json:"siteNote,omitempty"`
	RobotCode   string     `json:"robotCode,omitempty"`
	RobotCodes  []string   `json:"robotCodes,omitempty"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	EndedAt     *time.Time `json:"endedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

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

type RecordingChunk struct {
	ID                 string            `json:"id"`
	RecordingSessionID string            `json:"recordingSessionId"`
	MissionID          string            `json:"missionId"`
	MissionCode        string            `json:"missionCode"`
	RobotCode          string            `json:"robotCode"`
	ChunkIndex         int               `json:"chunkIndex"`
	Status             string            `json:"status"`
	StartedAt          time.Time         `json:"startedAt"`
	EndedAt            time.Time         `json:"endedAt"`
	DurationSeconds    int               `json:"durationSeconds"`
	ManifestObjectKey  string            `json:"manifestObjectKey"`
	MediaObjectKeys    map[string]string `json:"mediaObjectKeys"`
	AvailableFileTypes map[string]bool   `json:"availableFileTypes,omitempty"`
	CreatedAt          time.Time         `json:"createdAt"`
	UpdatedAt          time.Time         `json:"updatedAt"`
}

type RecordingFinalizationJob struct {
	ID                 string         `json:"id"`
	RecordingChunkID   string         `json:"recordingChunkId"`
	RecordingSessionID string         `json:"recordingSessionId"`
	MissionID          string         `json:"missionId"`
	RobotID            string         `json:"robotId"`
	Status             string         `json:"status"`
	Reason             string         `json:"reason,omitempty"`
	Attempts           int            `json:"attempts"`
	LockedBy           string         `json:"lockedBy,omitempty"`
	LockedUntil        *time.Time     `json:"lockedUntil,omitempty"`
	LastError          string         `json:"lastError,omitempty"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
	CompletedAt        *time.Time     `json:"completedAt,omitempty"`
	Chunk              RecordingChunk `json:"chunk"`
}

type RecordingTickResult struct {
	Chunk    RecordingChunk `json:"chunk"`
	Manifest map[string]any `json:"manifest"`
}

func StreamRoomID(missionCode string, robotCode string) string {
	missionCode = strings.TrimSpace(missionCode)
	robotCode = strings.TrimSpace(robotCode)
	if missionCode == "" || robotCode == "" {
		return missionCode
	}
	return missionCode + "__" + robotCode
}
