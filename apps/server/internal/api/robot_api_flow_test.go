package api

import (
	"net/http"
	"testing"
	"time"

	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/sfu"
)

func TestRobotAPIFlow(t *testing.T) {
	server := newAPIFlowTestServer(t)
	robot := server.createRobot(t, "Robot API Robot")
	supportRobot := server.createRobot(t, "Robot API Support")

	heartbeatPayload := requestJSON[dto.RobotHeartbeatResponse](t, server.baseURL, http.MethodPost, "/api/v1/robot/heartbeat", robot.token, dto.RobotHeartbeatRequest{
		State:  "online",
		SentAt: time.Now().UTC(),
	})
	if heartbeatPayload.RobotCode != robot.code {
		t.Fatalf("expected robot API heartbeat to expose token-authenticated robotCode, got %#v", heartbeatPayload)
	}

	requestJSON[dto.RobotHeartbeatResponse](t, server.baseURL, http.MethodPost, "/api/v1/robot/heartbeat", supportRobot.token, dto.RobotHeartbeatRequest{
		State:  "online",
		SentAt: time.Now().UTC(),
	})
	heartbeatWithRobotCodeStatus := requestStatus(t, server.baseURL, http.MethodPost, "/api/v1/robot/heartbeat", robot.token, map[string]any{
		"robotCode": robot.code,
		"state":     "online",
	})
	if heartbeatWithRobotCodeStatus != http.StatusBadRequest {
		t.Fatalf("expected heartbeat robotCode body to be rejected, got %d", heartbeatWithRobotCodeStatus)
	}

	mission := server.createStartedMission(t, robot, supportRobot)
	missionPayload := requestJSON[dto.RobotMissionResponse](t, server.baseURL, http.MethodGet, "/api/v1/robot/mission", robot.token, nil)
	if missionPayload.MissionStatus != "active" {
		t.Fatalf("expected active robot mission, got %#v", missionPayload)
	}
	if missionPayload.SFU == nil {
		t.Fatalf("expected active robot mission SFU payload, got %#v", missionPayload)
	}
	if missionPayload.SFU.SignalingURL != "ws://center.local/api/v1/robot/sfu/ws?room="+mission.code {
		t.Fatalf("expected robot API signaling URL, got %#v", missionPayload.SFU)
	}
	if missionPayload.SFU.ICETransportPolicy != "relay" {
		t.Fatalf("expected relay ICE policy, got %#v", missionPayload.SFU)
	}
	if len(missionPayload.TurnServers) != 1 {
		t.Fatalf("expected one TURN server, got %#v", missionPayload.TurnServers)
	}
	assertStringListEqual(t, missionPayload.Tracks, []string{
		sfu.StreamRoleTrackVideo1,
		sfu.StreamRoleTrackVideo2,
		sfu.StreamRoleTrackAudio1,
		sfu.StreamRoleTrackAudio2,
	})
	assertStringListEqual(t, missionPayload.DataChannels, []string{
		sfu.StreamRoleChannelTelemetry,
		sfu.StreamRoleChannelSpatial,
		sfu.StreamRoleChannelEvent,
		sfu.StreamRoleChannelControl,
	})

	missionRobotCodeQueryStatus := requestStatus(t, server.baseURL, http.MethodGet, "/api/v1/robot/mission?robotCode="+supportRobot.code, robot.token, nil)
	if missionRobotCodeQueryStatus != http.StatusBadRequest {
		t.Fatalf("expected robotCode query to be rejected, got %d", missionRobotCodeQueryStatus)
	}
	supportMissionPayload := requestJSON[dto.RobotMissionResponse](t, server.baseURL, http.MethodGet, "/api/v1/robot/mission", supportRobot.token, nil)
	if supportMissionPayload.MissionStatus != "active" || supportMissionPayload.MissionCode != mission.code {
		t.Fatalf("expected active support robot mission in shared room, got %#v", supportMissionPayload)
	}

	requestJSON[dto.MissionEnvelopeResponse](t, server.baseURL, http.MethodPost, "/api/v1/operator/missions/"+mission.code+"/end", "", nil)
	endedGatewayPayload := requestJSON[dto.RobotMissionResponse](t, server.baseURL, http.MethodGet, "/api/v1/robot/mission", supportRobot.token, nil)
	if endedGatewayPayload.MissionStatus != "none" {
		t.Fatalf("expected no active support robot mission after end, got %#v", endedGatewayPayload)
	}
	if endedGatewayPayload.MissionCode != "" || endedGatewayPayload.SFU != nil || len(endedGatewayPayload.TurnServers) != 0 {
		t.Fatalf("expected inactive robot mission payload to omit active connection fields, got %#v", endedGatewayPayload)
	}
}
