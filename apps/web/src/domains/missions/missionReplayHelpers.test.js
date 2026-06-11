import { describe, expect, it } from "vitest";
import {
  createRobotDisplayNamesByCode,
  getRobotDisplayName,
  makeFileAvailabilityLabel,
  makeLoadedChunkLabel,
  makeReplayContinuityNotice,
  makeReplayRefreshStatusLabel,
  makeRobotSummaryStatusLabel,
  makeRobotSummaryTone,
  shouldAutoRefreshReplayRecordings
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

  it("auto-refreshes active missions and pending recording summaries", () => {
    expect(shouldAutoRefreshReplayRecordings({
      missionStatus: "active",
      selectedRobotSummary: null
    })).toBe(true);
    expect(shouldAutoRefreshReplayRecordings({
      missionStatus: "ended",
      selectedRobotSummary: { recordingChunkCount: 1, finalizingChunkCount: 0 }
    })).toBe(true);
    expect(shouldAutoRefreshReplayRecordings({
      missionStatus: "ended",
      selectedRobotSummary: { recordingChunkCount: 0, finalizingChunkCount: 0 }
    })).toBe(false);
  });

  it("explains replay continuity states", () => {
    expect(makeReplayContinuityNotice({ recordingChunkCount: 1 })).toBe("현재 녹화 중인 청크는 저장 완료 후 재생 버튼이 표시됩니다.");
    expect(makeReplayContinuityNotice({ finalizingChunkCount: 1 })).toBe("저장 마무리 중인 청크는 업로드가 끝나면 재생할 수 있습니다.");
    expect(makeReplayContinuityNotice({ partialChunkCount: 1 })).toBe("부분 저장 청크는 저장된 파일만 재생할 수 있습니다.");
    expect(makeReplayContinuityNotice({ uploadedChunkCount: 1 })).toBe("");
  });

  it("labels replay refresh status", () => {
    expect(makeReplayRefreshStatusLabel({ refreshing: true })).toBe("갱신 중");
    expect(makeReplayRefreshStatusLabel({ autoRefreshEnabled: true, loadedAt: "2026-06-11T00:00:00Z" })).toBe("자동 갱신 중");
    expect(makeReplayRefreshStatusLabel({ autoRefreshEnabled: false, loadedAt: "2026-06-11T00:00:00Z" })).toBe("최근 갱신됨");
    expect(makeReplayRefreshStatusLabel({ autoRefreshEnabled: true })).toBe("자동 갱신 대기");
  });
});
