import { describe, expect, it } from "vitest";
import {
  filterRecordingsByMissionCode,
  getPlayableRecordingVideoEntries,
  makeRecordingStateForTarget,
  sortRecordingChunksLatestFirst
} from "./recordingHelpers.js";

describe("recordingHelpers", () => {
  it("marks active recording chunks as recording", () => {
    const state = makeRecordingStateForTarget([
      {
        missionCode: "mission-001",
        robotCode: "robot-001",
        status: "recording"
      }
    ], "mission-001", "robot-001");

    expect(state).toMatchObject({
      isActive: true,
      label: "녹화 중",
      tone: "recording"
    });
  });

  it("does not mix recording state across robots", () => {
    const state = makeRecordingStateForTarget([
      {
        missionCode: "mission-001",
        robotCode: "robot-002",
        status: "recording"
      }
    ], "mission-001", "robot-001");

    expect(state).toMatchObject({
      isActive: false,
      label: "녹화 대기",
      tone: "idle"
    });
  });

  it("does not show stopped chunks as active recording", () => {
    const state = makeRecordingStateForTarget([
      {
        missionCode: "mission-001",
        robotCode: "robot-001",
        status: "stopped"
      }
    ], "mission-001", "robot-001");

    expect(state).toMatchObject({
      isActive: false,
      label: "부분 저장",
      tone: "idle"
    });
  });

  it("shows finalizing chunks as save finalization", () => {
    const state = makeRecordingStateForTarget([
      {
        missionCode: "mission-001",
        robotCode: "robot-001",
        status: "finalizing"
      }
    ], "mission-001", "robot-001");

    expect(state).toMatchObject({
      isActive: false,
      label: "저장 마무리",
      tone: "recording"
    });
  });

  it("filters recordings by mission code", () => {
    const recordings = [
      { id: "recording-1", missionCode: "mission-001" },
      { id: "recording-2", missionCode: "mission-002" },
      { id: "recording-3", missionCode: "mission-001" }
    ];

    expect(filterRecordingsByMissionCode(recordings, "mission-001").map((recording) => recording.id)).toEqual([
      "recording-1",
      "recording-3"
    ]);
    expect(filterRecordingsByMissionCode(recordings, "")).toEqual([]);
  });

  it("sorts recording chunks latest first", () => {
    const recordings = [
      { id: "recording-1", chunkIndex: 1, updatedAt: "2026-05-01T00:00:00Z" },
      { id: "recording-3", chunkIndex: 3, updatedAt: "2026-05-01T00:02:00Z" },
      { id: "recording-2", chunkIndex: 2, updatedAt: "2026-05-01T00:01:00Z" }
    ];

    expect(sortRecordingChunksLatestFirst(recordings).map((recording) => recording.id)).toEqual([
      "recording-3",
      "recording-2",
      "recording-1"
    ]);
  });

  it("extracts playable video files only", () => {
    const playableEntries = getPlayableRecordingVideoEntries({
      files: [
        {
          type: "rgb_audio_mp4",
          status: "available",
          contentType: "video/mp4",
          url: "http://storage/rgb.mp4"
        },
        {
          type: "sensor_jsonl",
          status: "available",
          contentType: "application/x-ndjson",
          url: "http://storage/sensor.jsonl"
        },
        {
          type: "thermal_mp4",
          status: "recording",
          contentType: "video/mp4",
          url: ""
        }
      ]
    });

    expect(playableEntries).toHaveLength(1);
    expect(playableEntries[0].type).toBe("rgb_audio_mp4");
  });
});
