package recording

import (
	"testing"
	"time"

	"robot-center/apps/server/internal/config"
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
		"channel.spatial":   "",
		"channel.event":     "",
		"channel.control":   "",
		"telemetry":         "",
		"sensor":            "",
		"spatial":           "",
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

func TestResetRecorderTrackRuntimeClearsStaleRobotTracksOnly(t *testing.T) {
	worker := NewWorker(config.RecorderWorkerConfig{})
	worker.subscriberStatuses["mission-001"] = recorderSessionStatus{
		missionCode: "mission-001",
		robotCodes: map[string]struct{}{
			"robot-001": {},
			"robot-002": {},
		},
		trackLabels: map[string]struct{}{
			"robot-001:track.video_1": {},
			"robot-001:video":         {},
			"robot-002:track.video_1": {},
		},
		dataChannelLabels: map[string]struct{}{
			"channel.telemetry": {},
		},
		robotStatuses: map[string]recorderRobotRuntime{
			"robot-001": {
				trackLabels: map[string]struct{}{
					"track.video_1": {},
					"video":         {},
				},
				dataChannelLabels: map[string]struct{}{
					"channel.telemetry": {},
				},
				lastTrackAt: time.Now().UTC(),
			},
			"robot-002": {
				trackLabels: map[string]struct{}{
					"track.video_1": {},
				},
				lastTrackAt: time.Now().UTC(),
			},
		},
		lastTrackLabel: "robot-001:video",
	}

	worker.resetRecorderTrackRuntime("mission-001", "robot-001")

	status := worker.subscriberStatuses["mission-001"]
	if _, ok := status.trackLabels["robot-001:track.video_1"]; ok {
		t.Fatalf("expected robot-001 canonical stale track to be cleared")
	}
	if _, ok := status.trackLabels["robot-001:video"]; ok {
		t.Fatalf("expected robot-001 fallback stale track to be cleared")
	}
	if _, ok := status.trackLabels["robot-002:track.video_1"]; !ok {
		t.Fatalf("expected other robot track to remain")
	}
	if len(status.robotStatuses["robot-001"].trackLabels) != 0 {
		t.Fatalf("expected robot-001 runtime tracks to be empty, got %#v", status.robotStatuses["robot-001"].trackLabels)
	}
	if len(status.robotStatuses["robot-001"].dataChannelLabels) != 1 {
		t.Fatalf("expected robot-001 data channel runtime to be preserved")
	}
	if status.lastTrackLabel != "" {
		t.Fatalf("expected stale lastTrackLabel to be cleared, got %q", status.lastTrackLabel)
	}
}
