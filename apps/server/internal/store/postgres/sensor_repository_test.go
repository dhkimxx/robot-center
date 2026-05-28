package postgres

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"robot-center/apps/server/internal/domain"
)

func TestSensorRepositoryPersistsDescriptorsSamplesAndLatestValues(t *testing.T) {
	store := newPostgresTestStore(t)
	fixture := createActiveMissionFixture(t, store)
	ctx := context.Background()
	receivedAt := time.Date(2026, 5, 28, 1, 2, 3, 0, time.UTC)
	sentAt := receivedAt.Add(-time.Second)
	batteryValue := 87.5

	samples, err := store.SaveSensorEnvelope(ctx, domain.SensorEnvelope{
		MessageID:   "telemetry-001",
		MessageType: "telemetry",
		RobotCode:   fixture.Robot.RobotCode,
		MissionID:   fixture.Mission.ID,
		ChannelRole: "channel.telemetry",
		Sequence:    10,
		SentAt:      &sentAt,
		ReceivedAt:  receivedAt,
		RawPayload:  json.RawMessage(`{"messageType":"telemetry"}`),
		Descriptors: []domain.SensorDescriptor{
			{
				SensorID:    "telemetry.position_1",
				ChannelRole: "channel.telemetry",
				DisplayName: "GPS",
				SensorType:  "position",
				ValueType:   "object",
				Enabled:     true,
				Metadata:    json.RawMessage(`{"frame":"wgs84"}`),
			},
			{
				SensorID:    "telemetry.battery_1",
				ChannelRole: "channel.telemetry",
				DisplayName: "Battery",
				SensorType:  "battery",
				ValueType:   "number",
				Unit:        "percent",
				Enabled:     true,
			},
		},
		Samples: []domain.SensorSample{
			{
				SensorID:    "telemetry.position_1",
				ObjectValue: json.RawMessage(`{"latitude":37.402181,"longitude":127.106818}`),
			},
			{
				SensorID:     "telemetry.battery_1",
				NumericValue: &batteryValue,
			},
		},
	})
	if err != nil {
		t.Fatalf("save sensor envelope: %v", err)
	}
	if len(samples) != 2 {
		t.Fatalf("expected 2 samples, got %#v", samples)
	}

	descriptors, err := store.ListSensorDescriptors(ctx, fixture.Mission.ID, fixture.Robot.RobotCode)
	if err != nil {
		t.Fatalf("list descriptors: %v", err)
	}
	if len(descriptors) != 2 {
		t.Fatalf("expected 2 descriptors, got %#v", descriptors)
	}

	latest, err := store.ListLatestSensorSamples(ctx, fixture.Mission.ID, fixture.Robot.RobotCode)
	if err != nil {
		t.Fatalf("list latest samples: %v", err)
	}
	if len(latest) != 2 {
		t.Fatalf("expected 2 latest sensor rows, got %#v", latest)
	}
	if !latestSensorHasNumericValue(latest, "telemetry.battery_1", batteryValue) {
		t.Fatalf("expected latest battery value %.1f, got %#v", batteryValue, latest)
	}
	if !latestSensorHasObjectValue(latest, "telemetry.position_1", "latitude") {
		t.Fatalf("expected latest position object value, got %#v", latest)
	}

	if _, err := store.SaveSensorEnvelope(ctx, domain.SensorEnvelope{
		MessageID:   "telemetry-002",
		RobotCode:   fixture.Robot.RobotCode,
		MissionID:   fixture.Mission.ID,
		ChannelRole: "channel.telemetry",
		Sequence:    11,
		ReceivedAt:  receivedAt.Add(time.Second),
		RawPayload:  json.RawMessage(`{"messageType":"telemetry"}`),
		Descriptors: []domain.SensorDescriptor{
			{
				SensorID:    "telemetry.battery_1",
				DisplayName: "Main Battery",
				SensorType:  "battery",
				ValueType:   "number",
				Unit:        "percent",
				Enabled:     true,
			},
		},
		Samples: []domain.SensorSample{
			{
				SensorID:     "telemetry.battery_1",
				NumericValue: &batteryValue,
			},
		},
	}); err != nil {
		t.Fatalf("upsert sensor descriptor: %v", err)
	}

	descriptors, err = store.ListSensorDescriptors(ctx, fixture.Mission.ID, fixture.Robot.RobotCode)
	if err != nil {
		t.Fatalf("list descriptors after upsert: %v", err)
	}
	if len(descriptors) != 2 {
		t.Fatalf("expected descriptor upsert not to create duplicates, got %#v", descriptors)
	}
	if !descriptorHasDisplayName(descriptors, "telemetry.battery_1", "Main Battery") {
		t.Fatalf("expected updated battery display name, got %#v", descriptors)
	}
}

func latestSensorHasNumericValue(latest []domain.SensorLatest, sensorID string, expected float64) bool {
	for _, item := range latest {
		if item.Descriptor.SensorID == sensorID && item.LatestSample != nil && item.LatestSample.NumericValue != nil {
			return *item.LatestSample.NumericValue == expected
		}
	}
	return false
}

func latestSensorHasObjectValue(latest []domain.SensorLatest, sensorID string, key string) bool {
	for _, item := range latest {
		if item.Descriptor.SensorID != sensorID || item.LatestSample == nil {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal(item.LatestSample.ObjectValue, &payload); err != nil {
			return false
		}
		if _, ok := payload[key]; ok {
			return true
		}
	}
	return false
}

func descriptorHasDisplayName(descriptors []domain.SensorDescriptor, sensorID string, displayName string) bool {
	for _, descriptor := range descriptors {
		if descriptor.SensorID == sensorID && descriptor.DisplayName == displayName {
			return true
		}
	}
	return false
}
