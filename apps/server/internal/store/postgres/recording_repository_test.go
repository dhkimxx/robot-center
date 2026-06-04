package postgres

import (
	"context"
	"database/sql"
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

func TestRecordingRepositoryKeepsSessionOpenWhenLaterChunkIsRecording(t *testing.T) {
	store := newPostgresTestStore(t)
	fixture := createActiveMissionFixture(t, store)
	ctx := context.Background()
	now := time.Date(2026, 5, 28, 3, 0, 0, 0, time.UTC)
	firstChunk := createRecordingChunkFixture(t, store, fixture, now)

	target, err := store.FindRecordingTarget(ctx, fixture.Mission.MissionCode, fixture.Robot.RobotCode)
	if err != nil {
		t.Fatalf("find recording target: %v", err)
	}
	secondWindow := domain.NewRecordingChunkWindow(firstChunk.StartedAt, firstChunk.EndedAt.Add(time.Second), 600)
	if _, err := store.CreateRecordingChunk(ctx, repo.CreateRecordingChunkInput{
		RecordingSessionID: firstChunk.RecordingSessionID,
		MissionID:          target.Mission.ID,
		MissionCode:        target.Mission.MissionCode,
		RobotID:            target.RobotID,
		RobotCode:          target.RobotCode,
		Window:             secondWindow,
		MediaObjectKeys:    domain.NewRecordingObjectKeys(target.Mission.MissionCode, target.RobotCode, secondWindow.StartedAt, secondWindow.EndedAt),
		CreatedAt:          secondWindow.StartedAt,
		UpdatedAt:          secondWindow.StartedAt,
	}); err != nil {
		t.Fatalf("create second recording chunk: %v", err)
	}

	manifestSize := int64(2048)
	if _, err := store.MarkRecordingChunkUploaded(ctx, firstChunk.ID, repo.RecordingUploadMetadata{
		SizeBytes: &manifestSize,
	}); err != nil {
		t.Fatalf("mark first chunk uploaded: %v", err)
	}

	status, endedAt := recordingSessionState(t, store, firstChunk.RecordingSessionID)
	if status != "recording" {
		t.Fatalf("session status = %q, want recording", status)
	}
	if endedAt.Valid {
		t.Fatalf("session ended_at = %s, want NULL", endedAt.Time)
	}
}

func TestRecordingRepositoryRecoversSessionClosedWithOpenChunk(t *testing.T) {
	store := newPostgresTestStore(t)
	fixture := createActiveMissionFixture(t, store)
	ctx := context.Background()
	now := time.Date(2026, 5, 28, 3, 0, 0, 0, time.UTC)
	chunk := createRecordingChunkFixture(t, store, fixture, now)

	_, err := store.sqlDB.Exec(`
		UPDATE recording_sessions
		SET status = 'finalizing', ended_at = $2
		WHERE id = $1::uuid
	`, chunk.RecordingSessionID, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("force close recording session: %v", err)
	}

	target, err := store.FindRecordingTarget(ctx, fixture.Mission.MissionCode, fixture.Robot.RobotCode)
	if err != nil {
		t.Fatalf("find recording target: %v", err)
	}
	recordingSession, err := store.FindOrCreateRecordingSession(ctx, target.Mission.ID, target.RobotID, 600, now.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("find or create recording session: %v", err)
	}
	if recordingSession.ID != chunk.RecordingSessionID {
		t.Fatalf("session id = %q, want recovered %q", recordingSession.ID, chunk.RecordingSessionID)
	}

	status, endedAt := recordingSessionState(t, store, chunk.RecordingSessionID)
	if status != "recording" {
		t.Fatalf("session status = %q, want recording", status)
	}
	if endedAt.Valid {
		t.Fatalf("session ended_at = %s, want NULL", endedAt.Time)
	}
}

func TestRecordingRepositoryListSkipsClosedSessionOpenChunks(t *testing.T) {
	store := newPostgresTestStore(t)
	fixture := createActiveMissionFixture(t, store)
	ctx := context.Background()
	now := time.Date(2026, 5, 28, 3, 0, 0, 0, time.UTC)
	chunk := createRecordingChunkFixture(t, store, fixture, now)

	_, err := store.sqlDB.Exec(`
		UPDATE recording_sessions
		SET status = 'finalizing', ended_at = $2
		WHERE id = $1::uuid
	`, chunk.RecordingSessionID, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("force close recording session: %v", err)
	}

	chunks, err := store.ListRecordingChunks(ctx)
	if err != nil {
		t.Fatalf("list recording chunks: %v", err)
	}
	for _, listedChunk := range chunks {
		if listedChunk.ID == chunk.ID {
			t.Fatalf("stale recording chunk from closed session was listed: %#v", listedChunk)
		}
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

func recordingSessionState(t *testing.T, store *Store, recordingSessionID string) (string, sql.NullTime) {
	t.Helper()
	var status string
	var endedAt sql.NullTime
	if err := store.sqlDB.QueryRow(`
		SELECT status, ended_at
		FROM recording_sessions
		WHERE id = $1::uuid
	`, recordingSessionID).Scan(&status, &endedAt); err != nil {
		t.Fatalf("query recording session state: %v", err)
	}
	return status, endedAt
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
