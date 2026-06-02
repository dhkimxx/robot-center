package dto

import (
	"encoding/json"
	"testing"
	"time"

	"robot-center/apps/server/internal/domain"
)

func TestRobotResponseUsesDomainConnectionState(t *testing.T) {
	now := time.Date(2026, 5, 27, 11, 0, 0, 0, time.UTC)
	staleSeenAt := now.Add(-2 * time.Minute)

	response := Robot(domain.Robot{
		RobotCode:   "robot-003",
		DeviceState: domain.RobotDeviceStateOnline,
		LastSeenAt:  &staleSeenAt,
	}, now, 30*time.Second)

	if response.Status != domain.RobotConnectionStateOffline {
		t.Fatalf("status = %q, want %q", response.Status, domain.RobotConnectionStateOffline)
	}
}

func TestRobotResponseWrapperShape(t *testing.T) {
	now := time.Date(2026, 6, 2, 8, 0, 0, 0, time.UTC)
	robot := domain.Robot{
		ID:          "robot-id",
		RobotCode:   "robot-001",
		DisplayName: "Robot 1",
		DeviceState: domain.RobotDeviceStateOffline,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	connectionInfo := domain.RobotConnectionInfo{
		ServerURL:  "http://center.local",
		RobotCode:  "robot-001",
		RobotToken: "robot-token",
	}

	assertJSONHasRobotField(t, RobotEnvelope(robot, now, domain.DefaultRobotHeartbeatTTL), "robot")
	assertJSONHasRobotField(t, RobotsPayload([]domain.Robot{robot}, now, domain.DefaultRobotHeartbeatTTL), "robots")
	assertJSONHasRobotField(t, CreateRobotPayload(robot, connectionInfo, now, domain.DefaultRobotHeartbeatTTL), "robot")
	assertJSONHasRobotField(t, CreateRobotPayload(robot, connectionInfo, now, domain.DefaultRobotHeartbeatTTL), "connectionInfo")
	assertJSONHasRobotField(t, RobotConnectionInfoPayload(connectionInfo), "connectionInfo")
}

func TestRobotRequestShape(t *testing.T) {
	assertJSONHasRobotField(t, CreateRobotRequest{DisplayName: "Robot 1", ModelName: "Mock"}, "displayName")
	assertJSONHasRobotField(t, CreateRobotRequest{DisplayName: "Robot 1", ModelName: "Mock"}, "modelName")
	assertJSONHasRobotField(t, UpdateRobotRequest{DisplayName: "Robot 1", ModelName: "Mock"}, "displayName")
	assertJSONHasRobotField(t, UpdateRobotRequest{DisplayName: "Robot 1", ModelName: "Mock"}, "modelName")
}

func assertJSONHasRobotField(t *testing.T, value any, field string) {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal value: %v", err)
	}
	var fields map[string]any
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatalf("unmarshal value: %v", err)
	}
	if _, ok := fields[field]; !ok {
		t.Fatalf("expected field %q in JSON, got %s", field, string(payload))
	}
}
