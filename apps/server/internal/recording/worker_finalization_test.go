package recording

import (
	"context"
	"errors"
	"robot-center/apps/server/internal/config"
	"robot-center/apps/server/internal/domain"
	"testing"
	"time"
)

func TestWorkerTickFinalizesClaimedJobWhenMissionAlreadyEnded(t *testing.T) {
	chunk := domain.RecordingChunk{
		ID:                "chunk-001",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-001_manifest.json",
		MediaObjectKeys:   map[string]string{},
	}
	appServerClient := &fakeAppServerClient{
		claimedJobs: []domain.RecordingFinalizationJob{
			{ID: "job-001", RecordingChunkID: chunk.ID, Attempts: 2, Chunk: chunk},
		},
	}
	objectStorage := &fakeObjectStorage{manifestSize: 42}
	mediaUploader := &fakeMediaUploader{}
	worker := newWorkerWithCollaborators(
		config.RecorderWorkerConfig{RecordingChunkDuration: 10 * time.Minute},
		appServerClient,
		objectStorage,
	)
	worker.mediaUploader = mediaUploader

	worker.tick(context.Background())

	if len(mediaUploader.finalizedChunks) != 1 {
		t.Fatalf("finalized chunks = %d, want 1", len(mediaUploader.finalizedChunks))
	}
	if appServerClient.completedJobID != "job-001" {
		t.Fatalf("completed job id = %q, want job-001", appServerClient.completedJobID)
	}
	if appServerClient.markedChunkID != chunk.ID {
		t.Fatalf("marked chunk id = %q, want %q", appServerClient.markedChunkID, chunk.ID)
	}
	if appServerClient.markedChunkContext.WorkerID != worker.workerID || appServerClient.markedChunkContext.Attempt != 2 {
		t.Fatalf("marked chunk context = %#v, want worker=%q attempt=2", appServerClient.markedChunkContext, worker.workerID)
	}
	if appServerClient.completedJobContext.WorkerID != worker.workerID || appServerClient.completedJobContext.Attempt != 2 {
		t.Fatalf("completed job context = %#v, want worker=%q attempt=2", appServerClient.completedJobContext, worker.workerID)
	}
}

func TestWorkerMarksClaimedJobPartialWhenNoMediaExists(t *testing.T) {
	chunk := domain.RecordingChunk{
		ID:                "chunk-001",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-001_manifest.json",
		MediaObjectKeys:   map[string]string{},
	}
	appServerClient := &fakeAppServerClient{
		claimedJobs: []domain.RecordingFinalizationJob{
			{ID: "job-001", RecordingChunkID: chunk.ID, Attempts: 3, Chunk: chunk},
		},
	}
	objectStorage := &fakeObjectStorage{manifestSize: 42}
	mediaUploader := &fakeMediaUploader{noMedia: true}
	worker := newWorkerWithCollaborators(
		config.RecorderWorkerConfig{RecordingChunkDuration: 10 * time.Minute},
		appServerClient,
		objectStorage,
	)
	worker.mediaUploader = mediaUploader

	worker.tick(context.Background())

	if appServerClient.partialJobID != "job-001" {
		t.Fatalf("partial job id = %q, want job-001", appServerClient.partialJobID)
	}
	if appServerClient.partialJobContext.WorkerID != worker.workerID || appServerClient.partialJobContext.Attempt != 3 {
		t.Fatalf("partial job context = %#v, want worker=%q attempt=3", appServerClient.partialJobContext, worker.workerID)
	}
	if appServerClient.markedChunkID != "" {
		t.Fatalf("marked chunk id = %q, want empty when no media exists", appServerClient.markedChunkID)
	}
}

func TestWorkerMarksClaimedJobFailedWithClaimContext(t *testing.T) {
	chunk := domain.RecordingChunk{
		ID:                "chunk-001",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-001_manifest.json",
		MediaObjectKeys:   map[string]string{},
	}
	appServerClient := &fakeAppServerClient{
		claimedJobs: []domain.RecordingFinalizationJob{
			{ID: "job-001", RecordingChunkID: chunk.ID, Attempts: 4, Chunk: chunk},
		},
	}
	objectStorage := &fakeObjectStorage{manifestSize: 42}
	mediaUploader := &fakeMediaUploader{err: errors.New("mux failed")}
	worker := newWorkerWithCollaborators(
		config.RecorderWorkerConfig{RecordingChunkDuration: 10 * time.Minute},
		appServerClient,
		objectStorage,
	)
	worker.mediaUploader = mediaUploader

	worker.tick(context.Background())

	if appServerClient.failedJobID != "job-001" {
		t.Fatalf("failed job id = %q, want job-001", appServerClient.failedJobID)
	}
	if appServerClient.failedJobContext.WorkerID != worker.workerID || appServerClient.failedJobContext.Attempt != 4 {
		t.Fatalf("failed job context = %#v, want worker=%q attempt=4", appServerClient.failedJobContext, worker.workerID)
	}
	if appServerClient.markedChunkID != "" {
		t.Fatalf("marked chunk id = %q, want empty after upload failure", appServerClient.markedChunkID)
	}
}

