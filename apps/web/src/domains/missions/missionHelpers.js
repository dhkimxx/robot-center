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

const streamingStatusFreshMs = 30_000;
const closedMissionStatuses = new Set(["completed", "ended", "cancelled"]);

export function isClosedMission(mission) {
  return closedMissionStatuses.has(mission?.status);
}

export function isFreshMissionStreamingStatus(mission, robotCode, streamingStatus, nowMs = Date.now()) {
  if (!mission || mission.status !== "active" || !streamingStatus) {
    return false;
  }
  if (streamingStatus.missionId !== mission.id || streamingStatus.robotCode !== robotCode) {
    return false;
  }
  if (streamingStatus.status !== "streaming" && streamingStatus.status !== "publishing") {
    return false;
  }
  const sentAtMs = Date.parse(streamingStatus.sentAt ?? "");
  return Number.isFinite(sentAtMs) && nowMs - sentAtMs <= streamingStatusFreshMs;
}

export function findMissionStreamingStatus(mission, robotCode, streamingStatuses) {
  return streamingStatuses.find((status) => status.missionId === mission?.id && status.robotCode === robotCode) ?? null;
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

export function createMissionRobotTargets(mission, robots, streamingStatuses) {
  if (!mission) {
    return [];
  }
  const statusesForMission = streamingStatuses.filter((status) => status.missionId === mission.id);
  const robotCodes = new Set();
  getMissionRobotCodes(mission).forEach((robotCode) => robotCodes.add(robotCode));
  statusesForMission.forEach((status) => {
    if (status.robotCode) {
      robotCodes.add(status.robotCode);
    }
  });

  return Array.from(robotCodes).map((robotCode) => {
    const streamingStatus = findMissionStreamingStatus(mission, robotCode, statusesForMission);
    const isStreaming = isFreshMissionStreamingStatus(mission, robotCode, streamingStatus);
    return {
      isStreaming,
      key: makeMissionRobotKey(mission.missionCode, robotCode),
      liveLabel: makeMissionRobotLiveLabel(mission, isStreaming),
      mission,
      missionRoomId: makeMissionRoomId(mission),
      robot: robots.find((robot) => robot.robotCode === robotCode) ?? null,
      robotCode,
      roomId: makeMissionRoomId(mission),
      streamingStatus
    };
  });
}

export function getMissionRobotDetails(mission, robots, streamingStatuses = []) {
  return getMissionRobotCodes(mission).map((robotCode) => {
    const robot = robots.find((candidate) => candidate.robotCode === robotCode) ?? null;
    const streamingStatus = findMissionStreamingStatus(mission, robotCode, streamingStatuses);
    const isStreaming = isFreshMissionStreamingStatus(mission, robotCode, streamingStatus);
    return {
      deviceStatus: robot?.status ?? "offline",
      displayName: robot?.displayName ?? robotCode,
      isStreaming,
      liveLabel: makeMissionRobotLiveLabel(mission, isStreaming),
      robotCode,
      streamingStatus
    };
  });
}

export function getBusyRobotReasonForMissionCreate(robotCode, missions = [], streamingStatuses = [], nowMs = Date.now()) {
  const activeMission = missions.find((mission) => mission.status === "active" && getMissionRobotCodes(mission).includes(robotCode));
  if (activeMission) {
    return `진행 중 임무 ${activeMission.missionCode}`;
  }

  const freshStreamingStatus = streamingStatuses.find((status) => {
    if (status.robotCode !== robotCode || (status.status !== "streaming" && status.status !== "publishing")) {
      return false;
    }
    const sentAtMs = Date.parse(status.sentAt ?? "");
    return Number.isFinite(sentAtMs) && nowMs - sentAtMs <= streamingStatusFreshMs;
  });
  if (freshStreamingStatus) {
    return "실시간 송출 중";
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
