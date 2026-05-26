const knownMetricFields = [
  { key: "coPpm", label: "CO", unit: "ppm" },
  { key: "oxygenPercent", label: "O2", unit: "%" },
  { key: "temperatureCelsius", label: "온도", unit: "C" },
  { key: "humidityPercent", label: "습도", unit: "%" },
  { key: "batteryPercent", label: "배터리", unit: "%" }
];

function getSampleObjectValue(sample) {
  if (!sample) {
    return null;
  }
  return sample.objectValue ?? sample.vectorValue ?? sample.rawPayload?.values ?? sample.rawPayload?.payload ?? null;
}

function getSampleScalarValue(sample) {
  if (!sample) {
    return null;
  }
  if (sample.numericValue !== null && sample.numericValue !== undefined) {
    return sample.numericValue;
  }
  if (sample.textValue) {
    return sample.textValue;
  }
  if (sample.boolValue !== null && sample.boolValue !== undefined) {
    return sample.boolValue ? "true" : "false";
  }
  return null;
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

function appendKnownObjectMetrics(metrics, item, objectValue) {
  for (const field of knownMetricFields) {
    if (objectValue?.[field.key] === null || objectValue?.[field.key] === undefined) {
      continue;
    }
    metrics.push({
      label: field.label,
      receivedAt: item.latestSample?.receivedAt,
      unit: field.unit,
      value: objectValue[field.key]
    });
  }
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
    sentAt: item.latestSample?.sentAt
  };
}

export function createSensorPanelSnapshot(sensorLatest, robotCode) {
  if (!Array.isArray(sensorLatest) || sensorLatest.length === 0) {
    return null;
  }
  const metrics = [];
  let receivedAt = "";

  for (const item of sensorLatest) {
    const sensorType = getSensorType(item);
    if (sensorType === "position") {
      continue;
    }
    const sample = item.latestSample;
    if (!sample) {
      continue;
    }
    if (!receivedAt || new Date(sample.receivedAt) > new Date(receivedAt)) {
      receivedAt = sample.receivedAt;
    }
    const objectValue = getSampleObjectValue(sample);
    if (objectValue && typeof objectValue === "object" && !Array.isArray(objectValue)) {
      const beforeCount = metrics.length;
      appendKnownObjectMetrics(metrics, item, objectValue);
      if (metrics.length !== beforeCount) {
        continue;
      }
    }
    metrics.push({
      label: item.displayName ?? item.sensorId,
      receivedAt: sample.receivedAt,
      unit: item.unit ?? "",
      value: getSampleScalarValue(sample) ?? summarizeSensorObject(objectValue) ?? sample.objectKey ?? "-"
    });
  }

  return {
    receivedAt,
    robotCode,
    sensors: metrics.slice(0, 8)
  };
}

function summarizeSensorObject(value) {
  if (!value || typeof value !== "object") {
    return null;
  }
  const entries = Object.entries(value).slice(0, 2);
  if (entries.length === 0) {
    return null;
  }
  return entries.map(([key, item]) => `${key}:${formatCompactValue(item)}`).join(" ");
}

function formatCompactValue(value) {
  if (typeof value === "number") {
    return Number(value).toFixed(1);
  }
  if (typeof value === "string") {
    return value;
  }
  if (typeof value === "boolean") {
    return value ? "true" : "false";
  }
  return "object";
}
