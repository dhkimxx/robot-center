package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"robot-center/apps/server/internal/domain"
)

func TestMemoryStoreRejectsStaleFinalizationJobStatusUpdate(t *testing.T) {
	repository := NewMemoryStore("http://127.0.0.1:18080")
	now := time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)
	lockedUntil := now.Add(time.Minute)
	chunk := domain.RecordingChunk{
		ID:                 "chunk-001",
		RecordingSessionID: "session-001",
		MissionID:          "mission-id-001",
		MissionCode:        "mission-001",
		RobotCode:          "robot-001",
		Status:             "finalizing",
		AvailableFileTypes: map[string]bool{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	repository.recordingChunksByID[chunk.ID] = chunk
	repository.finalizationJobsByID["job-001"] = domain.RecordingFinalizationJob{
		ID:                 "job-001",
		RecordingChunkID:   chunk.ID,
		RecordingSessionID: chunk.RecordingSessionID,
		MissionID:          chunk.MissionID,
		Status:             "processing",
		Attempts:           2,
		LockedBy:           "worker-current",
		LockedUntil:        &lockedUntil,
		CreatedAt:          now,
		UpdatedAt:          now,
		Chunk:              chunk,
	}

	err := repository.MarkRecordingFinalizationJobFailed(context.Background(), "job-001", "worker-stale", 1, "late failure")
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("stale job failure error = %v, want ErrInvalidState", err)
	}
	if job := repository.finalizationJobsByID["job-001"]; job.Status != "processing" {
		t.Fatalf("stale job update changed status to %q, want processing", job.Status)
	}
	if updatedChunk := repository.recordingChunksByID[chunk.ID]; updatedChunk.Status != "finalizing" {
		t.Fatalf("stale job update changed chunk status to %q, want finalizing", updatedChunk.Status)
	}

	if err := repository.MarkRecordingFinalizationJobFailed(context.Background(), "job-001", "worker-current", 2, "real failure"); err != nil {
		t.Fatalf("current job failure returned error: %v", err)
	}
	if job := repository.finalizationJobsByID["job-001"]; job.Status != "failed" {
		t.Fatalf("current job status = %q, want failed", job.Status)
	}
	if updatedChunk := repository.recordingChunksByID[chunk.ID]; updatedChunk.Status != "failed" {
		t.Fatalf("current job chunk status = %q, want failed", updatedChunk.Status)
	}
}

func TestMemoryStoreRejectsStaleFinalizationUploadCallback(t *testing.T) {
	repository := NewMemoryStore("http://127.0.0.1:18080")
	now := time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)
	lockedUntil := now.Add(time.Minute)
	chunk := domain.RecordingChunk{
		ID:                 "chunk-001",
		RecordingSessionID: "session-001",
		MissionID:          "mission-id-001",
		MissionCode:        "mission-001",
		RobotCode:          "robot-001",
		Status:             "finalizing",
		AvailableFileTypes: map[string]bool{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	repository.recordingChunksByID[chunk.ID] = chunk
	repository.finalizationJobsByID["job-001"] = domain.RecordingFinalizationJob{
		ID:                 "job-001",
		RecordingChunkID:   chunk.ID,
		RecordingSessionID: chunk.RecordingSessionID,
		MissionID:          chunk.MissionID,
		Status:             "processing",
		Attempts:           2,
		LockedBy:           "worker-current",
		LockedUntil:        &lockedUntil,
		CreatedAt:          now,
		UpdatedAt:          now,
		Chunk:              chunk,
	}

	_, err := repository.MarkRecordingFileUploaded(context.Background(), chunk.ID, "rgb_audio_mp4", RecordingUploadMetadata{
		WorkerID: "worker-stale",
		Attempt:  1,
	})
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("stale upload callback error = %v, want ErrInvalidState", err)
	}
	if repository.recordingChunksByID[chunk.ID].AvailableFileTypes["rgb_audio_mp4"] {
		t.Fatal("stale upload callback marked file as available")
	}

	_, err = repository.MarkRecordingFileUploaded(context.Background(), chunk.ID, "rgb_audio_mp4", RecordingUploadMetadata{
		WorkerID: "worker-current",
		Attempt:  2,
	})
	if err != nil {
		t.Fatalf("current upload callback returned error: %v", err)
	}
	if !repository.recordingChunksByID[chunk.ID].AvailableFileTypes["rgb_audio_mp4"] {
		t.Fatal("current upload callback did not mark file as available")
	}
}
