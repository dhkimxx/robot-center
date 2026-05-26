import { describe, expect, it } from "vitest";
import {
  createSensorPanelSnapshot,
  createTelemetryFromSensorLatest
} from "./sensorLatestMapper.js";

describe("sensorLatestMapper", () => {
  const sensorLatest = [
    {
      displayName: "GPS",
      missionId: "mission-id",
      robotCode: "robot-001",
      sensorId: "telemetry.position_1",
      sensorType: "position",
      latestSample: {
        objectValue: {
          latitude: 37.5,
          longitude: 127.0
        },
        receivedAt: "2026-05-26T01:00:00Z"
      }
    },
    {
      displayName: "Gas",
      missionId: "mission-id",
      robotCode: "robot-001",
      sensorId: "telemetry.gas_1",
      sensorType: "gas",
      latestSample: {
        objectValue: {
          coPpm: 9,
          oxygenPercent: 20.8
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
      {
        label: "CO",
        receivedAt: "2026-05-26T01:00:01Z",
        unit: "ppm",
        value: 9
      },
      {
        label: "O2",
        receivedAt: "2026-05-26T01:00:01Z",
        unit: "%",
        value: 20.8
      }
    ]);
  });
});
