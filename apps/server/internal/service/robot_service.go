package service

import (
	"context"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
)

type RobotService struct {
	repository store.RobotRepository
}

func (s *RobotService) CreateRobot(ctx context.Context, input store.CreateRobotInput) (domain.Robot, domain.RobotConnectionInfo, error) {
	return s.repository.CreateRobot(ctx, input)
}

func (s *RobotService) ListRobots(ctx context.Context) ([]domain.Robot, error) {
	return s.repository.ListRobots(ctx)
}

func (s *RobotService) UpdateRobot(ctx context.Context, robotCode string, input store.UpdateRobotInput) (domain.Robot, error) {
	return s.repository.UpdateRobot(ctx, robotCode, input)
}

func (s *RobotService) ArchiveRobot(ctx context.Context, robotCode string) (domain.Robot, error) {
	return s.repository.ArchiveRobot(ctx, robotCode)
}

func (s *RobotService) GetRobotConnectionInfo(ctx context.Context, robotCode string) (domain.RobotConnectionInfo, error) {
	return s.repository.GetRobotConnectionInfo(ctx, robotCode)
}

func (s *RobotService) RotateRobotConnectionToken(ctx context.Context, robotCode string) (domain.RobotConnectionInfo, error) {
	return s.repository.RotateRobotConnectionToken(ctx, robotCode)
}

func (s *RobotService) ApplyHeartbeat(ctx context.Context, input store.HeartbeatInput, bearerToken string) (domain.Robot, error) {
	return s.repository.ApplyHeartbeat(ctx, input, bearerToken)
}

func (s *RobotService) ResolveRobotByBearerToken(ctx context.Context, bearerToken string) (domain.Robot, error) {
	return s.repository.ResolveRobotByBearerToken(ctx, bearerToken)
}
