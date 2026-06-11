package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	repo "robot-center/apps/server/internal/store/port"
)

func TestStorageAdminRepositoryPrunesOnlyCompletedStorageMetadata(t *testing.T) {
	store := newPostgresTestStore(t)
	fixture := createActiveMissionFixture(t, store)
	ctx := context.Background()
	now := time.Date(2026, 6, 11, 9, 0, 0, 0, time.UTC)

	activeChunk := createRecordingChunkFixture(t, store, fixture, now)
	activeSize := int64(1024)
	if _, err := store.MarkRecordingFileUploaded(ctx, activeChunk.ID, "rgb_audio_mp4", repo.RecordingUploadMetadata{
		SizeBytes: &activeSize,
		Checksum:  "sha256:active-rgb",
	}); err != nil {
		t.Fatalf("mark active file uploaded: %v", err)
	}

	completedChunk := createRecordingChunkFixture(t, store, fixture, now.Add(11*time.Minute))
	completedSize := int64(2048)
	completedChunk, err := store.MarkRecordingChunkUploaded(ctx, completedChunk.ID, repo.RecordingUploadMetadata{
		SizeBytes: &completedSize,
		Checksum:  "sha256:completed-manifest",
	})
	if err != nil {
		t.Fatalf("mark completed chunk uploaded: %v", err)
	}

	candidates, err := store.ListPrunableObjectStorageMetadata(ctx)
	if err != nil {
		t.Fatalf("list prunable object metadata: %v", err)
	}
	if len(candidates) != 1 || candidates[0].ObjectKey != completedChunk.ManifestObjectKey || candidates[0].SizeBytes != completedSize {
		t.Fatalf("unexpected prune candidates: %#v", candidates)
	}

	resetResult, err := store.ResetPrunedObjectStorageMetadata(ctx, []string{completedChunk.ManifestObjectKey})
	if err != nil {
		t.Fatalf("reset pruned object metadata: %v", err)
	}
	if resetResult.StorageObjectRowsDeleted != 1 || resetResult.RecordingChunksReset != 1 {
		t.Fatalf("unexpected reset result: %#v", resetResult)
	}
	if activeObjects := countStorageObjectsForChunk(t, store, activeChunk.ID); activeObjects != 1 {
		t.Fatalf("expected active chunk storage object to remain, got %d", activeObjects)
	}
	if completedObjects := countStorageObjectsForChunk(t, store, completedChunk.ID); completedObjects != 0 {
		t.Fatalf("expected completed chunk storage objects to be removed, got %d", completedObjects)
	}
	assertChunkStorageReset(t, store, completedChunk.ID)
}

func assertChunkStorageReset(t *testing.T, store *Store, chunkID string) {
	t.Helper()
	var status string
	var manifestObjectID sql.NullString
	var availableFileTypes string
	if err := store.sqlDB.QueryRow(`
		SELECT status, manifest_object_id::text, COALESCE(metadata->'availableFileTypes', '{}'::jsonb)::text
		FROM recording_chunks
		WHERE id = $1::uuid
	`, chunkID).Scan(&status, &manifestObjectID, &availableFileTypes); err != nil {
		t.Fatalf("query chunk storage metadata: %v", err)
	}
	if status != "partial" || manifestObjectID.Valid || availableFileTypes != "{}" {
		t.Fatalf("expected pruned chunk storage metadata reset, status=%q manifest=%#v available=%s", status, manifestObjectID, availableFileTypes)
	}
}
