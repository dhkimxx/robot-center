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
      ]
    };

    const result = mapLiveDataChannelPayload("channel.telemetry", JSON.stringify(payload));

    expect(result.ok).toBe(true);
    expect(result.telemetry).toEqual(payload);
    expect(result.sensor.payload).toMatchObject({
      batteryPercent: 91,
      coPpm: 7,
      oxygenPercent: 20.8,
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
    expect(result.sensor).toEqual(payload);
  });

  it("returns failed result for malformed JSON", () => {
    const result = mapLiveDataChannelPayload("channel.telemetry", "{");

    expect(result.ok).toBe(false);
    expect(result.eventMessage).toBe("데이터 해석 실패");
  });
});
