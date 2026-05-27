package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
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

func TestMissionServiceEndMissionQueuesRecordingFinalizationJobsInTransaction(t *testing.T) {
	outsideRepository := &recordingStoreSpy{}
	transactionRepository := &recordingStoreSpy{
		endMissionResult: domain.Mission{MissionCode: "mission-001", Status: "ended"},
	}
	transactionRunner := &recordingTransactionRunnerSpy{repository: transactionRepository}
	service := &MissionService{
		repository:          outsideRepository,
		recordingRepository: outsideRepository,
		transactionRunner:   transactionRunner,
	}

	mission, err := service.EndMission(context.Background(), "mission-001")
	if err != nil {
		t.Fatalf("EndMission returned error: %v", err)
	}
	if mission.MissionCode != "mission-001" || mission.Status != "ended" {
		t.Fatalf("mission = %#v", mission)
	}
	if !transactionRunner.called || !transactionRunner.committed {
		t.Fatalf("transaction runner called=%v committed=%v", transactionRunner.called, transactionRunner.committed)
	}
	if transactionRepository.endMissionInput != "mission-001" {
		t.Fatalf("transaction EndMission input = %q", transactionRepository.endMissionInput)
	}
	if transactionRepository.queuedFinalizationJobs != 1 {
		t.Fatalf("transaction queued finalization jobs = %d, want 1", transactionRepository.queuedFinalizationJobs)
	}
	if outsideRepository.endMissionInput != "" || outsideRepository.queuedFinalizationJobs != 0 {
		t.Fatalf("outside repository was used: end=%q queued=%d", outsideRepository.endMissionInput, outsideRepository.queuedFinalizationJobs)
	}
}

func TestRecordingServiceApplyRecordingTickNormalizesInput(t *testing.T) {
	repository := &recordingRepositorySpy{}
	service := &RecordingService{repository: repository}

	if _, err := service.ApplyRecordingTick(context.Background(), store.RecordingTickInput{
		MissionCode: " mission-001 ",
		RobotCode:   " robot-001 ",
	}); err != nil {
		t.Fatalf("ApplyRecordingTick returned error: %v", err)
	}

	if repository.targetInput.missionCode != "mission-001" {
		t.Fatalf("MissionCode = %q, want mission-001", repository.targetInput.missionCode)
	}
	if repository.targetInput.robotCode != "robot-001" {
		t.Fatalf("RobotCode = %q, want robot-001", repository.targetInput.robotCode)
	}
	if repository.sessionInput.chunkDurationSeconds != domain.DefaultRecordingChunkDurationSeconds {
		t.Fatalf("ChunkDurationSeconds = %d, want %d", repository.sessionInput.chunkDurationSeconds, domain.DefaultRecordingChunkDurationSeconds)
	}
	if repository.sessionInput.startedAt.IsZero() {
		t.Fatal("TickAt was not populated")
	}
}

func TestRecordingServiceApplyRecordingTickPreservesExplicitInput(t *testing.T) {
	repository := &recordingRepositorySpy{}
	service := &RecordingService{repository: repository}
	tickAt := time.Date(2026, 5, 23, 1, 2, 3, 0, time.UTC)

	if _, err := service.ApplyRecordingTick(context.Background(), store.RecordingTickInput{
		MissionCode:          "mission-001",
		RobotCode:            "robot-001",
		ChunkDurationSeconds: 120,
		TickAt:               tickAt,
	}); err != nil {
		t.Fatalf("ApplyRecordingTick returned error: %v", err)
	}

	if repository.sessionInput.chunkDurationSeconds != 120 {
		t.Fatalf("ChunkDurationSeconds = %d, want 120", repository.sessionInput.chunkDurationSeconds)
	}
	if !repository.sessionInput.startedAt.Equal(tickAt) {
		t.Fatalf("TickAt = %s, want %s", repository.sessionInput.startedAt, tickAt)
	}
}

func TestRecordingServiceApplyRecordingTickCreatesChunkWithDomainRules(t *testing.T) {
	repository := &recordingRepositorySpy{
		sessionStartedAt: time.Date(2026, 5, 23, 1, 10, 0, 0, time.UTC),
	}
	service := &RecordingService{repository: repository}
	tickAt := time.Date(2026, 5, 23, 1, 22, 0, 0, time.UTC)

	result, err := service.ApplyRecordingTick(context.Background(), store.RecordingTickInput{
		MissionCode:          "mission-001",
		RobotCode:            "robot-001",
		ChunkDurationSeconds: 600,
		TickAt:               tickAt,
	})
	if err != nil {
		t.Fatalf("ApplyRecordingTick returned error: %v", err)
	}

	if repository.findChunkInput.recordingSessionID != "session-001" {
		t.Fatalf("FindRecordingChunk session = %q", repository.findChunkInput.recordingSessionID)
	}
	if repository.findChunkInput.chunkIndex != 1 {
		t.Fatalf("FindRecordingChunk chunkIndex = %d, want 1", repository.findChunkInput.chunkIndex)
	}
	if repository.createInput.MediaObjectKeys["manifest"] == "" {
		t.Fatal("expected manifest object key")
	}
	if result.Manifest["chunkId"] != "chunk-001" {
		t.Fatalf("manifest chunkId = %v", result.Manifest["chunkId"])
	}
}

