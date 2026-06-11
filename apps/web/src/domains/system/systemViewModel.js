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
    topTables: createDatabaseTopTables(rawTables),
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

export function makeObjectStorageDisabledReason({ isProduction, recorderRuntimeStatus }) {
  if (isProduction) {
    return "운영 환경에서는 객체 스토리지 정리를 실행할 수 없습니다.";
  }
  if (!recorderRuntimeStatus || recorderRuntimeStatus.status !== "ok") {
    return "녹화 런타임 상태를 확인할 수 없어 객체 스토리지 정리를 실행할 수 없습니다.";
  }
  if (!recorderRuntimeStatus.clearable) {
    return makeRecorderRuntimeBlockingLabel(recorderRuntimeStatus.blockingReason);
  }
  return "";
}

export function createSystemClearActions({
  canClearEventData = false,
  canClearObjectStorage = false,
  canClearRecorderRuntime = false,
  canClearSensorData = false,
  canPruneObjectStorage = false,
  canPruneRecorderRuntime = false,
  clearingActionID = "",
  databaseUsage,
  isProduction = false,
  objectStorageUsage,
  recorderRuntimeStatus,
  statusReady = true
} = {}) {
  const sensorCategory = findDatabaseCategory(databaseUsage, "sensors");
  const eventCategory = findDatabaseCategory(databaseUsage, "events");
  const storageObjectsTable = findDatabaseTable(databaseUsage, "storage_objects");
  const statusDisabledReason = statusReady ? "" : "시스템 상태를 확인한 뒤 삭제를 실행할 수 있습니다.";
  const objectStorageDisabledReason = makeObjectStorageDisabledReason({ isProduction, recorderRuntimeStatus });
  const recorderRuntimeDisabledReason = makeRecorderRuntimeDisabledReason({ isProduction, recorderRuntimeStatus });
  const objectStoragePruneDisabledReason = isProduction ? "운영 환경에서는 객체 스토리지 운영 중 정리를 실행할 수 없습니다." : "";
  const recorderRuntimePruneDisabledReason = makeRecorderRuntimePruneDisabledReason({ isProduction, recorderRuntimeStatus });
  const productionSensorDisabledReason = isProduction ? "운영 환경에서는 센서 데이터 정리를 실행할 수 없습니다." : "";
  const productionEventDisabledReason = isProduction ? "운영 환경에서는 이벤트 데이터 정리를 실행할 수 없습니다." : "";

  return [
    createSystemClearAction({
      canRun: canClearObjectStorage,
      clearingActionID,
      description: "녹화 파일과 파일 상태 메타데이터를 정리합니다.",
      disabledReason: statusDisabledReason || objectStorageDisabledReason,
      id: "objectStorage",
      impact: "녹화 파일은 복구할 수 없습니다. 진행 중인 녹화가 있으면 실행하지 않습니다.",
      subject: "녹화 파일 저장소",
      targetMetrics: [
        { label: "삭제 대상", value: `${formatInteger(objectStorageUsage?.objectCount)}개 파일` },
        { label: "파일 용량", value: formatStorageByteCount(objectStorageUsage?.bucketUsedBytes) },
        { label: "파일 메타데이터", value: `${formatInteger(storageObjectsTable?.rowCount)}건` }
      ],
      title: "객체 스토리지 전체 삭제"
    }),
    createSystemClearAction({
      buttonLabel: "운영 중 정리",
      busyLabel: "정리 중",
      canRun: canPruneObjectStorage,
      clearingActionID,
      confirmLabel: "운영 중 정리",
      description: "진행 중인 녹화 파일은 제외하고 완료된 녹화 파일과 파일 상태 메타데이터를 정리합니다.",
      disabledReason: statusDisabledReason || objectStoragePruneDisabledReason,
      id: "objectStoragePrune",
      impact: "진행 중이거나 마무리 중인 녹화 파일은 삭제하지 않습니다.",
      subject: "완료된 녹화 파일 저장소",
      targetMetrics: [
        { label: "최대 후보", value: `${formatInteger(objectStorageUsage?.objectCount)}개 파일` },
        { label: "파일 용량", value: formatStorageByteCount(objectStorageUsage?.bucketUsedBytes) },
        { label: "파일 메타데이터", value: `${formatInteger(storageObjectsTable?.rowCount)}건` }
      ],
      title: "객체 스토리지 운영 중 정리"
    }),
    createSystemClearAction({
      canRun: canClearSensorData,
      clearingActionID,
      description: "센서 정의, 최신 센서값, 센서 샘플을 정리합니다.",
      disabledReason: statusDisabledReason || productionSensorDisabledReason,
      id: "sensorData",
      impact: "새 telemetry가 들어오면 센서 데이터는 다시 생성됩니다.",
      subject: "telemetry 저장 데이터",
      targetMetrics: [
        { label: "삭제 대상", value: `${formatInteger(sensorCategory?.rowCount)}건` },
        { label: "DB 사용량", value: formatStorageByteCount(sensorCategory?.totalBytes) },
        { label: "관련 테이블", value: `${formatInteger(sensorCategory?.tableCount)}개` }
      ],
      title: "센서 데이터 전체 삭제"
    }),
    createSystemClearAction({
      canRun: canClearEventData,
      clearingActionID,
      description: "저장된 임무 이벤트와 객체 탐지 이벤트를 정리합니다.",
      disabledReason: statusDisabledReason || productionEventDisabledReason,
      id: "eventData",
      impact: "새 event가 들어오면 이벤트 데이터는 다시 생성됩니다.",
      subject: "mission.event / detection.object",
      targetMetrics: [
        { label: "삭제 대상", value: `${formatInteger(eventCategory?.rowCount)}건` },
        { label: "DB 사용량", value: formatStorageByteCount(eventCategory?.totalBytes) },
        { label: "관련 테이블", value: `${formatInteger(eventCategory?.tableCount)}개` }
      ],
      title: "이벤트 데이터 전체 삭제"
    }),
    createSystemClearAction({
      canRun: canClearRecorderRuntime,
      clearingActionID,
      description: "녹화 서비스가 로컬에 임시로 만든 런타임 파일을 정리합니다.",
      disabledReason: statusDisabledReason || recorderRuntimeDisabledReason,
      id: "recorderRuntime",
      impact: "진행 중인 녹화 작업이 있으면 실행하지 않습니다. 저장 완료된 객체 스토리지 파일은 삭제하지 않습니다.",
      subject: "녹화 런타임 파일",
      targetMetrics: [
        { label: "삭제 대상", value: `${formatInteger(recorderRuntimeStatus?.files)}개 파일` },
        { label: "파일 용량", value: formatStorageByteCount(recorderRuntimeStatus?.usedBytes) },
        { label: "청크 디렉터리", value: `${formatInteger(recorderRuntimeStatus?.recordingDirectories)}개` }
      ],
      title: "녹화 런타임 파일 전체 삭제"
    }),
    createSystemClearAction({
      buttonLabel: "운영 중 정리",
      busyLabel: "정리 중",
      canRun: canPruneRecorderRuntime,
      clearingActionID,
      confirmLabel: "운영 중 정리",
      description: "작성 중인 chunk와 마무리 대기 chunk는 제외하고 오래된 로컬 임시 파일만 정리합니다.",
      disabledReason: statusDisabledReason || recorderRuntimePruneDisabledReason,
      id: "recorderRuntimePrune",
      impact: "현재 녹화 중인 chunk 디렉터리는 삭제하지 않습니다. 저장 완료된 객체 스토리지 파일도 삭제하지 않습니다.",
      subject: "오래된 녹화 런타임 파일",
      targetMetrics: [
        { label: "최대 후보", value: `${formatInteger(recorderRuntimeStatus?.files)}개 파일` },
        { label: "파일 용량", value: formatStorageByteCount(recorderRuntimeStatus?.usedBytes) },
        { label: "청크 디렉터리", value: `${formatInteger(recorderRuntimeStatus?.recordingDirectories)}개` }
      ],
      title: "녹화 런타임 운영 중 정리"
    })
  ];
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

export function makeRecorderRuntimePruneDisabledReason({ isProduction, recorderRuntimeStatus }) {
  if (isProduction) {
    return "운영 환경에서는 녹화 런타임 운영 중 정리를 실행할 수 없습니다.";
  }
  if (!recorderRuntimeStatus || recorderRuntimeStatus.status !== "ok") {
    return "녹화 런타임 상태를 확인할 수 없어 운영 중 정리를 실행할 수 없습니다.";
  }
  return "";
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

export function createDatabaseTopTables(tables, limit = 5) {
  return tables
    .filter((table) => !isInternalDatabaseTable(table.tableName))
    .filter((table) => table.totalBytes > 0 || table.rowCount > 0)
    .sort((left, right) => {
      if (left.totalBytes !== right.totalBytes) {
        return right.totalBytes - left.totalBytes;
      }
      return right.rowCount - left.rowCount;
    })
    .slice(0, limit)
    .map((table) => ({
      ...table,
      label: databaseTableLabels[table.tableName] ?? table.tableName
    }));
}

function createSystemClearAction({
  buttonLabel = "전체 삭제",
  busyLabel = "삭제 중",
  canRun,
  clearingActionID,
  confirmLabel = buttonLabel,
  description,
  disabledReason,
  id,
  impact,
  subject,
  targetMetrics,
  title
}) {
  const handlerDisabledReason = canRun ? "" : "정리 기능이 현재 화면에 연결되지 않았습니다.";
  const resolvedDisabledReason = disabledReason || handlerDisabledReason;
  return {
    buttonLabel,
    busy: clearingActionID === id,
    busyLabel,
    confirmLabel,
    description,
    disabled: Boolean(resolvedDisabledReason) || clearingActionID === id,
    disabledReason: resolvedDisabledReason,
    id,
    impact,
    subject,
    targetMetrics,
    title
  };
}

function findDatabaseCategory(databaseUsage, categoryID) {
  return (databaseUsage?.categories ?? []).find((category) => category.id === categoryID);
}

function findDatabaseTable(databaseUsage, tableName) {
  return (databaseUsage?.tables ?? []).find((table) => table.tableName === tableName);
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

const databaseTableLabels = {
  browser_sessions: "브라우저 세션",
  control_acks: "제어 응답",
  control_commands: "제어 명령",
  events: "이벤트 로그",
  mission_robots: "임무 로봇 할당",
  missions: "임무",
  recorder_sessions: "녹화 연결 세션",
  recording_chunks: "녹화 청크",
  recording_finalization_jobs: "녹화 마무리 작업",
  recording_sessions: "녹화 세션",
  robot_sessions: "로봇 세션",
  robot_stream_sessions: "로봇 송출 세션",
  robot_tokens: "로봇 토큰",
  robots: "로봇",
  sensor_descriptors: "센서 정의",
  sensor_latest_samples: "최신 센서값",
  sensor_samples: "센서 샘플",
  storage_objects: "녹화 파일 메타데이터",
  users: "사용자"
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
