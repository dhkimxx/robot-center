package port

import (
	"context"
	"time"

	"robot-center/apps/server/internal/domain"
)

type RecordingStore interface {
	FindRecordingTarget(ctx context.Context, missionCode string, robotCode string) (RecordingTarget, error)
	FindOrCreateRecordingSession(ctx context.Context, missionID string, robotID string, chunkDurationSeconds int, startedAt time.Time) (RecordingSession, error)
	FindRecordingChunk(ctx context.Context, recordingSessionID string, chunkIndex int) (domain.RecordingChunk, bool, error)
	CreateRecordingChunk(ctx context.Context, input CreateRecordingChunkInput) (domain.RecordingChunk, error)
	MarkRecordingChunkUploaded(ctx context.Context, chunkID string, metadata RecordingUploadMetadata) (domain.RecordingChunk, error)
	MarkRecordingFileUploaded(ctx context.Context, chunkID string, fileType string, metadata RecordingUploadMetadata) (domain.RecordingChunk, error)
	QueueRecordingFinalizationJobsForInactiveMissions(ctx context.Context) (int64, error)
	ClaimRecordingFinalizationJobs(ctx context.Context, workerID string, limit int, lockDuration time.Duration) ([]domain.RecordingFinalizationJob, error)
	MarkRecordingFinalizationJobCompleted(ctx context.Context, jobID string, workerID string, attempt int) error
	MarkRecordingFinalizationJobPartial(ctx context.Context, jobID string, workerID string, attempt int, reason string) error
	MarkRecordingFinalizationJobFailed(ctx context.Context, jobID string, workerID string, attempt int, reason string) error
	ListRecordingChunks(ctx context.Context) ([]domain.RecordingChunk, error)
}
