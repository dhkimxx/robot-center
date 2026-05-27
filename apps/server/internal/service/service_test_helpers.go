package service

import (
	"context"
	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
	"time"
)

type recordingRepositorySpy struct {
	targetInput struct {
		missionCode string
		robotCode   string
	}
	sessionInput struct {
		missionID            string
		robotID              string
		chunkDurationSeconds int
		startedAt            time.Time
	}
	sessionStartedAt time.Time
	findChunkInput   struct {
		recordingSessionID string
		chunkIndex         int
	}
	createInput   store.CreateRecordingChunkInput
	existingChunk domain.RecordingChunk
	existingFound bool

	markChunkUploadedInput struct {
		chunkID  string
		metadata store.RecordingUploadMetadata
	}
	markFileUploadedInput struct {
		chunkID  string
		fileType string
		metadata store.RecordingUploadMetadata
	}

	markChunkUploadedErr error
	markFileUploadedErr  error

	queuedFinalizationJobs int64
	queueFinalizationErr   error
}

func (r *recordingRepositorySpy) FindRecordingTarget(_ context.Context, missionCode string, robotCode string) (store.RecordingTarget, error) {
	r.targetInput.missionCode = missionCode
	r.targetInput.robotCode = robotCode
	startedAt := time.Date(2026, 5, 23, 1, 0, 0, 0, time.UTC)
	return store.RecordingTarget{
		Mission: domain.Mission{
			ID:          "mission-id-001",
			MissionCode: missionCode,
			Status:      "active",
			RobotCode:   "robot-001",
			StartedAt:   &startedAt,
		},
		RobotID:   "robot-id-001",
		RobotCode: "robot-001",
	}, nil
}

func (r *recordingRepositorySpy) FindOrCreateRecordingSession(_ context.Context, missionID string, robotID string, chunkDurationSeconds int, startedAt time.Time) (store.RecordingSession, error) {
	r.sessionInput.missionID = missionID
	r.sessionInput.robotID = robotID
	r.sessionInput.chunkDurationSeconds = chunkDurationSeconds
	r.sessionInput.startedAt = startedAt
	sessionStartedAt := r.sessionStartedAt
	if sessionStartedAt.IsZero() {
		sessionStartedAt = startedAt
	}
	return store.RecordingSession{ID: "session-001", StartedAt: sessionStartedAt}, nil
}

func (r *recordingRepositorySpy) FindRecordingChunk(_ context.Context, recordingSessionID string, chunkIndex int) (domain.RecordingChunk, bool, error) {
	r.findChunkInput.recordingSessionID = recordingSessionID
	r.findChunkInput.chunkIndex = chunkIndex
	return r.existingChunk, r.existingFound, nil
}

func (r *recordingRepositorySpy) CreateRecordingChunk(_ context.Context, input store.CreateRecordingChunkInput) (domain.RecordingChunk, error) {
	r.createInput = input
	return domain.RecordingChunk{
		ID:                 "chunk-001",
		RecordingSessionID: input.RecordingSessionID,
		MissionID:          input.MissionID,
		MissionCode:        input.MissionCode,
		RobotCode:          input.RobotCode,
		ChunkIndex:         input.Window.Index,
		Status:             "recording",
		StartedAt:          input.Window.StartedAt,
		EndedAt:            input.Window.EndedAt,
		DurationSeconds:    input.Window.DurationSeconds,
		ManifestObjectKey:  input.MediaObjectKeys["manifest"],
		MediaObjectKeys:    input.MediaObjectKeys,
		CreatedAt:          input.CreatedAt,
		UpdatedAt:          input.UpdatedAt,
	}, nil
}

func (r *recordingRepositorySpy) MarkRecordingChunkUploaded(_ context.Context, chunkID string, metadata store.RecordingUploadMetadata) (domain.RecordingChunk, error) {
	r.markChunkUploadedInput.chunkID = chunkID
	r.markChunkUploadedInput.metadata = metadata
	if r.markChunkUploadedErr != nil {
		return domain.RecordingChunk{}, r.markChunkUploadedErr
	}
	return domain.RecordingChunk{ID: chunkID, Status: "uploaded"}, nil
}

func (r *recordingRepositorySpy) MarkRecordingFileUploaded(_ context.Context, chunkID string, fileType string, metadata store.RecordingUploadMetadata) (domain.RecordingChunk, error) {
	r.markFileUploadedInput.chunkID = chunkID
	r.markFileUploadedInput.fileType = fileType
	r.markFileUploadedInput.metadata = metadata
	if r.markFileUploadedErr != nil {
		return domain.RecordingChunk{}, r.markFileUploadedErr
	}
	return domain.RecordingChunk{ID: chunkID}, nil
}

func (r *recordingRepositorySpy) ListRecordingChunks(_ context.Context) ([]domain.RecordingChunk, error) {
	return nil, nil
}

