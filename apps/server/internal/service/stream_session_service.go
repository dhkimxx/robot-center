package service

import (
	"context"
	"strings"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/sfu"
	"robot-center/apps/server/internal/store"
)

type StreamSessionService struct {
	repository store.StreamSessionStore
}

func (s *StreamSessionService) HandlePublisherEvent(ctx context.Context, event sfu.PublisherEvent) error {
	if s == nil || s.repository == nil {
		return nil
	}
	switch event.Type {
	case sfu.PublisherEventMediaStarted:
		_, err := s.repository.StartRobotStreamSession(ctx, store.StartRobotStreamSessionInput{
			MissionCode:     event.RoomID,
			RobotCode:       event.RobotCode,
			PublisherPeerID: event.PublisherPeerID,
			StartedAt:       event.ObservedAt,
		})
		return err
	case sfu.PublisherEventMediaActive:
		return s.repository.TouchRobotStreamSession(ctx, store.TouchRobotStreamSessionInput{
			PublisherPeerID: event.PublisherPeerID,
			ObservedAt:      event.ObservedAt,
		})
	case sfu.PublisherEventEnded:
		return s.repository.EndRobotStreamSession(ctx, store.EndRobotStreamSessionInput{
			MissionCode:     event.RoomID,
			RobotCode:       event.RobotCode,
			PublisherPeerID: event.PublisherPeerID,
			Reason:          normalizeStreamEndReason(event.Reason),
			EndedAt:         event.ObservedAt,
		})
	default:
		return nil
	}
}

func (s *StreamSessionService) ListRobotStreamSessionsForMission(ctx context.Context, missionCode string) ([]domain.RobotStreamSession, error) {
	if s == nil || s.repository == nil {
		return []domain.RobotStreamSession{}, nil
	}
	return s.repository.ListRobotStreamSessionsForMission(ctx, strings.TrimSpace(missionCode))
}

func normalizeStreamEndReason(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return "closed"
	}
	return reason
}
