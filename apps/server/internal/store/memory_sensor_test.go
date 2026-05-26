package store

import (
	"context"
	"testing"
	"time"

	"robot-center/apps/server/internal/domain"
)

func TestMemoryStoreSaveSensorEnvelopeCreatesDescriptorForSampleOnlyPayload(t *testing.T) {
	ctx := context.Background()
	repository := NewMemoryStore("http://127.0.0.1:18080")
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	value := 42.5

	samples, err := repository.SaveSensorEnvelope(ctx, domain.SensorEnvelope{
		RobotCode:   "robot-001",
		MissionID:   "mission-001",
		ChannelRole: "channel.telemetry",
		ReceivedAt:  now,
		Samples: []domain.SensorSample{
			{
				SensorID:     "telemetry.gas_1",
				NumericValue: &value,
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveSensorEnvelope returned error: %v", err)
	}
	if len(samples) != 1 || samples[0].DescriptorID == "" {
		t.Fatalf("samples = %#v, want one sample linked to an auto descriptor", samples)
	}

	descriptors, err := repository.ListSensorDescriptors(ctx, "mission-001", "robot-001")
	if err != nil {
		t.Fatalf("ListSensorDescriptors returned error: %v", err)
	}
	if len(descriptors) != 1 {
		t.Fatalf("descriptors = %#v, want one auto descriptor", descriptors)
	}
	if descriptors[0].ID != samples[0].DescriptorID || !descriptors[0].Enabled {
		t.Fatalf("descriptor = %#v, sample descriptor id = %q", descriptors[0], samples[0].DescriptorID)
	}
	if descriptors[0].SensorType != "gas" || descriptors[0].ValueType != "number" {
		t.Fatalf("descriptor type/valueType = %q/%q, want gas/number", descriptors[0].SensorType, descriptors[0].ValueType)
	}

	latest, err := repository.ListLatestSensorSamples(ctx, "mission-001", "robot-001")
	if err != nil {
		t.Fatalf("ListLatestSensorSamples returned error: %v", err)
	}
	if len(latest) != 1 || latest[0].LatestSample == nil || latest[0].LatestSample.DescriptorID != descriptors[0].ID {
		t.Fatalf("latest = %#v, want descriptor with linked latest sample", latest)
	}
}

func TestMemoryStoreListLatestSensorSamplesKeepsSameSensorIDPerRobot(t *testing.T) {
	ctx := context.Background()
	repository := NewMemoryStore("http://127.0.0.1:18080")
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	firstBattery := 81.0
	secondBattery := 67.0

	saveBatterySample(t, repository, "robot-001", now, firstBattery)
	saveBatterySample(t, repository, "robot-002", now.Add(time.Second), secondBattery)

	latest, err := repository.ListLatestSensorSamples(ctx, "mission-001", "")
	if err != nil {
		t.Fatalf("ListLatestSensorSamples returned error: %v", err)
	}
	if len(latest) != 2 {
		t.Fatalf("latest = %#v, want one latest row per robot", latest)
	}
	for _, item := range latest {
		if item.LatestSample == nil {
			t.Fatalf("latest item = %#v, want sample", item)
		}
		if item.Descriptor.RobotCode != item.LatestSample.RobotCode {
			t.Fatalf("latest item = %#v, descriptor and sample robot codes must match", item)
		}
		if item.Descriptor.SensorID != "telemetry.battery_1" || item.LatestSample.SensorID != "telemetry.battery_1" {
			t.Fatalf("latest item = %#v, want battery sensor for both descriptor and sample", item)
		}
	}
}

func saveBatterySample(t *testing.T, repository *MemoryStore, robotCode string, receivedAt time.Time, batteryPercent float64) {
	t.Helper()

	_, err := repository.SaveSensorEnvelope(context.Background(), domain.SensorEnvelope{
		RobotCode:   robotCode,
		MissionID:   "mission-001",
		ChannelRole: "channel.telemetry",
		ReceivedAt:  receivedAt,
		Samples: []domain.SensorSample{
			{
				SensorID:     "telemetry.battery_1",
				ReceivedAt:   receivedAt,
				NumericValue: &batteryPercent,
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveSensorEnvelope(%s) returned error: %v", robotCode, err)
	}
}
