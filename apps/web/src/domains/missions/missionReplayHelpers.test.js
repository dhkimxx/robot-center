import { describe, expect, it } from "vitest";
import {
  createRobotDisplayNamesByCode,
  getRobotDisplayName,
  makeFileAvailabilityLabel,
  makeLoadedChunkLabel,
  makeRobotSummaryStatusLabel,
  makeRobotSummaryTone
} from "./missionReplayHelpers.js";

describe("missionReplayHelpers", () => {
  it("uses robot display names when available", () => {
    const namesByCode = createRobotDisplayNamesByCode([
      { robotCode: "robot-001", displayName: "Alpha" }
    ]);

    expect(getRobotDisplayName(namesByCode, "robot-001")).toBe("Alpha");
    expect(getRobotDisplayName(namesByCode, "robot-002")).toBe("robot-002");
  });

  it("summarizes partial recording robots before completed state", () => {
    const summary = {
      chunkCount: 10,
      partialChunkCount: 2,
      recordingChunkCount: 0,
      finalizingChunkCount: 0,
      uploadedChunkCount: 8,
      availableFileCounts: {
        rgb_audio_mp4: 7
      }
    };

    expect(makeRobotSummaryTone(summary)).toBe("warning");
    expect(makeRobotSummaryStatusLabel(summary)).toBe("부분 저장");
    expect(makeFileAvailabilityLabel(summary, "rgb_audio_mp4", "RGB")).toBe("RGB 7/10");
    expect(makeLoadedChunkLabel(80, 130)).toBe("80/130개 로드");
  });
});
