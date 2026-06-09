export function createRoomPeerSummaries(room) {
  const summaries = [];
  const seen = new Set();
  const addPeer = (peer) => {
    const key = makePeerSummaryKey(peer);
    if (!key || seen.has(key)) {
      return;
    }
    seen.add(key);
    summaries.push(peer);
  };

  (room?.publishers ?? []).forEach((publisher) => {
    if (publisher?.robotCode) {
      addPeer({
        peerId: publisher.publisherPeerId,
        role: "robot",
        robotCode: publisher.robotCode
      });
    }
  });
  (room?.peers ?? []).forEach((peer) => {
    if (peer?.role === "robot") {
      addPeer(peer);
      return;
    }
    if (peer?.role === "operator" || peer?.role === "recorder") {
      addPeer(peer);
    }
  });
  return summaries;
}

export function countRoomRobotPublishers(room) {
  const publisherRobotCodes = new Set(
    (room?.publishers ?? [])
      .map((publisher) => publisher?.robotCode)
      .filter(Boolean)
  );
  if (publisherRobotCodes.size > 0) {
    return publisherRobotCodes.size;
  }
  return new Set(
    (room?.peers ?? [])
      .filter((peer) => peer?.role === "robot" && peer.robotCode)
      .map((peer) => peer.robotCode)
  ).size;
}

export function countRoomPublishedTracks(room) {
  const publisherTrackCount = (room?.publishers ?? []).reduce((sum, publisher) => (
    sum + Math.max(0, Number(publisher?.trackCount) || 0)
  ), 0);
  if (publisherTrackCount > 0) {
    return publisherTrackCount;
  }
  return room?.publishedTracks?.length ?? 0;
}

export function makeRoomStreamingState(room) {
  const isPublishing = (room?.publishers ?? []).some((publisher) => (
    publisher?.state === "publishing" && Math.max(0, Number(publisher?.trackCount) || 0) > 0
  ));
  if (isPublishing) {
    return "송출 중";
  }
  if ((room?.publishers ?? []).length > 0 || countRoomRobotPublishers(room) > 0) {
    return "연결됨";
  }
  return "대기";
}

export function normalizeObjectStorageUsage(rawUsage) {
  if (!rawUsage) {
    return null;
  }
  const totalBytes = readStorageNumber(rawUsage.totalBytes);
  const usedBytes = readStorageNumber(rawUsage.usedBytes);
  const rawPercent = readStorageNumber(rawUsage.usedPercent);
  const calculatedPercent = totalBytes > 0 ? (usedBytes / totalBytes) * 100 : 0;
  return {
    availableBytes: readStorageNumber(rawUsage.availableBytes),
    bucket: rawUsage.bucket ?? "",
    bucketUsedBytes: readStorageNumber(rawUsage.bucketUsedBytes),
    objectCount: Math.max(0, Math.round(readStorageNumber(rawUsage.objectCount))),
    status: rawUsage.status ?? "unavailable",
    totalBytes,
    usedBytes,
    usedPercent: Number.isFinite(rawPercent) && rawPercent > 0 ? rawPercent : calculatedPercent
  };
}

export function normalizeDatabaseUsage(rawUsage) {
  if (!rawUsage) {
    return null;
  }
  const rawTables = (rawUsage.tables ?? []).map((table) => ({
    rowCount: Math.max(0, Math.round(readStorageNumber(table.rowCount))),
    tableName: table.tableName ?? "",
    totalBytes: readStorageNumber(table.totalBytes)
  }));
  return {
    categories: createDatabaseUsageCategories(rawTables),
    databaseName: rawUsage.databaseName ?? "",
    databaseSizeBytes: readStorageNumber(rawUsage.databaseSizeBytes),
    status: rawUsage.status ?? "unavailable",
    tables: rawTables,
    trackedTableBytes: readStorageNumber(rawUsage.trackedTableBytes)
  };
}

