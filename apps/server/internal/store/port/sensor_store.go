package port

import (
	"context"

	"robot-center/apps/server/internal/domain"
)

type SensorStore interface {
	SaveSensorEnvelope(ctx context.Context, envelope domain.SensorEnvelope) ([]domain.SensorSample, error)
	ListSensorDescriptors(ctx context.Context, missionID string, robotCode string) ([]domain.SensorDescriptor, error)
	ListSensorSamples(ctx context.Context, missionID string, robotCode string, sensorID string, limit int) ([]domain.SensorSample, error)
	ListLatestSensorSamples(ctx context.Context, missionID string, robotCode string) ([]domain.SensorLatest, error)
}
