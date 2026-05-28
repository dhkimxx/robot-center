package recording

import (
	"testing"
	"time"

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

func TestRecorderStorageDataChannelLabelKeepsSensorSampleChannels(t *testing.T) {
	cases := map[string]string{
		"channel.telemetry": "channel.telemetry",
		"channel.spatial":   "channel.spatial",
		"channel.event":     "",
		"channel.control":   "",
	}
	for input, want := range cases {
		if got := recorderStorageDataChannelLabel(input); got != want {
			t.Fatalf("recorderStorageDataChannelLabel(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestRecorderDataChannelFileLabelKeepsTelemetryJSONLOnly(t *testing.T) {
	cases := map[string]string{
		"channel.telemetry": "telemetry",
		"channel.spatial":   "",
	}
	for input, want := range cases {
		if got := recorderDataChannelFileLabel(input); got != want {
			t.Fatalf("recorderDataChannelFileLabel(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSubscriberDataChannelStatusesExposeLifecycle(t *testing.T) {
	observedAt := time.Now().UTC()
	status := recorderSessionStatus{
		dataChannelLabels: map[string]struct{}{},
		dataChannelStates: map[string]recorderDataChannelRuntime{},
	}
	runtime := ensureRecorderDataChannelRuntime(&status, "channel.telemetry", observedAt)
	runtime.state = "open"
	runtime.openedAt = observedAt
	runtime.lastMessageAt = observedAt
	runtime.messageCount = 3
	status.dataChannelStates["channel.telemetry"] = runtime

	dataChannels := subscriberDataChannelStatuses(status)
	if len(dataChannels) != 1 {
		t.Fatalf("expected one data channel status, got %#v", dataChannels)
	}
	if dataChannels[0].Label != "channel.telemetry" || dataChannels[0].State != "open" || dataChannels[0].MessageCount != 3 {
		t.Fatalf("unexpected data channel status: %#v", dataChannels[0])
	}
	if dataChannels[0].DetectedAt == nil || dataChannels[0].OpenedAt == nil || dataChannels[0].LastMessageAt == nil {
		t.Fatalf("expected data channel lifecycle timestamps, got %#v", dataChannels[0])
	}
}
