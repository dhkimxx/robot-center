package service

import (
	"context"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
)

type SensorService struct {
	repository store.SensorRepository
}

func (s *SensorService) SaveSensorEnvelope(ctx context.Context, envelope domain.SensorEnvelope) ([]domain.SensorSample, error) {
	return s.repository.SaveSensorEnvelope(ctx, envelope)
}

func (s *SensorService) ListSensorDescriptors(ctx context.Context, missionID string, robotCode string) ([]domain.SensorDescriptor, error) {
	return s.repository.ListSensorDescriptors(ctx, missionID, robotCode)
}

func (s *SensorService) ListSensorSamples(ctx context.Context, missionID string, robotCode string, sensorID string, limit int) ([]domain.SensorSample, error) {
	return s.repository.ListSensorSamples(ctx, missionID, robotCode, sensorID, limit)
}

func (s *SensorService) ListLatestSensorSamples(ctx context.Context, missionID string, robotCode string) ([]domain.SensorLatest, error) {
	return s.repository.ListLatestSensorSamples(ctx, missionID, robotCode)
}
