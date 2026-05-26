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
        { sensorId: "telemetry.environment_1", kind: "environment", displayName: "Environment" },
        { sensorId: "telemetry.battery_1", kind: "battery", displayName: "Battery" }
      ],
      payload: {
        networkState: "mock-local"
      },
      samples: [
        {
          sensorId: "telemetry.environment_1",
          values: {
            ch4Ppm: 2,
            coPpm: 9,
            humidityPercent: 48,
            oxygenPercent: 20.7,
            temperatureCelsius: 29.5
          }
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
      "O2",
      "CH4",
      "온도",
      "습도",
      "배터리",
      "네트워크"
    ]);
  });

  it("keeps spatial samples when telemetry and spatial payloads are merged", () => {
    const merged = mergeSensorSnapshots(
      {
        descriptors: [{ sensorId: "telemetry.battery_1", kind: "battery", displayName: "Battery" }],
        samples: [{ sensorId: "telemetry.battery_1", values: { batteryPercent: 90 } }],
        sentAt: "2026-05-26T01:00:00Z"
      },
      {
        descriptors: [{ sensorId: "spatial.imu_1", kind: "imu", displayName: "IMU" }],
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
        sentAt: "2026-05-26T01:00:01Z"
      }
    );

    const metrics = createSensorMetrics(merged);

    expect(metrics.some((metric) => metric.label === "배터리")).toBe(true);
    expect(metrics.some((metric) => metric.label === "IMU 선가속도 X")).toBe(true);
    expect(merged.sentAt).toBe("2026-05-26T01:00:01Z");
  });

  it("creates metrics from sensor-latest API objects without a fixed four-metric limit", () => {
    const metrics = createSensorMetricsFromSensorLatest([
      {
        displayName: "Environment",
        sensorId: "telemetry.environment_1",
        sensorType: "environment",
        latestSample: {
          objectValue: {
            ch4Ppm: 2,
            coPpm: 9,
            humidityPercent: 48,
            oxygenPercent: 20.7,
            temperatureCelsius: 29.5
          }
        }
      },
      {
        displayName: "Battery",
        sensorId: "telemetry.battery_1",
        sensorType: "battery",
        latestSample: {
          objectValue: {
            batteryPercent: 91
          }
        }
      }
    ]);

    expect(metrics).toHaveLength(6);
    expect(metrics.map((metric) => metric.label)).toContain("배터리");
  });
});
