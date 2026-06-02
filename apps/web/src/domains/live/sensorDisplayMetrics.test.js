import { describe, expect, it } from "vitest";
import {
  createSensorMetrics,
  createSensorMetricsFromSensorLatest,
  mergeSensorSnapshots
} from "./sensorDisplayMetrics.js";

describe("sensorDisplayMetrics", () => {
  it("creates dynamic metrics from canonical telemetry descriptors and samples", () => {
    const metrics = createSensorMetrics({
      descriptors: [
        { sensorId: "telemetry.gas.slot_f", sensorType: "gas", label: "HUM", unit: "RH" },
        { sensorId: "telemetry.gas.slot_d", sensorType: "gas", label: "CH4", unit: "%LEL" },
        { sensorId: "telemetry.gas.slot_b", sensorType: "gas", label: "H2S", unit: "ppm" },
        { sensorId: "telemetry.gas.slot_e", sensorType: "gas", label: "TEMP", unit: "degC" },
        { sensorId: "telemetry.gas.slot_a", sensorType: "gas", label: "CO", unit: "ppm" },
        { sensorId: "telemetry.gas.slot_c", sensorType: "gas", label: "O2", unit: "%Vol" },
        { sensorId: "telemetry.battery_1", sensorType: "battery", label: "Battery" }
      ],
      samples: [
        {
          sensorId: "telemetry.gas.slot_f",
          values: { concentration: 48 }
        },
        {
          sensorId: "telemetry.gas.slot_d",
          values: { concentration: 2 }
        },
        {
          sensorId: "telemetry.gas.slot_b",
          values: { concentration: 2.1 }
        },
        {
          sensorId: "telemetry.gas.slot_e",
          values: { concentration: 29.5 }
        },
        {
          sensorId: "telemetry.gas.slot_a",
          values: { concentration: 9 }
        },
        {
          sensorId: "telemetry.gas.slot_c",
          values: { concentration: 20.7 }
        },
        {
          sensorId: "telemetry.battery_1",
          values: {
            batteryPercent: 91
          }
        }
      ]
    });

    expect(metrics.map((metric) => metric.label)).toEqual([
      "CO",
      "H2S",
      "O2",
      "CH4",
      "TEMP",
      "HUM",
      "배터리"
    ]);
  });

  it("keeps spatial samples when telemetry and spatial payloads are merged", () => {
    const merged = mergeSensorSnapshots(
      {
        descriptors: [{ sensorId: "telemetry.battery_1", sensorType: "battery", label: "Battery" }],
        samples: [{ sensorId: "telemetry.battery_1", values: { batteryPercent: 90 } }],
        receivedAt: "2026-05-26T01:00:00Z"
      },
      {
        descriptors: [{ sensorId: "spatial.imu_1", sensorType: "imu", label: "IMU" }],
        samples: [
          {
            sensorId: "spatial.imu_1",
            values: {
              linearAcceleration: {
                x: 0.1,
                y: 0.2,
                z: 9.8
              }
            }
          }
        ],
        receivedAt: "2026-05-26T01:00:01Z"
      }
    );

    const metrics = createSensorMetrics(merged);

    expect(metrics.some((metric) => metric.label === "배터리")).toBe(true);
    expect(metrics.some((metric) => metric.label === "IMU 선가속도 X")).toBe(true);
    expect(merged.receivedAt).toBe("2026-05-26T01:00:01Z");
  });

  it("shows gas module concentration and unit without alarm interpretation", () => {
    const metrics = createSensorMetrics({
      descriptors: [
        {
          label: "CO",
          sensorId: "telemetry.gas.channel_1",
          sensorType: "gas",
          unit: "ppm"
        }
      ],
      samples: [
        {
          sensorId: "telemetry.gas.channel_1",
          values: {
            alarm: "normal",
            alarm_code: 0,
            concentration: 35,
            high_alarm: 30,
            low_alarm: 10,
            scale_code: 1,
            unit: "ppm",
            valid: true
          }
        }
      ]
    });

    expect(metrics).toHaveLength(1);
    expect(metrics[0]).toMatchObject({
      alarmLevel: "normal",
      label: "CO",
      unit: "ppm",
      value: 35
    });
  });

  it("creates metrics from sensor-latest API objects without a fixed four-metric limit", () => {
    const metrics = createSensorMetricsFromSensorLatest([
      {
        label: "TEMP",
        sensorId: "telemetry.gas.channel_5",
        sensorType: "gas",
        unit: "degC",
        latestSample: {
          values: {
            concentration: 29.5
          }
        }
      },
      {
        label: "HUM",
        sensorId: "telemetry.gas.channel_6",
        sensorType: "gas",
        unit: "RH",
        latestSample: {
          values: {
            concentration: 48
          }
        }
      },
      {
        label: "Battery",
        sensorId: "telemetry.battery_1",
        sensorType: "battery",
        latestSample: {
          values: {
            batteryPercent: 91
          }
        }
      }
    ]);

    expect(metrics).toHaveLength(3);
    expect(metrics.map((metric) => metric.label)).toContain("배터리");
  });

  it("prefers backend readings for sensor-latest metrics", () => {
    const metrics = createSensorMetricsFromSensorLatest([
      {
        label: "Battery",
        sensorId: "telemetry.battery_1",
        sensorType: "battery",
        unit: "percent",
        latestSample: {
          receivedAt: "2026-05-26T01:00:01Z",
          values: {
            batteryPercent: 91
          }
        },
        readings: [
          {
            fieldPath: "batteryPercent",
            label: "배터리",
            order: 30,
            unit: "%",
            value: 91
          }
        ]
      }
    ]);

    expect(metrics).toEqual([
      expect.objectContaining({
        label: "배터리",
        receivedAt: "2026-05-26T01:00:01Z",
        unit: "%",
        value: 91
      })
    ]);
  });
});
