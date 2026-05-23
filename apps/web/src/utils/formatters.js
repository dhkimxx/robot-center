import { missionTypes } from "../config/controlCenterConfig.js";

const freshTelemetryThresholdMs = 15_000;

export function formatDateTime(value) {
  if (!value) {
    return "-";
  }
  return new Intl.DateTimeFormat("ko-KR", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit"
  }).format(new Date(value));
}

export function formatNumber(value, digits = 1) {
  if (value === null || value === undefined || Number.isNaN(Number(value))) {
    return "-";
  }
  return Number(value).toFixed(digits);
}

export function formatElapsedTime(value, now = Date.now()) {
  if (!value) {
    return "-";
  }
  const timestamp = new Date(value).getTime();
  if (Number.isNaN(timestamp)) {
    return "-";
  }
  const elapsedSeconds = Math.max(0, Math.floor((now - timestamp) / 1000));
  if (elapsedSeconds < 60) {
    return `${elapsedSeconds}초 전`;
  }
  const elapsedMinutes = Math.floor(elapsedSeconds / 60);
  if (elapsedMinutes < 60) {
    return `${elapsedMinutes}분 전`;
  }
  return `${Math.floor(elapsedMinutes / 60)}시간 전`;
}

export function getTelemetryPositionState(telemetry, now = Date.now()) {
  const position = telemetry?.payload?.position ?? telemetry?.rawPayload?.payload?.position ?? telemetry;
  const latitude = position?.latitude;
  const longitude = position?.longitude;
  const hasPosition = latitude !== null
    && latitude !== undefined
    && longitude !== null
    && longitude !== undefined
    && !Number.isNaN(Number(latitude))
    && !Number.isNaN(Number(longitude));
  const timestamp = telemetry?.sentAt ?? telemetry?.receivedAt ?? telemetry?.rawPayload?.sentAt ?? position?.fixTime;
  const timestampMs = timestamp ? new Date(timestamp).getTime() : Number.NaN;
  const ageMs = Number.isNaN(timestampMs) ? null : Math.max(0, now - timestampMs);
  const isFresh = hasPosition && ageMs !== null && ageMs <= freshTelemetryThresholdMs;

  return {
    accuracyMeter: position?.accuracyMeter ?? telemetry?.accuracyMeter,
    ageMs,
    hasPosition,
    isFresh,
    latitude,
    longitude,
    provider: position?.provider ?? telemetry?.rawPayload?.payload?.position?.provider ?? "-",
    statusLabel: !hasPosition ? "GPS 대기" : isFresh ? "현재 위치" : "마지막 위치",
    timestamp
  };
}

export function missionTypeLabel(value) {
  return missionTypes.find((missionType) => missionType.value === value)?.label ?? value;
}

export function makeStatusLabel(status) {
  const labels = {
    offline: "오프라인",
    online: "온라인",
    assigned: "배정",
    streaming: "송출 중",
    reconnecting: "재연결",
    fault: "장애",
    ready: "준비",
    active: "진행 중",
    ended: "종료",
    recording: "녹화 중",
    uploaded: "업로드 완료"
  };
  return labels[status] ?? status;
}

export function makeLiveStatusLabel(status) {
  const labels = {
    connected: "연결됨",
    completed: "연결됨",
    checking: "연결 확인",
    connecting: "연결 중",
    disconnected: "연결 안 됨",
    failed: "실패",
    closed: "종료",
    "signaling connected": "대기 중",
    "signaling closed": "종료",
    "signaling error": "신호 오류"
  };
  return labels[status] ?? status;
}

export function makeLiveChannelLabel(label) {
  const labels = {
    audio: "오디오",
    sensor: "센서",
    telemetry: "위치",
    thermal: "Thermal",
    rgb: "RGB"
  };
  return labels[label] ?? label;
}

export function makePeerRoleLabel(role) {
  const labels = {
    operator: "관제",
    recorder: "녹화",
    robot: "로봇"
  };
  return labels[role] ?? "연결";
}

export function makeStatusTone(status) {
  if (["ok", "online", "streaming", "connected", "completed", "uploaded", "현재 위치"].includes(status)) {
    return "ok";
  }
  if (["failed", "fault", "offline", "신호 오류"].includes(status)) {
    return "danger";
  }
  return "waiting";
}

export function formatDurationSeconds(value) {
  if (!value && value !== 0) {
    return "-";
  }
  const seconds = Number(value);
  if (!Number.isFinite(seconds)) {
    return "-";
  }
  if (seconds < 60) {
    return `${seconds}초`;
  }
  return `${Math.round(seconds / 60)}분`;
}
