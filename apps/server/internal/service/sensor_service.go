package service

import (
	"context"
	"strings"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
)

const clearSensorDataConfirmation = "CLEAR_SENSOR_DATA"

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

func (s *SensorService) ClearSensorData(ctx context.Context, environment string, confirmation string) (store.SensorDataClearResult, error) {
	if strings.EqualFold(strings.TrimSpace(environment), "production") {
		return store.SensorDataClearResult{}, ErrSystemActionForbidden
	}
	if strings.TrimSpace(confirmation) != clearSensorDataConfirmation {
		return store.SensorDataClearResult{}, ErrSystemActionConfirmationRequired
	}
	return s.repository.ClearSensorData(ctx)
}
