package service

import (
	"context"
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
	findChunkInput struct {
		recordingSessionID string
		chunkIndex         int
	}
	createInput   store.CreateRecordingChunkInput
	existingChunk domain.RecordingChunk
	existingFound bool
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

func (r *recordingRepositorySpy) FindOrCreateRecordingSession(_ context.Context, missionID string, robotID string, chunkDurationSeconds int, startedAt time.Time) (string, error) {
	r.sessionInput.missionID = missionID
	r.sessionInput.robotID = robotID
	r.sessionInput.chunkDurationSeconds = chunkDurationSeconds
	r.sessionInput.startedAt = startedAt
	return "session-001", nil
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

func (r *recordingRepositorySpy) MarkRecordingChunkUploaded(_ context.Context, _ string, _ store.RecordingUploadMetadata) (domain.RecordingChunk, error) {
	return domain.RecordingChunk{}, nil
}

func (r *recordingRepositorySpy) MarkRecordingFileUploaded(_ context.Context, _ string, _ string, _ store.RecordingUploadMetadata) (domain.RecordingChunk, error) {
	return domain.RecordingChunk{}, nil
}

func (r *recordingRepositorySpy) ListRecordingChunks(_ context.Context) ([]domain.RecordingChunk, error) {
	return nil, nil
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
	repository := &recordingRepositorySpy{}
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
	if repository.findChunkInput.chunkIndex != 2 {
		t.Fatalf("FindRecordingChunk chunkIndex = %d, want 2", repository.findChunkInput.chunkIndex)
	}
	if repository.createInput.MediaObjectKeys["manifest"] == "" {
		t.Fatal("expected manifest object key")
	}
	if result.Manifest["chunkId"] != "chunk-001" {
		t.Fatalf("manifest chunkId = %v", result.Manifest["chunkId"])
	}
}