func TestWorkerTickRetriesFailedChunkFinalization(t *testing.T) {
	target := domain.Mission{
		MissionCode: "mission-001",
		RobotCode:   "robot-001",
	}
	observedAt := time.Date(2026, 5, 26, 1, 0, 0, 0, time.UTC)
	chunk1 := domain.RecordingChunk{
		ID:                "chunk-001",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-001_manifest.json",
		MediaObjectKeys:   map[string]string{},
		EndedAt:           observedAt.Add(10 * time.Minute),
	}
	chunk2 := domain.RecordingChunk{
		ID:                "chunk-002",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-002_manifest.json",
		MediaObjectKeys:   map[string]string{},
		EndedAt:           observedAt.Add(20 * time.Minute),
	}
	appServerClient := &fakeAppServerClient{
		targets: []domain.Mission{target},
		tickResults: []domain.RecordingTickResult{
			{Chunk: chunk1},
			{Chunk: chunk2},
			{Chunk: chunk2},
		},
	}
	objectStorage := &fakeObjectStorage{manifestSize: 42}
	mediaUploader := &fakeMediaUploader{err: errors.New("temporary upload failure")}
	worker := newWorkerWithCollaborators(
		config.RecorderWorkerConfig{RecordingChunkDuration: 10 * time.Minute},
		appServerClient,
		objectStorage,
	)
	worker.mediaUploader = mediaUploader

	worker.tick(context.Background())
	if _, _, err := worker.ensureActiveRecordingChunk(context.Background(), recorderMediaKey(target.MissionCode, target.RobotCode), observedAt); err != nil {
		t.Fatalf("first chunk open failed: %v", err)
	}
	if _, _, err := worker.ensureActiveRecordingChunk(context.Background(), recorderMediaKey(target.MissionCode, target.RobotCode), observedAt.Add(11*time.Minute)); err != nil {
		t.Fatalf("second chunk open failed: %v", err)
	}
	worker.processPendingRecordingChunkFinalizations(context.Background())

	if objectStorage.manifestUploads != 0 {
		t.Fatalf("manifest uploads = %d, want 0 after failed media upload", objectStorage.manifestUploads)
	}
	if len(worker.pendingFinalizations) != 1 {
		t.Fatalf("pending finalizations = %d, want 1", len(worker.pendingFinalizations))
	}

	mediaUploader.err = nil
	worker.processPendingRecordingChunkFinalizations(context.Background())

	if objectStorage.manifestUploads != 1 {
		t.Fatalf("manifest uploads = %d, want 1 after retry", objectStorage.manifestUploads)
	}
	if appServerClient.markedChunkID != chunk1.ID {
		t.Fatalf("marked chunk id = %q, want %q", appServerClient.markedChunkID, chunk1.ID)
	}
	if len(worker.pendingFinalizations) != 0 {
		t.Fatalf("pending finalizations = %d, want 0 after retry", len(worker.pendingFinalizations))
	}
}

func TestWorkerDropsPendingFinalizationAfterAppServerConflict(t *testing.T) {
	chunk := domain.RecordingChunk{
		ID:                "chunk-001",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-001_manifest.json",
		MediaObjectKeys:   map[string]string{},
	}
	mediaKey := recorderMediaKey(chunk.MissionCode, chunk.RobotCode)
	mediaUploader := &fakeMediaUploader{err: errAppServerConflict}
	worker := newWorkerWithCollaborators(
		config.RecorderWorkerConfig{RecordingChunkDuration: 10 * time.Minute},
		&fakeAppServerClient{},
		&fakeObjectStorage{manifestSize: 42},
	)
	worker.mediaUploader = mediaUploader
	worker.pendingFinalizations[recordingChunkFinalizationKey(mediaKey, chunk.ID)] = recordingChunkFinalization{
		mediaKey: mediaKey,
		chunk:    chunk,
	}

	worker.processPendingRecordingChunkFinalizations(context.Background())

	if len(worker.pendingFinalizations) != 0 {
		t.Fatalf("pending finalizations = %d, want 0 after conflict", len(worker.pendingFinalizations))
	}
	if len(mediaUploader.finalizedChunks) != 1 {
		t.Fatalf("finalized chunks = %d, want 1", len(mediaUploader.finalizedChunks))
	}
}

func TestWorkerRemovesPendingFinalizationWhenClaimedJobHandlesSameChunk(t *testing.T) {
	chunk := domain.RecordingChunk{
		ID:                "chunk-001",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-001_manifest.json",
		MediaObjectKeys:   map[string]string{},
	}
	mediaKey := recorderMediaKey(chunk.MissionCode, chunk.RobotCode)
	appServerClient := &fakeAppServerClient{
		claimedJobs: []domain.RecordingFinalizationJob{
			{ID: "job-001", RecordingChunkID: chunk.ID, Attempts: 2, Chunk: chunk},
		},
	}
	worker := newWorkerWithCollaborators(
		config.RecorderWorkerConfig{RecordingChunkDuration: 10 * time.Minute},
		appServerClient,
		&fakeObjectStorage{manifestSize: 42},
	)
	worker.mediaUploader = &fakeMediaUploader{}
	worker.pendingFinalizations[recordingChunkFinalizationKey(mediaKey, chunk.ID)] = recordingChunkFinalization{
		mediaKey: mediaKey,
		chunk:    chunk,
	}

	worker.processClaimedRecordingFinalizationJobs(context.Background())

	if len(worker.pendingFinalizations) != 0 {
		t.Fatalf("pending finalizations = %d, want 0 after claimed job", len(worker.pendingFinalizations))
	}
	if appServerClient.completedJobID != "job-001" {
		t.Fatalf("completed job id = %q, want job-001", appServerClient.completedJobID)
	}
}
