export function createInitialMissionForm() {
  return {
    name: "P0 통합 시연",
    missionType: "mountain_rescue",
    robotCodes: [],
    siteNote: "관제/영상/센서/녹화 흐름 검증",
    robotCode: ""
  };
}

export function getMissionRobotCodes(mission) {
  if (!mission) {
    return [];
  }

  const codes = new Set();
  if (Array.isArray(mission.robotCodes)) {
    mission.robotCodes.forEach((robotCode) => {
      if (robotCode) {
        codes.add(robotCode);
      }
    });
  }
  if (Array.isArray(mission.robots)) {
    mission.robots.forEach((robot) => {
      const robotCode = typeof robot === "string" ? robot : robot?.robotCode;
      if (robotCode) {
        codes.add(robotCode);
      }
    });
  }
  if (mission.robotCode) {
    codes.add(mission.robotCode);
  }
  return Array.from(codes);
}

export function makeMissionRoomId(mission) {
  return mission?.roomId ?? mission?.missionCode ?? "";
}

export function makeMissionRobotKey(missionCode, robotCode) {
  return `${missionCode}:${robotCode}`;
}

export function makeMissionConnectionKey(missionCode) {
  return `mission:${missionCode}`;
}

export function getMissionCodeFromRobotKey(targetKey) {
  return String(targetKey ?? "").split(":")[0] ?? "";
}

const closedMissionStatuses = new Set(["completed", "ended", "cancelled"]);
const missionStatusOrder = { active: 0, ready: 1, completed: 2, ended: 2, cancelled: 3 };

export function isClosedMission(mission) {
  return closedMissionStatuses.has(mission?.status);
}

export function sortMissionsByLifecycle(missions) {
  return [...missions].sort((left, right) => {
    const leftOrder = missionStatusOrder[left.status] ?? 9;
    const rightOrder = missionStatusOrder[right.status] ?? 9;
    if (leftOrder !== rightOrder) {
      return leftOrder - rightOrder;
    }
    return (right.startedAt ?? right.createdAt ?? "").localeCompare(left.startedAt ?? left.createdAt ?? "");
  });
}

export function groupMissionsByLifecycle(missions) {
  const orderedMissions = sortMissionsByLifecycle(missions);
  return {
    openMissions: orderedMissions.filter((mission) => !isClosedMission(mission)),
    closedMissions: orderedMissions.filter(isClosedMission)
  };
}

export function makeMissionRobotLiveLabel(mission, isStreaming) {
  if (isStreaming) {
    return "송출 중";
  }
  if (isClosedMission(mission)) {
    return "임무 종료";
  }
  if (mission?.status === "active") {
    return "송출 대기";
  }
  return "배정 대기";
}

function findMissionLiveStatusRobot(liveStatus, robotCode) {
  return liveStatus?.robots?.find((robot) => robot.robotCode === robotCode) ?? null;
}

function isStreamingFromLiveStatus(liveStatusRobot) {
  return liveStatusRobot?.stream?.state === "streaming";
}

export function countStreamingRobotsFromLiveStatuses(liveStatuses = {}) {
  return Object.values(liveStatuses).reduce((count, liveStatus) => {
    if (liveStatus?.missionStatus !== "active") {
      return count;
    }
    return count + (liveStatus.robots ?? []).filter((robot) => robot.stream?.state === "streaming").length;
  }, 0);
}

export function createMissionRobotTargets(mission, robots, liveStatus = null) {
  if (!mission) {
    return [];
  }
  const robotCodes = new Set();
  getMissionRobotCodes(mission).forEach((robotCode) => robotCodes.add(robotCode));
  (liveStatus?.robots ?? []).forEach((robot) => {
    if (robot.robotCode) {
      robotCodes.add(robot.robotCode);
    }
  });

  return Array.from(robotCodes).map((robotCode) => {
    const liveStatusRobot = findMissionLiveStatusRobot(liveStatus, robotCode);
    const isStreaming = isStreamingFromLiveStatus(liveStatusRobot);
    return {
      isStreaming,
      key: makeMissionRobotKey(mission.missionCode, robotCode),
      liveLabel: makeMissionRobotLiveLabel(mission, isStreaming),
      mission,
      missionRoomId: makeMissionRoomId(mission),
      robot: robots.find((robot) => robot.robotCode === robotCode) ?? null,
      robotCode,
      roomId: makeMissionRoomId(mission),
      liveStatus: liveStatusRobot
    };
  });
}

export function getMissionRobotDetails(mission, robots, liveStatus = null) {
  return getMissionRobotCodes(mission).map((robotCode) => {
    const robot = robots.find((candidate) => candidate.robotCode === robotCode) ?? null;
    const liveStatusRobot = findMissionLiveStatusRobot(liveStatus, robotCode);
    const isStreaming = isStreamingFromLiveStatus(liveStatusRobot);
    return {
      deviceStatus: robot?.status ?? "offline",
      displayName: robot?.displayName ?? robotCode,
      isStreaming,
      liveLabel: makeMissionRobotLiveLabel(mission, isStreaming),
      liveStatus: liveStatusRobot,
      robotCode
    };
  });
}

export function getBusyRobotReasonForMissionCreate(robotCode, missions = []) {
  const activeMission = missions.find((mission) => mission.status === "active" && getMissionRobotCodes(mission).includes(robotCode));
  if (activeMission) {
    return `진행 중 임무 ${activeMission.missionCode}`;
  }
  return "";
}

export function formatMissionRobotCount(robotDetails) {
  if (robotDetails.length === 0) {
    return "미배정";
  }
  if (robotDetails.length === 1) {
    return robotDetails[0].robotCode;
  }
  return `로봇 ${robotDetails.length}대`;
}
