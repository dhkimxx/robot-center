package dto

import (
	"encoding/json"
	"testing"
	"time"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/sfu"
)

func TestRobotGatewayRequestShape(t *testing.T) {
	request := RobotHeartbeatRequest{
		State:          "online",
		BatteryPercent: 90,
		NetworkQuality: "good",
		SentAt:         time.Date(2026, 6, 2, 8, 0, 0, 0, time.UTC),
	}

	for _, field := range []string{"state", "batteryPercent", "networkQuality", "sentAt"} {
		assertRobotGatewayJSONHasField(t, request, field)
	}
}

func TestRobotGatewayHeartbeatPayloadShape(t *testing.T) {
	now := time.Date(2026, 6, 2, 8, 0, 0, 0, time.UTC)
	payload := RobotHeartbeatPayload(domain.Robot{
		RobotCode:   "robot-001",
		DeviceState: domain.RobotDeviceStateOnline,
	}, now)

	fields := robotGatewayJSONFields(t, payload)
	assertRobotGatewayJSONKeys(t, fields, []string{"robotCode", "serverTime", "status"})
	if fields["robotCode"] != "robot-001" || fields["status"] != "online" {
		t.Fatalf("unexpected heartbeat payload: %#v", fields)
	}
}

func TestRobotGatewayMissionPayloadShape(t *testing.T) {
	now := time.Date(2026, 6, 2, 8, 0, 0, 0, time.UTC)
	payload := RobotMissionPayload(RobotMissionInput{
		Mission: domain.Mission{
			MissionCode: "mission-001",
			Status:      "active",
		},
		SignalingURL: "ws://center.local/api/v1/robot/sfu/ws?room=mission-001",
		TURNURL:      "turn:center.local:3478?transport=udp",
		TURNUsername: "robot",
		TURNPassword: "robot-pass",
		Now:          now,
	})

	fields := robotGatewayJSONFields(t, payload)
	assertRobotGatewayJSONKeys(t, fields, []string{"dataChannels", "missionCode", "missionStatus", "serverTime", "sfu", "tracks", "turnServers"})
	if fields["missionCode"] != "mission-001" || fields["missionStatus"] != "active" {
		t.Fatalf("unexpected mission payload: %#v", fields)
	}
	assertRobotGatewayStringListEqual(t, fields["tracks"], []string{
		sfu.StreamRoleTrackVideo1,
		sfu.StreamRoleTrackVideo2,
		sfu.StreamRoleTrackAudio1,
		sfu.StreamRoleTrackAudio2,
	})
	assertRobotGatewayStringListEqual(t, fields["dataChannels"], []string{
		sfu.StreamRoleChannelTelemetry,
		sfu.StreamRoleChannelSpatial,
		sfu.StreamRoleChannelEvent,
		sfu.StreamRoleChannelControl,
	})
}

func TestRobotGatewayMissionNonePayloadShape(t *testing.T) {
	now := time.Date(2026, 6, 2, 8, 0, 0, 0, time.UTC)
	fields := robotGatewayJSONFields(t, RobotMissionNonePayload(now))

	assertRobotGatewayJSONKeys(t, fields, []string{"missionStatus", "serverTime"})
	if fields["missionStatus"] != "none" {
		t.Fatalf("unexpected none payload: %#v", fields)
	}
}

func assertRobotGatewayJSONHasField(t *testing.T, value any, field string) {
	t.Helper()
	fields := robotGatewayJSONFields(t, value)
	if _, ok := fields[field]; !ok {
		t.Fatalf("expected field %q in JSON, got %#v", field, fields)
	}
}

func robotGatewayJSONFields(t *testing.T, value any) map[string]any {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal value: %v", err)
	}
	var fields map[string]any
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatalf("unmarshal value: %v", err)
	}
	return fields
}

func assertRobotGatewayJSONKeys(t *testing.T, fields map[string]any, expected []string) {
	t.Helper()
	if len(fields) != len(expected) {
		t.Fatalf("expected keys %#v, got %#v", expected, fields)
	}
	for _, field := range expected {
		if _, ok := fields[field]; !ok {
			t.Fatalf("expected key %q in %#v", field, fields)
		}
	}
}

func assertRobotGatewayStringListEqual(t *testing.T, value any, expected []string) {
	t.Helper()
	items, ok := value.([]any)
	if !ok || len(items) != len(expected) {
		t.Fatalf("expected string list %#v, got %#v", expected, value)
	}
	for index, expectedValue := range expected {
		actualValue, ok := items[index].(string)
		if !ok || actualValue != expectedValue {
			t.Fatalf("expected string list %#v, got %#v", expected, value)
		}
	}
}
