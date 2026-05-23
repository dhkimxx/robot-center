package recording

import (
	"testing"

	"robot-center/apps/server/internal/domain"
)

func TestGroupRecordingTargetsByMissionUsesOneRoomPerMission(t *testing.T) {
	targetsByRoom := groupRecordingTargetsByMission([]domain.Mission{
		{MissionCode: "mission-001", RobotCode: "robot-001"},
		{MissionCode: "mission-001", RobotCode: "robot-002"},
		{MissionCode: "mission-002", RobotCode: "robot-003"},
		{MissionCode: "mission-003"},
	})

	missionOneTargets := targetsByRoom["mission-001"]
	if len(missionOneTargets) != 2 {
		t.Fatalf("expected mission-001 recorder room to handle 2 robots, got %#v", missionOneTargets)
	}
	if _, ok := targetsByRoom["mission-001__robot-001"]; ok {
		t.Fatalf("expected recorder room key to be missionCode, got robot-scoped room")
	}
	if len(targetsByRoom["mission-002"]) != 1 {
		t.Fatalf("expected mission-002 recorder room, got %#v", targetsByRoom)
	}
	if _, ok := targetsByRoom["mission-003"]; ok {
		t.Fatalf("expected target without robotCode to be skipped")
	}
}

func TestRecorderMediaKeySeparatesRobotStorageWithinMissionRoom(t *testing.T) {
	robotOneKey := recorderMediaKey("mission-001", "robot-001")
	robotTwoKey := recorderMediaKey("mission-001", "robot-002")
	if robotOneKey == robotTwoKey {
		t.Fatalf("expected robot-specific media keys, got %q", robotOneKey)
	}
	missionCode, robotCode := splitRecorderMediaKey(robotOneKey)
	if missionCode != "mission-001" || robotCode != "robot-001" {
		t.Fatalf("expected media key to preserve mission and robot, got mission=%q robot=%q", missionCode, robotCode)
	}
}
