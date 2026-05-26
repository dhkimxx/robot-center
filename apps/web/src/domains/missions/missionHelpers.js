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

const observedStreamFreshMs = 30_000;
const closedMissionStatuses = new Set(["completed", "ended", "cancelled"]);
const missionStatusOrder = { active: 0, ready: 1, completed: 2, ended: 2, cancelled: 3 };
const inactiveObservedIceStates = new Set(["failed", "disconnected", "closed"]);

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

function observedRooms(observedStreams = []) {
  return Array.isArray(observedStreams) ? observedStreams : observedStreams?.rooms ?? [];
}

export function findMissionObservedPublisher(mission, robotCode, observedStreams = []) {
  const roomId = makeMissionRoomId(mission);
  if (!roomId || !robotCode) {
    return null;
  }
  const room = observedRooms(observedStreams).find((candidate) => candidate.roomId === roomId);
  return room?.publishers?.find((publisher) => publisher.robotCode === robotCode) ?? null;
}

export function isFreshMissionObservedPublisher(mission, robotCode, observedPublisher, nowMs = Date.now()) {
  if (!mission || mission.status !== "active" || !observedPublisher) {
    return false;
  }
  if (observedPublisher.robotCode !== robotCode || inactiveObservedIceStates.has(observedPublisher.iceState)) {
    return false;
  }
  const freshAtMs = Date.parse(
    observedPublisher.lastTrackAt
      ?? observedPublisher.lastDataAt
      ?? ""
  );
  return Number.isFinite(freshAtMs) && Math.abs(nowMs - freshAtMs) <= observedStreamFreshMs;
}

export function countFreshObservedPublishers(observedStreams = [], nowMs = Date.now()) {
  return observedRooms(observedStreams).reduce((count, room) => {
    const publishers = room.publishers ?? [];
    return count + publishers.filter((publisher) => {
      if (inactiveObservedIceStates.has(publisher.iceState)) {
        return false;
      }
      const freshAtMs = Date.parse(publisher.lastTrackAt ?? publisher.lastDataAt ?? "");
      return Number.isFinite(freshAtMs) && Math.abs(nowMs - freshAtMs) <= observedStreamFreshMs;
    }).length;
  }, 0);
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

export function createMissionRobotTargets(mission, robots, observedStreams = [], liveStatus = null) {
  if (!mission) {
    return [];
  }
  const observedPublishersForMission = observedRooms(observedStreams)
    .find((room) => room.roomId === makeMissionRoomId(mission))
    ?.publishers ?? [];
  const robotCodes = new Set();
  getMissionRobotCodes(mission).forEach((robotCode) => robotCodes.add(robotCode));
  observedPublishersForMission.forEach((publisher) => {
    if (publisher.robotCode) {
      robotCodes.add(publisher.robotCode);
    }
  });

  return Array.from(robotCodes).map((robotCode) => {
    const observedPublisher = findMissionObservedPublisher(mission, robotCode, observedStreams);
    const liveStatusRobot = findMissionLiveStatusRobot(liveStatus, robotCode);
    const isStreaming = liveStatusRobot
      ? isStreamingFromLiveStatus(liveStatusRobot)
      : isFreshMissionObservedPublisher(mission, robotCode, observedPublisher);
    return {
      isStreaming,
      key: makeMissionRobotKey(mission.missionCode, robotCode),
      liveLabel: makeMissionRobotLiveLabel(mission, isStreaming),
      mission,
      missionRoomId: makeMissionRoomId(mission),
      robot: robots.find((robot) => robot.robotCode === robotCode) ?? null,
      robotCode,
      roomId: makeMissionRoomId(mission),
      observedPublisher,
      liveStatus: liveStatusRobot
    };
  });
}

export function getMissionRobotDetails(mission, robots, observedStreams = []) {
  return getMissionRobotCodes(mission).map((robotCode) => {
    const robot = robots.find((candidate) => candidate.robotCode === robotCode) ?? null;
    const observedPublisher = findMissionObservedPublisher(mission, robotCode, observedStreams);
    const isStreaming = isFreshMissionObservedPublisher(mission, robotCode, observedPublisher);
    return {
      deviceStatus: robot?.status ?? "offline",
      displayName: robot?.displayName ?? robotCode,
      isStreaming,
      liveLabel: makeMissionRobotLiveLabel(mission, isStreaming),
      observedPublisher,
      robotCode
    };
  });
}

export function getBusyRobotReasonForMissionCreate(robotCode, missions = [], nowMs = Date.now(), observedStreams = []) {
  const activeMission = missions.find((mission) => mission.status === "active" && getMissionRobotCodes(mission).includes(robotCode));
  if (activeMission) {
    return `진행 중 임무 ${activeMission.missionCode}`;
  }

  const freshObservedPublisher = countFreshObservedPublishers(
    observedRooms(observedStreams)
      .map((room) => ({
        ...room,
        publishers: (room.publishers ?? []).filter((publisher) => publisher.robotCode === robotCode)
      })),
    nowMs
  ) > 0;
  if (freshObservedPublisher) {
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
