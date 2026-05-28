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
type RecordingStore = port.RecordingStore
type StorageAdminStore = port.StorageAdminStore

type RobotRepository = port.RobotStore
type MissionRepository = port.MissionStore
type SensorRepository = port.SensorStore
type RecordingRepository = port.RecordingStore
type StorageAdminRepository = port.StorageAdminStore
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
type StorageMetadataResetResult = port.StorageMetadataResetResult

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
