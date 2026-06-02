package api

import (
	"net/http"
	"testing"

	"robot-center/apps/server/internal/api/dto"
)

func TestOperatorAPIFlow(t *testing.T) {
	server := newAPIFlowTestServer(t)
	robot := server.createRobot(t, "Test Robot")

	connectionInfoPayload := requestJSON[dto.RobotConnectionInfoEnvelopeResponse](t, server.baseURL, http.MethodGet, "/api/v1/operator/robots/"+robot.code+"/connection-info", "", nil)
	if connectionInfoPayload.ConnectionInfo.RobotCode != robot.code || connectionInfoPayload.ConnectionInfo.RobotToken == "" {
		t.Fatalf("expected operator connection info for robot %s, got %#v", robot.code, connectionInfoPayload)
	}

	updateRobotPayload := requestJSON[dto.RobotEnvelopeResponse](t, server.baseURL, http.MethodPatch, "/api/v1/operator/robots/"+robot.code, "", dto.UpdateRobotRequest{
		DisplayName: "Updated Test Robot",
		ModelName:   "Updated Android Mock",
	})
	if updateRobotPayload.Robot.DisplayName != "Updated Test Robot" {
		t.Fatalf("expected updated robot name, got %#v", updateRobotPayload)
	}

	rotateTokenPayload := requestJSON[dto.RobotConnectionInfoEnvelopeResponse](t, server.baseURL, http.MethodPost, "/api/v1/operator/robots/"+robot.code+"/connection-token", "", nil)
	if rotateTokenPayload.ConnectionInfo.RobotToken == robot.token {
		t.Fatalf("expected rotated robot token, got %#v", rotateTokenPayload)
	}

	supportRobot := server.createRobot(t, "Support Robot")
	idleRobot := server.createRobot(t, "Idle Robot")
	requestJSON[dto.RobotEnvelopeResponse](t, server.baseURL, http.MethodDelete, "/api/v1/operator/robots/"+idleRobot.code, "", nil)
	robotsPayload := requestJSON[dto.RobotsResponse](t, server.baseURL, http.MethodGet, "/api/v1/operator/robots", "", nil)
	if robotListHasCode(robotsPayload.Robots, idleRobot.code) {
		t.Fatalf("expected archived robot to be hidden, got %#v", robotsPayload)
	}

	mission := server.createMission(t, []string{robot.code, supportRobot.code})
	missionsPayload := requestJSON[dto.MissionsResponse](t, server.baseURL, http.MethodGet, "/api/v1/operator/missions", "", nil)
	if len(missionsPayload.Missions) != 1 {
		t.Fatalf("expected one mission row for multi-robot mission, got %#v", missionsPayload)
	}
	listedMission := missionsPayload.Missions[0]
	assertStringListEqual(t, listedMission.RobotCodes, []string{robot.code, supportRobot.code})

	startedMission := server.startMission(t, mission.code)
	if startedMission.Status != "active" {
		t.Fatalf("expected active mission, got %#v", startedMission)
	}
	assertStringListEqual(t, startedMission.RobotCodes, []string{robot.code, supportRobot.code})

	conflictStatus, conflictPayload := requestRawJSONAs[dto.MissionConflictEnvelopeResponse](t, server.baseURL, http.MethodPost, "/api/v1/operator/missions", "", dto.CreateMissionRequest{
		Name:        "Conflicting Mission",
		MissionType: "mountain_rescue",
		RobotCode:   robot.code,
	})
	if conflictStatus != http.StatusConflict {
		t.Fatalf("expected mission create conflict status, got %d payload %#v", conflictStatus, conflictPayload)
	}
	if len(conflictPayload.Conflicts) != 1 {
		t.Fatalf("expected one conflict, got %#v", conflictPayload)
	}
	conflict := conflictPayload.Conflicts[0]
	if conflict.RobotCode != robot.code || conflict.ActiveMissionCode != mission.code {
		t.Fatalf("expected conflict robot %s active in %s, got %#v", robot.code, mission.code, conflict)
	}

	liveStatus := requestJSON[dto.MissionLiveStatusResponse](t, server.baseURL, http.MethodGet, "/api/v1/operator/missions/"+mission.code+"/live-status", "", nil)
	if len(liveStatus.Robots) != 2 {
		t.Fatalf("expected two live status robots, got %#v", liveStatus)
	}

	endMissionPayload := requestJSON[dto.MissionEnvelopeResponse](t, server.baseURL, http.MethodPost, "/api/v1/operator/missions/"+mission.code+"/end", "", nil)
	endedMission := endMissionPayload.Mission
	if endedMission.Status != "ended" {
		t.Fatalf("expected ended mission, got %#v", endedMission)
	}
	assertStringListEqual(t, endedMission.RobotCodes, []string{robot.code, supportRobot.code})
}
