import { getMissionRobotCodes } from "../missions/missionHelpers.js";

const onlineRobotStatuses = new Set(["online", "streaming", "active"]);
const robotStatusOrder = {
  streaming: 0,
  active: 1,
  online: 2,
  ready: 3,
  idle: 4,
  fault: 5,
  error: 5,
  failed: 5,
  offline: 6
};

export function createInitialRobotForm() {
  return {
    displayName: "현장 로봇 1",
    modelName: "Field Robot"
  };
}

export function createRobotEditForm(robot) {
  return {
    displayName: robot?.displayName ?? "",
    modelName: robot?.modelName ?? ""
  };
}

export function shouldRefreshRobotEditForm({ nextRobotCode, previousRobotCode, robotModal }) {
  if (robotModal === "edit") {
    return false;
  }
  return previousRobotCode !== nextRobotCode;
}

export function isOnlineRobot(robot) {
  return onlineRobotStatuses.has(robot?.status);
}

export function groupRobotsByAvailability(robots = []) {
  const sortedRobots = sortRobotsForManagement(robots);
  return {
    offlineRobots: sortedRobots.filter((robot) => !isOnlineRobot(robot)),
    onlineRobots: sortedRobots.filter(isOnlineRobot)
  };
}

export function selectDefaultRobotForManagement(robots = []) {
  return sortRobotsForManagement(robots)[0] ?? null;
}

export function sortRobotsForManagement(robots = []) {
  return [...robots].sort((left, right) => {
    const leftOrder = robotStatusOrder[left.status] ?? 9;
    const rightOrder = robotStatusOrder[right.status] ?? 9;
    if (leftOrder !== rightOrder) {
      return leftOrder - rightOrder;
    }
    const leftSeenAt = left.lastSeenAt ?? "";
    const rightSeenAt = right.lastSeenAt ?? "";
    if (leftSeenAt !== rightSeenAt) {
      return rightSeenAt.localeCompare(leftSeenAt);
    }
    return String(left.robotCode ?? "").localeCompare(String(right.robotCode ?? ""));
  });
}

export function findRobotOpenMission(robotCode, missions = []) {
  if (!robotCode) {
    return null;
  }
  return missions.find((mission) => (
    ["ready", "active"].includes(mission.status)
    && getMissionRobotCodes(mission).includes(robotCode)
  )) ?? null;
}

export function makeRobotStatusTone(status) {
  if (["online", "streaming", "active"].includes(status)) {
    return "success";
  }
  if (["ready", "idle"].includes(status)) {
    return "info";
  }
  if (["fault", "error", "failed"].includes(status)) {
    return "danger";
  }
  return "neutral";
}
