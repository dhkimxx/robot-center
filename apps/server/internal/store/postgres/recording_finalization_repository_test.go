package postgres

import (
	"context"
	"testing"
	"time"

	"robot-center/apps/server/internal/domain"
	repo "robot-center/apps/server/internal/store/port"
)

func TestRecordingFinalizationRepositoryQueuesAndClaimsInactiveMissionChunks(t *testing.T) {
	store := newPostgresTestStore(t)
	fixture := createActiveMissionFixture(t, store)
	ctx := context.Background()
	now := time.Date(2026, 5, 28, 2, 0, 0, 0, time.UTC)

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

	if _, err := store.EndMission(ctx, fixture.Mission.MissionCode); err != nil {
		t.Fatalf("end mission: %v", err)
	}
	queued, err := store.QueueRecordingFinalizationJobsForInactiveMissions(ctx)
	if err != nil {
		t.Fatalf("queue finalization jobs: %v", err)
	}
	if queued != 1 {
		t.Fatalf("queued jobs = %d, want 1", queued)
	}

	claimedJobs, err := store.ClaimRecordingFinalizationJobs(ctx, "worker-a", 10, time.Minute)
	if err != nil {
		t.Fatalf("claim finalization jobs: %v", err)
	}
	if len(claimedJobs) != 1 {
		t.Fatalf("expected one claimed job, got %#v", claimedJobs)
	}
	if claimedJobs[0].RecordingChunkID != chunk.ID || claimedJobs[0].Attempts != 1 || claimedJobs[0].LockedBy != "worker-a" {
		t.Fatalf("unexpected claimed job: %#v", claimedJobs[0])
	}

	concurrentJobs, err := store.ClaimRecordingFinalizationJobs(ctx, "worker-b", 10, time.Minute)
	if err != nil {
		t.Fatalf("claim concurrent finalization jobs: %v", err)
	}
	if len(concurrentJobs) != 0 {
		t.Fatalf("locked job should not be claimed concurrently, got %#v", concurrentJobs)
	}

	if err := store.MarkRecordingFinalizationJobFailed(ctx, claimedJobs[0].ID, "worker-a", claimedJobs[0].Attempts, "muxing failed"); err != nil {
		t.Fatalf("mark finalization job failed: %v", err)
	}
	updatedChunk, found, err := store.FindRecordingChunk(ctx, recordingSession.ID, window.Index)
	if err != nil {
		t.Fatalf("find updated chunk: %v", err)
	}
	if !found || updatedChunk.Status != "failed" {
		t.Fatalf("expected failed chunk after failed finalization, found=%v chunk=%#v", found, updatedChunk)
	}
}
