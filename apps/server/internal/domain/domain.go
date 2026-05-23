package domain

import (
	"encoding/json"
	"strings"
	"time"
)

type Robot struct {
	ID              string     `json:"id"`
	RobotCode       string     `json:"robotCode"`
	DisplayName     string     `json:"displayName"`
	ModelName       string     `json:"modelName,omitempty"`
	Status          string     `json:"status"`
	LastSeenAt      *time.Time `json:"lastSeenAt,omitempty"`
	LastStreamingAt *time.Time `json:"lastStreamingAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
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

type StreamingTrack struct {
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Codec       string `json:"codec"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	FPS         int    `json:"fps,omitempty"`
	BitrateKbps int    `json:"bitrateKbps,omitempty"`
}

type StreamingStatus struct {
	RobotCode             string           `json:"robotCode"`
	MissionID             string           `json:"missionId"`
	RoomID                string           `json:"roomId"`
	Status                string           `json:"status"`
	PublishedTracks       []StreamingTrack `json:"publishedTracks"`
	PublishedDataChannels []string         `json:"publishedDataChannels"`
	SentAt                time.Time        `json:"sentAt"`
}

type TelemetrySnapshot struct {
	ID                  string          `json:"id"`
	RobotCode           string          `json:"robotCode"`
	MissionID           string          `json:"missionId"`
	MessageID           string          `json:"messageId,omitempty"`
	MessageType         string          `json:"messageType,omitempty"`
	Sequence            int64           `json:"sequence,omitempty"`
	SentAt              *time.Time      `json:"sentAt,omitempty"`
	ReceivedAt          time.Time       `json:"receivedAt"`
	BatteryPercent      *float64        `json:"batteryPercent,omitempty"`
	NetworkState        string          `json:"networkState,omitempty"`
	PositionAvailable   bool            `json:"positionAvailable"`
	Latitude            *float64        `json:"latitude,omitempty"`
	Longitude           *float64        `json:"longitude,omitempty"`
	AltitudeMeter       *float64        `json:"altitudeMeter,omitempty"`
	AccuracyMeter       *float64        `json:"accuracyMeter,omitempty"`
	HeadingDegree       *float64        `json:"headingDegree,omitempty"`
	SpeedMeterPerSecond *float64        `json:"speedMeterPerSecond,omitempty"`
	RawPayload          json.RawMessage `json:"rawPayload"`
}

type SensorReading struct {
	ID                 string          `json:"id"`
	RobotCode          string          `json:"robotCode"`
	MissionID          string          `json:"missionId"`
	MessageID          string          `json:"messageId,omitempty"`
	MessageType        string          `json:"messageType,omitempty"`
	Sequence           int64           `json:"sequence,omitempty"`
	SentAt             *time.Time      `json:"sentAt,omitempty"`
	ReceivedAt         time.Time       `json:"receivedAt"`
	BatteryPercent     *float64        `json:"batteryPercent,omitempty"`
	TemperatureCelsius *float64        `json:"temperatureCelsius,omitempty"`
	HumidityPercent    *float64        `json:"humidityPercent,omitempty"`
	OxygenPercent      *float64        `json:"oxygenPercent,omitempty"`
	COPpm              *float64        `json:"coPpm,omitempty"`
	CH4Ppm             *float64        `json:"ch4Ppm,omitempty"`
	RawPayload         json.RawMessage `json:"rawPayload"`
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
