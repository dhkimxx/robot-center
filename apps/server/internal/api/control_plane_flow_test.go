package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"robot-center/apps/server/internal/config"
	"robot-center/apps/server/internal/sfu"
	"robot-center/apps/server/internal/testsupport/postgrestest"
	"strings"
	"testing"
	"time"
)

func TestControlPlaneFlow(t *testing.T) {
	postgresContainer := postgrestest.Start(t)

	recorderHealth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	}))
	defer recorderHealth.Close()

	appServer, err := NewServerFromConfig(context.Background(), config.AppServerConfig{
		PostgresDSN:               postgresContainer.DSN,
		PublicURL:                 "http://center.local",
		RecorderWorkerURL:         recorderHealth.URL,
		SFUWebSocketPublicBaseURL: "ws://center.local",
		TURNPublicURL:             "turn:127.0.0.1:3478?transport=udp",
		TURNInternalURL:           "turn:127.0.0.1:3478?transport=udp",
		TURNUsername:              "robot",
		TURNPassword:              "robot-pass",
		MinIOEndpoint:             "http://127.0.0.1:9000",
		MinIOBucket:               "robot-center-poc",
	})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(appServer.Handler())
	defer server.Close()

	systemStatus := requestJSON[map[string]any](t, server.URL, http.MethodGet, "/api/system/status", "", nil)
	components := systemStatus["components"].([]any)
	if !componentHasStatus(components, "recorder-worker", "ok") {
		t.Fatalf("expected recorder-worker component status ok, got %#v", components)
	}
	openAPI := requestJSON[map[string]any](t, server.URL, http.MethodGet, "/api/docs/openapi.json", "", nil)
	info := openAPI["info"].(map[string]any)
	if !strings.Contains(info["description"].(string), "관제 서버") {
		t.Fatalf("expected Korean OpenAPI description, got %#v", info)
	}
	forbiddenOpenAPIWords := []string{"개발 서버", "테스트 슬롯", "테스트용", "Mock Robot"}
	for _, forbiddenWord := range forbiddenOpenAPIWords {
		if strings.Contains(info["description"].(string), forbiddenWord) {
			t.Fatalf("OpenAPI description should be API-reference oriented, got %#v", info)
		}
	}
	paths := openAPI["paths"].(map[string]any)
	if _, ok := paths["/api/v1/robot/mission"]; !ok {
		t.Fatalf("expected robot mission API in OpenAPI paths, got %#v", paths)
	}
	for _, privatePath := range []string{"/api/system/status", "/api/recorder/tick"} {
		if _, ok := paths[privatePath]; ok {
			t.Fatalf("OpenAPI should not publish removed or private path %s", privatePath)
		}
	}
	swaggerResponse, err := http.Get(server.URL + "/api/docs")
	if err != nil {
		t.Fatalf("request Swagger UI: %v", err)
	}
	defer swaggerResponse.Body.Close()
	if swaggerResponse.StatusCode != http.StatusOK || !strings.Contains(swaggerResponse.Header.Get("Content-Type"), "text/html") {
		t.Fatalf("expected Swagger UI HTML response, got %s %s", swaggerResponse.Status, swaggerResponse.Header.Get("Content-Type"))
	}
	createRobotPayload := requestJSON[map[string]any](t, server.URL, http.MethodPost, "/api/robots", "", map[string]any{
		"displayName": "Test Robot",
		"modelName":   "Android Mock",
	})
	robot := createRobotPayload["robot"].(map[string]any)
	connectionInfo := createRobotPayload["connectionInfo"].(map[string]any)
	robotCode := robot["robotCode"].(string)
	robotToken := connectionInfo["robotToken"].(string)

	updateRobotPayload := requestJSON[map[string]any](t, server.URL, http.MethodPatch, "/api/robots/"+robotCode, "", map[string]any{
		"displayName": "Updated Test Robot",
		"modelName":   "Updated Android Mock",
	})
	updatedRobot := updateRobotPayload["robot"].(map[string]any)
	if updatedRobot["displayName"] != "Updated Test Robot" {
		t.Fatalf("expected updated robot name, got %#v", updatedRobot)
	}

	rotateTokenPayload := requestJSON[map[string]any](t, server.URL, http.MethodPost, "/api/robots/"+robotCode+"/connection-token", "", nil)
	rotatedConnectionInfo := rotateTokenPayload["connectionInfo"].(map[string]any)
	robotToken = rotatedConnectionInfo["robotToken"].(string)

	supportRobotPayload := requestJSON[map[string]any](t, server.URL, http.MethodPost, "/api/robots", "", map[string]any{
		"displayName": "Support Robot",
		"modelName":   "Android Mock",
	})
	supportRobot := supportRobotPayload["robot"].(map[string]any)
	supportConnectionInfo := supportRobotPayload["connectionInfo"].(map[string]any)
	supportRobotCode := supportRobot["robotCode"].(string)
	supportRobotToken := supportConnectionInfo["robotToken"].(string)

	idleRobotPayload := requestJSON[map[string]any](t, server.URL, http.MethodPost, "/api/robots", "", map[string]any{
		"displayName": "Idle Robot",
		"modelName":   "Retire Target",
	})
	idleRobot := idleRobotPayload["robot"].(map[string]any)
	idleRobotCode := idleRobot["robotCode"].(string)
	requestJSON[map[string]any](t, server.URL, http.MethodDelete, "/api/robots/"+idleRobotCode, "", nil)
	robotsPayload := requestJSON[map[string]any](t, server.URL, http.MethodGet, "/api/robots", "", nil)
	if robotListHasCode(robotsPayload["robots"].([]any), idleRobotCode) {
		t.Fatalf("expected archived robot to be hidden, got %#v", robotsPayload)
	}

	heartbeatPayload := requestJSON[map[string]any](t, server.URL, http.MethodPost, "/api/v1/robot/heartbeat", robotToken, map[string]any{
		"state":  "online",
		"sentAt": time.Now().UTC().Format(time.RFC3339Nano),
	})
	if heartbeatPayload["robotCode"] != robotCode {
		t.Fatalf("expected robot API heartbeat to expose token-authenticated robotCode, got %#v", heartbeatPayload)
	}
	if _, ok := heartbeatPayload["robotId"]; ok {
		t.Fatalf("robot API heartbeat should not expose internal robotId, got %#v", heartbeatPayload)
	}
	requestJSON[map[string]any](t, server.URL, http.MethodPost, "/api/v1/robot/heartbeat", supportRobotToken, map[string]any{
		"state":  "online",
		"sentAt": time.Now().UTC().Format(time.RFC3339Nano),
	})
	heartbeatWithRobotCodeStatus, _ := requestRawJSON(t, server.URL, http.MethodPost, "/api/v1/robot/heartbeat", robotToken, map[string]any{
		"robotCode": robotCode,
		"state":     "online",
	})
	if heartbeatWithRobotCodeStatus != http.StatusBadRequest {
		t.Fatalf("expected heartbeat robotCode body to be rejected, got %d", heartbeatWithRobotCodeStatus)
	}

	createMissionPayload := requestJSON[map[string]any](t, server.URL, http.MethodPost, "/api/missions", "", map[string]any{
		"name":        "Integration Mission",
		"missionType": "mountain_rescue",
		"siteNote":    "test",
		"robotCodes":  []string{robotCode, supportRobotCode},
	})
	mission := createMissionPayload["mission"].(map[string]any)
	missionCode := mission["missionCode"].(string)
	assertStringListEqual(t, mission["robotCodes"], []string{robotCode, supportRobotCode})
	if mission["robotCode"] != robotCode {
		t.Fatalf("expected legacy robotCode to use first robot, got %#v", mission)
	}

	missionsPayload := requestJSON[map[string]any](t, server.URL, http.MethodGet, "/api/missions", "", nil)
	missions := missionsPayload["missions"].([]any)
	if len(missions) != 1 {
		t.Fatalf("expected one mission row for multi-robot mission, got %#v", missionsPayload)
	}
	listedMission := missions[0].(map[string]any)
	assertStringListEqual(t, listedMission["robotCodes"], []string{robotCode, supportRobotCode})

	startMissionPayload := requestJSON[map[string]any](t, server.URL, http.MethodPost, "/api/missions/"+missionCode+"/start", "", nil)
	startedMission := startMissionPayload["mission"].(map[string]any)
	if startedMission["status"] != "active" {
		t.Fatalf("expected active mission, got %#v", startedMission)
	}
	assertStringListEqual(t, startedMission["robotCodes"], []string{robotCode, supportRobotCode})

	conflictStatus, conflictPayload := requestRawJSON(t, server.URL, http.MethodPost, "/api/missions", "", map[string]any{
		"name":        "Conflicting Mission",
		"missionType": "mountain_rescue",
		"robotCode":   robotCode,
	})
	if conflictStatus != http.StatusConflict {
		t.Fatalf("expected mission create conflict status, got %d payload %#v", conflictStatus, conflictPayload)
	}
	conflicts := conflictPayload["conflicts"].([]any)
	if len(conflicts) != 1 {
		t.Fatalf("expected one conflict, got %#v", conflictPayload)
	}
	conflict := conflicts[0].(map[string]any)
	if conflict["robotCode"] != robotCode || conflict["activeMissionCode"] != missionCode {
		t.Fatalf("expected conflict robot %s active in %s, got %#v", robotCode, missionCode, conflict)
	}

	missionPayload := requestJSON[map[string]any](t, server.URL, http.MethodGet, "/api/v1/robot/mission", robotToken, nil)
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
	if sfuPayload["signalingUrl"] != "ws://center.local/api/v1/robot/sfu/ws?room="+missionCode {
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
	missionRobotCodeQueryStatus, _ := requestRawJSON(t, server.URL, http.MethodGet, "/api/v1/robot/mission?robotCode="+supportRobotCode, robotToken, nil)
	if missionRobotCodeQueryStatus != http.StatusBadRequest {
		t.Fatalf("expected robotCode query to be rejected, got %d", missionRobotCodeQueryStatus)
	}
	supportMissionPayload := requestJSON[map[string]any](t, server.URL, http.MethodGet, "/api/v1/robot/mission", supportRobotToken, nil)
	if supportMissionPayload["missionStatus"] != "active" || supportMissionPayload["missionCode"] != missionCode {
		t.Fatalf("expected active support robot mission in shared room, got %#v", supportMissionPayload)
	}

	missionID := startedMission["id"].(string)
	requestJSON[map[string]any](t, server.URL, http.MethodPost, "/api/sensor-samples", "", map[string]any{
		"messageId":   "telemetry-canonical-1",
		"messageType": "telemetry",
		"channelRole": "channel.telemetry",
		"robotCode":   robotCode,
		"missionId":   missionID,
		"descriptors": []map[string]any{
			{
				"sensorId":   "telemetry.position_1",
				"sensorType": "position",
				"label":      "GPS",
				"enabled":    true,
			},
			{
				"sensorId":   "telemetry.gas.channel_1",
				"sensorType": "gas",
				"label":      "Gas",
				"enabled":    true,
			},
		},
		"samples": []map[string]any{
			{
				"sensorId":  "telemetry.position_1",
				"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
				"values": map[string]any{
					"latitude":  37.402181,
					"longitude": 127.106818,
				},
			},
			{
				"sensorId":  "telemetry.gas.channel_1",
				"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
				"values": map[string]any{
					"concentration": 12.3,
				},
			},
		},
	})

	missingSensorTypeStatus, _ := requestRawJSON(t, server.URL, http.MethodPost, "/api/sensor-samples", "", map[string]any{
		"messageId":   "telemetry-missing-sensor-type",
		"messageType": "telemetry",
		"channelRole": "channel.telemetry",
		"robotCode":   robotCode,
		"missionId":   missionID,
		"descriptors": []map[string]any{
			{
				"sensorId": "telemetry.gas.channel_5",
				"label":    "TEMP",
				"enabled":  true,
			},
		},
	})
	if missingSensorTypeStatus != http.StatusBadRequest {
		t.Fatalf("expected descriptor without sensorType to be rejected, got %d", missingSensorTypeStatus)
	}

	invalidSensorTypeStatus, _ := requestRawJSON(t, server.URL, http.MethodPost, "/api/sensor-samples", "", map[string]any{
		"messageId":   "telemetry-invalid-sensor-type",
		"messageType": "telemetry",
		"channelRole": "channel.telemetry",
		"robotCode":   robotCode,
		"missionId":   missionID,
		"descriptors": []map[string]any{
			{
				"sensorId":   "telemetry.custom_1",
				"sensorType": "custom",
				"label":      "Custom",
				"enabled":    true,
			},
		},
	})
	if invalidSensorTypeStatus != http.StatusBadRequest {
		t.Fatalf("expected descriptor with invalid sensorType to be rejected, got %d", invalidSensorTypeStatus)
	}

	sampleWithoutDescriptorStatus, _ := requestRawJSON(t, server.URL, http.MethodPost, "/api/sensor-samples", "", map[string]any{
		"messageId":   "telemetry-sample-without-descriptor",
		"messageType": "telemetry",
		"channelRole": "channel.telemetry",
		"robotCode":   robotCode,
		"missionId":   missionID,
		"samples": []map[string]any{
			{
				"sensorId":  "telemetry.unregistered_1",
				"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
				"values": map[string]any{
					"value": 1,
				},
			},
		},
	})
	if sampleWithoutDescriptorStatus != http.StatusBadRequest {
		t.Fatalf("expected sample without descriptor to be rejected, got %d", sampleWithoutDescriptorStatus)
	}

	payloadOnlyStatus, _ := requestRawJSON(t, server.URL, http.MethodPost, "/api/sensor-samples", "", map[string]any{
		"messageId":   "telemetry-payload-only",
		"messageType": "telemetry",
		"channelRole": "channel.telemetry",
		"robotCode":   robotCode,
		"missionId":   missionID,
		"payload": map[string]any{
			"batteryPercent": 82,
		},
	})
	if payloadOnlyStatus != http.StatusBadRequest {
		t.Fatalf("expected payload-only sensor envelope to be rejected, got %d", payloadOnlyStatus)
	}

	sensorLatestPayload := requestJSON[map[string]any](t, server.URL, http.MethodGet, "/api/sensor-latest?missionId="+missionID+"&robotCode="+robotCode, "", nil)
	latestSensors := sensorLatestPayload["sensors"].([]any)
	if len(latestSensors) != 2 {
		t.Fatalf("expected two latest sensor rows, got %#v", sensorLatestPayload)
	}
	if !sensorListHasID(latestSensors, "telemetry.position_1") {
		t.Fatalf("expected position sensor latest row, got %#v", sensorLatestPayload)
	}

	recordingTickPayload := requestJSON[map[string]any](t, server.URL, http.MethodPost, "/api/recorder/tick", "", map[string]any{
		"missionCode":          missionCode,
		"robotCode":            robotCode,
		"chunkDurationSeconds": 600,
		"tickAt":               time.Now().UTC().Format(time.RFC3339Nano),
	})
	chunk := recordingTickPayload["chunk"].(map[string]any)
	if chunk["status"] != "recording" {
		t.Fatalf("expected recording chunk, got %#v", chunk)
	}
	liveStatusPayload := requestJSON[map[string]any](t, server.URL, http.MethodGet, "/api/missions/"+missionCode+"/live-status", "", nil)
	liveStatusRobots := liveStatusPayload["robots"].([]any)
	liveStatusRobot := liveStatusRobots[0].(map[string]any)
	liveStreamStatus := liveStatusRobot["stream"].(map[string]any)
	liveRecordingStatus := liveStatusRobot["recording"].(map[string]any)
	if liveStreamStatus["state"] != "waiting" {
		t.Fatalf("expected live stream waiting without SFU publisher, got %#v", liveStatusPayload)
	}
	if liveRecordingStatus["state"] != "idle" || liveRecordingStatus["latestChunkStatus"] != "recording" {
		t.Fatalf("expected live recording idle with recording chunk metadata, got %#v", liveStatusPayload)
	}

	uploadedPayload := requestJSON[map[string]any](t, server.URL, http.MethodPost, "/api/recorder/chunks/"+chunk["id"].(string)+"/uploaded", "", nil)
	uploadedChunk := uploadedPayload["chunk"].(map[string]any)
	if uploadedChunk["status"] != "uploaded" {
		t.Fatalf("expected uploaded chunk, got %#v", uploadedChunk)
	}
	fileUploadedPayload := requestJSON[map[string]any](t, server.URL, http.MethodPost, "/api/recorder/chunks/"+chunk["id"].(string)+"/files/rgb_audio_mp4/uploaded", "", nil)
	fileUploadedChunk := fileUploadedPayload["chunk"].(map[string]any)
	fileTypes := fileUploadedChunk["availableFileTypes"].(map[string]any)
	if fileTypes["rgb_audio_mp4"] != true {
		t.Fatalf("expected rgb mp4 available flag, got %#v", fileUploadedChunk)
	}

	recordingsPayload := requestJSON[map[string]any](t, server.URL, http.MethodGet, "/api/recordings", "", nil)
	recordings := recordingsPayload["recordings"].([]any)
	if len(recordings) != 1 {
		t.Fatalf("expected one recording, got %#v", recordingsPayload)
	}
	recording := recordings[0].(map[string]any)
	if recording["recordingSessionId"] == "" {
		t.Fatalf("expected recordingSessionId in response, got %#v", recording)
	}
	files := recording["files"].([]any)
	if !fileHasAvailableURL(files, "manifest") {
		t.Fatalf("expected manifest file with available URL, got %#v", files)
	}
	if !fileHasAvailableURL(files, "rgb_audio_mp4") {
		t.Fatalf("expected rgb mp4 file with available URL, got %#v", files)
	}

	endMissionPayload := requestJSON[map[string]any](t, server.URL, http.MethodPost, "/api/missions/"+missionCode+"/end", "", nil)
	endedMission := endMissionPayload["mission"].(map[string]any)
	if endedMission["status"] != "ended" {
		t.Fatalf("expected ended mission, got %#v", endedMission)
	}
	assertStringListEqual(t, endedMission["robotCodes"], []string{robotCode, supportRobotCode})
	endedGatewayPayload := requestJSON[map[string]any](t, server.URL, http.MethodGet, "/api/v1/robot/mission", supportRobotToken, nil)
	if endedGatewayPayload["missionStatus"] != "none" {
		t.Fatalf("expected no active support robot mission after end, got %#v", endedGatewayPayload)
	}
}
