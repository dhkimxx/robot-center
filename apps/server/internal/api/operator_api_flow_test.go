package api

import (
	"net/http"
	"testing"
)

func TestOperatorAPIFlow(t *testing.T) {
	server := newAPIFlowTestServer(t)
	robot := server.createRobot(t, "Test Robot")

	connectionInfoPayload := requestJSON[map[string]any](t, server.baseURL, http.MethodGet, "/api/v1/operator/robots/"+robot.code+"/connection-info", "", nil)
	connectionInfo := connectionInfoPayload["connectionInfo"].(map[string]any)
	if connectionInfo["robotCode"] != robot.code || connectionInfo["robotToken"] == "" {
		t.Fatalf("expected operator connection info for robot %s, got %#v", robot.code, connectionInfo)
	}

	updateRobotPayload := requestJSON[map[string]any](t, server.baseURL, http.MethodPatch, "/api/v1/operator/robots/"+robot.code, "", map[string]any{
		"displayName": "Updated Test Robot",
		"modelName":   "Updated Android Mock",
	})
	updatedRobot := updateRobotPayload["robot"].(map[string]any)
	if updatedRobot["displayName"] != "Updated Test Robot" {
		t.Fatalf("expected updated robot name, got %#v", updatedRobot)
	}

	rotateTokenPayload := requestJSON[map[string]any](t, server.baseURL, http.MethodPost, "/api/v1/operator/robots/"+robot.code+"/connection-token", "", nil)
	rotatedConnectionInfo := rotateTokenPayload["connectionInfo"].(map[string]any)
	if rotatedConnectionInfo["robotToken"] == robot.token {
		t.Fatalf("expected rotated robot token, got %#v", rotatedConnectionInfo)
	}

	supportRobot := server.createRobot(t, "Support Robot")
	idleRobot := server.createRobot(t, "Idle Robot")
	requestJSON[map[string]any](t, server.baseURL, http.MethodDelete, "/api/v1/operator/robots/"+idleRobot.code, "", nil)
	robotsPayload := requestJSON[map[string]any](t, server.baseURL, http.MethodGet, "/api/v1/operator/robots", "", nil)
	if robotListHasCode(robotsPayload["robots"].([]any), idleRobot.code) {
		t.Fatalf("expected archived robot to be hidden, got %#v", robotsPayload)
	}

	mission := server.createMission(t, []string{robot.code, supportRobot.code})
	missionsPayload := requestJSON[map[string]any](t, server.baseURL, http.MethodGet, "/api/v1/operator/missions", "", nil)
	missions := missionsPayload["missions"].([]any)
	if len(missions) != 1 {
		t.Fatalf("expected one mission row for multi-robot mission, got %#v", missionsPayload)
	}
	listedMission := missions[0].(map[string]any)
	assertStringListEqual(t, listedMission["robotCodes"], []string{robot.code, supportRobot.code})

	startedMission := server.startMission(t, mission.code)
	if startedMission["status"] != "active" {
		t.Fatalf("expected active mission, got %#v", startedMission)
	}
	assertStringListEqual(t, startedMission["robotCodes"], []string{robot.code, supportRobot.code})

	conflictStatus, conflictPayload := requestRawJSON(t, server.baseURL, http.MethodPost, "/api/v1/operator/missions", "", map[string]any{
		"name":        "Conflicting Mission",
		"missionType": "mountain_rescue",
		"robotCode":   robot.code,
	})
	if conflictStatus != http.StatusConflict {
		t.Fatalf("expected mission create conflict status, got %d payload %#v", conflictStatus, conflictPayload)
	}
	conflicts := conflictPayload["conflicts"].([]any)
	if len(conflicts) != 1 {
		t.Fatalf("expected one conflict, got %#v", conflictPayload)
	}
	conflict := conflicts[0].(map[string]any)
	if conflict["robotCode"] != robot.code || conflict["activeMissionCode"] != mission.code {
		t.Fatalf("expected conflict robot %s active in %s, got %#v", robot.code, mission.code, conflict)
	}

	liveStatus := requestJSON[map[string]any](t, server.baseURL, http.MethodGet, "/api/v1/operator/missions/"+mission.code+"/live-status", "", nil)
	if len(liveStatus["robots"].([]any)) != 2 {
		t.Fatalf("expected two live status robots, got %#v", liveStatus)
	}

	endMissionPayload := requestJSON[map[string]any](t, server.baseURL, http.MethodPost, "/api/v1/operator/missions/"+mission.code+"/end", "", nil)
	endedMission := endMissionPayload["mission"].(map[string]any)
	if endedMission["status"] != "ended" {
		t.Fatalf("expected ended mission, got %#v", endedMission)
	}
	assertStringListEqual(t, endedMission["robotCodes"], []string{robot.code, supportRobot.code})
}
