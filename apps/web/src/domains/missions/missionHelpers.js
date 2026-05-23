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
    const streamingStatus = statusesForMission.find((status) => status.robotCode === robotCode) ?? null;
    return {
      key: makeMissionRobotKey(mission.missionCode, robotCode),
      mission,
      missionRoomId: makeMissionRoomId(mission),
      robot: robots.find((robot) => robot.robotCode === robotCode) ?? null,
      robotCode,
      roomId: streamingStatus?.roomId || makeMissionRoomId(mission),
      streamingStatus
    };
  });
}

export function getMissionRobotDetails(mission, robots) {
  return getMissionRobotCodes(mission).map((robotCode) => {
    const robot = robots.find((candidate) => candidate.robotCode === robotCode) ?? null;
    return {
      displayName: robot?.displayName ?? robotCode,
      robotCode,
      status: robot?.status ?? "offline"
    };
  });
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
