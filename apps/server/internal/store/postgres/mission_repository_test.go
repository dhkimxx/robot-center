package postgres

import (
	"context"
	"errors"
	"testing"

	repo "robot-center/apps/server/internal/store/port"
)

func TestMissionRepositoryPersistsMultiRobotMissionLifecycle(t *testing.T) {
	store := newPostgresTestStore(t)
	ctx := context.Background()
	robotA := createRobotFixture(t, store, "Mission Robot A")
	robotB := createRobotFixture(t, store, "Mission Robot B")

	mission, err := store.CreateMission(ctx, repo.CreateMissionInput{
		Name:        "Multi Robot Mission",
		MissionType: "collapse_site",
		SiteNote:    "two robot assignment",
		RobotCodes:  []string{robotA.Robot.RobotCode, robotB.Robot.RobotCode},
	})
	if err != nil {
		t.Fatalf("create mission: %v", err)
	}
	if mission.RobotCode != robotA.Robot.RobotCode || len(mission.RobotCodes) != 2 {
		t.Fatalf("expected ordered mission robot codes, got %#v", mission)
	}

	startedMission, err := store.StartMission(ctx, mission.MissionCode)
	if err != nil {
		t.Fatalf("start mission: %v", err)
	}
	if startedMission.Status != "active" || startedMission.StartedAt == nil {
		t.Fatalf("expected active mission with startedAt, got %#v", startedMission)
	}
	for _, robotCode := range []string{robotA.Robot.RobotCode, robotB.Robot.RobotCode} {
		if err := store.ValidateActiveMissionRobot(ctx, mission.MissionCode, robotCode); err != nil {
			t.Fatalf("expected active mission robot %s: %v", robotCode, err)
		}
	}

	endedMission, err := store.EndMission(ctx, mission.MissionCode)
	if err != nil {
		t.Fatalf("end mission: %v", err)
	}
	if endedMission.Status != "ended" || endedMission.EndedAt == nil {
		t.Fatalf("expected ended mission with endedAt, got %#v", endedMission)
	}
	for _, robotCode := range []string{robotA.Robot.RobotCode, robotB.Robot.RobotCode} {
		if err := store.ValidateActiveMissionRobot(ctx, mission.MissionCode, robotCode); !errors.Is(err, repo.ErrInvalidState) {
			t.Fatalf("expected ended robot assignment %s to be inactive, got %v", robotCode, err)
		}
	}
}

func TestMissionRepositoryRejectsBusyRobotAtCreateAndStart(t *testing.T) {
	store := newPostgresTestStore(t)
	ctx := context.Background()
	busyRobot := createRobotFixture(t, store, "Busy Robot")
	freeRobot := createRobotFixture(t, store, "Free Robot")

	activeMission, err := store.CreateMission(ctx, repo.CreateMissionInput{
		Name:        "Active Mission",
		MissionType: "mountain_rescue",
		RobotCodes:  []string{busyRobot.Robot.RobotCode},
	})
	if err != nil {
		t.Fatalf("create active mission: %v", err)
	}
	if _, err := store.StartMission(ctx, activeMission.MissionCode); err != nil {
		t.Fatalf("start active mission: %v", err)
	}

	if _, err := store.CreateMission(ctx, repo.CreateMissionInput{
		Name:        "Create Conflict Mission",
		MissionType: "underground",
		RobotCodes:  []string{busyRobot.Robot.RobotCode},
	}); !isMissionStartConflictForRobot(err, busyRobot.Robot.RobotCode, activeMission.MissionCode) {
		t.Fatalf("expected create conflict for busy robot, got %v", err)
	}

	startConflictMission, err := store.CreateMission(ctx, repo.CreateMissionInput{
		Name:        "Start Conflict Mission",
		MissionType: "underground",
		RobotCodes:  []string{freeRobot.Robot.RobotCode},
	})
	if err != nil {
		t.Fatalf("create start conflict mission: %v", err)
	}
	_, err = store.sqlDB.ExecContext(ctx, `
		UPDATE mission_robots
		SET robot_id = (SELECT id FROM robots WHERE robot_code = $1)
		WHERE mission_id = (SELECT id FROM missions WHERE mission_code = $2)
	`, busyRobot.Robot.RobotCode, startConflictMission.MissionCode)
	if err != nil {
		t.Fatalf("force start conflict fixture: %v", err)
	}

	if _, err := store.StartMission(ctx, startConflictMission.MissionCode); !isMissionStartConflictForRobot(err, busyRobot.Robot.RobotCode, activeMission.MissionCode) {
		t.Fatalf("expected start conflict for busy robot, got %v", err)
	}
}

func isMissionStartConflictForRobot(err error, robotCode string, activeMissionCode string) bool {
	var conflictErr *repo.MissionStartConflictError
	if !errors.As(err, &conflictErr) {
		return false
	}
	for _, conflict := range conflictErr.Conflicts {
		if conflict.RobotCode == robotCode && conflict.ActiveMissionCode == activeMissionCode {
			return true
		}
	}
	return false
}
