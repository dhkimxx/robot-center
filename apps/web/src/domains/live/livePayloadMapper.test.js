import { describe, expect, it } from "vitest";
import { mapLiveDataChannelPayload } from "./livePayloadMapper.js";

describe("mapLiveDataChannelPayload", () => {
  it("maps telemetry channel to telemetry and sensor projection", () => {
    const payload = {
      channelRole: "channel.telemetry",
      messageType: "telemetry",
      samples: [
        {
          sensorId: "telemetry.environment_1",
          values: {
            coPpm: 7,
            oxygenPercent: 20.8,
            temperatureCelsius: 28.5
          }
        },
        {
          sensorId: "telemetry.battery_1",
          values: {
            batteryPercent: 91
          }
        }
      ],
      payload: {
        position: {
          latitude: 37.5,
          longitude: 127.0
        }
      }
    };

    const result = mapLiveDataChannelPayload("channel.telemetry", JSON.stringify(payload));

    expect(result.ok).toBe(true);
    expect(result.telemetry).toEqual(payload);
    expect(result.sensor.payload).toMatchObject({
      batteryPercent: 91,
      coPpm: 7,
      oxygenPercent: 20.8,
      position: {
        latitude: 37.5,
        longitude: 127.0
      },
      temperatureCelsius: 28.5
    });
  });

  it("maps legacy sensor message to sensor only", () => {
    const payload = {
      messageType: "sensor",
      payload: {
        coPpm: 3
      }
    };

    const result = mapLiveDataChannelPayload("sensor", JSON.stringify(payload));

    expect(result.ok).toBe(true);
    expect(result.telemetry).toBeUndefined();
    expect(result.sensor).toEqual(payload);
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

  it("returns failed result for malformed JSON", () => {
    const result = mapLiveDataChannelPayload("channel.telemetry", "{");

    expect(result.ok).toBe(false);
    expect(result.eventMessage).toBe("데이터 해석 실패");
  });
});