export function normalizeRecorderRuntimeStatus(rawStatus) {
  if (!rawStatus) {
    return null;
  }
  const totalBytes = readStorageNumber(rawStatus.totalBytes);
  const usedBytes = readStorageNumber(rawStatus.usedBytes);
  const rawPercent = readStorageNumber(rawStatus.usedPercent);
  const calculatedPercent = totalBytes > 0 ? (usedBytes / totalBytes) * 100 : 0;
  return {
    availableBytes: readStorageNumber(rawStatus.availableBytes),
    blockingReason: rawStatus.blockingReason ?? "",
    clearable: Boolean(rawStatus.clearable),
    files: Math.max(0, Math.round(readStorageNumber(rawStatus.files))),
    recordingDirectories: Math.max(0, Math.round(readStorageNumber(rawStatus.recordingDirectories))),
    status: rawStatus.status ?? "unavailable",
    totalBytes,
    usedBytes,
    usedPercent: Number.isFinite(rawPercent) && rawPercent > 0 ? rawPercent : calculatedPercent
  };
}

export function makeRecorderRuntimeDisabledReason({ isProduction, recorderRuntimeStatus }) {
  if (isProduction) {
    return "운영 환경에서는 녹화 런타임 파일 정리를 실행할 수 없습니다.";
  }
  if (!recorderRuntimeStatus || recorderRuntimeStatus.status !== "ok") {
    return "녹화 런타임 상태를 확인할 수 없어 정리를 실행할 수 없습니다.";
  }
  if (!recorderRuntimeStatus.clearable) {
    return makeRecorderRuntimeBlockingLabel(recorderRuntimeStatus.blockingReason);
  }
  return "";
}

export function makeRecorderRuntimeBlockingLabel(reason) {
  const labels = {
    "active audio writer": "녹화 오디오 파일 작성 중이라 정리를 실행할 수 없습니다.",
    "active recording chunk": "녹화 청크가 작성 중이라 정리를 실행할 수 없습니다.",
    "active recording target": "진행 중인 녹화 대상이 있어 정리를 실행할 수 없습니다.",
    "pending finalization": "녹화 파일 마무리 작업이 남아 있어 정리를 실행할 수 없습니다.",
    "production environment": "운영 환경에서는 녹화 런타임 파일 정리를 실행할 수 없습니다."
  };
  return labels[reason] ?? "녹화 런타임 파일 정리를 지금 실행할 수 없습니다.";
}

export function makeSystemStatusLabel(status) {
  const labels = {
    configured: "설정됨",
    degraded: "점검 필요",
    error: "오류",
    failed: "실패",
    ok: "정상",
    ready: "준비"
  };
  return labels[status] ?? status;
}

export function makeSystemStatusTone(status) {
  if (["ok", "ready", "configured"].includes(status)) {
    return "success";
  }
  if (["degraded", "warning"].includes(status)) {
    return "warning";
  }
  if (["error", "failed"].includes(status)) {
    return "danger";
  }
  return "neutral";
}

export function makePeerLabel(peer) {
  if (peer.role === "robot") {
    return peer.robotCode ? `로봇 ${peer.robotCode}` : "로봇";
  }
  if (peer.role === "operator") {
    return peer.selectedRobotCode ? `관제 ${peer.selectedRobotCode}` : "관제";
  }
  if (peer.role === "recorder") {
    return "녹화";
  }
  return peer.role;
}

export function makeStorageChartColor(percent) {
  if (percent >= 90) {
    return "#fb7185";
  }
  if (percent >= 75) {
    return "#fbbf24";
  }
  return "#38bdf8";
}

export function formatStoragePercent(value) {
  const percent = clampStoragePercent(value);
  if (percent === 0 || percent >= 10) {
    return `${percent.toFixed(0)}%`;
  }
  return `${percent.toFixed(1)}%`;
}

export function formatStorageByteCount(value) {
  const bytes = Math.max(0, readStorageNumber(value));
  if (bytes === 0) {
    return "0 B";
  }
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  let amount = bytes;
  let unitIndex = 0;
  while (amount >= 1024 && unitIndex < units.length - 1) {
    amount /= 1024;
    unitIndex += 1;
  }
  const digits = unitIndex === 0 || amount >= 100 ? 0 : amount >= 10 ? 1 : 2;
  return `${amount.toFixed(digits)} ${units[unitIndex]}`;
}

