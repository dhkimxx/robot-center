package service

import (
	"context"

	"robot-center/apps/server/internal/store"
)

type SystemService struct {
	repository store.SystemRepository
}

func (s *SystemService) GetDatabaseUsage(ctx context.Context) (store.DatabaseUsageResult, error) {
	return s.repository.GetDatabaseUsage(ctx)
}
