package port

import (
	"context"

	"robot-center/apps/server/internal/domain"
)

type MissionStore interface {
	CreateMission(ctx context.Context, input CreateMissionInput) (domain.Mission, error)
	ListMissions(ctx context.Context) ([]domain.Mission, error)
	StartMission(ctx context.Context, missionCode string) (domain.Mission, error)
	EndMission(ctx context.Context, missionCode string) (domain.Mission, error)
	FindActiveMissionForRobot(ctx context.Context, robotCode string, bearerToken string) (domain.Mission, bool, error)
	ValidateActiveMissionRobot(ctx context.Context, missionCode string, robotCode string) error
	RecordingTargets(ctx context.Context) ([]domain.Mission, error)
}
