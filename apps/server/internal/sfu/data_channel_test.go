package sfu

import (
	"encoding/json"
	"testing"
)

func TestDataChannelPayloadWithContextPreservesEventEnvelope(t *testing.T) {
	raw := []byte(`{"messageType":"event","events":[{"eventType":"detection.object","values":{"trackId":"track.video_1","detections":[]}}]}`)

	got := dataChannelPayloadWithContext("mission-001", "robot-001", StreamRoleChannelEvent, raw)

	var payload map[string]any
	if err := json.Unmarshal(got, &payload); err != nil {
		t.Fatalf("expected contextual event payload to be valid JSON: %v", err)
	}
	if payload["robotCode"] != "robot-001" {
		t.Fatalf("robotCode = %v, want robot-001", payload["robotCode"])
	}
	if payload["missionCode"] != "mission-001" {
		t.Fatalf("missionCode = %v, want mission-001", payload["missionCode"])
	}
	if payload["channelRole"] != StreamRoleChannelEvent {
		t.Fatalf("channelRole = %v, want %s", payload["channelRole"], StreamRoleChannelEvent)
	}
	events, ok := payload["events"].([]any)
	if !ok || len(events) != 1 {
		t.Fatalf("events = %#v, want one event item", payload["events"])
	}
}
