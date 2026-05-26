package store

import (
	"context"
	"time"

	"robot-center/apps/server/internal/domain"
)

type Store interface {
	RobotRepository
	MissionRepository
	SensorRepository
	RecordingRepository
}

// TransactionRunner is the service-level boundary for composite repository flows.
// Repository implementations decide how their storage operations join the transaction.
type TransactionRunner interface {
	WithTransaction(ctx context.Context, run func(ctx context.Context, repository Store) error) error
}

type RobotRepository interface {
	CreateRobot(ctx context.Context, input CreateRobotInput) (domain.Robot, domain.RobotConnectionInfo, error)
	ListRobots(ctx context.Context) ([]domain.Robot, error)
	UpdateRobot(ctx context.Context, robotCode string, input UpdateRobotInput) (domain.Robot, error)
	ArchiveRobot(ctx context.Context, robotCode string) (domain.Robot, error)
	GetRobotConnectionInfo(ctx context.Context, robotCode string) (domain.RobotConnectionInfo, error)
	RotateRobotConnectionToken(ctx context.Context, robotCode string) (domain.RobotConnectionInfo, error)
	ApplyHeartbeat(ctx context.Context, input HeartbeatInput, bearerToken string) (domain.Robot, error)
}

type MissionRepository interface {
	CreateMission(ctx context.Context, input CreateMissionInput) (domain.Mission, error)
	ListMissions(ctx context.Context) ([]domain.Mission, error)
	StartMission(ctx context.Context, missionCode string) (domain.Mission, error)
	EndMission(ctx context.Context, missionCode string) (domain.Mission, error)
	FindActiveMissionForRobot(ctx context.Context, robotCode string, bearerToken string) (domain.Mission, bool, error)
	ValidateActiveMissionRobot(ctx context.Context, missionCode string, robotCode string) error
	RecordingTargets(ctx context.Context) ([]domain.Mission, error)
}

type SensorRepository interface {
	SaveSensorEnvelope(ctx context.Context, envelope domain.SensorEnvelope) ([]domain.SensorSample, error)
	ListSensorDescriptors(ctx context.Context, missionID string, robotCode string) ([]domain.SensorDescriptor, error)
	ListSensorSamples(ctx context.Context, missionID string, robotCode string, sensorID string, limit int) ([]domain.SensorSample, error)
	ListLatestSensorSamples(ctx context.Context, missionID string, robotCode string) ([]domain.SensorLatest, error)
}

type RecordingRepository interface {
	FindRecordingTarget(ctx context.Context, missionCode string, robotCode string) (RecordingTarget, error)
	FindOrCreateRecordingSession(ctx context.Context, missionID string, robotID string, chunkDurationSeconds int, startedAt time.Time) (RecordingSession, error)
	FindRecordingChunk(ctx context.Context, recordingSessionID string, chunkIndex int) (domain.RecordingChunk, bool, error)
	CreateRecordingChunk(ctx context.Context, input CreateRecordingChunkInput) (domain.RecordingChunk, error)
	MarkRecordingChunkUploaded(ctx context.Context, chunkID string, metadata RecordingUploadMetadata) (domain.RecordingChunk, error)
	MarkRecordingFileUploaded(ctx context.Context, chunkID string, fileType string, metadata RecordingUploadMetadata) (domain.RecordingChunk, error)
	ListRecordingChunks(ctx context.Context) ([]domain.RecordingChunk, error)
}
