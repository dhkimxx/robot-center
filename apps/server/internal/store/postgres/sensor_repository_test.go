package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"robot-center/apps/server/internal/domain"
	repo "robot-center/apps/server/internal/store/port"
)

func TestSensorRepositoryPersistsDescriptorsSamplesAndLatestValues(t *testing.T) {
	store := newPostgresTestStore(t)
	fixture := createActiveMissionFixture(t, store)
	ctx := context.Background()
	receivedAt := time.Date(2026, 5, 28, 1, 2, 3, 0, time.UTC)
	sampleTimestamp := receivedAt.Add(-time.Second)
	batteryValue := 87.5

	samples, err := store.SaveSensorEnvelope(ctx, domain.SensorEnvelope{
		MessageID:   "telemetry-001",
		MessageType: "telemetry",
		RobotCode:   fixture.Robot.RobotCode,
		MissionID:   fixture.Mission.ID,
		ChannelRole: "channel.telemetry",
		ReceivedAt:  receivedAt,
		RawPayload:  json.RawMessage(`{"messageType":"telemetry"}`),
		Descriptors: []domain.SensorDescriptor{
			{
				SensorID:    "telemetry.position_1",
				ChannelRole: "channel.telemetry",
				Label:       "GPS",
				SensorType:  "position",
				Enabled:     true,
			},
			{
				SensorID:    "telemetry.battery_1",
				ChannelRole: "channel.telemetry",
				Label:       "Battery",
				SensorType:  "battery",
				Unit:        "percent",
				Enabled:     true,
			},
		},
		Samples: []domain.SensorSample{
			{
				SensorID:  "telemetry.position_1",
				Timestamp: &sampleTimestamp,
				Values:    json.RawMessage(`{"latitude":37.402181,"longitude":127.106818}`),
			},
			{
				SensorID: "telemetry.battery_1",
				Values:   json.RawMessage(`{"batteryPercent":87.5}`),
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
	if !latestSensorHasValue(latest, "telemetry.battery_1", "batteryPercent") {
		t.Fatalf("expected latest battery value %.1f, got %#v", batteryValue, latest)
	}
	if !latestSensorHasValue(latest, "telemetry.position_1", "latitude") {
		t.Fatalf("expected latest position object value, got %#v", latest)
	}

	if _, err := store.SaveSensorEnvelope(ctx, domain.SensorEnvelope{
		MessageID:   "telemetry-002",
		RobotCode:   fixture.Robot.RobotCode,
		MissionID:   fixture.Mission.ID,
		ChannelRole: "channel.telemetry",
		ReceivedAt:  receivedAt.Add(time.Second),
		RawPayload:  json.RawMessage(`{"messageType":"telemetry"}`),
		Descriptors: []domain.SensorDescriptor{
			{
				SensorID:   "telemetry.battery_1",
				Label:      "Main Battery",
				SensorType: "battery",
				Unit:       "percent",
				Enabled:    true,
			},
		},
		Samples: []domain.SensorSample{
			{
				SensorID: "telemetry.battery_1",
				Values:   json.RawMessage(`{"batteryPercent":87.5}`),
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
	if !descriptorHasLabel(descriptors, "telemetry.battery_1", "Main Battery") {
		t.Fatalf("expected updated battery label, got %#v", descriptors)
	}

	if _, err := store.SaveSensorEnvelope(ctx, domain.SensorEnvelope{
		MessageID:   "telemetry-older",
		RobotCode:   fixture.Robot.RobotCode,
		MissionID:   fixture.Mission.ID,
		ChannelRole: "channel.telemetry",
		ReceivedAt:  receivedAt.Add(-time.Minute),
		RawPayload:  json.RawMessage(`{"messageType":"telemetry"}`),
		Samples: []domain.SensorSample{
			{
				SensorID: "telemetry.battery_1",
				Values:   json.RawMessage(`{"batteryPercent":12}`),
			},
		},
	}); err != nil {
		t.Fatalf("save older sensor sample: %v", err)
	}
	latest, err = store.ListLatestSensorSamples(ctx, fixture.Mission.ID, fixture.Robot.RobotCode)
	if err != nil {
		t.Fatalf("list latest after older sample: %v", err)
	}
	if latestSensorNumericValue(latest, "telemetry.battery_1", "batteryPercent") == 12 {
		t.Fatalf("expected older sample not to overwrite latest cache, got %#v", latest)
	}
	history, err := store.ListSensorSamples(ctx, fixture.Mission.ID, fixture.Robot.RobotCode, "telemetry.battery_1", 10)
	if err != nil {
		t.Fatalf("list battery history: %v", err)
	}
	if !sensorSamplesHaveValue(history, "batteryPercent", 12) {
		t.Fatalf("expected older sample to remain in history, got %#v", history)
	}

	_, err = store.SaveSensorEnvelope(ctx, domain.SensorEnvelope{
		MessageID:   "telemetry-003",
		RobotCode:   fixture.Robot.RobotCode,
		MissionID:   fixture.Mission.ID,
		ChannelRole: "channel.telemetry",
		ReceivedAt:  receivedAt.Add(2 * time.Second),
		RawPayload:  json.RawMessage(`{"messageType":"telemetry"}`),
		Samples: []domain.SensorSample{
			{
				SensorID: "telemetry.unregistered_1",
				Values:   json.RawMessage(`{"value":1}`),
			},
		},
	})
	if !errors.Is(err, repo.ErrInvalidState) {
		t.Fatalf("expected sample without descriptor to be rejected, got %v", err)
	}
}

func TestSensorRepositoryListsArchivedRobotSensorHistory(t *testing.T) {
	store := newPostgresTestStore(t)
	fixture := createActiveMissionFixture(t, store)
	ctx := context.Background()
	receivedAt := time.Date(2026, 5, 28, 3, 4, 5, 0, time.UTC)

	if _, err := store.SaveSensorEnvelope(ctx, domain.SensorEnvelope{
		MessageID:   "telemetry-archived-robot",
		RobotCode:   fixture.Robot.RobotCode,
		MissionID:   fixture.Mission.ID,
		ChannelRole: "channel.telemetry",
		ReceivedAt:  receivedAt,
		RawPayload:  json.RawMessage(`{"messageType":"telemetry"}`),
		Descriptors: []domain.SensorDescriptor{
			{
				SensorID:   "telemetry.battery_1",
				Label:      "Battery",
				SensorType: "battery",
				Unit:       "percent",
				Enabled:    true,
			},
		},
		Samples: []domain.SensorSample{
			{
				SensorID: "telemetry.battery_1",
				Values:   json.RawMessage(`{"batteryPercent":74}`),
			},
		},
	}); err != nil {
		t.Fatalf("save sensor envelope: %v", err)
	}
	if _, err := store.EndMission(ctx, fixture.Mission.MissionCode); err != nil {
		t.Fatalf("end mission before archive: %v", err)
	}
	if _, err := store.ArchiveRobot(ctx, fixture.Robot.RobotCode); err != nil {
		t.Fatalf("archive robot: %v", err)
	}

	latest, err := store.ListLatestSensorSamples(ctx, fixture.Mission.ID, fixture.Robot.RobotCode)
	if err != nil {
		t.Fatalf("list archived robot latest samples: %v", err)
	}
	if !latestSensorHasValue(latest, "telemetry.battery_1", "batteryPercent") {
		t.Fatalf("expected archived robot latest sensor history, got %#v", latest)
	}

	samples, err := store.ListSensorSamples(ctx, fixture.Mission.ID, fixture.Robot.RobotCode, "telemetry.battery_1", 10)
	if err != nil {
		t.Fatalf("list archived robot samples: %v", err)
	}
	if len(samples) != 1 {
		t.Fatalf("expected archived robot sensor sample history, got %#v", samples)
	}
}

func TestSensorRepositoryClearSensorDataDeletesLatestCache(t *testing.T) {
	store := newPostgresTestStore(t)
	fixture := createActiveMissionFixture(t, store)
	ctx := context.Background()

	if _, err := store.SaveSensorEnvelope(ctx, domain.SensorEnvelope{
		MessageID:   "telemetry-clear",
		RobotCode:   fixture.Robot.RobotCode,
		MissionID:   fixture.Mission.ID,
		ChannelRole: "channel.telemetry",
		ReceivedAt:  time.Date(2026, 5, 28, 5, 6, 7, 0, time.UTC),
		RawPayload:  json.RawMessage(`{"messageType":"telemetry"}`),
		Descriptors: []domain.SensorDescriptor{
			{
				SensorID:   "telemetry.battery_1",
				Label:      "Battery",
				SensorType: "battery",
				Unit:       "percent",
				Enabled:    true,
			},
		},
		Samples: []domain.SensorSample{
			{
				SensorID: "telemetry.battery_1",
				Values:   json.RawMessage(`{"batteryPercent":80}`),
			},
		},
	}); err != nil {
		t.Fatalf("save sensor envelope: %v", err)
	}

	result, err := store.ClearSensorData(ctx)
	if err != nil {
		t.Fatalf("clear sensor data: %v", err)
	}
	if result.SensorLatestSamplesDeleted != 1 || result.SensorSamplesDeleted != 1 || result.SensorDescriptorsDeleted != 1 {
		t.Fatalf("unexpected clear result: %#v", result)
	}
	if rows := countRows(t, store, "sensor_latest_samples"); rows != 0 {
		t.Fatalf("expected latest cache to be empty, got %d", rows)
	}
}

func latestSensorHasValue(latest []domain.SensorLatest, sensorID string, key string) bool {
	for _, item := range latest {
		if item.Descriptor.SensorID != sensorID || item.LatestSample == nil {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal(item.LatestSample.Values, &payload); err != nil {
			return false
		}
		if _, ok := payload[key]; ok {
			return true
		}
	}
	return false
}

func descriptorHasLabel(descriptors []domain.SensorDescriptor, sensorID string, label string) bool {
	for _, descriptor := range descriptors {
		if descriptor.SensorID == sensorID && descriptor.Label == label {
			return true
		}
	}
	return false
}

func latestSensorNumericValue(latest []domain.SensorLatest, sensorID string, key string) float64 {
	for _, item := range latest {
		if item.Descriptor.SensorID != sensorID || item.LatestSample == nil {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal(item.LatestSample.Values, &payload); err != nil {
			return 0
		}
		if value, ok := payload[key].(float64); ok {
			return value
		}
	}
	return 0
}

func sensorSamplesHaveValue(samples []domain.SensorSample, key string, expected float64) bool {
	for _, sample := range samples {
		var payload map[string]any
		if err := json.Unmarshal(sample.Values, &payload); err != nil {
			continue
		}
		if value, ok := payload[key].(float64); ok && value == expected {
			return true
		}
	}
	return false
}

func countRows(t *testing.T, store *Store, tableName string) int {
	t.Helper()

	var count int
	if err := store.sqlDB.QueryRow(`SELECT COUNT(*) FROM ` + tableName).Scan(&count); err != nil {
		t.Fatalf("count %s rows: %v", tableName, err)
	}
	return count
}
