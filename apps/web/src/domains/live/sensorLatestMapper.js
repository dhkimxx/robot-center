import { createSensorMetricsFromSensorLatest } from "./sensorDisplayMetrics.js";

const liveSensorDelayThresholdMs = 10_000;

function getSampleObjectValue(sample) {
  if (!sample) {
    return null;
  }
  return sample.values ?? null;
}

function getSensorType(item) {
  return item?.sensorType ?? item?.descriptor?.sensorType ?? "";
}

function getSensorId(item) {
  return item?.sensorId ?? item?.descriptor?.sensorId ?? "";
}

function findPositionSensor(sensorLatest) {
  return (sensorLatest ?? []).find((item) => {
    const sensorType = getSensorType(item);
    const sensorId = getSensorId(item);
    return sensorType === "position" || sensorId.includes("position");
  });
}

export function createTelemetryFromSensorLatest(sensorLatest, robotCode) {
  const item = findPositionSensor(sensorLatest);
  const position = getSampleObjectValue(item?.latestSample);
  if (position?.latitude === null || position?.latitude === undefined || position?.longitude === null || position?.longitude === undefined) {
    return null;
  }
  return {
    missionId: item.missionId,
    payload: {
      position,
      positionAvailable: true
    },
    receivedAt: item.latestSample?.receivedAt,
    robotCode: item.robotCode ?? robotCode,
    timestamp: item.latestSample?.timestamp
  };
}

export function createSensorPanelSnapshot(sensorLatest, robotCode) {
  if (!Array.isArray(sensorLatest) || sensorLatest.length === 0) {
    return null;
  }
  let receivedAt = "";

  for (const item of sensorLatest) {
    const sample = item.latestSample;
    if (!sample) {
      continue;
    }
    if (!receivedAt || new Date(sample.receivedAt) > new Date(receivedAt)) {
      receivedAt = sample.receivedAt;
    }
  }

  return {
    receivedAt,
    robotCode,
    sensors: createSensorMetricsFromSensorLatest(sensorLatest)
  };
}

export function createSensorPanelState({ liveSensor, snapshotSensor, snapshotState } = {}) {
  if (liveSensor) {
    const sourceLabel = isDelayedSensor(liveSensor) ? "수신 지연" : "실시간 수신";
    return {
      sensor: liveSensor,
      source: sourceLabel === "수신 지연" ? "delayed" : "live",
      sourceLabel
    };
  }

  if (snapshotState?.status === "error" && snapshotSensor) {
    return {
      sensor: snapshotSensor,
      source: "snapshot-error",
      sourceLabel: "최근 저장값 갱신 실패"
    };
  }

  if (snapshotSensor) {
    return {
      sensor: snapshotSensor,
      source: "snapshot",
      sourceLabel: "최근 저장값"
    };
  }

  if (snapshotState?.status === "loading") {
    return {
      sensor: null,
      source: "loading",
      sourceLabel: "최근 저장값 확인 중"
    };
  }

  return {
    sensor: null,
    source: "none",
    sourceLabel: "센서 대기"
  };
}

function isDelayedSensor(sensor) {
  const receivedAt = Date.parse(sensor?.receivedAt ?? "");
  if (!Number.isFinite(receivedAt)) {
    return false;
  }
  return Date.now() - receivedAt > liveSensorDelayThresholdMs;
}