export function formatInteger(value) {
  return Math.max(0, Math.round(readStorageNumber(value))).toLocaleString();
}

export function createDatabaseUsageCategories(tables) {
  const categoriesByID = new Map();
  for (const table of tables) {
    const category = resolveDatabaseUsageCategory(table.tableName);
    const current = categoriesByID.get(category.id) ?? {
      ...category,
      rowCount: 0,
      tableCount: 0,
      totalBytes: 0
    };
    current.rowCount += table.rowCount;
    current.tableCount += 1;
    current.totalBytes += table.totalBytes;
    categoriesByID.set(category.id, current);
  }

  return Array.from(categoriesByID.values())
    .filter((category) => category.totalBytes > 0 || category.rowCount > 0)
    .sort((left, right) => {
      if (left.id === "internal" || right.id === "internal") {
        return left.id === "internal" ? 1 : -1;
      }
      if (left.totalBytes !== right.totalBytes) {
        return right.totalBytes - left.totalBytes;
      }
      return left.sortOrder - right.sortOrder;
    });
}

function resolveDatabaseUsageCategory(tableName) {
  const categoryID = databaseTableCategoryIDs[tableName] ?? (
    isInternalDatabaseTable(tableName) ? "internal" : "other"
  );
  return databaseUsageCategories[categoryID];
}

function isInternalDatabaseTable(tableName) {
  return tableName === "spatial_ref_sys"
    || tableName === "geography_columns"
    || tableName === "geometry_columns"
    || tableName.endsWith("_migrations")
    || tableName.startsWith("schema_");
}

const databaseUsageCategories = {
  control: {
    id: "control",
    label: "제어 데이터",
    sortOrder: 50
  },
  events: {
    id: "events",
    label: "이벤트 데이터",
    sortOrder: 20
  },
  internal: {
    id: "internal",
    label: "시스템 내부 데이터",
    sortOrder: 90
  },
  operations: {
    id: "operations",
    label: "로봇/임무 운영 데이터",
    sortOrder: 40
  },
  other: {
    id: "other",
    label: "기타 관제 데이터",
    sortOrder: 80
  },
  recordings: {
    id: "recordings",
    label: "녹화 데이터",
    sortOrder: 30
  },
  sensors: {
    id: "sensors",
    label: "센서 데이터",
    sortOrder: 10
  }
};

const databaseTableCategoryIDs = {
  browser_sessions: "operations",
  control_acks: "control",
  control_commands: "control",
  events: "events",
  mission_robots: "operations",
  missions: "operations",
  recorder_sessions: "operations",
  recording_chunks: "recordings",
  recording_finalization_jobs: "recordings",
  recording_sessions: "recordings",
  robot_sessions: "operations",
  robot_stream_sessions: "operations",
  robot_tokens: "operations",
  robots: "operations",
  sensor_descriptors: "sensors",
  sensor_latest_samples: "sensors",
  sensor_samples: "sensors",
  storage_objects: "recordings",
  users: "operations"
};

function makePeerSummaryKey(peer) {
  if (!peer?.role) {
    return "";
  }
  if (peer.role === "robot") {
    return peer.robotCode ? `robot:${peer.robotCode}` : `robot-peer:${peer.peerId}`;
  }
  if (peer.role === "operator") {
    return peer.selectedRobotCode ? `operator:${peer.selectedRobotCode}` : `operator-peer:${peer.peerId}`;
  }
  if (peer.role === "recorder") {
    return "recorder";
  }
  return `${peer.role}:${peer.peerId ?? ""}`;
}

function clampStoragePercent(value) {
  const numberValue = Number(value);
  if (!Number.isFinite(numberValue)) {
    return 0;
  }
  return Math.min(100, Math.max(0, numberValue));
}

function readStorageNumber(value) {
  const numberValue = Number(value);
  if (!Number.isFinite(numberValue)) {
    return 0;
  }
  return numberValue;
}
