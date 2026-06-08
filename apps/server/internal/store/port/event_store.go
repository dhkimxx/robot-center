package port

import (
	"context"

	"robot-center/apps/server/internal/domain"
)

type EventQuery struct {
	MissionID         string
	RobotCode         string
	EventType         string
	EventCategory     string
	TrackID           string
	IncludeDetections bool
	Limit             int
}

type EventStore interface {
	SaveMissionEventEnvelope(ctx context.Context, envelope domain.MissionEventEnvelope) ([]domain.MissionEvent, error)
	ListMissionEvents(ctx context.Context, query EventQuery) ([]domain.MissionEvent, error)
}
