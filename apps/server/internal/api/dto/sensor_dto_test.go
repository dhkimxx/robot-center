package dto

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSensorResponseWrapperShape(t *testing.T) {
	descriptorPayload, err := json.Marshal(SensorDescriptorsResponse{SensorDescriptors: []SensorDescriptorResponse{}})
	if err != nil {
		t.Fatalf("marshal descriptor payload: %v", err)
	}
	samplePayload, err := json.Marshal(SensorSamplesResponse{SensorSamples: []SensorSampleResponse{}})
	if err != nil {
		t.Fatalf("marshal sample payload: %v", err)
	}
	latestPayload, err := json.Marshal(SensorLatestListResponse{
		MissionID: "mission-001",
		RobotCode: "robot-001",
		Sensors:   []SensorLatestResponse{},
	})
	if err != nil {
		t.Fatalf("marshal latest payload: %v", err)
	}

	assertJSONHasField(t, descriptorPayload, "sensorDescriptors")
	assertJSONHasField(t, samplePayload, "sensorSamples")
	assertJSONHasField(t, latestPayload, "missionId")
	assertJSONHasField(t, latestPayload, "robotCode")
	assertJSONHasField(t, latestPayload, "sensors")
}

func TestSensorEnvelopeRequestShape(t *testing.T) {
	timestamp := time.Date(2026, 6, 2, 8, 0, 0, 0, time.UTC)
	payload, err := json.Marshal(SensorEnvelopeRequest{
		MessageID:   "msg-001",
		MessageType: "telemetry",
		RobotCode:   "robot-001",
		MissionID:   "mission-001",
		ChannelRole: "channel.telemetry",
		Descriptors: []SensorDescriptorRequest{
			{
				SensorID:    "gas",
				ChannelRole: "channel.telemetry",
				Label:       "Gas",
				SensorType:  "gas",
				Unit:        "ppm",
				Enabled:     true,
			},
		},
		Samples: []SensorSampleRequest{
			{
				SensorID:    "gas",
				ChannelRole: "channel.telemetry",
				MessageID:   "sample-001",
				Timestamp:   &timestamp,
				Values:      map[string]any{"co": 1},
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	for _, field := range []string{"messageId", "messageType", "robotCode", "missionId", "channelRole", "descriptors", "samples"} {
		assertJSONHasField(t, payload, field)
	}
}

func assertJSONHasField(t *testing.T, payload []byte, field string) {
	t.Helper()
	var fields map[string]any
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if _, ok := fields[field]; !ok {
		t.Fatalf("expected field %q in JSON, got %s", field, string(payload))
	}
}
