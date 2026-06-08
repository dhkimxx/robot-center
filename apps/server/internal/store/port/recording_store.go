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
	SummarizeMissionRecordings(ctx context.Context, missionCode string) (MissionRecordingSummary, error)
	ListMissionRecordingChunks(ctx context.Context, query MissionRecordingChunkQuery) (MissionRecordingChunkPage, error)
}

type MissionRecordingChunkQuery struct {
	MissionCode string
	RobotCode   string
	Limit       int
	Offset      int
}

type MissionRecordingChunkPage struct {
	Chunks []domain.RecordingChunk
	Limit  int
	Offset int
	Total  int
}

type MissionRecordingSummary struct {
	MissionCode string
	TotalChunks int
	Robots      []MissionRecordingRobotSummary
}

type MissionRecordingRobotSummary struct {
	RobotCode            string
	ChunkCount           int
	UploadedChunkCount   int
	RecordingChunkCount  int
	FinalizingChunkCount int
	PartialChunkCount    int
	FirstStartedAt       *time.Time
	LastEndedAt          *time.Time
	AvailableFileCounts  map[string]int
	MissingFileCounts    map[string]int
}
