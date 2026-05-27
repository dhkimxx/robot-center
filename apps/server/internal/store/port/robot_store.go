package port

import (
	"context"

	"robot-center/apps/server/internal/domain"
)

type RobotStore interface {
	CreateRobot(ctx context.Context, input CreateRobotInput) (domain.Robot, domain.RobotConnectionInfo, error)
	ListRobots(ctx context.Context) ([]domain.Robot, error)
	UpdateRobot(ctx context.Context, robotCode string, input UpdateRobotInput) (domain.Robot, error)
	ArchiveRobot(ctx context.Context, robotCode string) (domain.Robot, error)
	GetRobotConnectionInfo(ctx context.Context, robotCode string) (domain.RobotConnectionInfo, error)
	RotateRobotConnectionToken(ctx context.Context, robotCode string) (domain.RobotConnectionInfo, error)
	ApplyHeartbeat(ctx context.Context, input HeartbeatInput, bearerToken string) (domain.Robot, error)
	ResolveRobotByBearerToken(ctx context.Context, bearerToken string) (domain.Robot, error)
}
