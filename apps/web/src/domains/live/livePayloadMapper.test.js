import { describe, expect, it } from "vitest";
import { mapLiveDataChannelPayload } from "./livePayloadMapper.js";

describe("mapLiveDataChannelPayload", () => {
  it("maps telemetry channel to telemetry and sensor projection", () => {
    const payload = {
      channelRole: "channel.telemetry",
      messageType: "telemetry",
      samples: [
        {
          sensorId: "telemetry.temperature_1",
          timestamp: "2026-05-28T01:00:00Z",
          values: {
            temperatureCelsius: 28.5
          }
        },
        {
          sensorId: "telemetry.humidity_1",
          timestamp: "2026-05-28T01:00:00Z",
          values: {
            humidityPercent: 48
          }
        },
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
      batteryPercent: 91,
      humidityPercent: 48,
      temperatureCelsius: 28.5
    });
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
