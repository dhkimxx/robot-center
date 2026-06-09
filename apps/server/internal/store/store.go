package store

import (
	"context"

	"robot-center/apps/server/internal/store/port"
	postgresstore "robot-center/apps/server/internal/store/postgres"
)

type Store = port.Store

type RobotStore = port.RobotStore
type MissionStore = port.MissionStore
type SensorStore = port.SensorStore
type EventStore = port.EventStore
type RecordingStore = port.RecordingStore
type StreamSessionStore = port.StreamSessionStore
type StorageAdminStore = port.StorageAdminStore
type SystemStore = port.SystemStore

type RobotRepository = port.RobotStore
type MissionRepository = port.MissionStore
type SensorRepository = port.SensorStore
type EventRepository = port.EventStore
type RecordingRepository = port.RecordingStore
type StreamSessionRepository = port.StreamSessionStore
type StorageAdminRepository = port.StorageAdminStore
type SystemRepository = port.SystemStore
type TransactionRunner = port.TransactionRunner

type MissionStartConflict = port.MissionStartConflict
type MissionStartConflictError = port.MissionStartConflictError

type CreateRobotInput = port.CreateRobotInput
type UpdateRobotInput = port.UpdateRobotInput
type CreateMissionInput = port.CreateMissionInput
type HeartbeatInput = port.HeartbeatInput

type RecordingTickInput = port.RecordingTickInput
type RecordingTarget = port.RecordingTarget
type RecordingSession = port.RecordingSession
type CreateRecordingChunkInput = port.CreateRecordingChunkInput
type RecordingUploadMetadata = port.RecordingUploadMetadata
type MissionRecordingChunkQuery = port.MissionRecordingChunkQuery
type MissionRecordingChunkPage = port.MissionRecordingChunkPage
type MissionRecordingSummary = port.MissionRecordingSummary
type MissionRecordingRobotSummary = port.MissionRecordingRobotSummary
type StartRobotStreamSessionInput = port.StartRobotStreamSessionInput
type TouchRobotStreamSessionInput = port.TouchRobotStreamSessionInput
type EndRobotStreamSessionInput = port.EndRobotStreamSessionInput
type StorageMetadataResetResult = port.StorageMetadataResetResult
type SensorDataClearResult = port.SensorDataClearResult
type EventDataClearResult = port.EventDataClearResult
type DatabaseUsageResult = port.DatabaseUsageResult
type DatabaseTableUsage = port.DatabaseTableUsage
type EventQuery = port.EventQuery

type PostgresConfig = postgresstore.Config
type PostgresStore = postgresstore.Store

var (
	ErrNotFound     = port.ErrNotFound
	ErrUnauthorized = port.ErrUnauthorized
	ErrInvalidState = port.ErrInvalidState
)

func NewPostgresStore(ctx context.Context, cfg PostgresConfig) (*PostgresStore, error) {
	return postgresstore.NewStore(ctx, postgresstore.Config(cfg))
}
