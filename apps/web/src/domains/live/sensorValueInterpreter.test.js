import { describe, expect, it } from "vitest";
import { interpretSensorSampleValue } from "./sensorValueInterpreter.js";

describe("sensorValueInterpreter", () => {
  it("uses gas strategy for concentration readings", () => {
    const readings = interpretSensorSampleValue(
      {
        label: "CO",
        sensorId: "telemetry.gas.channel_1",
        sensorType: "gas",
        unit: "ppm"
      },
      {
        values: {
          alarm: "normal",
          concentration: 35,
          unit: "ppm",
          valid: true
        }
      }
    );

    expect(readings).toHaveLength(1);
    expect(readings[0]).toMatchObject({
      fieldPath: "concentration",
      label: "CO",
      unit: "ppm",
      value: 35
    });
  });

  it("flattens default object values and hides metadata fields", () => {
    const readings = interpretSensorSampleValue(
      {
        label: "Battery",
        sensorId: "telemetry.battery_1",
        sensorType: "battery",
        unit: "percent"
      },
      {
        values: {
          batteryPercent: 91,
          positionAvailable: true
        }
      }
    );

    expect(readings).toHaveLength(1);
    expect(readings[0]).toMatchObject({
      fieldPath: "batteryPercent",
      label: "배터리",
      unit: "%",
      value: 91
    });
  });
});
