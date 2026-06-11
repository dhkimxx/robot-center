export const replayChunkPageSize = 80;
export const replayAutoRefreshIntervalMs = 10000;

export function createEmptyReplaySummaryState() {
  return {
    error: "",
    loadedAt: "",
    refreshing: false,
    status: "idle",
    summary: null
  };
}

export function createEmptyReplayChunkState() {
  return {
    chunks: [],
    error: "",
    loadedAt: "",
    page: null,
    refreshing: false,
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

export function hasPendingReplayChunks(robotSummary) {
  return (robotSummary?.recordingChunkCount ?? 0) > 0
    || (robotSummary?.finalizingChunkCount ?? 0) > 0;
}

export function shouldAutoRefreshReplayRecordings({ missionStatus, selectedRobotSummary } = {}) {
  return missionStatus === "active" || hasPendingReplayChunks(selectedRobotSummary);
}

export function makeReplayContinuityNotice(robotSummary) {
  if ((robotSummary?.recordingChunkCount ?? 0) > 0) {
    return "현재 녹화 중인 청크는 저장 완료 후 재생 버튼이 표시됩니다.";
  }
  if ((robotSummary?.finalizingChunkCount ?? 0) > 0) {
    return "저장 마무리 중인 청크는 업로드가 끝나면 재생할 수 있습니다.";
  }
  if ((robotSummary?.partialChunkCount ?? 0) > 0) {
    return "부분 저장 청크는 저장된 파일만 재생할 수 있습니다.";
  }
  return "";
}

export function makeReplayRefreshStatusLabel({ autoRefreshEnabled, loadedAt, refreshing } = {}) {
  if (refreshing) {
    return "갱신 중";
  }
  if (loadedAt) {
    return autoRefreshEnabled ? "자동 갱신 중" : "최근 갱신됨";
  }
  return autoRefreshEnabled ? "자동 갱신 대기" : "수동 갱신";
}
