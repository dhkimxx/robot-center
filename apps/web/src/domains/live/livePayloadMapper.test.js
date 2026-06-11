import { describe, expect, it } from "vitest";
import { mapLiveDataChannelPayload } from "./livePayloadMapper.js";

describe("mapLiveDataChannelPayload", () => {
  it("maps telemetry channel to telemetry and sensor projection", () => {
    const payload = {
      channelRole: "channel.telemetry",
      messageType: "telemetry",
      samples: [
        {
          sensorId: "telemetry.battery_1",
          timestamp: "2026-05-28T01:00:01Z",
          values: {
            batteryPercent: 91
          }
        }
      ]
    };

    const result = mapLiveDataChannelPayload("channel.telemetry", JSON.stringify(payload));

    expect(result.ok).toBe(true);
    expect(result.telemetry).toMatchObject({
      ...payload,
      receivedAt: "2026-05-28T01:00:01Z"
    });
    expect(result.sensor.payload).toMatchObject({
      batteryPercent: 91
    });
  });

  it("maps gas-only telemetry samples to a live sensor projection", () => {
    const payload = {
      channelRole: "channel.telemetry",
      messageType: "telemetry",
      descriptors: [
        {
          enabled: true,
          label: "CO",
          sensorId: "telemetry.gas.channel_1",
          sensorType: "gas",
          unit: "ppm"
        }
      ],
      samples: [
        {
          sensorId: "telemetry.gas.channel_1",
          timestamp: "2026-06-11T08:35:40.490Z",
          values: {
            alarm: "normal",
            concentration: 12.8,
            valid: true
          }
        }
      ]
    };

    const result = mapLiveDataChannelPayload("channel.telemetry", JSON.stringify(payload));

    expect(result.ok).toBe(true);
    expect(result.sensor).toMatchObject({
      descriptors: payload.descriptors,
      receivedAt: "2026-06-11T08:35:40.490Z",
      samples: payload.samples
    });
    expect(result.sensor.payload).toBeUndefined();
  });

  it("maps event channel to event message", () => {
    const result = mapLiveDataChannelPayload("channel.event", JSON.stringify({
      channelRole: "channel.event",
      event: {
        message: "mock robot streaming"
      }
    }));

    expect(result.ok).toBe(true);
    expect(result.eventMessage).toBe("이벤트: mock robot streaming");
  });

  it("maps detection object events to RGB overlay projection", () => {
    const result = mapLiveDataChannelPayload("channel.event", JSON.stringify({
      channelRole: "channel.event",
      messageType: "event",
      events: [
        {
          eventType: "detection.object",
          timestamp: "2026-06-08T01:00:00.000Z",
          values: {
            trackId: "track.video_1",
            detections: [
              {
                className: "person",
                confidence: 0.92,
                bbox: {
                  x: 0.1,
                  y: 0.2,
                  width: 0.3,
                  height: 0.4
                }
              }
            ]
          }
        }
      ]
    }));

    expect(result.ok).toBe(true);
    expect(result.eventMessage).toBe("");
    expect(result.detectionOverlays).toHaveLength(1);
    expect(result.detectionOverlays[0]).toMatchObject({
      timestamp: "2026-06-08T01:00:00.000Z",
      trackId: "track.video_1",
      trackSlot: "rgb",
      detections: [
        {
          className: "person",
          confidence: 0.92,
          bbox: {
            x: 0.1,
            y: 0.2,
            width: 0.3,
            height: 0.4
          }
        }
      ]
    });
  });

  it("maps empty detection snapshots to an overlay clear projection", () => {
    const result = mapLiveDataChannelPayload("channel.event", JSON.stringify({
      channelRole: "channel.event",
      messageType: "event",
      events: [
        {
          eventType: "detection.object",
          timestamp: "2026-06-08T01:00:01.000Z",
          values: {
            trackId: "track.video_2",
            detections: []
          }
        }
      ]
    }));

    expect(result.ok).toBe(true);
    expect(result.detectionOverlays).toEqual([
      expect.objectContaining({
        detections: [],
        trackId: "track.video_2",
        trackSlot: "thermal"
      })
    ]);
  });

  it("maps mission events to live event panel items", () => {
    const result = mapLiveDataChannelPayload("channel.event", JSON.stringify({
      channelRole: "channel.event",
      messageType: "event",
      events: [
        {
          eventType: "mission.event",
          timestamp: "2026-06-08T01:03:00.000Z",
          values: {
            severity: "notice",
            title: "목표 지점 도착",
            description: "waypoint-3 도착",
            category: "navigation",
            code: "waypoint.arrived"
          }
        }
      ]
    }));

    expect(result.ok).toBe(true);
    expect(result.liveEvents).toEqual([
      expect.objectContaining({
        at: "2026-06-08T01:03:00.000Z",
        category: "navigation",
        code: "waypoint.arrived",
        description: "waypoint-3 도착",
        eventType: "mission.event",
        message: "목표 지점 도착",
        severity: "notice"
      })
    ]);
  });

  it("maps spatial channel to event message and sensor payload", () => {
    const payload = {
      channelRole: "channel.spatial",
      samples: [
        {
          sensorId: "spatial.imu_1",
          timestamp: "2026-05-28T01:00:00Z",
          values: {
            linearAcceleration: {
              x: 0.1
            }
          }
        }
      ]
    };

    const result = mapLiveDataChannelPayload("channel.spatial", JSON.stringify(payload));

    expect(result.ok).toBe(true);
    expect(result.eventMessage).toBe("공간 데이터 수신");
    expect(result.sensor).toMatchObject({
      ...payload,
      receivedAt: "2026-05-28T01:00:00Z"
    });
  });

  it("returns failed result for malformed JSON", () => {
    const result = mapLiveDataChannelPayload("channel.telemetry", "{");

    expect(result.ok).toBe(false);
    expect(result.eventMessage).toBe("데이터 해석 실패");
  });
});
