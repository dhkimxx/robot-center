package service

import (
	"context"
	"strings"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
)

const clearEventDataConfirmation = "CLEAR_EVENT_DATA"

type EventService struct {
	repository store.EventRepository
}

func (s *EventService) SaveMissionEventEnvelope(ctx context.Context, envelope domain.MissionEventEnvelope) ([]domain.MissionEvent, error) {
	return s.repository.SaveMissionEventEnvelope(ctx, envelope)
}

func (s *EventService) ListMissionEvents(ctx context.Context, query store.EventQuery) ([]domain.MissionEvent, error) {
	return s.repository.ListMissionEvents(ctx, query)
}

func (s *EventService) ClearEventData(ctx context.Context, environment string, confirmation string) (store.EventDataClearResult, error) {
	if strings.EqualFold(strings.TrimSpace(environment), "production") {
		return store.EventDataClearResult{}, ErrSystemActionForbidden
	}
	if strings.TrimSpace(confirmation) != clearEventDataConfirmation {
		return store.EventDataClearResult{}, ErrSystemActionConfirmationRequired
	}
	return s.repository.ClearEventData(ctx)
}
