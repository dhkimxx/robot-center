package service

import (
	"context"
	"strings"
	"time"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
)

type Services struct {
	Robots    *RobotService
	Missions  *MissionService
	Streaming *StreamingService
	Sensors   *SensorService
	Recording *RecordingService

	transactionRunner store.TransactionRunner
}

func NewServices(repository store.Store) *Services {
	transactionRunner, _ := repository.(store.TransactionRunner)
	return &Services{
		Robots:            &RobotService{repository: repository},
		Missions:          &MissionService{repository: repository},
		Streaming:         &StreamingService{repository: repository},
		Sensors:           &SensorService{repository: repository},
		Recording:         &RecordingService{repository: repository, transactionRunner: transactionRunner},
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

type RobotService struct {
	repository store.RobotRepository
}

func (s *RobotService) CreateRobot(ctx context.Context, input store.CreateRobotInput) (domain.Robot, domain.RobotConnectionInfo, error) {
	return s.repository.CreateRobot(ctx, input)
}

func (s *RobotService) ListRobots(ctx context.Context) ([]domain.Robot, error) {
	return s.repository.ListRobots(ctx)
}

func (s *RobotService) UpdateRobot(ctx context.Context, robotCode string, input store.UpdateRobotInput) (domain.Robot, error) {
	return s.repository.UpdateRobot(ctx, robotCode, input)
}

func (s *RobotService) ArchiveRobot(ctx context.Context, robotCode string) (domain.Robot, error) {
	return s.repository.ArchiveRobot(ctx, robotCode)
}

func (s *RobotService) GetRobotConnectionInfo(ctx context.Context, robotCode string) (domain.RobotConnectionInfo, error) {
	return s.repository.GetRobotConnectionInfo(ctx, robotCode)
}

func (s *RobotService) RotateRobotConnectionToken(ctx context.Context, robotCode string) (domain.RobotConnectionInfo, error) {
	return s.repository.RotateRobotConnectionToken(ctx, robotCode)
}

func (s *RobotService) ApplyHeartbeat(ctx context.Context, input store.HeartbeatInput, bearerToken string) (domain.Robot, error) {
	return s.repository.ApplyHeartbeat(ctx, input, bearerToken)
}

type MissionService struct {
	repository store.MissionRepository
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
	return s.repository.EndMission(ctx, missionCode)
}

func (s *MissionService) FindActiveMissionForRobot(ctx context.Context, robotCode string, bearerToken string) (domain.Mission, bool, error) {
	return s.repository.FindActiveMissionForRobot(ctx, robotCode, bearerToken)
}

func (s *MissionService) ValidateActiveMissionRobot(ctx context.Context, missionCode string, robotCode string) error {
	return s.repository.ValidateActiveMissionRobot(ctx, missionCode, robotCode)
}

func (s *MissionService) RecordingTargets(ctx context.Context) ([]domain.Mission, error) {
	return s.repository.RecordingTargets(ctx)
}

type StreamingService struct {
	repository store.StreamingRepository
}

func (s *StreamingService) ApplyStreamingStatus(ctx context.Context, status domain.StreamingStatus, bearerToken string) (domain.Robot, error) {
	return s.repository.ApplyStreamingStatus(ctx, status, bearerToken)
}

func (s *StreamingService) ListStreamingStatuses(ctx context.Context) ([]domain.StreamingStatus, error) {
	return s.repository.ListStreamingStatuses(ctx)
}

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

type RecordingService struct {
	repository        store.RecordingRepository
	transactionRunner store.TransactionRunner
}

func (s *RecordingService) ApplyRecordingTick(ctx context.Context, input store.RecordingTickInput) (domain.RecordingTickResult, error) {
	normalizedInput := normalizeRecordingTickInput(input)
	if s.transactionRunner == nil {
		return s.applyRecordingTick(ctx, s.repository, normalizedInput)
	}
	var result domain.RecordingTickResult
	err := s.transactionRunner.WithTransaction(ctx, func(ctx context.Context, repository store.Store) error {
		var applyErr error
		result, applyErr = s.applyRecordingTick(ctx, repository, normalizedInput)
		return applyErr
	})
	return result, err
}

func normalizeRecordingTickInput(input store.RecordingTickInput) store.RecordingTickInput {
	input.MissionCode = strings.TrimSpace(input.MissionCode)
	input.RobotCode = strings.TrimSpace(input.RobotCode)
	if input.ChunkDurationSeconds <= 0 {
		input.ChunkDurationSeconds = domain.DefaultRecordingChunkDurationSeconds
	}
	if input.TickAt.IsZero() {
		input.TickAt = time.Now().UTC()
	}
	return input
}

func (s *RecordingService) applyRecordingTick(ctx context.Context, repository store.RecordingRepository, input store.RecordingTickInput) (domain.RecordingTickResult, error) {
	target, err := repository.FindRecordingTarget(ctx, input.MissionCode, input.RobotCode)
	if err != nil {
		return domain.RecordingTickResult{}, err
	}
	if target.Mission.Status != "active" {
		return domain.RecordingTickResult{}, store.ErrInvalidState
	}
	if input.RobotCode == "" {
		input.RobotCode = strings.TrimSpace(target.RobotCode)
	}
	if input.RobotCode == "" || strings.TrimSpace(target.RobotCode) != input.RobotCode {
		return domain.RecordingTickResult{}, store.ErrInvalidState
	}

	recordingSessionID, err := repository.FindOrCreateRecordingSession(ctx, target.Mission.ID, target.RobotID, input.ChunkDurationSeconds, input.TickAt)
	if err != nil {
		return domain.RecordingTickResult{}, err
	}

	chunkBase := input.TickAt
	if target.Mission.StartedAt != nil {
		chunkBase = *target.Mission.StartedAt
	}
	chunkWindow := domain.NewRecordingChunkWindow(chunkBase, input.TickAt, input.ChunkDurationSeconds)
	existingChunk, found, err := repository.FindRecordingChunk(ctx, recordingSessionID, chunkWindow.Index)
	if err != nil {
		return domain.RecordingTickResult{}, err
	}
	if found {
		return domain.RecordingTickResult{
			Chunk:    existingChunk,
			Manifest: domain.NewRecordingManifest(existingChunk),
		}, nil
	}

	mediaKeys := domain.NewRecordingObjectKeys(target.Mission.MissionCode, input.RobotCode, chunkWindow.StartedAt, chunkWindow.EndedAt)
	chunk, err := repository.CreateRecordingChunk(ctx, store.CreateRecordingChunkInput{
		RecordingSessionID: recordingSessionID,
		MissionID:          target.Mission.ID,
		MissionCode:        target.Mission.MissionCode,
		RobotID:            target.RobotID,
		RobotCode:          input.RobotCode,
		Window:             chunkWindow,
		MediaObjectKeys:    mediaKeys,
		CreatedAt:          input.TickAt,
		UpdatedAt:          input.TickAt,
	})
	if err != nil {
		return domain.RecordingTickResult{}, err
	}
	return domain.RecordingTickResult{
		Chunk:    chunk,
		Manifest: domain.NewRecordingManifest(chunk),
	}, nil
}

func (s *RecordingService) MarkRecordingChunkUploaded(ctx context.Context, chunkID string, metadata store.RecordingUploadMetadata) (domain.RecordingChunk, error) {
	return s.withRecordingTransaction(ctx, func(ctx context.Context, repository store.RecordingRepository) (domain.RecordingChunk, error) {
		return repository.MarkRecordingChunkUploaded(ctx, chunkID, metadata)
	})
}

func (s *RecordingService) MarkRecordingFileUploaded(ctx context.Context, chunkID string, fileType string, metadata store.RecordingUploadMetadata) (domain.RecordingChunk, error) {
	return s.withRecordingTransaction(ctx, func(ctx context.Context, repository store.RecordingRepository) (domain.RecordingChunk, error) {
		return repository.MarkRecordingFileUploaded(ctx, chunkID, fileType, metadata)
	})
}

func (s *RecordingService) ListRecordingChunks(ctx context.Context) ([]domain.RecordingChunk, error) {
	return s.repository.ListRecordingChunks(ctx)
}

func (s *RecordingService) withRecordingTransaction(ctx context.Context, run func(ctx context.Context, repository store.RecordingRepository) (domain.RecordingChunk, error)) (domain.RecordingChunk, error) {
	if s.transactionRunner == nil {
		return run(ctx, s.repository)
	}
	var result domain.RecordingChunk
	err := s.transactionRunner.WithTransaction(ctx, func(ctx context.Context, repository store.Store) error {
		var runErr error
		result, runErr = run(ctx, repository)
		return runErr
	})
	return result, err
}
