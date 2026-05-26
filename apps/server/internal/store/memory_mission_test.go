package store

import (
	"context"
	"errors"
	"testing"

	"robot-center/apps/server/internal/domain"
)

func TestMemoryStoreStartMissionRejectsSecondActiveMissionForRobot(t *testing.T) {
	ctx := context.Background()
	repository := NewMemoryStore("http://127.0.0.1:18080")

	robot, _, err := repository.CreateRobot(ctx, CreateRobotInput{
		DisplayName: "Conflict Test Robot",
		ModelName:   "Mock",
	})
	if err != nil {
		t.Fatalf("CreateRobot returned error: %v", err)
	}

	firstMission := createReadyMissionForRobot(t, repository, robot.RobotCode, "First Mission")
	secondMission := createReadyMissionForRobot(t, repository, robot.RobotCode, "Second Mission")

	if _, err := repository.StartMission(ctx, firstMission.MissionCode); err != nil {
		t.Fatalf("StartMission(first) returned error: %v", err)
	}

	_, err = repository.StartMission(ctx, secondMission.MissionCode)
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("StartMission(second) error = %v, want %v", err, ErrInvalidState)
	}
	var conflictError *MissionStartConflictError
	if !errors.As(err, &conflictError) {
		t.Fatalf("StartMission(second) error = %T, want MissionStartConflictError", err)
	}
	if len(conflictError.Conflicts) != 1 {
		t.Fatalf("conflicts = %#v, want one conflict", conflictError.Conflicts)
	}
	if conflictError.Conflicts[0].RobotCode != robot.RobotCode || conflictError.Conflicts[0].ActiveMissionCode != firstMission.MissionCode {
		t.Fatalf("conflict = %#v, want robot %s active in %s", conflictError.Conflicts[0], robot.RobotCode, firstMission.MissionCode)
	}

	missions, err := repository.ListMissions(ctx)
	if err != nil {
		t.Fatalf("ListMissions returned error: %v", err)
	}
	assertMissionStatus(t, missions, firstMission.MissionCode, "active")
	assertMissionStatus(t, missions, secondMission.MissionCode, "ready")
}

func TestMemoryStoreValidateActiveMissionRobot(t *testing.T) {
	ctx := context.Background()
	repository := NewMemoryStore("http://127.0.0.1:18080")

	robot, _, err := repository.CreateRobot(ctx, CreateRobotInput{
		DisplayName: "Validated Robot",
		ModelName:   "Mock",
	})
	if err != nil {
		t.Fatalf("CreateRobot returned error: %v", err)
	}
	mission := createReadyMissionForRobot(t, repository, robot.RobotCode, "Validated Mission")
	if err := repository.ValidateActiveMissionRobot(ctx, mission.MissionCode, robot.RobotCode); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("ValidateActiveMissionRobot before start error = %v, want %v", err, ErrInvalidState)
	}
	if _, err := repository.StartMission(ctx, mission.MissionCode); err != nil {
		t.Fatalf("StartMission returned error: %v", err)
	}
	if err := repository.ValidateActiveMissionRobot(ctx, mission.MissionCode, robot.RobotCode); err != nil {
		t.Fatalf("ValidateActiveMissionRobot after start returned error: %v", err)
	}
	if err := repository.ValidateActiveMissionRobot(ctx, mission.MissionCode, "robot-missing"); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("ValidateActiveMissionRobot missing robot error = %v, want %v", err, ErrInvalidState)
	}
}

func createReadyMissionForRobot(t *testing.T, repository *MemoryStore, robotCode string, name string) domain.Mission {
	t.Helper()

	mission, err := repository.CreateMission(context.Background(), CreateMissionInput{
		Name:        name,
		MissionType: "mountain_rescue",
		RobotCode:   robotCode,
	})
	if err != nil {
		t.Fatalf("CreateMission(%q) returned error: %v", name, err)
	}
	return mission
}

func assertMissionStatus(t *testing.T, missions []domain.Mission, missionCode string, wantStatus string) {
	t.Helper()

	for _, mission := range missions {
		if mission.MissionCode == missionCode {
			if mission.Status != wantStatus {
				t.Fatalf("mission %s status = %q, want %q", missionCode, mission.Status, wantStatus)
			}
			return
		}
	}
	t.Fatalf("mission %s not found", missionCode)
}
