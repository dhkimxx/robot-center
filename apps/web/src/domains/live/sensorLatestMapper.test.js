import { describe, expect, it } from "vitest";
import {
  createSensorPanelSnapshot,
  createTelemetryFromSensorLatest
} from "./sensorLatestMapper.js";

describe("sensorLatestMapper", () => {
  const sensorLatest = [
    {
      label: "GPS",
      missionId: "mission-id",
      robotCode: "robot-001",
      sensorId: "telemetry.position_1",
      sensorType: "position",
      latestSample: {
        values: {
          latitude: 37.5,
          longitude: 127.0
        },
        receivedAt: "2026-05-26T01:00:00Z"
      }
    },
    {
      label: "CO",
      missionId: "mission-id",
      robotCode: "robot-001",
      sensorId: "telemetry.gas.channel_1",
      sensorType: "gas",
      unit: "ppm",
      latestSample: {
        values: {
          concentration: 9
        },
        receivedAt: "2026-05-26T01:00:01Z"
      }
    },
    {
      label: "O2",
      missionId: "mission-id",
      robotCode: "robot-001",
      sensorId: "telemetry.gas.channel_3",
      sensorType: "gas",
      unit: "%Vol",
      latestSample: {
        values: {
          concentration: 20.8
        },
        receivedAt: "2026-05-26T01:00:01Z"
      }
    }
  ];

  it("creates telemetry-compatible position state", () => {
    const telemetry = createTelemetryFromSensorLatest(sensorLatest, "robot-001");

    expect(telemetry.payload.position).toEqual({
      latitude: 37.5,
      longitude: 127.0
    });
    expect(telemetry.payload.positionAvailable).toBe(true);
  });

  it("creates dynamic sensor panel metrics", () => {
    const sensor = createSensorPanelSnapshot(sensorLatest, "robot-001");

    expect(sensor.sensors).toEqual([
      expect.objectContaining({
        label: "CO",
        receivedAt: "2026-05-26T01:00:01Z",
        unit: "ppm",
        value: 9
      }),
      expect.objectContaining({
        label: "O2",
        receivedAt: "2026-05-26T01:00:01Z",
        unit: "%Vol",
        value: 20.8
      })
    ]);
  });
});
