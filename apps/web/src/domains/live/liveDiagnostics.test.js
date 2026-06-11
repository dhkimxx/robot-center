import { describe, expect, it } from "vitest";
import {
  getRobotLiveStatusSummary,
  makeLiveRobotDiagnostics,
  makeRecordingStateFromLiveStatus
} from "./liveDiagnostics.js";

describe("liveDiagnostics", () => {
  it("summarizes robot live status for dropdown chips", () => {
    const summary = getRobotLiveStatusSummary({
      liveSessions: {},
      target: {
        isStreaming: true,
        key: "mission-001:robot-001",
        liveStatus: {
          connection: { state: "online" },
          recording: { state: "recording" },
          stream: {
            dataChannelCount: 2,
            state: "streaming",
            trackCount: 3
          }
        }
      }
    });

    expect(summary.streamLabel).toBe("송출 중");
    expect(summary.recordingLabel).toBe("녹화 중");
    expect(summary.connectionLabel).toBe("연결됨");
    expect(summary.channelLabel).toBe("미디어 3 / 데이터 2 / 관제 -");
  });

  it("creates selected robot diagnostics from live-status and browser session", () => {
    const now = new Date("2026-06-11T05:00:10Z").getTime();
    const diagnostics = makeLiveRobotDiagnostics({
      now,
      session: {
        status: "connected",
        videoStreams: {
          audio: {},
          rgb: {},
          thermal: null
        }
      },
      target: {
        liveStatus: {
          recording: {
            state: "recording",
            latestChunk: {
              chunkIndex: 7,
              endedAt: "2026-06-11T05:10:00Z",
              startedAt: "2026-06-11T05:00:00Z",
              status: "recording"
            }
          },
          stream: {
            dataChannelCount: 1,
            lastDataAt: "2026-06-11T05:00:06Z",
            lastMediaAt: "2026-06-11T05:00:05Z",
            roomId: "mission-001",
            state: "streaming",
            trackCount: 3
          }
        }
      }
    });

    expect(diagnostics).toEqual([
      expect.objectContaining({
        detail: "브라우저 수신 RGB, Audio",
        key: "operator",
        tone: "ok",
        value: "연결됨"
      }),
      expect.objectContaining({
        detail: "최근 미디어 5초 전 · mission-001",
        key: "media",
        tone: "ok",
        value: "미디어 3개"
      }),
      expect.objectContaining({
        detail: "최근 데이터 4초 전",
        key: "data",
        tone: "ok",
        value: "데이터 1개"
      }),
      expect.objectContaining({
        detail: expect.stringContaining("chunk #7"),
        key: "recording",
        tone: "ok",
        value: "녹화 중"
      })
    ]);
  });

  it("keeps missing publisher data as waiting instead of hiding it", () => {
    const diagnostics = makeLiveRobotDiagnostics({
      session: { status: "connected", videoStreams: {} },
      target: {
        liveStatus: {
          recording: { state: "idle" },
          stream: {
            dataChannelCount: 0,
            state: "streaming",
            trackCount: 2
          }
        }
      }
    });

    expect(diagnostics.find((diagnostic) => diagnostic.key === "data")).toMatchObject({
      detail: "DataChannel 대기",
      tone: "waiting",
      value: "데이터 대기"
    });
  });

  it("maps recording states to reusable display states", () => {
    expect(makeRecordingStateFromLiveStatus({ state: "uploaded" })).toEqual({
      isActive: false,
      label: "저장 완료",
      tone: "available"
    });
    expect(makeRecordingStateFromLiveStatus({ state: "failed" })).toEqual({
      isActive: false,
      label: "녹화 오류",
      tone: "danger"
    });
  });
});
