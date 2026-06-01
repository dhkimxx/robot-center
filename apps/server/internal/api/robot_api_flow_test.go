package api

import (
	"net/http"
	"testing"
	"time"

	"robot-center/apps/server/internal/sfu"
)

func TestRobotAPIFlow(t *testing.T) {
	server := newAPIFlowTestServer(t)
	robot := server.createRobot(t, "Robot API Robot")
	supportRobot := server.createRobot(t, "Robot API Support")

	heartbeatPayload := requestJSON[map[string]any](t, server.baseURL, http.MethodPost, "/api/v1/robot/heartbeat", robot.token, map[string]any{
		"state":  "online",
		"sentAt": time.Now().UTC().Format(time.RFC3339Nano),
	})
	if heartbeatPayload["robotCode"] != robot.code {
		t.Fatalf("expected robot API heartbeat to expose token-authenticated robotCode, got %#v", heartbeatPayload)
	}
	if _, ok := heartbeatPayload["robotId"]; ok {
		t.Fatalf("robot API heartbeat should not expose internal robotId, got %#v", heartbeatPayload)
	}

	requestJSON[map[string]any](t, server.baseURL, http.MethodPost, "/api/v1/robot/heartbeat", supportRobot.token, map[string]any{
		"state":  "online",
		"sentAt": time.Now().UTC().Format(time.RFC3339Nano),
	})
	heartbeatWithRobotCodeStatus, _ := requestRawJSON(t, server.baseURL, http.MethodPost, "/api/v1/robot/heartbeat", robot.token, map[string]any{
		"robotCode": robot.code,
		"state":     "online",
	})
	if heartbeatWithRobotCodeStatus != http.StatusBadRequest {
		t.Fatalf("expected heartbeat robotCode body to be rejected, got %d", heartbeatWithRobotCodeStatus)
	}

	mission := server.createStartedMission(t, robot, supportRobot)
	missionPayload := requestJSON[map[string]any](t, server.baseURL, http.MethodGet, "/api/v1/robot/mission", robot.token, nil)
	if missionPayload["missionStatus"] != "active" {
		t.Fatalf("expected active robot mission, got %#v", missionPayload)
	}
	for _, internalField := range []string{"missionId", "robotCode", "roomId", "legacyRoomId", "videoPolicy"} {
		if _, ok := missionPayload[internalField]; ok {
			t.Fatalf("robot API mission should not expose %s, got %#v", internalField, missionPayload)
		}
	}
	sfuPayload := missionPayload["sfu"].(map[string]any)
	if _, ok := sfuPayload["publisherToken"]; ok {
		t.Fatalf("publisherToken should not be exposed in the P0 robot contract, got %#v", missionPayload)
	}
	if sfuPayload["signalingUrl"] != "ws://center.local/api/v1/robot/sfu/ws?room="+mission.code {
		t.Fatalf("expected robot API signaling URL, got %#v", sfuPayload)
	}
	assertStringListEqual(t, missionPayload["tracks"], []string{
		sfu.StreamRoleTrackVideo1,
		sfu.StreamRoleTrackVideo2,
		sfu.StreamRoleTrackAudio1,
		sfu.StreamRoleTrackAudio2,
	})
	assertStringListEqual(t, missionPayload["dataChannels"], []string{
		sfu.StreamRoleChannelTelemetry,
		sfu.StreamRoleChannelSpatial,
		sfu.StreamRoleChannelEvent,
		sfu.StreamRoleChannelControl,
	})

	missionRobotCodeQueryStatus, _ := requestRawJSON(t, server.baseURL, http.MethodGet, "/api/v1/robot/mission?robotCode="+supportRobot.code, robot.token, nil)
	if missionRobotCodeQueryStatus != http.StatusBadRequest {
		t.Fatalf("expected robotCode query to be rejected, got %d", missionRobotCodeQueryStatus)
	}
	supportMissionPayload := requestJSON[map[string]any](t, server.baseURL, http.MethodGet, "/api/v1/robot/mission", supportRobot.token, nil)
	if supportMissionPayload["missionStatus"] != "active" || supportMissionPayload["missionCode"] != mission.code {
		t.Fatalf("expected active support robot mission in shared room, got %#v", supportMissionPayload)
	}

	requestJSON[map[string]any](t, server.baseURL, http.MethodPost, "/api/v1/operator/missions/"+mission.code+"/end", "", nil)
	endedGatewayPayload := requestJSON[map[string]any](t, server.baseURL, http.MethodGet, "/api/v1/robot/mission", supportRobot.token, nil)
	if endedGatewayPayload["missionStatus"] != "none" {
		t.Fatalf("expected no active support robot mission after end, got %#v", endedGatewayPayload)
	}
}
