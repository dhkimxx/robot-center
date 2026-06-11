import { describe, expect, it } from "vitest";
import {
  createMissionRobotTargets,
  getBusyRobotReasonForMissionCreate,
  getMissionRobotDetails,
  groupMissionsByLifecycle,
  makeLiveRecordingTimingLabel,
  makeLiveStreamTimingLabel
} from "./missionHelpers.js";

describe("missionHelpers", () => {
  it("creates mission robot targets with missionCode room id", () => {
    const mission = {
      id: "mission-id-001",
      missionCode: "mission-001",
      status: "active",
      robotCodes: ["robot-001"]
    };
    const robots = [
      {
        displayName: "Robot 1",
        robotCode: "robot-001"
      }
    ];
    const liveStatus = {
      missionCode: "mission-001",
      missionStatus: "active",
      robots: [
        {
          robotCode: "robot-001",
          stream: { state: "streaming", trackCount: 1, dataChannelCount: 1 },
          recording: { state: "recording" },
          connection: { state: "online" }
        }
      ]
    };

    const targets = createMissionRobotTargets(mission, robots, liveStatus);

    expect(targets).toHaveLength(1);
    expect(targets[0]).toMatchObject({
      key: "mission-001:robot-001",
      missionRoomId: "mission-001",
      roomId: "mission-001",
      robotCode: "robot-001",
      isStreaming: true,
      liveLabel: "송출 중"
    });
  });

  it("does not show streaming for closed missions even when robot has a stale status", () => {
    const mission = {
      id: "mission-id-001",
      missionCode: "mission-001",
      status: "ended",
      robotCodes: ["robot-001"]
    };
    const details = getMissionRobotDetails(mission, [{ robotCode: "robot-001", status: "streaming" }]);

    expect(details[0]).toMatchObject({
      deviceStatus: "streaming",
      isStreaming: false,
      liveLabel: "임무 종료"
    });
  });

  it("marks robots in active missions as busy for new mission creation", () => {
    const reason = getBusyRobotReasonForMissionCreate("robot-001", [
      {
        missionCode: "mission-002",
        status: "active",
        robotCodes: ["robot-001"]
      }
    ]);

    expect(reason).toBe("진행 중 임무 mission-002");
  });

  it("uses live-status as the control target status source", () => {
    const mission = {
      id: "mission-id-001",
      missionCode: "mission-001",
      status: "active",
      robotCodes: ["robot-001"]
    };
    const liveStatus = {
      missionCode: "mission-001",
      missionStatus: "active",
      robots: [
        {
          robotCode: "robot-001",
          stream: { state: "waiting" },
          recording: { state: "idle" },
          connection: { state: "online" }
        }
      ]
    };

    const targets = createMissionRobotTargets(mission, [], liveStatus);

    expect(targets[0]).toMatchObject({
      isStreaming: false,
      liveStatus: {
        stream: { state: "waiting" },
        recording: { state: "idle" }
      }
    });
  });

  it("groups open missions separately from closed missions", () => {
    const groups = groupMissionsByLifecycle([
      { missionCode: "mission-ended", status: "ended", startedAt: "2026-05-20T00:00:00Z" },
      { missionCode: "mission-active", status: "active", startedAt: "2026-05-21T00:00:00Z" },
      { missionCode: "mission-ready", status: "ready", createdAt: "2026-05-22T00:00:00Z" }
    ]);

    expect(groups.openMissions.map((mission) => mission.missionCode)).toEqual(["mission-active", "mission-ready"]);
    expect(groups.closedMissions.map((mission) => mission.missionCode)).toEqual(["mission-ended"]);
  });

  it("summarizes live stream timing for current and interrupted publishers", () => {
    const now = new Date("2026-06-04T04:35:20Z").getTime();

    const streamingLabel = makeLiveStreamTimingLabel({
      state: "streaming",
      startedAt: "2026-06-04T04:35:00Z",
      lastTrackAt: "2026-06-04T04:35:15Z",
      diagnostics: { reconnectCount: 1 }
    }, now);

    expect(streamingLabel).toContain("최근 수신 5초 전");
    expect(streamingLabel).not.toContain("재접속");

    expect(makeLiveStreamTimingLabel({
      state: "waiting",
      diagnostics: {
        lastSessionMediaAt: "2026-06-04T04:34:00Z",
        previousEndedAt: "2026-06-04T04:34:30Z"
      }
    }, now)).toBe("송출 대기");

    expect(makeLiveStreamTimingLabel(null, now)).toBe("상태 확인 중");
  });

  it("summarizes the active recording chunk window", () => {
    expect(makeLiveRecordingTimingLabel({
      state: "recording",
      latestChunk: {
        chunkIndex: 0,
        startedAt: "2026-06-04T04:30:00Z",
        endedAt: "2026-06-04T04:40:00Z",
        status: "recording"
      }
    })).toContain("녹화 구간 #0");
  });
});
