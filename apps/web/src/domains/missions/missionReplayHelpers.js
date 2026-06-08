export const replayChunkPageSize = 80;

export function createEmptyReplaySummaryState() {
  return {
    error: "",
    status: "idle",
    summary: null
  };
}

export function createEmptyReplayChunkState() {
  return {
    chunks: [],
    error: "",
    page: null,
    robotCode: "",
    status: "idle"
  };
}

export function createRobotDisplayNamesByCode(robots) {
  return new Map(
    robots
      .filter((robot) => robot.robotCode)
      .map((robot) => [robot.robotCode, robot.displayName || robot.robotCode])
  );
}

export function getRobotDisplayName(robotDisplayNamesByCode, robotCode) {
  if (!robotCode) {
    return "";
  }
  return robotDisplayNamesByCode.get(robotCode) ?? robotCode;
}

export function makeRobotReplayMeta(robotSummary, displayName) {
  const countLabel = `${robotSummary.chunkCount}개 청크`;
  return displayName && displayName !== robotSummary.robotCode
    ? `${robotSummary.robotCode} · ${countLabel}`
    : countLabel;
}

export function makeRobotSummaryTone(robotSummary) {
  if (robotSummary.partialChunkCount > 0) {
    return "warning";
  }
  if (robotSummary.recordingChunkCount > 0 || robotSummary.finalizingChunkCount > 0) {
    return "info";
  }
  if (robotSummary.uploadedChunkCount > 0) {
    return "success";
  }
  return "neutral";
}

export function makeRobotSummaryStatusLabel(robotSummary) {
  if (robotSummary.partialChunkCount > 0) {
    return "부분 저장";
  }
  if (robotSummary.recordingChunkCount > 0) {
    return "녹화 중";
  }
  if (robotSummary.finalizingChunkCount > 0) {
    return "저장 마무리";
  }
  if (robotSummary.uploadedChunkCount > 0) {
    return "저장 완료";
  }
  return "대기";
}

export function makeFileAvailabilityLabel(robotSummary, fileType, label) {
  const availableCount = robotSummary.availableFileCounts?.[fileType] ?? 0;
  return `${label} ${availableCount}/${robotSummary.chunkCount}`;
}

export function makeLoadedChunkLabel(loadedCount, totalCount) {
  return `${loadedCount}/${totalCount}개 로드`;
}
