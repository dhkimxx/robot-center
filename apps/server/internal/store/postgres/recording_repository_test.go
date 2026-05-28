package postgres

import (
	"context"
	"testing"
	"time"

	"robot-center/apps/server/internal/domain"
	repo "robot-center/apps/server/internal/store/port"
)

func TestRecordingRepositoryMarksUploadedFilesAndManifest(t *testing.T) {
	store := newPostgresTestStore(t)
	fixture := createActiveMissionFixture(t, store)
	ctx := context.Background()
	now := time.Date(2026, 5, 28, 3, 0, 0, 0, time.UTC)
	chunk := createRecordingChunkFixture(t, store, fixture, now)

	rgbSize := int64(1024)
	updatedChunk, err := store.MarkRecordingFileUploaded(ctx, chunk.ID, "rgb_audio_mp4", repo.RecordingUploadMetadata{
		SizeBytes: &rgbSize,
		Checksum:  "sha256:rgb",
	})
	if err != nil {
		t.Fatalf("mark rgb file uploaded: %v", err)
	}
	if updatedChunk.Status != "recording" || !updatedChunk.AvailableFileTypes["rgb_audio_mp4"] {
		t.Fatalf("expected rgb file to be available without closing chunk, got %#v", updatedChunk)
	}

	manifestSize := int64(2048)
	uploadedChunk, err := store.MarkRecordingChunkUploaded(ctx, chunk.ID, repo.RecordingUploadMetadata{
		SizeBytes: &manifestSize,
		Checksum:  "sha256:manifest",
	})
	if err != nil {
		t.Fatalf("mark chunk uploaded: %v", err)
	}
	if uploadedChunk.Status != "uploaded" || uploadedChunk.ManifestObjectKey == "" || !uploadedChunk.AvailableFileTypes["manifest"] {
		t.Fatalf("expected uploaded chunk with manifest availability, got %#v", uploadedChunk)
	}
	if !uploadedChunk.AvailableFileTypes["rgb_audio_mp4"] {
		t.Fatalf("expected previous rgb availability to be preserved, got %#v", uploadedChunk.AvailableFileTypes)
	}

	storageObjectCount := countStorageObjectsForChunk(t, store, uploadedChunk.ID)
	if storageObjectCount != 2 {
		t.Fatalf("expected two storage objects for file and manifest, got %d", storageObjectCount)
	}
}

func createRecordingChunkFixture(t *testing.T, store *Store, fixture activeMissionFixture, now time.Time) domain.RecordingChunk {
	t.Helper()
	ctx := context.Background()
	target, err := store.FindRecordingTarget(ctx, fixture.Mission.MissionCode, fixture.Robot.RobotCode)
	if err != nil {
		t.Fatalf("find recording target: %v", err)
	}
	recordingSession, err := store.FindOrCreateRecordingSession(ctx, target.Mission.ID, target.RobotID, 600, now)
	if err != nil {
		t.Fatalf("find or create recording session: %v", err)
	}
	window := domain.NewRecordingChunkWindow(recordingSession.StartedAt, now, 600)
	chunk, err := store.CreateRecordingChunk(ctx, repo.CreateRecordingChunkInput{
		RecordingSessionID: recordingSession.ID,
		MissionID:          target.Mission.ID,
		MissionCode:        target.Mission.MissionCode,
		RobotID:            target.RobotID,
		RobotCode:          target.RobotCode,
		Window:             window,
		MediaObjectKeys:    domain.NewRecordingObjectKeys(target.Mission.MissionCode, target.RobotCode, window.StartedAt, window.EndedAt),
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	if err != nil {
		t.Fatalf("create recording chunk: %v", err)
	}
	return chunk
}

func countStorageObjectsForChunk(t *testing.T, store *Store, chunkID string) int {
	t.Helper()
	var count int
	if err := store.sqlDB.QueryRow(`
		SELECT COUNT(*)
		FROM storage_objects
		WHERE recording_chunk_id = $1::uuid
	`, chunkID).Scan(&count); err != nil {
		t.Fatalf("count storage objects: %v", err)
	}
	return count
}
