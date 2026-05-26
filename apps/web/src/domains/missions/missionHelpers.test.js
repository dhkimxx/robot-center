import { describe, expect, it } from "vitest";
import {
  createMissionRobotTargets,
  getBusyRobotReasonForMissionCreate,
  getMissionRobotDetails,
  groupMissionsByLifecycle
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
    const observedStreams = [
      {
        roomId: "mission-001",
        publishers: [
          {
            robotCode: "robot-001",
            state: "publishing",
            tracks: ["robot-001:track.video_1"],
            trackCount: 1,
            dataChannels: ["channel.telemetry"],
            dataChannelCount: 1,
            lastTrackAt: new Date().toISOString(),
            updatedAt: new Date().toISOString()
          }
        ]
      }
    ];

    const targets = createMissionRobotTargets(mission, robots, [], observedStreams);

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
    const details = getMissionRobotDetails(mission, [{ robotCode: "robot-001", status: "streaming" }], [], [
      {
        roomId: "mission-001",
        publishers: [
          {
            robotCode: "robot-001",
            state: "publishing",
            trackCount: 1,
            lastTrackAt: new Date().toISOString()
          }
        ]
      }
    ]);

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
    ], []);

    expect(reason).toBe("진행 중 임무 mission-002");
  });

  it("ignores streaming status when the room id does not match the mission room", () => {
    const mission = {
      id: "mission-id-001",
      missionCode: "mission-001",
      status: "active",
      robotCodes: ["robot-001"]
    };
    const details = getMissionRobotDetails(mission, [{ robotCode: "robot-001", status: "online" }], [
      {
        missionId: "mission-id-001",
        robotCode: "robot-001",
        roomId: "mission-001__robot-001",
        status: "streaming",
        sentAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      }
    ]);

    expect(details[0]).toMatchObject({
      isStreaming: false,
      liveLabel: "송출 대기"
    });
  });

  it("does not use streaming-status reports as live truth", () => {
    const mission = {
      id: "mission-id-001",
      missionCode: "mission-001",
      status: "active",
      robotCodes: ["robot-001"]
    };
    const details = getMissionRobotDetails(mission, [{ robotCode: "robot-001", status: "online" }], [
      {
        missionId: "mission-id-001",
        robotCode: "robot-001",
        roomId: "mission-001",
        status: "streaming",
        sentAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      }
    ], []);

    expect(details[0]).toMatchObject({
      isStreaming: false,
      liveLabel: "송출 대기"
    });
  });

  it("does not use streaming-status reports as mission create guards", () => {
    const nowMs = Date.parse("2026-05-26T00:00:00.000Z");
    const reason = getBusyRobotReasonForMissionCreate("robot-001", [], [
      {
        robotCode: "robot-001",
        roomId: "mission-001",
        status: "streaming",
        sentAt: "2026-05-26T00:10:00.000Z",
        updatedAt: "2026-05-25T23:58:00.000Z"
      }
    ], nowMs);

    expect(reason).toBe("");
  });

  it("uses observed publishers for mission create guards", () => {
    const now = new Date("2026-05-26T00:00:00.000Z");
    const reason = getBusyRobotReasonForMissionCreate("robot-001", [], [], now.getTime(), [
      {
        roomId: "mission-001",
        publishers: [
          {
            robotCode: "robot-001",
            iceState: "connected",
            lastTrackAt: now.toISOString()
          }
        ]
      }
    ]);

    expect(reason).toBe("실시간 송출 중");
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

    const targets = createMissionRobotTargets(mission, [], [], [
      {
        roomId: "mission-001",
        publishers: [
          {
            robotCode: "robot-001",
            iceState: "connected",
            lastTrackAt: new Date().toISOString()
          }
        ]
      }
    ], liveStatus);

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
});
