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
	    const streamingStatuses = [
	      {
	        missionId: "mission-id-001",
	        robotCode: "robot-001",
	        roomId: "mission-001",
	        status: "streaming",
	        sentAt: new Date().toISOString(),
	        updatedAt: new Date().toISOString()
	      }
	    ];

    const targets = createMissionRobotTargets(mission, robots, streamingStatuses);

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
    const details = getMissionRobotDetails(mission, [{ robotCode: "robot-001", status: "streaming" }], [
      {
	        missionId: "mission-id-001",
	        robotCode: "robot-001",
	        roomId: "mission-001",
	        status: "streaming",
	        sentAt: new Date().toISOString(),
	        updatedAt: new Date().toISOString()
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

	  it("uses server updatedAt for fresh streaming create guards", () => {
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
