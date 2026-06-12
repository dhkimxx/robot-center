package port

import (
	"context"

	"robot-center/apps/server/internal/domain"
)

type LiveStatusStore interface {
	GetMissionLiveStatusSnapshot(ctx context.Context, missionCode string) (MissionLiveStatusSnapshot, error)
}

type MissionLiveStatusSnapshot struct {
	Mission         domain.Mission
	Robots          []domain.Robot
	RecordingChunks []domain.RecordingChunk
	StreamSessions  []domain.RobotStreamSession
}
