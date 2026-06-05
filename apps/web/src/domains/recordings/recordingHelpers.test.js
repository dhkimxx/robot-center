import { describe, expect, it } from "vitest";
import {
  getPlayableRecordingVideoEntries,
  makeRecordingStateForTarget
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
