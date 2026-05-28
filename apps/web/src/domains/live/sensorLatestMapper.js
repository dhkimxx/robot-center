import { createSensorMetricsFromSensorLatest } from "./sensorDisplayMetrics.js";

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
