package postgres

import (
	"context"
	"testing"
	"time"

	repo "robot-center/apps/server/internal/store/port"
)

func TestLiveStatusRepositoryLoadsMissionScopedSnapshot(t *testing.T) {
	store := newPostgresTestStore(t)
	fixture := createActiveMissionFixture(t, store)
	otherFixture := createActiveMissionFixture(t, store)
	ctx := context.Background()
	base := time.Date(2026, 6, 12, 1, 0, 0, 0, time.UTC)

	firstChunk := createRecordingChunkFixture(t, store, fixture, base)
	secondChunk := createRecordingChunkFixture(t, store, fixture, base.Add(10*time.Minute))
	createRecordingChunkFixture(t, store, otherFixture, base.Add(20*time.Minute))
	if _, err := store.MarkRecordingChunkUploaded(ctx, firstChunk.ID, repo.RecordingUploadMetadata{}); err != nil {
		t.Fatalf("mark first chunk uploaded: %v", err)
	}
	if _, err := store.StartRobotStreamSession(ctx, repo.StartRobotStreamSessionInput{
		MissionCode:     fixture.Mission.MissionCode,
		RobotCode:       fixture.Robot.RobotCode,
		PublisherPeerID: "publisher-main",
		StartedAt:       base,
	}); err != nil {
		t.Fatalf("start stream session: %v", err)
	}
	if _, err := store.StartRobotStreamSession(ctx, repo.StartRobotStreamSessionInput{
		MissionCode:     otherFixture.Mission.MissionCode,
		RobotCode:       otherFixture.Robot.RobotCode,
		PublisherPeerID: "publisher-other",
		StartedAt:       base,
	}); err != nil {
		t.Fatalf("start other stream session: %v", err)
	}

	snapshot, err := store.GetMissionLiveStatusSnapshot(ctx, fixture.Mission.MissionCode)
	if err != nil {
		t.Fatalf("get live status snapshot: %v", err)
	}
	if snapshot.Mission.MissionCode != fixture.Mission.MissionCode {
		t.Fatalf("snapshot mission = %#v, want %s", snapshot.Mission, fixture.Mission.MissionCode)
	}
	if len(snapshot.Robots) != 1 || snapshot.Robots[0].RobotCode != fixture.Robot.RobotCode {
		t.Fatalf("expected only assigned robot, got %#v", snapshot.Robots)
	}
	if len(snapshot.RecordingChunks) != 1 || snapshot.RecordingChunks[0].ID != secondChunk.ID {
		t.Fatalf("expected latest live chunk %s only, got %#v", secondChunk.ID, snapshot.RecordingChunks)
	}
	if len(snapshot.StreamSessions) != 1 || snapshot.StreamSessions[0].PublisherPeerID != "publisher-main" {
		t.Fatalf("expected mission-scoped stream session, got %#v", snapshot.StreamSessions)
	}
}
