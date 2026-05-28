package service

import (
	"context"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
)

type MissionService struct {
	repository          store.MissionRepository
	recordingRepository store.RecordingRepository
	transactionRunner   store.TransactionRunner
}

func (s *MissionService) CreateMission(ctx context.Context, input store.CreateMissionInput) (domain.Mission, error) {
	return s.repository.CreateMission(ctx, input)
}

func (s *MissionService) ListMissions(ctx context.Context) ([]domain.Mission, error) {
	return s.repository.ListMissions(ctx)
}

func (s *MissionService) StartMission(ctx context.Context, missionCode string) (domain.Mission, error) {
	return s.repository.StartMission(ctx, missionCode)
}

func (s *MissionService) EndMission(ctx context.Context, missionCode string) (domain.Mission, error) {
	if s.transactionRunner == nil {
		mission, err := s.repository.EndMission(ctx, missionCode)
		if err != nil {
			return domain.Mission{}, err
		}
		if s.recordingRepository != nil {
			if _, err := s.recordingRepository.QueueRecordingFinalizationJobsForInactiveMissions(ctx); err != nil {
				return domain.Mission{}, err
			}
		}
		return mission, nil
	}

	var mission domain.Mission
	err := s.transactionRunner.WithTransaction(ctx, func(ctx context.Context, repository store.Store) error {
		var endErr error
		mission, endErr = repository.EndMission(ctx, missionCode)
		if endErr != nil {
			return endErr
		}
		_, queueErr := repository.QueueRecordingFinalizationJobsForInactiveMissions(ctx)
		return queueErr
	})
	return mission, err
}

func (s *MissionService) FindActiveMissionForRobot(ctx context.Context, bearerToken string) (domain.Mission, bool, error) {
	return s.repository.FindActiveMissionForRobot(ctx, bearerToken)
}

func (s *MissionService) ValidateActiveMissionRobot(ctx context.Context, missionCode string, robotCode string) error {
	return s.repository.ValidateActiveMissionRobot(ctx, missionCode, robotCode)
}

func (s *MissionService) RecordingTargets(ctx context.Context) ([]domain.Mission, error) {
	return s.repository.RecordingTargets(ctx)
}
