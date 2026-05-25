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

  it("maps canonical track.video_2 to thermal slot", () => {
    const slot = findTrackSlot({
      track: {
        id: "robot-001:track.video_2",
        kind: "video",
        label: "robot-001:track.video_2"
      },
      streams: []
    }, 0);

    expect(slot).toBe("thermal");
  });
});
