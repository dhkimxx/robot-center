package api

import (
	"net/http"
	"testing"
	"time"

	"robot-center/apps/server/internal/api/dto"
)

func TestControlPlaneFlowSmoke(t *testing.T) {
	server := newAPIFlowTestServer(t)
	robot := server.createRobot(t, "Smoke Robot")
	mission := server.createStartedMission(t, robot)

	requestJSON[dto.RobotHeartbeatResponse](t, server.baseURL, http.MethodPost, "/api/v1/robot/heartbeat", robot.token, dto.RobotHeartbeatRequest{
		State:  "online",
		SentAt: time.Now().UTC(),
	})

	robotMission := requestJSON[dto.RobotMissionResponse](t, server.baseURL, http.MethodGet, "/api/v1/robot/mission", robot.token, nil)
	if robotMission.MissionStatus != "active" || robotMission.MissionCode != mission.code {
		t.Fatalf("expected active robot mission, got %#v", robotMission)
	}

	recordingTick := requestJSON[dto.RecorderRecordingTickResponse](t, server.baseURL, http.MethodPost, "/api/v1/recorder/tick", "", dto.RecorderTickRequest{
		MissionCode:          mission.code,
		RobotCode:            robot.code,
		ChunkDurationSeconds: 600,
		TickAt:               time.Now().UTC(),
	})
	if recordingTick.Chunk.Status != "recording" {
		t.Fatalf("expected recording chunk, got %#v", recordingTick.Chunk)
	}

	liveStatus := requestJSON[dto.MissionLiveStatusResponse](t, server.baseURL, http.MethodGet, "/api/v1/operator/missions/"+mission.code+"/live-status", "", nil)
	if len(liveStatus.Robots) != 1 {
		t.Fatalf("expected one live status robot, got %#v", liveStatus)
	}

	endedPayload := requestJSON[dto.MissionEnvelopeResponse](t, server.baseURL, http.MethodPost, "/api/v1/operator/missions/"+mission.code+"/end", "", nil)
	if endedPayload.Mission.Status != "ended" {
		t.Fatalf("expected ended mission, got %#v", endedPayload)
	}
}
