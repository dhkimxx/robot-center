package service

import (
	"context"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
)

type EventService struct {
	repository store.EventRepository
}

func (s *EventService) SaveMissionEventEnvelope(ctx context.Context, envelope domain.MissionEventEnvelope) ([]domain.MissionEvent, error) {
	return s.repository.SaveMissionEventEnvelope(ctx, envelope)
}

func (s *EventService) ListMissionEvents(ctx context.Context, query store.EventQuery) ([]domain.MissionEvent, error) {
	return s.repository.ListMissionEvents(ctx, query)
}
