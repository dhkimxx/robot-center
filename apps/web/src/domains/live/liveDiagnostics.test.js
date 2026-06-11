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
    expect(summary.channelLabel).toBe("영상/음성 3채널 / 센서 2채널 / 관제 -");
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
        detail: "관제 화면 수신 RGB, 음성",
        key: "operator",
        tone: "ok",
        value: "연결됨"
      }),
      expect.objectContaining({
        detail: "최근 영상/음성 5초 전 · 송출 채널 3개",
        key: "media",
        tone: "ok",
        value: "송출 중"
      }),
      expect.objectContaining({
        detail: "최근 수신 4초 전 · 수신 채널 1개",
        key: "data",
        tone: "ok",
        value: "수신 중"
      }),
      expect.objectContaining({
        detail: expect.stringContaining("구간 #7"),
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
      detail: "센서/이벤트 수신 대기",
      tone: "waiting",
      value: "수신 대기"
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
