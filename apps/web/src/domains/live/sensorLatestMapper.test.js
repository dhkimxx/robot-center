import { describe, expect, it } from "vitest";
import {
  createSensorPanelState,
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

  it("prefers live sensor values over stored snapshots", () => {
    const snapshotSensor = createSensorPanelSnapshot(sensorLatest, "robot-001");
    const liveSensor = {
      receivedAt: new Date().toISOString(),
      robotCode: "robot-001",
      sensors: [
        {
          label: "CO",
          receivedAt: new Date().toISOString(),
          unit: "ppm",
          value: 11
        }
      ]
    };

    const state = createSensorPanelState({
      liveSensor,
      snapshotSensor,
      snapshotState: { status: "ready" }
    });

    expect(state.sensor).toBe(liveSensor);
    expect(state.source).toBe("live");
    expect(state.sourceLabel).toBe("실시간 수신");
  });

  it("marks stale snapshots when snapshot refresh fails", () => {
    const snapshotSensor = createSensorPanelSnapshot(sensorLatest, "robot-001");

    const state = createSensorPanelState({
      liveSensor: null,
      snapshotSensor,
      snapshotState: { error: "failed", status: "error" }
    });

    expect(state.sensor).toBe(snapshotSensor);
    expect(state.source).toBe("snapshot-error");
    expect(state.sourceLabel).toBe("최근 저장값 갱신 실패");
  });
});
