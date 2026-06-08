package domain

import (
	"encoding/json"
	"time"
)

const (
	EventTypeDetectionObject = "detection.object"
	EventTypeMissionEvent    = "mission.event"

	EventCategoryDetection = "detection"
	EventCategoryMission   = "mission"
	EventCategoryAlarm     = "alarm"
	EventCategorySystem    = "system"

	EventSourceChannel = "channel.event"
)

type MissionEventEnvelope struct {
	MessageID   string          `json:"messageId,omitempty"`
	MessageType string          `json:"messageType,omitempty"`
	RobotCode   string          `json:"robotCode"`
	MissionID   string          `json:"missionId"`
	ReceivedAt  time.Time       `json:"receivedAt"`
	Events      []MissionEvent  `json:"events"`
	RawMessage  json.RawMessage `json:"rawMessage"`
}

type MissionEvent struct {
	ID             string          `json:"id,omitempty"`
	MissionID      string          `json:"missionId"`
	RobotCode      string          `json:"robotCode,omitempty"`
	EventID        string          `json:"eventId,omitempty"`
	EventType      string          `json:"eventType"`
	EventCategory  string          `json:"eventCategory"`
	TrackID        string          `json:"trackId,omitempty"`
	Severity       string          `json:"severity"`
	Title          string          `json:"title"`
	Description    string          `json:"description,omitempty"`
	Timestamp      time.Time       `json:"timestamp"`
	ReceivedAt     time.Time       `json:"receivedAt"`
	DetectionCount *int            `json:"detectionCount,omitempty"`
	Values         json.RawMessage `json:"values"`
	RawMessage     json.RawMessage `json:"rawMessage,omitempty"`
	CreatedAt      time.Time       `json:"createdAt,omitempty"`
	UpdatedAt      time.Time       `json:"updatedAt,omitempty"`
}
