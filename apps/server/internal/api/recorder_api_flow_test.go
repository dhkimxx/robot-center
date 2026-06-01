package api

import (
	"net/http"
	"testing"
	"time"
)

func TestRecorderAPIFlow(t *testing.T) {
	server := newAPIFlowTestServer(t)
	robot := server.createRobot(t, "Recorder Robot")
	mission := server.createStartedMission(t, robot)

	targetsPayload := requestJSON[map[string]any](t, server.baseURL, http.MethodGet, "/api/v1/recorder/recording-targets", "", nil)
	targets := targetsPayload["targets"].([]any)
	if len(targets) != 1 {
		t.Fatalf("expected one recording target, got %#v", targetsPayload)
	}

	requestJSON[map[string]any](t, server.baseURL, http.MethodPost, "/api/v1/recorder/sensor-samples", "", map[string]any{
		"messageId":   "telemetry-canonical-1",
		"messageType": "telemetry",
		"channelRole": "channel.telemetry",
		"robotCode":   robot.code,
		"missionId":   mission.id,
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

	missingSensorTypeStatus, _ := requestRawJSON(t, server.baseURL, http.MethodPost, "/api/v1/recorder/sensor-samples", "", map[string]any{
		"messageId":   "telemetry-missing-sensor-type",
		"messageType": "telemetry",
		"channelRole": "channel.telemetry",
		"robotCode":   robot.code,
		"missionId":   mission.id,
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

	invalidSensorTypeStatus, _ := requestRawJSON(t, server.baseURL, http.MethodPost, "/api/v1/recorder/sensor-samples", "", map[string]any{
		"messageId":   "telemetry-invalid-sensor-type",
		"messageType": "telemetry",
		"channelRole": "channel.telemetry",
		"robotCode":   robot.code,
		"missionId":   mission.id,
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

	sampleWithoutDescriptorStatus, _ := requestRawJSON(t, server.baseURL, http.MethodPost, "/api/v1/recorder/sensor-samples", "", map[string]any{
		"messageId":   "telemetry-sample-without-descriptor",
		"messageType": "telemetry",
		"channelRole": "channel.telemetry",
		"robotCode":   robot.code,
		"missionId":   mission.id,
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

	payloadOnlyStatus, _ := requestRawJSON(t, server.baseURL, http.MethodPost, "/api/v1/recorder/sensor-samples", "", map[string]any{
		"messageId":   "telemetry-payload-only",
		"messageType": "telemetry",
		"channelRole": "channel.telemetry",
		"robotCode":   robot.code,
		"missionId":   mission.id,
		"payload": map[string]any{
			"batteryPercent": 82,
		},
	})
	if payloadOnlyStatus != http.StatusBadRequest {
		t.Fatalf("expected payload-only sensor envelope to be rejected, got %d", payloadOnlyStatus)
	}

	sensorLatestPayload := requestJSON[map[string]any](t, server.baseURL, http.MethodGet, "/api/v1/operator/sensor-latest?missionId="+mission.id+"&robotCode="+robot.code, "", nil)
	latestSensors := sensorLatestPayload["sensors"].([]any)
	if len(latestSensors) != 2 {
		t.Fatalf("expected two latest sensor rows, got %#v", sensorLatestPayload)
	}
	if !sensorListHasID(latestSensors, "telemetry.position_1") {
		t.Fatalf("expected position sensor latest row, got %#v", sensorLatestPayload)
	}

	recordingTickPayload := requestJSON[map[string]any](t, server.baseURL, http.MethodPost, "/api/v1/recorder/tick", "", map[string]any{
		"missionCode":          mission.code,
		"robotCode":            robot.code,
		"chunkDurationSeconds": 600,
		"tickAt":               time.Now().UTC().Format(time.RFC3339Nano),
	})
	chunk := recordingTickPayload["chunk"].(map[string]any)
	if chunk["status"] != "recording" {
		t.Fatalf("expected recording chunk, got %#v", chunk)
	}

	uploadedPayload := requestJSON[map[string]any](t, server.baseURL, http.MethodPost, "/api/v1/recorder/chunks/"+chunk["id"].(string)+"/uploaded", "", nil)
	uploadedChunk := uploadedPayload["chunk"].(map[string]any)
	if uploadedChunk["status"] != "uploaded" {
		t.Fatalf("expected uploaded chunk, got %#v", uploadedChunk)
	}
	fileUploadedPayload := requestJSON[map[string]any](t, server.baseURL, http.MethodPost, "/api/v1/recorder/chunks/"+chunk["id"].(string)+"/files/rgb_audio_mp4/uploaded", "", nil)
	fileUploadedChunk := fileUploadedPayload["chunk"].(map[string]any)
	fileTypes := fileUploadedChunk["availableFileTypes"].(map[string]any)
	if fileTypes["rgb_audio_mp4"] != true {
		t.Fatalf("expected rgb mp4 available flag, got %#v", fileUploadedChunk)
	}

	recordingsPayload := requestJSON[map[string]any](t, server.baseURL, http.MethodGet, "/api/v1/operator/recordings", "", nil)
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
}
