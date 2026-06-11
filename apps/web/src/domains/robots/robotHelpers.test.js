import { describe, expect, it } from "vitest";
import {
  findRobotOpenMission,
  groupRobotsByAvailability,
  isOnlineRobot,
  makeRobotStatusTone,
  selectDefaultRobotForManagement,
  shouldRefreshRobotEditForm
} from "./robotHelpers.js";

describe("robotHelpers", () => {
  it("does not refresh edit form while edit modal is open", () => {
    expect(shouldRefreshRobotEditForm({
      nextRobotCode: "robot-002",
      previousRobotCode: "robot-001",
      robotModal: "edit"
    })).toBe(false);
  });

  it("does not refresh edit form when selected robot code is unchanged", () => {
    expect(shouldRefreshRobotEditForm({
      nextRobotCode: "robot-001",
      previousRobotCode: "robot-001",
      robotModal: null
    })).toBe(false);
  });

  it("refreshes edit form when selected robot changes outside edit mode", () => {
    expect(shouldRefreshRobotEditForm({
      nextRobotCode: "robot-002",
      previousRobotCode: "robot-001",
      robotModal: null
    })).toBe(true);
  });

  it("groups robots by operator-facing availability", () => {
    const groups = groupRobotsByAvailability([
      { robotCode: "robot-offline", status: "offline" },
      { robotCode: "robot-online", status: "online" },
      { robotCode: "robot-streaming", status: "streaming" },
      { robotCode: "robot-fault", status: "fault" }
    ]);

    expect(groups.onlineRobots.map((robot) => robot.robotCode)).toEqual(["robot-streaming", "robot-online"]);
    expect(groups.offlineRobots.map((robot) => robot.robotCode)).toEqual(["robot-fault", "robot-offline"]);
    expect(isOnlineRobot({ status: "streaming" })).toBe(true);
    expect(isOnlineRobot({ status: "offline" })).toBe(false);
  });

  it("selects an online robot first for management defaults", () => {
    const robot = selectDefaultRobotForManagement([
      { robotCode: "robot-offline", status: "offline", lastSeenAt: "2026-06-11T06:00:00Z" },
      { robotCode: "robot-online", status: "online", lastSeenAt: "2026-06-11T05:00:00Z" }
    ]);

    expect(robot?.robotCode).toBe("robot-online");
  });

  it("finds the current ready or active mission for a robot", () => {
    const mission = findRobotOpenMission("robot-002", [
      { missionCode: "mission-ended", robotCodes: ["robot-002"], status: "ended" },
      { missionCode: "mission-active", robots: [{ robotCode: "robot-002" }], status: "active" }
    ]);

    expect(mission?.missionCode).toBe("mission-active");
  });

  it("maps robot status to badge tone", () => {
    expect(makeRobotStatusTone("online")).toBe("success");
    expect(makeRobotStatusTone("ready")).toBe("info");
    expect(makeRobotStatusTone("fault")).toBe("danger");
    expect(makeRobotStatusTone("offline")).toBe("neutral");
  });
});
