package api

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"robot-center/apps/server/internal/api/dto"
)

func TestRecorderAPIFlow(t *testing.T) {
	server := newAPIFlowTestServer(t)
	robot := server.createRobot(t, "Recorder Robot")
	mission := server.createStartedMission(t, robot)

	targetsPayload := requestJSON[dto.RecorderRecordingTargetsResponse](t, server.baseURL, http.MethodGet, "/api/v1/recorder/recording-targets", "", nil)
	if len(targetsPayload.Targets) != 1 {
		t.Fatalf("expected one recording target, got %#v", targetsPayload)
	}

	positionTimestamp := time.Now().UTC()
	gasTimestamp := time.Now().UTC()
	requestJSON[dto.SensorSamplesResponse](t, server.baseURL, http.MethodPost, "/api/v1/recorder/sensor-samples", "", dto.SensorEnvelopeRequest{
		MessageID:   "telemetry-canonical-1",
		MessageType: "telemetry",
		ChannelRole: "channel.telemetry",
		RobotCode:   robot.code,
		MissionID:   mission.id,
		Descriptors: []dto.SensorDescriptorRequest{
			{
				SensorID:   "telemetry.position_1",
				SensorType: "position",
				Label:      "GPS",
				Enabled:    true,
			},
			{
				SensorID:   "telemetry.gas.channel_1",
				SensorType: "gas",
				Label:      "Gas",
				Enabled:    true,
			},
		},
		Samples: []dto.SensorSampleRequest{
			{
				SensorID:  "telemetry.position_1",
				Timestamp: &positionTimestamp,
				Values: map[string]any{
					"latitude":  37.402181,
					"longitude": 127.106818,
				},
			},
			{
				SensorID:  "telemetry.gas.channel_1",
				Timestamp: &gasTimestamp,
				Values: map[string]any{
					"concentration": 12.3,
				},
			},
		},
	})

	missingSensorTypeStatus := requestStatus(t, server.baseURL, http.MethodPost, "/api/v1/recorder/sensor-samples", "", map[string]any{
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

	invalidSensorTypeStatus := requestStatus(t, server.baseURL, http.MethodPost, "/api/v1/recorder/sensor-samples", "", map[string]any{
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

	sampleWithoutDescriptorStatus := requestStatus(t, server.baseURL, http.MethodPost, "/api/v1/recorder/sensor-samples", "", map[string]any{
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

	payloadOnlyStatus := requestStatus(t, server.baseURL, http.MethodPost, "/api/v1/recorder/sensor-samples", "", map[string]any{
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

	sensorLatestPayload := requestJSON[dto.SensorLatestListResponse](t, server.baseURL, http.MethodGet, "/api/v1/operator/sensor-latest?missionId="+mission.id+"&robotCode="+robot.code, "", nil)
	if len(sensorLatestPayload.Sensors) != 2 {
		t.Fatalf("expected two latest sensor rows, got %#v", sensorLatestPayload)
	}
	if !sensorListHasID(sensorLatestPayload.Sensors, "telemetry.position_1") {
		t.Fatalf("expected position sensor latest row, got %#v", sensorLatestPayload)
	}

	eventTimestamp := time.Now().UTC()
	eventPayload := requestJSON[dto.MissionEventsResponse](t, server.baseURL, http.MethodPost, "/api/v1/recorder/events", "", dto.EventEnvelopeRequest{
		MessageID:   "event-canonical-1",
		MessageType: "event",
		ChannelRole: "channel.event",
		RobotCode:   robot.code,
		MissionID:   mission.id,
		Events: []dto.EventItemRequest{
			{
				EventID:   "detection-rgb-person",
				EventType: "detection.object",
				Timestamp: &eventTimestamp,
				Values:    []byte(`{"trackId":"track.video_1","detections":[{"className":"person","confidence":0.92,"bbox":{"x":0.1,"y":0.2,"width":0.3,"height":0.4}}]}`),
			},
			{
				EventID:   "detection-thermal-empty",
				EventType: "detection.object",
				Timestamp: &eventTimestamp,
				Values:    []byte(`{"trackId":"track.video_2","detections":[]}`),
			},
			{
				EventID:   "mission-waypoint",
				EventType: "mission.event",
				Timestamp: &eventTimestamp,
				Values:    []byte(`{"severity":"notice","title":"목표 지점 도착","description":"waypoint-3 도착","category":"navigation","code":"waypoint.arrived"}`),
			},
			{
				EventID:   "mission-code-only",
				EventType: "mission.event",
				Timestamp: &eventTimestamp,
				Values:    []byte(`{"severity":"WARNING","category":"diagnostic","code":"battery.low"}`),
			},
		},
	})
	if len(eventPayload.Events) != 4 {
		t.Fatalf("expected four event rows, got %#v", eventPayload)
	}
	if !eventListHasType(eventPayload.Events, "detection.object") || !eventListHasType(eventPayload.Events, "mission.event") {
		t.Fatalf("expected detection and mission events, got %#v", eventPayload)
	}
	if !eventListHasTitleAndSeverity(eventPayload.Events, "battery.low", "warning") {
		t.Fatalf("expected mission event title fallback to code and severity normalization, got %#v", eventPayload)
	}
	if !eventListHasDetectionCount(eventPayload.Events, "track.video_1", 1) {
		t.Fatalf("expected non-empty detection snapshot to persist with detectionCount=1, got %#v", eventPayload)
	}
	if !eventListHasDetectionCount(eventPayload.Events, "track.video_2", 0) {
		t.Fatalf("expected empty detection snapshot to persist with detectionCount=0, got %#v", eventPayload)
	}

	operatorEventsPayload := requestJSON[dto.OperatorMissionEventsResponse](t, server.baseURL, http.MethodGet, "/api/v1/operator/missions/"+mission.code+"/events", "", nil)
	if len(operatorEventsPayload.Events) != 2 || !eventListHasTitleAndSeverity(operatorEventsPayload.Events, "battery.low", "warning") {
		t.Fatalf("expected default operator event feed to exclude detection.object, got %#v", operatorEventsPayload)
	}
	operatorDetectionsPayload := requestJSON[dto.OperatorMissionEventsResponse](t, server.baseURL, http.MethodGet, "/api/v1/operator/missions/"+mission.code+"/events?eventType=detection.object", "", nil)
	if len(operatorDetectionsPayload.Events) != 2 || !eventListHasDetectionCount(operatorDetectionsPayload.Events, "track.video_1", 1) || !eventListHasDetectionCount(operatorDetectionsPayload.Events, "track.video_2", 0) {
		t.Fatalf("expected explicit detection query to return snapshot, got %#v", operatorDetectionsPayload)
	}
	assertInvalidDetectionObjectEventsRejected(t, server, robot, mission)
	missingEventValuesStatus := requestStatus(t, server.baseURL, http.MethodPost, "/api/v1/recorder/events", "", map[string]any{
		"messageId":   "event-missing-values",
		"messageType": "event",
		"channelRole": "channel.event",
		"robotCode":   robot.code,
		"missionId":   mission.id,
		"events": []map[string]any{
			{
				"eventType": "mission.event",
			},
		},
	})
	if missingEventValuesStatus != http.StatusBadRequest {
		t.Fatalf("expected event without values to be rejected, got %d", missingEventValuesStatus)
	}

	clearSensorPayload := requestJSON[dto.ClearSensorDataResponse](t, server.baseURL, http.MethodPost, "/api/v1/system/sensors/clear", "", dto.ClearSensorDataRequest{
		Confirmation: "CLEAR_SENSOR_DATA",
	})
	if clearSensorPayload.SensorData.SensorSamplesDeleted != 2 {
		t.Fatalf("expected two sensor samples deleted, got %#v", clearSensorPayload)
	}
	sensorLatestAfterClearPayload := requestJSON[dto.SensorLatestListResponse](t, server.baseURL, http.MethodGet, "/api/v1/operator/sensor-latest?missionId="+mission.id+"&robotCode="+robot.code, "", nil)
	if latestAfterClear := sensorLatestAfterClearPayload.Sensors; len(latestAfterClear) != 0 {
		t.Fatalf("expected sensor latest to be empty after clear, got %#v", sensorLatestAfterClearPayload)
	}

	recordingTickPayload := requestJSON[dto.RecorderRecordingTickResponse](t, server.baseURL, http.MethodPost, "/api/v1/recorder/tick", "", dto.RecorderTickRequest{
		MissionCode:          mission.code,
		RobotCode:            robot.code,
		ChunkDurationSeconds: 600,
		TickAt:               time.Now().UTC(),
	})
	chunk := recordingTickPayload.Chunk
	if chunk.Status != "recording" {
		t.Fatalf("expected recording chunk, got %#v", chunk)
	}

	uploadedPayload := requestJSON[dto.RecorderRecordingChunkEnvelopeResponse](t, server.baseURL, http.MethodPost, "/api/v1/recorder/chunks/"+chunk.ID+"/uploaded", "", nil)
	uploadedChunk := uploadedPayload.Chunk
	if uploadedChunk.Status != "uploaded" {
		t.Fatalf("expected recorder chunk uploaded state, got %#v", uploadedChunk)
	}
	operatorRecordingsBeforeVideoPayload := requestJSON[dto.OperatorRecordingsResponse](t, server.baseURL, http.MethodGet, "/api/v1/operator/recordings", "", nil)
	if len(operatorRecordingsBeforeVideoPayload.Recordings) != 1 || operatorRecordingsBeforeVideoPayload.Recordings[0].Status != "partial" {
		t.Fatalf("expected operator playback state to be partial before video files are available, got %#v", operatorRecordingsBeforeVideoPayload)
	}
	fileUploadedPayload := requestJSON[dto.RecorderRecordingChunkEnvelopeResponse](t, server.baseURL, http.MethodPost, "/api/v1/recorder/chunks/"+chunk.ID+"/files/rgb_audio_mp4/uploaded", "", nil)
	fileUploadedChunk := fileUploadedPayload.Chunk
	if fileUploadedChunk.Status != "uploaded" {
		t.Fatalf("expected uploaded chunk after video file is available, got %#v", fileUploadedChunk)
	}
	if fileUploadedChunk.AvailableFileTypes["rgb_audio_mp4"] != true {
		t.Fatalf("expected rgb mp4 available flag, got %#v", fileUploadedChunk)
	}

	recordingsPayload := requestJSON[dto.OperatorRecordingsResponse](t, server.baseURL, http.MethodGet, "/api/v1/operator/recordings", "", nil)
	if len(recordingsPayload.Recordings) != 1 {
		t.Fatalf("expected one recording, got %#v", recordingsPayload)
	}
	recording := recordingsPayload.Recordings[0]
	if recording.RecordingSessionID == "" {
		t.Fatalf("expected recordingSessionId in response, got %#v", recording)
	}
	if !fileHasAvailableURL(recording.Files, "manifest") {
		t.Fatalf("expected manifest file with available URL, got %#v", recording.Files)
	}
	if !fileHasAvailableURL(recording.Files, "rgb_audio_mp4") {
		t.Fatalf("expected rgb mp4 file with available URL, got %#v", recording.Files)
	}
	filteredRecordingsPayload := requestJSON[dto.OperatorRecordingsResponse](t, server.baseURL, http.MethodGet, "/api/v1/operator/recordings?missionCode="+url.QueryEscape(mission.code), "", nil)
	if len(filteredRecordingsPayload.Recordings) != 1 || filteredRecordingsPayload.Recordings[0].MissionCode != mission.code {
		t.Fatalf("expected one recording for mission filter, got %#v", filteredRecordingsPayload)
	}
	missionRecordingSummaryPayload := requestJSON[dto.OperatorMissionRecordingSummaryResponse](t, server.baseURL, http.MethodGet, "/api/v1/operator/missions/"+mission.code+"/recordings/summary", "", nil)
	if missionRecordingSummaryPayload.TotalChunks != 1 || len(missionRecordingSummaryPayload.Robots) != 1 {
		t.Fatalf("expected one mission recording summary row, got %#v", missionRecordingSummaryPayload)
	}
	if missionRecordingSummaryPayload.Robots[0].RobotCode != robot.code || missionRecordingSummaryPayload.Robots[0].AvailableFileCounts["rgb_audio_mp4"] != 1 {
		t.Fatalf("expected robot recording file summary, got %#v", missionRecordingSummaryPayload)
	}
	missionRecordingChunksPayload := requestJSON[dto.OperatorMissionRecordingChunksResponse](t, server.baseURL, http.MethodGet, "/api/v1/operator/missions/"+mission.code+"/recordings/chunks?robotCode="+url.QueryEscape(robot.code)+"&limit=1", "", nil)
	if missionRecordingChunksPayload.Page.Total != 1 || missionRecordingChunksPayload.Page.HasMore {
		t.Fatalf("expected one paged mission recording chunk, got %#v", missionRecordingChunksPayload)
	}
	if len(missionRecordingChunksPayload.Recordings) != 1 || missionRecordingChunksPayload.Recordings[0].RobotCode != robot.code {
		t.Fatalf("expected one robot recording chunk, got %#v", missionRecordingChunksPayload)
	}
	if !fileHasAvailableURL(missionRecordingChunksPayload.Recordings[0].Files, "rgb_audio_mp4") {
		t.Fatalf("expected paged rgb mp4 file URL, got %#v", missionRecordingChunksPayload.Recordings[0].Files)
	}
	missingMissionRecordingsPayload := requestJSON[dto.OperatorRecordingsResponse](t, server.baseURL, http.MethodGet, "/api/v1/operator/recordings?missionCode=missing-mission", "", nil)
	if len(missingMissionRecordingsPayload.Recordings) != 0 {
		t.Fatalf("expected no recordings for missing mission filter, got %#v", missingMissionRecordingsPayload)
	}
}

func assertInvalidDetectionObjectEventsRejected(t *testing.T, server apiFlowTestServer, robot testRobot, mission testMission) {
	t.Helper()

	invalidDetectionCases := []struct {
		name   string
		values string
	}{
		{name: "missing-track", values: `{"detections":[]}`},
		{name: "invalid-track", values: `{"trackId":"track.video_9","detections":[]}`},
		{name: "missing-detections", values: `{"trackId":"track.video_1"}`},
		{name: "missing-class", values: `{"trackId":"track.video_1","detections":[{"confidence":0.9,"bbox":{"x":0.1,"y":0.1,"width":0.2,"height":0.2}}]}`},
		{name: "invalid-confidence", values: `{"trackId":"track.video_1","detections":[{"className":"person","confidence":1.1,"bbox":{"x":0.1,"y":0.1,"width":0.2,"height":0.2}}]}`},
		{name: "missing-bbox", values: `{"trackId":"track.video_1","detections":[{"className":"person","confidence":0.9}]}`},
		{name: "invalid-bbox-range", values: `{"trackId":"track.video_1","detections":[{"className":"person","confidence":0.9,"bbox":{"x":0.9,"y":0.1,"width":0.2,"height":0.2}}]}`},
		{name: "non-object-values", values: `[]`},
	}
	for _, testCase := range invalidDetectionCases {
		invalidDetectionStatus := requestStatus(t, server.baseURL, http.MethodPost, "/api/v1/recorder/events", "", dto.EventEnvelopeRequest{
			MessageID:   "event-invalid-detection-" + testCase.name,
			MessageType: "event",
			ChannelRole: "channel.event",
			RobotCode:   robot.code,
			MissionID:   mission.id,
			Events: []dto.EventItemRequest{
				{
					EventType: "detection.object",
					Values:    []byte(testCase.values),
				},
			},
		})
		if invalidDetectionStatus != http.StatusBadRequest {
			t.Fatalf("expected invalid detection %s to be rejected, got %d", testCase.name, invalidDetectionStatus)
		}
	}
}
