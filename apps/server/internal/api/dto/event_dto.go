package dto

import (
	"encoding/json"
	"time"

	"robot-center/apps/server/internal/domain"
)

type EventEnvelopeRequest struct {
	MessageID   string             `json:"messageId"`
	MessageType string             `json:"messageType"`
	RobotCode   string             `json:"robotCode"`
	MissionID   string             `json:"missionId"`
	MissionCode string             `json:"missionCode"`
	ChannelRole string             `json:"channelRole" binding:"required"`
	Events      []EventItemRequest `json:"events" binding:"required"`
}

type EventItemRequest struct {
	EventID   string          `json:"eventId"`
	EventType string          `json:"eventType" binding:"required"`
	Timestamp *time.Time      `json:"timestamp,omitempty"`
	Values    json.RawMessage `json:"values" binding:"required" swaggertype:"object"`
}

type MissionEventResponse struct {
	ID             string          `json:"id"`
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
	Values         json.RawMessage `json:"values" swaggertype:"object"`
	CreatedAt      time.Time       `json:"createdAt,omitempty"`
	UpdatedAt      time.Time       `json:"updatedAt,omitempty"`
}

type MissionEventsResponse struct {
	Events []MissionEventResponse `json:"events"`
}

type OperatorMissionEventsResponse struct {
	MissionCode string                 `json:"missionCode"`
	Events      []MissionEventResponse `json:"events"`
}

func MissionEvent(event domain.MissionEvent) MissionEventResponse {
	return MissionEventResponse{
		ID:             event.ID,
		MissionID:      event.MissionID,
		RobotCode:      event.RobotCode,
		EventID:        event.EventID,
		EventType:      event.EventType,
		EventCategory:  event.EventCategory,
		TrackID:        event.TrackID,
		Severity:       event.Severity,
		Title:          event.Title,
		Description:    event.Description,
		Timestamp:      event.Timestamp,
		ReceivedAt:     event.ReceivedAt,
		DetectionCount: event.DetectionCount,
		Values:         event.Values,
		CreatedAt:      event.CreatedAt,
		UpdatedAt:      event.UpdatedAt,
	}
}

func MissionEvents(events []domain.MissionEvent) []MissionEventResponse {
	response := make([]MissionEventResponse, 0, len(events))
	for _, event := range events {
		response = append(response, MissionEvent(event))
	}
	return response
}

func MissionEventsPayload(events []domain.MissionEvent) MissionEventsResponse {
	return MissionEventsResponse{
		Events: MissionEvents(events),
	}
}

func OperatorMissionEventsPayload(missionCode string, events []domain.MissionEvent) OperatorMissionEventsResponse {
	return OperatorMissionEventsResponse{
		MissionCode: missionCode,
		Events:      MissionEvents(events),
	}
}
