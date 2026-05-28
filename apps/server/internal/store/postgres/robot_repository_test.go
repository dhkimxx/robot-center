package postgres

import (
	"context"
	"testing"
	"time"

	"robot-center/apps/server/internal/domain"
	repo "robot-center/apps/server/internal/store/port"
)

func TestRobotRepositoryPersistsRobotMissionAndHeartbeat(t *testing.T) {
	store := newPostgresTestStore(t)
	ctx := context.Background()

	robot, connectionInfo, err := store.CreateRobot(ctx, repo.CreateRobotInput{
		DisplayName: "Repository Robot",
		ModelName:   "Test Model",
	})
	if err != nil {
		t.Fatalf("create robot: %v", err)
	}
	if robot.RobotCode != "robot-001" || connectionInfo.RobotToken == "" {
		t.Fatalf("unexpected robot create result robot=%#v connection=%#v", robot, connectionInfo)
	}

	updatedRobot, err := store.ApplyHeartbeat(ctx, repo.HeartbeatInput{
		State:          string(domain.RobotDeviceStateOnline),
		BatteryPercent: 88,
		NetworkQuality: "test",
		SentAt:         time.Now().UTC(),
	}, connectionInfo.RobotToken)
	if err != nil {
		t.Fatalf("apply heartbeat: %v", err)
	}
	if !updatedRobot.IsOnline(time.Now().UTC(), domain.DefaultRobotHeartbeatTTL) {
		t.Fatalf("expected heartbeat robot to be online, got %#v", updatedRobot)
	}

	mission, err := store.CreateMission(ctx, repo.CreateMissionInput{
		Name:        "Repository Mission",
		MissionType: "mountain_rescue",
		RobotCodes:  []string{robot.RobotCode},
	})
	if err != nil {
		t.Fatalf("create mission: %v", err)
	}
	if mission.MissionCode != "mission-001" {
		t.Fatalf("unexpected mission code: %#v", mission)
	}

	startedMission, err := store.StartMission(ctx, mission.MissionCode)
	if err != nil {
		t.Fatalf("start mission: %v", err)
	}
	if startedMission.Status != "active" {
		t.Fatalf("expected active mission, got %#v", startedMission)
	}

	activeMission, found, err := store.FindActiveMissionForRobot(ctx, "", connectionInfo.RobotToken)
	if err != nil {
		t.Fatalf("find active mission by token: %v", err)
	}
	if !found || activeMission.MissionCode != mission.MissionCode {
		t.Fatalf("expected active mission %s, found=%v mission=%#v", mission.MissionCode, found, activeMission)
	}
}
