package dto

import (
	"encoding/json"
	"testing"
	"time"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
)

func TestMissionResponseWrapperShape(t *testing.T) {
	now := time.Date(2026, 6, 2, 8, 0, 0, 0, time.UTC)
	mission := domain.Mission{
		ID:          "mission-id",
		MissionCode: "mission-001",
		Name:        "Mission 1",
		MissionType: "mountain_rescue",
		Status:      "ready",
		RobotCodes:  []string{"robot-001"},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	assertMissionJSONHasField(t, MissionPayload(mission), "mission")
	assertMissionJSONHasField(t, MissionsPayload([]domain.Mission{mission}), "missions")
}

func TestCreateMissionRequestShape(t *testing.T) {
	request := CreateMissionRequest{
		Name:        "Mission 1",
		MissionType: "mountain_rescue",
		SiteNote:    "site",
		RobotCode:   "robot-001",
		RobotCodes:  []string{"robot-001", "robot-002"},
	}

	for _, field := range []string{"name", "missionType", "siteNote", "robotCode", "robotCodes"} {
		assertMissionJSONHasField(t, request, field)
	}
}

func TestMissionConflictPayloadShape(t *testing.T) {
	payload := MissionConflictPayload("mission start conflict", []store.MissionStartConflict{
		{
			RobotCode:         "robot-001",
			ActiveMissionCode: "mission-active",
		},
	})

	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if fields["error"] != "mission start conflict" {
		t.Fatalf("error = %q, want mission start conflict", fields["error"])
	}
	conflicts, ok := fields["conflicts"].([]any)
	if !ok || len(conflicts) != 1 {
		t.Fatalf("expected one conflict, got %s", string(encoded))
	}
	conflict := conflicts[0].(map[string]any)
	if conflict["robotCode"] != "robot-001" || conflict["activeMissionCode"] != "mission-active" {
		t.Fatalf("unexpected conflict shape: %s", string(encoded))
	}
}

func assertMissionJSONHasField(t *testing.T, value any, field string) {
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
