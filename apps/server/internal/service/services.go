package service

import (
	"context"

	"robot-center/apps/server/internal/store"
)

type Services struct {
	Robots    *RobotService
	Missions  *MissionService
	Sensors   *SensorService
	Recording *RecordingService
	Streams   *StreamSessionService
	Live      *LiveStatusService
	Storage   *ObjectStorageAdminService

	transactionRunner store.TransactionRunner
}

func NewServices(repository store.Store) *Services {
	transactionRunner, _ := repository.(store.TransactionRunner)
	return &Services{
		Robots:            &RobotService{repository: repository},
		Missions:          &MissionService{repository: repository, recordingRepository: repository, transactionRunner: transactionRunner},
		Sensors:           &SensorService{repository: repository},
		Recording:         &RecordingService{repository: repository, transactionRunner: transactionRunner},
		Streams:           &StreamSessionService{repository: repository},
		Live:              &LiveStatusService{},
		transactionRunner: transactionRunner,
	}
}

func (s *Services) WithTransaction(ctx context.Context, run func(ctx context.Context, services *Services) error) error {
	if s.transactionRunner == nil {
		return run(ctx, s)
	}
	return s.transactionRunner.WithTransaction(ctx, func(ctx context.Context, repository store.Store) error {
		return run(ctx, NewServices(repository))
	})
}
