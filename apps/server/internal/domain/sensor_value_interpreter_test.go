package domain

import (
	"encoding/json"
	"testing"
)

func TestInterpretSensorSampleValueUsesGasStrategy(t *testing.T) {
	readings := InterpretSensorSampleValue(
		SensorDescriptor{
			SensorID:   "telemetry.gas.channel_1",
			Label:      "CO",
			SensorType: "gas",
			Unit:       "ppm",
		},
		SensorSample{
			Values: json.RawMessage(`{"alarm":"normal","concentration":35,"unit":"ppm","valid":true}`),
		},
	)

	if len(readings) != 1 {
		t.Fatalf("expected one gas reading, got %#v", readings)
	}
	if readings[0].Label != "CO" || readings[0].FieldPath != "concentration" || readings[0].Value != float64(35) {
		t.Fatalf("unexpected gas reading: %#v", readings[0])
	}
	if readings[0].Order != 10.01 || readings[0].Unit != "ppm" {
		t.Fatalf("unexpected gas presentation metadata: %#v", readings[0])
	}
}

func TestInterpretSensorSampleValueFlattensDefaultObject(t *testing.T) {
	readings := InterpretSensorSampleValue(
		SensorDescriptor{
			SensorID:   "telemetry.battery_1",
			Label:      "Battery",
			SensorType: "battery",
			Unit:       "percent",
		},
		SensorSample{
			Values: json.RawMessage(`{"batteryPercent":91,"positionAvailable":true}`),
		},
	)

	if len(readings) != 1 {
		t.Fatalf("expected one visible reading, got %#v", readings)
	}
	if readings[0].Label != "배터리" || readings[0].FieldPath != "batteryPercent" || readings[0].Value != float64(91) {
		t.Fatalf("unexpected battery reading: %#v", readings[0])
	}
	if readings[0].Order != 30 || readings[0].Unit != "%" {
		t.Fatalf("unexpected battery presentation metadata: %#v", readings[0])
	}
}
