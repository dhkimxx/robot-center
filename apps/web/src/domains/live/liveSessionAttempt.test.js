import { describe, expect, it } from "vitest";
import { LiveSessionStatus } from "./liveConnectionStates.js";
import { createEmptyLiveSession } from "./liveHelpers.js";
import { applyLiveAttemptUpdate } from "./liveSessionAttempts.js";

describe("live session attempt guard", () => {
  it("ignores an old attempt close after a newer attempt starts", () => {
    const session = {
      ...createEmptyLiveSession(),
      attemptId: 17,
      status: LiveSessionStatus.CONNECTING
    };

    const nextSession = applyLiveAttemptUpdate(session, 16, (current) => ({
      ...current,
      status: LiveSessionStatus.IDLE
    }));

    expect(nextSession).toBe(session);
    expect(nextSession.status).toBe(LiveSessionStatus.CONNECTING);
  });

  it("keeps failed status when a stale close arrives later", () => {
    const session = {
      ...createEmptyLiveSession(),
      attemptId: 17,
      status: LiveSessionStatus.FAILED
    };

    const nextSession = applyLiveAttemptUpdate(session, 16, (current) => ({
      ...current,
      status: LiveSessionStatus.SIGNALING_CLOSED
    }));

    expect(nextSession).toBe(session);
    expect(nextSession.status).toBe(LiveSessionStatus.FAILED);
  });

  it("allows current navigation close to reset streams", () => {
    const session = {
      ...createEmptyLiveSession(),
      attemptId: 17,
      status: LiveSessionStatus.CONNECTED,
      videoStreams: { rgb: { id: "rgb" }, thermal: { id: "thermal" }, audio: { id: "audio" } }
    };

    const nextSession = applyLiveAttemptUpdate(session, 17, (current) => ({
      ...current,
      status: LiveSessionStatus.IDLE,
      videoStreams: { rgb: null, thermal: null, audio: null }
    }));

    expect(nextSession.status).toBe(LiveSessionStatus.IDLE);
    expect(nextSession.videoStreams).toEqual({ rgb: null, thermal: null, audio: null });
  });

  it("ignores stale track or data updates", () => {
    const session = {
      ...createEmptyLiveSession(),
      attemptId: 17,
      telemetry: { robotCode: "robot-002" },
      videoStreams: { rgb: { id: "current-rgb" }, thermal: null, audio: null }
    };

    const nextSession = applyLiveAttemptUpdate(session, 16, (current) => ({
      ...current,
      telemetry: { robotCode: "robot-001" },
      videoStreams: { ...current.videoStreams, rgb: { id: "stale-rgb" } }
    }));

    expect(nextSession).toBe(session);
    expect(nextSession.telemetry).toEqual({ robotCode: "robot-002" });
    expect(nextSession.videoStreams.rgb).toEqual({ id: "current-rgb" });
  });
});
