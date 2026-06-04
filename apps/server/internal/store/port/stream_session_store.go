package port

import (
	"context"

	"robot-center/apps/server/internal/domain"
)

type StreamSessionStore interface {
	StartRobotStreamSession(ctx context.Context, input StartRobotStreamSessionInput) (domain.RobotStreamSession, error)
	TouchRobotStreamSession(ctx context.Context, input TouchRobotStreamSessionInput) error
	EndRobotStreamSession(ctx context.Context, input EndRobotStreamSessionInput) error
	ListRobotStreamSessionsForMission(ctx context.Context, missionCode string) ([]domain.RobotStreamSession, error)
}
