package service

import (
	"context"
	"errors"
	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
	"testing"
	"time"
)

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

func TestRecordingServiceListMissionRecordingChunksNormalizesQuery(t *testing.T) {
	repository := &recordingRepositorySpy{}
	service := &RecordingService{repository: repository}

	if _, err := service.ListMissionRecordingChunks(context.Background(), store.MissionRecordingChunkQuery{
		MissionCode: " mission-001 ",
		RobotCode:   " robot-001 ",
		Limit:       999,
		Offset:      -10,
	}); err != nil {
		t.Fatalf("ListMissionRecordingChunks returned error: %v", err)
	}

	if repository.queuedFinalizationJobs != 1 {
		t.Fatalf("expected inactive mission finalization queue check, got %d", repository.queuedFinalizationJobs)
	}
	if repository.listMissionRecordingChunksInput.MissionCode != "mission-001" {
		t.Fatalf("MissionCode = %q, want mission-001", repository.listMissionRecordingChunksInput.MissionCode)
	}
	if repository.listMissionRecordingChunksInput.RobotCode != "robot-001" {
		t.Fatalf("RobotCode = %q, want robot-001", repository.listMissionRecordingChunksInput.RobotCode)
	}
	if repository.listMissionRecordingChunksInput.Limit != 300 {
		t.Fatalf("Limit = %d, want 300", repository.listMissionRecordingChunksInput.Limit)
	}
	if repository.listMissionRecordingChunksInput.Offset != 0 {
		t.Fatalf("Offset = %d, want 0", repository.listMissionRecordingChunksInput.Offset)
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
