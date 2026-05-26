package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"robot-center/apps/server/internal/config"
	"robot-center/apps/server/internal/sfu"
)

func TestControlPlaneFlow(t *testing.T) {
	recorderHealth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	}))
	defer recorderHealth.Close()

	server := httptest.NewServer(NewServer(config.AppServerConfig{
		PublicURL:         "http://center.local",
		RecorderWorkerURL: recorderHealth.URL,
		SFUWebSocketURL:   "ws://center.local/sfu/ws",
		TURNURL:           "turn:127.0.0.1:3478?transport=udp",
		TURNUsername:      "robot",
		TURNPassword:      "robot-pass",
		MinIOEndpoint:     "http://127.0.0.1:9000",
		MinIOBucket:       "robot-center-poc",
	}).Handler())
	defer server.Close()

	systemStatus := requestJSON[map[string]any](t, server.URL, http.MethodGet, "/api/system/status", "", nil)
	components := systemStatus["components"].([]any)
	if !componentHasStatus(components, "recorder-worker", "ok") {
		t.Fatalf("expected recorder-worker component status ok, got %#v", components)
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

	requestJSON[map[string]any](t, server.URL, http.MethodPost, "/api/robot-gateway/heartbeat", robotToken, map[string]any{
		"robotCode": robotCode,
		"state":     "online",
		"sentAt":    time.Now().UTC().Format(time.RFC3339Nano),
	})
	requestJSON[map[string]any](t, server.URL, http.MethodPost, "/api/robot-gateway/heartbeat", supportRobotToken, map[string]any{
		"robotCode": supportRobotCode,
		"state":     "online",
		"sentAt":    time.Now().UTC().Format(time.RFC3339Nano),
	})

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

	conflictMissionPayload := requestJSON[map[string]any](t, server.URL, http.MethodPost, "/api/missions", "", map[string]any{
		"name":        "Conflicting Mission",
		"missionType": "mountain_rescue",
		"robotCode":   robotCode,
	})
	conflictMission := conflictMissionPayload["mission"].(map[string]any)
	conflictStatus, conflictPayload := requestRawJSON(t, server.URL, http.MethodPost, "/api/missions/"+conflictMission["missionCode"].(string)+"/start", "", nil)
	if conflictStatus != http.StatusConflict {
		t.Fatalf("expected mission start conflict status, got %d payload %#v", conflictStatus, conflictPayload)
	}
	conflicts := conflictPayload["conflicts"].([]any)
	if len(conflicts) != 1 {
		t.Fatalf("expected one conflict, got %#v", conflictPayload)
	}
	conflict := conflicts[0].(map[string]any)
	if conflict["robotCode"] != robotCode || conflict["activeMissionCode"] != missionCode {
		t.Fatalf("expected conflict robot %s active in %s, got %#v", robotCode, missionCode, conflict)
	}

	missionPayload := requestJSON[map[string]any](t, server.URL, http.MethodGet, "/api/robot-gateway/mission?robotCode="+robotCode, robotToken, nil)
	if missionPayload["missionStatus"] != "active" {
		t.Fatalf("expected active robot mission, got %#v", missionPayload)
	}
	if missionPayload["roomId"] != missionCode {
		t.Fatalf("expected gateway roomId to be missionCode, got %#v", missionPayload)
	}
	if missionPayload["legacyRoomId"] != missionCode+"__"+robotCode {
		t.Fatalf("expected legacy room id for compatibility, got %#v", missionPayload)
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
	supportMissionPayload := requestJSON[map[string]any](t, server.URL, http.MethodGet, "/api/robot-gateway/mission?robotCode="+supportRobotCode, supportRobotToken, nil)
	if supportMissionPayload["missionStatus"] != "active" || supportMissionPayload["roomId"] != missionCode {
		t.Fatalf("expected active support robot mission in shared room, got %#v", supportMissionPayload)
	}

	requestJSON[map[string]any](t, server.URL, http.MethodPost, "/api/robot-gateway/streaming-status", robotToken, map[string]any{
		"robotCode": robotCode,
		"missionId": startedMission["id"],
		"roomId":    missionCode,
		"status":    "streaming",
		"publishedTracks": []map[string]any{
			{"name": sfu.StreamRoleTrackVideo1, "kind": "video", "codec": "h264"},
			{"name": sfu.StreamRoleTrackVideo2, "kind": "video", "codec": "h264"},
			{"name": sfu.StreamRoleTrackAudio1, "kind": "audio", "codec": "opus"},
		},
		"publishedDataChannels": []string{
			sfu.StreamRoleChannelTelemetry,
			sfu.StreamRoleChannelEvent,
			sfu.StreamRoleChannelSpatial,
			sfu.StreamRoleChannelControl,
		},
		"sentAt": time.Now().UTC().Format(time.RFC3339Nano),
	})

	streamingPayload := requestJSON[map[string]any](t, server.URL, http.MethodGet, "/api/streaming-statuses", "", nil)
	streamingStatuses := streamingPayload["streamingStatuses"].([]any)
	if len(streamingStatuses) != 1 {
		t.Fatalf("expected one streaming status, got %#v", streamingStatuses)
	}

	missionID := startedMission["id"].(string)
	requestJSON[map[string]any](t, server.URL, http.MethodPost, "/api/sensor-samples", "", map[string]any{
		"messageId":   "telemetry-canonical-1",
		"messageType": "telemetry",
		"channelRole": "channel.telemetry",
		"robotCode":   robotCode,
		"missionId":   missionID,
		"sequence":    2,
		"sentAt":      time.Now().UTC().Format(time.RFC3339Nano),
		"descriptors": []map[string]any{
			{
				"sensorId":    "telemetry.position_1",
				"kind":        "position",
				"displayName": "GPS",
				"valueType":   "object",
				"enabled":     true,
			},
			{
				"sensorId":    "telemetry.gas_1",
				"kind":        "gas",
				"displayName": "Gas",
				"valueType":   "object",
				"enabled":     true,
			},
		},
		"samples": []map[string]any{
			{
				"sensorId":  "telemetry.position_1",
				"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
				"sequence":  2,
				"values": map[string]any{
					"latitude":  37.402181,
					"longitude": 127.106818,
				},
			},
			{
				"sensorId":  "telemetry.gas_1",
				"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
				"sequence":  2,
				"values": map[string]any{
					"coPpm":         12.3,
					"oxygenPercent": 20.8,
				},
			},
		},
	})

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
	endedGatewayPayload := requestJSON[map[string]any](t, server.URL, http.MethodGet, "/api/robot-gateway/mission?robotCode="+supportRobotCode, supportRobotToken, nil)
	if endedGatewayPayload["missionStatus"] != "none" {
		t.Fatalf("expected no active support robot mission after end, got %#v", endedGatewayPayload)
	}
}

func requestJSON[T any](t *testing.T, baseURL string, method string, path string, bearerToken string, body any) T {
	t.Helper()

	var requestBody *bytes.Reader
	if body == nil {
		requestBody = bytes.NewReader(nil)
	} else {
		rawBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		requestBody = bytes.NewReader(rawBody)
	}

	request, err := http.NewRequest(method, baseURL+path, requestBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if strings.TrimSpace(bearerToken) != "" {
		request.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("send request %s %s: %v", method, path, err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		t.Fatalf("%s %s returned %s", method, path, response.Status)
	}

	var payload T
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return payload
}

func requestRawJSON(t *testing.T, baseURL string, method string, path string, bearerToken string, body any) (int, map[string]any) {
	t.Helper()

	var requestBody *bytes.Reader
	if body == nil {
		requestBody = bytes.NewReader(nil)
	} else {
		rawBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		requestBody = bytes.NewReader(rawBody)
	}

	request, err := http.NewRequest(method, baseURL+path, requestBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if strings.TrimSpace(bearerToken) != "" {
		request.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("send request %s %s: %v", method, path, err)
	}
	defer response.Body.Close()

	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return response.StatusCode, payload
}

func componentHasStatus(components []any, name string, status string) bool {
	for _, item := range components {
		component, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if component["name"] == name && component["status"] == status {
			return true
		}
	}
	return false
}

func robotListHasCode(robots []any, robotCode string) bool {
	for _, item := range robots {
		robot, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if robot["robotCode"] == robotCode {
			return true
		}
	}
	return false
}

func sensorListHasID(sensors []any, sensorID string) bool {
	for _, item := range sensors {
		sensor, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if sensor["sensorId"] == sensorID {
			return true
		}
	}
	return false
}

func assertStringListEqual(t *testing.T, value any, expected []string) {
	t.Helper()

	items, ok := value.([]any)
	if !ok {
		t.Fatalf("expected string list, got %#v", value)
	}
	if len(items) != len(expected) {
		t.Fatalf("expected %d strings, got %#v", len(expected), value)
	}
	for index, expectedValue := range expected {
		actualValue, ok := items[index].(string)
		if !ok || actualValue != expectedValue {
			t.Fatalf("expected strings %#v, got %#v", expected, value)
		}
	}
}

func fileHasAvailableURL(files []any, fileType string) bool {
	for _, item := range files {
		file, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if file["type"] == fileType && file["status"] == "available" {
			urlValue, _ := file["url"].(string)
			return strings.Contains(urlValue, "http://center.local:9000/robot-center-poc/")
		}
	}
	return false
}
