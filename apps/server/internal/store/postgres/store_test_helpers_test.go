package postgres

import (
	"context"
	"testing"

	"robot-center/apps/server/internal/domain"
	repo "robot-center/apps/server/internal/store/port"
	"robot-center/apps/server/internal/testsupport/postgrestest"
)

type activeMissionFixture struct {
	Robot          domain.Robot
	ConnectionInfo domain.RobotConnectionInfo
	Mission        domain.Mission
}

type createdRobotFixture struct {
	Robot          domain.Robot
	ConnectionInfo domain.RobotConnectionInfo
}

func newPostgresTestStore(t *testing.T) *Store {
	t.Helper()

	postgresDSN := ""
	if sharedPostgresDSN != "" {
		postgresDSN = postgrestest.CreateDatabase(t, sharedPostgresDSN)
	} else {
		postgresDSN = postgrestest.Start(t).DSN
	}
	store, err := NewStore(context.Background(), Config{
		DSN:                postgresDSN,
		AppServerPublicURL: "http://test-server",
	})
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("close postgres store: %v", err)
		}
	})
	return store
}

func createActiveMissionFixture(t *testing.T, store *Store) activeMissionFixture {
	t.Helper()
	ctx := context.Background()

	robotFixture := createRobotFixture(t, store, "Repository Robot")

	mission, err := store.CreateMission(ctx, repo.CreateMissionInput{
		Name:        "Repository Mission",
		MissionType: "mountain_rescue",
		RobotCodes:  []string{robotFixture.Robot.RobotCode},
	})
	if err != nil {
		t.Fatalf("create mission: %v", err)
	}
	startedMission, err := store.StartMission(ctx, mission.MissionCode)
	if err != nil {
		t.Fatalf("start mission: %v", err)
	}

	return activeMissionFixture{
		Robot:          robotFixture.Robot,
		ConnectionInfo: robotFixture.ConnectionInfo,
		Mission:        startedMission,
	}
}

func createRobotFixture(t *testing.T, store *Store, displayName string) createdRobotFixture {
	t.Helper()

	robot, connectionInfo, err := store.CreateRobot(context.Background(), repo.CreateRobotInput{
		DisplayName: displayName,
		ModelName:   "Test Model",
	})
	if err != nil {
		t.Fatalf("create robot %q: %v", displayName, err)
	}
	return createdRobotFixture{
		Robot:          robot,
		ConnectionInfo: connectionInfo,
	}
}
