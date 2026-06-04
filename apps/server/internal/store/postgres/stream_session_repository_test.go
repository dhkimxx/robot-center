package postgres

import (
	"context"
	"testing"
	"time"

	repo "robot-center/apps/server/internal/store/port"
)

func TestStreamSessionRepositoryPersistsRobotPublisherLifecycle(t *testing.T) {
	store := newPostgresTestStore(t)
	ctx := context.Background()
	fixture := createActiveMissionFixture(t, store)
	startedAt := time.Date(2026, 6, 4, 4, 30, 0, 0, time.UTC)

	firstSession, err := store.StartRobotStreamSession(ctx, repo.StartRobotStreamSessionInput{
		MissionCode:     fixture.Mission.MissionCode,
		RobotCode:       fixture.Robot.RobotCode,
		PublisherPeerID: "publisher-peer-1",
		StartedAt:       startedAt,
	})
	if err != nil {
		t.Fatalf("start first stream session: %v", err)
	}
	if firstSession.State != "active" || firstSession.LastMediaAt == nil || !firstSession.LastMediaAt.Equal(startedAt) {
		t.Fatalf("unexpected first session: %#v", firstSession)
	}

	touchedAt := startedAt.Add(15 * time.Second)
	if err := store.TouchRobotStreamSession(ctx, repo.TouchRobotStreamSessionInput{
		PublisherPeerID: "publisher-peer-1",
		ObservedAt:      touchedAt,
	}); err != nil {
		t.Fatalf("touch stream session: %v", err)
	}

	secondStartedAt := startedAt.Add(time.Minute)
	secondSession, err := store.StartRobotStreamSession(ctx, repo.StartRobotStreamSessionInput{
		MissionCode:     fixture.Mission.MissionCode,
		RobotCode:       fixture.Robot.RobotCode,
		PublisherPeerID: "publisher-peer-2",
		StartedAt:       secondStartedAt,
	})
	if err != nil {
		t.Fatalf("start replacement stream session: %v", err)
	}
	if secondSession.State != "active" {
		t.Fatalf("second session state = %q, want active", secondSession.State)
	}

	endedAt := secondStartedAt.Add(30 * time.Second)
	if err := store.EndRobotStreamSession(ctx, repo.EndRobotStreamSessionInput{
		PublisherPeerID: "publisher-peer-2",
		Reason:          "peer_left",
		EndedAt:         endedAt,
	}); err != nil {
		t.Fatalf("end stream session: %v", err)
	}

	sessions, err := store.ListRobotStreamSessionsForMission(ctx, fixture.Mission.MissionCode)
	if err != nil {
		t.Fatalf("list stream sessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("session count = %d, want 2: %#v", len(sessions), sessions)
	}
	if sessions[0].PublisherPeerID != "publisher-peer-2" || sessions[0].EndReason != "peer_left" {
		t.Fatalf("unexpected latest session: %#v", sessions[0])
	}
	if sessions[1].PublisherPeerID != "publisher-peer-1" || sessions[1].EndReason != "replaced" {
		t.Fatalf("unexpected replaced session: %#v", sessions[1])
	}
	if sessions[1].LastMediaAt == nil || !sessions[1].LastMediaAt.Equal(touchedAt) {
		t.Fatalf("replaced session lastMediaAt = %#v, want %s", sessions[1].LastMediaAt, touchedAt)
	}
}
