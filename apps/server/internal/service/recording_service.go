package service

import (
	"context"
	"strings"
	"time"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
)

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

	recordingSession, err := repository.FindOrCreateRecordingSession(ctx, target.Mission.ID, target.RobotID, input.ChunkDurationSeconds, input.TickAt)
	if err != nil {
		return domain.RecordingTickResult{}, err
	}

	chunkBase := recordingSession.StartedAt
	if chunkBase.IsZero() {
		chunkBase = input.TickAt
	}
	chunkWindow := domain.NewRecordingChunkWindow(chunkBase, input.TickAt, input.ChunkDurationSeconds)
	existingChunk, found, err := repository.FindRecordingChunk(ctx, recordingSession.ID, chunkWindow.Index)
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
		RecordingSessionID: recordingSession.ID,
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
	if _, err := s.repository.QueueRecordingFinalizationJobsForInactiveMissions(ctx); err != nil {
		return nil, err
	}
	return s.repository.ListRecordingChunks(ctx)
}

func (s *RecordingService) SummarizeMissionRecordings(ctx context.Context, missionCode string) (store.MissionRecordingSummary, error) {
	if _, err := s.repository.QueueRecordingFinalizationJobsForInactiveMissions(ctx); err != nil {
		return store.MissionRecordingSummary{}, err
	}
	return s.repository.SummarizeMissionRecordings(ctx, strings.TrimSpace(missionCode))
}

func (s *RecordingService) ListMissionRecordingChunks(ctx context.Context, query store.MissionRecordingChunkQuery) (store.MissionRecordingChunkPage, error) {
	if _, err := s.repository.QueueRecordingFinalizationJobsForInactiveMissions(ctx); err != nil {
		return store.MissionRecordingChunkPage{}, err
	}
	query.MissionCode = strings.TrimSpace(query.MissionCode)
	query.RobotCode = strings.TrimSpace(query.RobotCode)
	if query.Limit <= 0 {
		query.Limit = 100
	}
	if query.Limit > 300 {
		query.Limit = 300
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	return s.repository.ListMissionRecordingChunks(ctx, query)
}

func (s *RecordingService) ClaimFinalizationJobs(ctx context.Context, workerID string, limit int, lockDuration time.Duration) ([]domain.RecordingFinalizationJob, error) {
	return s.repository.ClaimRecordingFinalizationJobs(ctx, workerID, limit, lockDuration)
}

func (s *RecordingService) MarkFinalizationJobCompleted(ctx context.Context, jobID string, workerID string, attempt int) error {
	return s.repository.MarkRecordingFinalizationJobCompleted(ctx, jobID, workerID, attempt)
}

func (s *RecordingService) MarkFinalizationJobPartial(ctx context.Context, jobID string, workerID string, attempt int, reason string) error {
	return s.repository.MarkRecordingFinalizationJobPartial(ctx, jobID, workerID, attempt, reason)
}

func (s *RecordingService) MarkFinalizationJobFailed(ctx context.Context, jobID string, workerID string, attempt int, reason string) error {
	return s.repository.MarkRecordingFinalizationJobFailed(ctx, jobID, workerID, attempt, reason)
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