func TestRecordingServiceMarkRecordingChunkUploadedUsesTransactionRepository(t *testing.T) {
	outsideRepository := &recordingStoreSpy{}
	transactionRepository := &recordingStoreSpy{}
	transactionRunner := &recordingTransactionRunnerSpy{repository: transactionRepository}
	service := &RecordingService{
		repository:        outsideRepository,
		transactionRunner: transactionRunner,
	}
	sizeBytes := int64(42)

	result, err := service.MarkRecordingChunkUploaded(context.Background(), "chunk-001", store.RecordingUploadMetadata{SizeBytes: &sizeBytes})
	if err != nil {
		t.Fatalf("MarkRecordingChunkUploaded returned error: %v", err)
	}

	if !transactionRunner.called {
		t.Fatal("expected transaction runner to be used")
	}
	if !transactionRunner.committed {
		t.Fatal("expected transaction to be committed")
	}
	if transactionRepository.markChunkUploadedInput.chunkID != "chunk-001" {
		t.Fatalf("transaction repository chunk id = %q", transactionRepository.markChunkUploadedInput.chunkID)
	}
	if outsideRepository.markChunkUploadedInput.chunkID != "" {
		t.Fatalf("outside repository was used outside transaction")
	}
	if result.ID != "chunk-001" || result.Status != "uploaded" {
		t.Fatalf("result = %#v", result)
	}
}

func TestRecordingServiceMarkRecordingFileUploadedUsesTransactionRepository(t *testing.T) {
	outsideRepository := &recordingStoreSpy{}
	transactionRepository := &recordingStoreSpy{}
	transactionRunner := &recordingTransactionRunnerSpy{repository: transactionRepository}
	service := &RecordingService{
		repository:        outsideRepository,
		transactionRunner: transactionRunner,
	}

	if _, err := service.MarkRecordingFileUploaded(context.Background(), "chunk-001", "rgb_audio_mp4", store.RecordingUploadMetadata{}); err != nil {
		t.Fatalf("MarkRecordingFileUploaded returned error: %v", err)
	}

	if !transactionRunner.called {
		t.Fatal("expected transaction runner to be used")
	}
	if !transactionRunner.committed {
		t.Fatal("expected transaction to be committed")
	}
	if transactionRepository.markFileUploadedInput.chunkID != "chunk-001" {
		t.Fatalf("transaction repository chunk id = %q", transactionRepository.markFileUploadedInput.chunkID)
	}
	if transactionRepository.markFileUploadedInput.fileType != "rgb_audio_mp4" {
		t.Fatalf("transaction repository fileType = %q", transactionRepository.markFileUploadedInput.fileType)
	}
	if outsideRepository.markFileUploadedInput.chunkID != "" {
		t.Fatalf("outside repository was used outside transaction")
	}
}

func TestRecordingServiceMarkRecordingChunkUploadedReturnsTransactionErrorWithoutCommit(t *testing.T) {
	expectedErr := errors.New("chunk update failed after storage upsert")
	outsideRepository := &recordingStoreSpy{}
	transactionRepository := &recordingStoreSpy{}
	transactionRepository.markChunkUploadedErr = expectedErr
	transactionRunner := &recordingTransactionRunnerSpy{repository: transactionRepository}
	service := &RecordingService{
		repository:        outsideRepository,
		transactionRunner: transactionRunner,
	}

	_, err := service.MarkRecordingChunkUploaded(context.Background(), "chunk-001", store.RecordingUploadMetadata{})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
	if !transactionRunner.called {
		t.Fatal("expected transaction runner to be used")
	}
	if transactionRunner.committed {
		t.Fatal("transaction committed despite repository error")
	}
	if outsideRepository.markChunkUploadedInput.chunkID != "" {
		t.Fatalf("outside repository was used outside transaction")
	}
}

func TestRecordingServiceMarkRecordingFileUploadedReturnsTransactionErrorWithoutCommit(t *testing.T) {
	expectedErr := errors.New("file status update failed after storage upsert")
	outsideRepository := &recordingStoreSpy{}
	transactionRepository := &recordingStoreSpy{}
	transactionRepository.markFileUploadedErr = expectedErr
	transactionRunner := &recordingTransactionRunnerSpy{repository: transactionRepository}
	service := &RecordingService{
		repository:        outsideRepository,
		transactionRunner: transactionRunner,
	}

	_, err := service.MarkRecordingFileUploaded(context.Background(), "chunk-001", "rgb_audio_mp4", store.RecordingUploadMetadata{})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
	if !transactionRunner.called {
		t.Fatal("expected transaction runner to be used")
	}
	if transactionRunner.committed {
		t.Fatal("transaction committed despite repository error")
	}
	if outsideRepository.markFileUploadedInput.chunkID != "" {
		t.Fatalf("outside repository was used outside transaction")
	}
}
