package api

import (
	"net/http"
	"testing"
	"time"
)

func TestControlPlaneFlowSmoke(t *testing.T) {
	server := newAPIFlowTestServer(t)
	robot := server.createRobot(t, "Smoke Robot")
	mission := server.createStartedMission(t, robot)

	requestJSON[map[string]any](t, server.baseURL, http.MethodPost, "/api/v1/robot/heartbeat", robot.token, map[string]any{
		"state":  "online",
		"sentAt": time.Now().UTC().Format(time.RFC3339Nano),
	})

	robotMission := requestJSON[map[string]any](t, server.baseURL, http.MethodGet, "/api/v1/robot/mission", robot.token, nil)
	if robotMission["missionStatus"] != "active" || robotMission["missionCode"] != mission.code {
		t.Fatalf("expected active robot mission, got %#v", robotMission)
	}

	recordingTick := requestJSON[map[string]any](t, server.baseURL, http.MethodPost, "/api/v1/recorder/tick", "", map[string]any{
		"missionCode":          mission.code,
		"robotCode":            robot.code,
		"chunkDurationSeconds": 600,
		"tickAt":               time.Now().UTC().Format(time.RFC3339Nano),
	})
	chunk := recordingTick["chunk"].(map[string]any)
	if chunk["status"] != "recording" {
		t.Fatalf("expected recording chunk, got %#v", chunk)
	}

	liveStatus := requestJSON[map[string]any](t, server.baseURL, http.MethodGet, "/api/v1/operator/missions/"+mission.code+"/live-status", "", nil)
	if len(liveStatus["robots"].([]any)) != 1 {
		t.Fatalf("expected one live status robot, got %#v", liveStatus)
	}

	endedPayload := requestJSON[map[string]any](t, server.baseURL, http.MethodPost, "/api/v1/operator/missions/"+mission.code+"/end", "", nil)
	endedMission := endedPayload["mission"].(map[string]any)
	if endedMission["status"] != "ended" {
		t.Fatalf("expected ended mission, got %#v", endedMission)
	}
}
