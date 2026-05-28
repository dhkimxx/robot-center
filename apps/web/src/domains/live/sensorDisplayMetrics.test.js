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
        { sensorId: "telemetry.gas.ch4", sensorType: "gas", displayName: "CH4", unit: "ppm" },
        { sensorId: "telemetry.gas.co", sensorType: "gas", displayName: "CO", unit: "ppm" },
        { sensorId: "telemetry.gas.o2", sensorType: "gas", displayName: "O2", unit: "%" },
        { sensorId: "telemetry.temperature_1", sensorType: "temperature", displayName: "Temperature" },
        { sensorId: "telemetry.humidity_1", sensorType: "humidity", displayName: "Humidity" },
        { sensorId: "telemetry.battery_1", sensorType: "battery", displayName: "Battery" }
      ],
      samples: [
        {
          sensorId: "telemetry.gas.ch4",
          values: { concentration: 2 }
        },
        {
          sensorId: "telemetry.gas.co",
          values: { concentration: 9 }
        },
        {
          sensorId: "telemetry.gas.o2",
          values: { concentration: 20.7 }
        },
        {
          sensorId: "telemetry.temperature_1",
          values: { temperatureCelsius: 29.5 }
        },
        {
          sensorId: "telemetry.humidity_1",
          values: { humidityPercent: 48 }
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
      "CH4",
      "CO",
      "O2",
      "온도",
      "습도",
      "배터리"
    ]);
  });

  it("keeps spatial samples when telemetry and spatial payloads are merged", () => {
    const merged = mergeSensorSnapshots(
      {
        descriptors: [{ sensorId: "telemetry.battery_1", sensorType: "battery", displayName: "Battery" }],
        samples: [{ sensorId: "telemetry.battery_1", values: { batteryPercent: 90 } }],
        receivedAt: "2026-05-26T01:00:00Z"
      },
      {
        descriptors: [{ sensorId: "spatial.imu_1", sensorType: "imu", displayName: "IMU" }],
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

  it("uses gas interpreter to expose concentration and alarm level", () => {
    const metrics = createSensorMetrics({
      descriptors: [
        {
          displayName: "CO",
          metadata: {
            criticalHigh: 50,
            warningHigh: 30
          },
          sensorId: "telemetry.gas.co",
          sensorType: "gas",
          unit: "ppm"
        }
      ],
      samples: [
        {
          sensorId: "telemetry.gas.co",
          values: {
            concentration: 35
          }
        }
      ]
    });

    expect(metrics).toHaveLength(1);
    expect(metrics[0]).toMatchObject({
      alarmLevel: "warning",
      label: "CO",
      unit: "ppm",
      value: 35
    });
  });

  it("creates metrics from sensor-latest API objects without a fixed four-metric limit", () => {
    const metrics = createSensorMetricsFromSensorLatest([
      {
        displayName: "Temperature",
        sensorId: "telemetry.temperature_1",
        sensorType: "temperature",
        latestSample: {
          values: {
            temperatureCelsius: 29.5
          }
        }
      },
      {
        displayName: "Humidity",
        sensorId: "telemetry.humidity_1",
        sensorType: "humidity",
        latestSample: {
          values: {
            humidityPercent: 48
          }
        }
      },
      {
        displayName: "Battery",
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
});
