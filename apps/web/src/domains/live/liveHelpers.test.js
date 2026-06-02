import { describe, expect, it } from "vitest";
import {
  findRobotCodeFromDataMessage,
  findTrackSlot
} from "./liveHelpers.js";

describe("liveHelpers", () => {
  it("extracts robotCode from canonical data channel payload", () => {
    const robotCode = findRobotCodeFromDataMessage(JSON.stringify({
      channelRole: "channel.telemetry",
      robotCode: "robot-001"
    }));

    expect(robotCode).toBe("robot-001");
  });

  it("maps canonical media track ids to live slots", () => {
    expect(findTrackSlot({
      track: { id: "robot-001:track.video_1", kind: "video" },
      streams: []
    })).toBe("rgb");

    expect(findTrackSlot({
      track: { id: "robot-001:track.video_2", kind: "video" },
      streams: []
    })).toBe("thermal");

    expect(findTrackSlot({
      track: { id: "robot-001:track.audio_1", kind: "audio" },
      streams: []
    })).toBe("audio");
  });

  it("does not map non-canonical media tracks by kind or order", () => {
    expect(findTrackSlot({
      track: { id: "webrtctransceiver0", kind: "video" },
      streams: []
    })).toBe("unmapped");

    expect(findTrackSlot({
      track: { id: "robot-001-audio", kind: "audio" },
      streams: []
    })).toBe("unmapped");
  });
});