func (r *recordingRepositorySpy) QueueRecordingFinalizationJobsForInactiveMissions(_ context.Context) (int64, error) {
	if r.queueFinalizationErr != nil {
		return 0, r.queueFinalizationErr
	}
	r.queuedFinalizationJobs++
	return r.queuedFinalizationJobs, nil
}

func (r *recordingRepositorySpy) ClaimRecordingFinalizationJobs(_ context.Context, _ string, _ int, _ time.Duration) ([]domain.RecordingFinalizationJob, error) {
	return nil, nil
}

func (r *recordingRepositorySpy) MarkRecordingFinalizationJobCompleted(_ context.Context, _ string, _ string, _ int) error {
	return nil
}

func (r *recordingRepositorySpy) MarkRecordingFinalizationJobPartial(_ context.Context, _ string, _ string, _ int, _ string) error {
	return nil
}

func (r *recordingRepositorySpy) MarkRecordingFinalizationJobFailed(_ context.Context, _ string, _ string, _ int, _ string) error {
	return nil
}

type recordingStoreSpy struct {
	recordingRepositorySpy
	endMissionInput  string
	endMissionResult domain.Mission
}

func (s *recordingStoreSpy) CreateRobot(_ context.Context, _ store.CreateRobotInput) (domain.Robot, domain.RobotConnectionInfo, error) {
	return domain.Robot{}, domain.RobotConnectionInfo{}, nil
}

func (s *recordingStoreSpy) ListRobots(_ context.Context) ([]domain.Robot, error) {
	return nil, nil
}

func (s *recordingStoreSpy) UpdateRobot(_ context.Context, _ string, _ store.UpdateRobotInput) (domain.Robot, error) {
	return domain.Robot{}, nil
}

func (s *recordingStoreSpy) ArchiveRobot(_ context.Context, _ string) (domain.Robot, error) {
	return domain.Robot{}, nil
}

func (s *recordingStoreSpy) GetRobotConnectionInfo(_ context.Context, _ string) (domain.RobotConnectionInfo, error) {
	return domain.RobotConnectionInfo{}, nil
}

func (s *recordingStoreSpy) RotateRobotConnectionToken(_ context.Context, _ string) (domain.RobotConnectionInfo, error) {
	return domain.RobotConnectionInfo{}, nil
}

func (s *recordingStoreSpy) ApplyHeartbeat(_ context.Context, _ store.HeartbeatInput, _ string) (domain.Robot, error) {
	return domain.Robot{}, nil
}

func (s *recordingStoreSpy) CreateMission(_ context.Context, _ store.CreateMissionInput) (domain.Mission, error) {
	return domain.Mission{}, nil
}

func (s *recordingStoreSpy) ListMissions(_ context.Context) ([]domain.Mission, error) {
	return nil, nil
}

func (s *recordingStoreSpy) StartMission(_ context.Context, _ string) (domain.Mission, error) {
	return domain.Mission{}, nil
}

func (s *recordingStoreSpy) EndMission(_ context.Context, missionCode string) (domain.Mission, error) {
	s.endMissionInput = missionCode
	if s.endMissionResult.MissionCode != "" {
		return s.endMissionResult, nil
	}
	return domain.Mission{MissionCode: missionCode, Status: "ended"}, nil
}

func (s *recordingStoreSpy) FindActiveMissionForRobot(_ context.Context, _ string, _ string) (domain.Mission, bool, error) {
	return domain.Mission{}, false, nil
}

func (s *recordingStoreSpy) ValidateActiveMissionRobot(_ context.Context, _ string, _ string) error {
	return nil
}

func (s *recordingStoreSpy) RecordingTargets(_ context.Context) ([]domain.Mission, error) {
	return nil, nil
}

func (s *recordingStoreSpy) SaveSensorEnvelope(_ context.Context, envelope domain.SensorEnvelope) ([]domain.SensorSample, error) {
	return envelope.Samples, nil
}

func (s *recordingStoreSpy) ListSensorDescriptors(_ context.Context, _ string, _ string) ([]domain.SensorDescriptor, error) {
	return nil, nil
}

func (s *recordingStoreSpy) ListSensorSamples(_ context.Context, _ string, _ string, _ string, _ int) ([]domain.SensorSample, error) {
	return nil, nil
}

func (s *recordingStoreSpy) ListLatestSensorSamples(_ context.Context, _ string, _ string) ([]domain.SensorLatest, error) {
	return nil, nil
}

type recordingTransactionRunnerSpy struct {
	repository store.Store
	called     bool
	committed  bool
}

func (r *recordingTransactionRunnerSpy) WithTransaction(ctx context.Context, run func(ctx context.Context, repository store.Store) error) error {
	r.called = true
	if err := run(ctx, r.repository); err != nil {
		return err
	}
	r.committed = true
	return nil
}
