const fieldDefinitions = {
  accuracyMeter: { label: "위치 오차", unit: "m", order: 46 },
  altitudeMeter: { label: "고도", unit: "m", order: 43 },
  batteryPercent: { label: "배터리", unit: "%", order: 30 },
  ch4Ppm: { label: "CH4", unit: "ppm", order: 14 },
  coPpm: { label: "CO", unit: "ppm", order: 10 },
  headingDegree: { label: "방향", unit: "deg", order: 44 },
  humidityPercent: { label: "습도", unit: "%", order: 22 },
  networkState: { label: "네트워크", unit: "", order: 35 },
  oxygenPercent: { label: "O2", unit: "%", order: 12 },
  speedMeterPerSecond: { label: "속도", unit: "m/s", order: 45 },
  temperatureCelsius: { label: "온도", unit: "C", order: 20 },
  yawDegree: { label: "Yaw", unit: "deg", order: 62 }
};

const segmentLabels = {
  angularVelocity: "각속도",
  linearAcceleration: "선가속도",
  x: "X",
  y: "Y",
  z: "Z"
};

const hiddenFieldNames = new Set(["latitude", "longitude", "positionAvailable"]);

const sensorTypeOrders = {
  gas: 10,
  environment: 12,
  temperature: 20,
  humidity: 22,
  battery: 30,
  network: 35,
  position: 40,
  imu: 50,
  odometry: 60,
  point_cloud: 70,
  spatial: 75,
  unknown: 90
};

export function createSensorMetrics(sensor) {
  if (!sensor) {
    return [];
  }
  if (Array.isArray(sensor.sensors) && sensor.sensors.length > 0) {
    return normalizeExistingMetrics(sensor.sensors);
  }

  const metrics = [];
  const seenMetricKeys = new Set();
  const descriptorBySensorId = createDescriptorMap(sensor.descriptors);

  for (const sample of sensor.samples ?? []) {
    appendSampleMetrics(metrics, seenMetricKeys, descriptorBySensorId.get(sample.sensorId), sample);
  }

  appendObjectMetrics(metrics, seenMetricKeys, {
    displayName: "Payload",
    sensorId: "payload",
    sensorType: "legacy",
    unit: ""
  }, sensor.payload, "", sensor.receivedAt ?? sensor.sentAt);

  return sortMetrics(metrics);
}

export function createSensorMetricsFromSensorLatest(sensorLatest) {
  if (!Array.isArray(sensorLatest)) {
    return [];
  }
  const metrics = [];
  const seenMetricKeys = new Set();

  for (const item of sensorLatest) {
    const descriptor = {
      displayName: item.displayName ?? item.descriptor?.displayName,
      sensorId: item.sensorId ?? item.descriptor?.sensorId,
      sensorType: item.sensorType ?? item.descriptor?.sensorType,
      unit: item.unit ?? item.descriptor?.unit,
      valueType: item.valueType ?? item.descriptor?.valueType
    };
    appendSampleMetrics(metrics, seenMetricKeys, descriptor, item.latestSample);
  }

  return sortMetrics(metrics);
}

export function mergeSensorSnapshots(previous, incoming) {
  if (!previous) {
    return incoming;
  }
  if (!incoming) {
    return previous;
  }

  return {
    ...previous,
    ...incoming,
    descriptors: mergeBySensorId(previous.descriptors, incoming.descriptors),
    payload: {
      ...(previous.payload ?? {}),
      ...(incoming.payload ?? {})
    },
    receivedAt: latestTimestamp(previous.receivedAt, incoming.receivedAt),
    samples: mergeBySensorId(previous.samples, incoming.samples),
    sentAt: latestTimestamp(previous.sentAt, incoming.sentAt)
  };
}

function normalizeExistingMetrics(metrics) {
  return sortMetrics(metrics.map((metric, index) => ({
    key: metric.key ?? `${metric.label}-${metric.unit ?? ""}-${index}`,
    label: metric.label,
    order: metric.order ?? index,
    receivedAt: metric.receivedAt,
    unit: metric.unit ?? "",
    value: metric.value
  })));
}

function createDescriptorMap(descriptors = []) {
  const descriptorBySensorId = new Map();
  descriptors.forEach((descriptor) => {
    if (descriptor?.sensorId) {
      descriptorBySensorId.set(descriptor.sensorId, descriptor);
    }
  });
  return descriptorBySensorId;
}

function appendSampleMetrics(metrics, seenMetricKeys, descriptor, sample) {
  if (!sample) {
    return;
  }
  const normalizedDescriptor = normalizeDescriptor(descriptor, sample);
  const value = getSampleValue(sample);
  const receivedAt = sample.receivedAt ?? sample.timestamp;

  if (value && typeof value === "object" && !Array.isArray(value)) {
    appendObjectMetrics(metrics, seenMetricKeys, normalizedDescriptor, value, "", receivedAt);
    return;
  }

  appendMetric(metrics, seenMetricKeys, normalizedDescriptor, {
    fieldPath: normalizedDescriptor.sensorId,
    receivedAt,
    value
  });
}

function appendObjectMetrics(metrics, seenMetricKeys, descriptor, value, basePath, receivedAt) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return;
  }

  Object.entries(value).forEach(([fieldName, fieldValue]) => {
    const fieldPath = basePath ? `${basePath}.${fieldName}` : fieldName;
    if (hiddenFieldNames.has(fieldName)) {
      return;
    }
    if (fieldValue && typeof fieldValue === "object" && !Array.isArray(fieldValue)) {
      appendObjectMetrics(metrics, seenMetricKeys, descriptor, fieldValue, fieldPath, receivedAt);
      return;
    }
    appendMetric(metrics, seenMetricKeys, descriptor, {
      fieldPath,
      receivedAt,
      value: fieldValue
    });
  });
}

function appendMetric(metrics, seenMetricKeys, descriptor, { fieldPath, receivedAt, value }) {
  if (value === null || value === undefined || value === "") {
    return;
  }
  const fieldName = getLastPathSegment(fieldPath);
  const fieldDefinition = fieldDefinitions[fieldName];
  const metricKey = fieldDefinition?.label ?? fieldPath;
  if (seenMetricKeys.has(metricKey)) {
    return;
  }
  seenMetricKeys.add(metricKey);

  metrics.push({
    key: `${descriptor.sensorId}:${fieldPath}`,
    label: makeMetricLabel(descriptor, fieldPath, fieldDefinition),
    order: fieldDefinition?.order ?? (sensorTypeOrders[descriptor.sensorType] ?? sensorTypeOrders.unknown),
    receivedAt,
    unit: fieldDefinition?.unit ?? descriptor.unit ?? "",
    value
  });
}

function normalizeDescriptor(descriptor, sample) {
  const sensorId = descriptor?.sensorId ?? sample?.sensorId ?? "sensor.unknown";
  return {
    displayName: descriptor?.displayName ?? sensorId,
    sensorId,
    sensorType: descriptor?.sensorType ?? descriptor?.kind ?? inferSensorType(sensorId),
    unit: descriptor?.unit ?? ""
  };
}

function getSampleValue(sample) {
  if (!sample) {
    return null;
  }
  if (sample.values !== null && sample.values !== undefined) {
    return sample.values;
  }
  if (sample.objectValue !== null && sample.objectValue !== undefined) {
    return sample.objectValue;
  }
  if (sample.vectorValue !== null && sample.vectorValue !== undefined) {
    return sample.vectorValue;
  }
  if (sample.rawPayload?.values !== null && sample.rawPayload?.values !== undefined) {
    return sample.rawPayload.values;
  }
  if (sample.rawPayload?.payload !== null && sample.rawPayload?.payload !== undefined) {
    return sample.rawPayload.payload;
  }
  if (sample.numericValue !== null && sample.numericValue !== undefined) {
    return sample.numericValue;
  }
  if (sample.textValue) {
    return sample.textValue;
  }
  if (sample.boolValue !== null && sample.boolValue !== undefined) {
    return sample.boolValue;
  }
  if (sample.objectKey) {
    return sample.objectKey;
  }
  return null;
}

function makeMetricLabel(descriptor, fieldPath, fieldDefinition) {
  if (fieldDefinition) {
    return fieldDefinition.label;
  }
  const segments = fieldPath.split(".");
  if (segments.length >= 2) {
    const readableSegments = segments.map((segment) => segmentLabels[segment] ?? segment);
    return `${descriptor.displayName} ${readableSegments.join(" ")}`;
  }
  if (fieldPath === descriptor.sensorId) {
    return descriptor.displayName;
  }
  return `${descriptor.displayName} ${segmentLabels[fieldPath] ?? fieldPath}`;
}

function getLastPathSegment(fieldPath) {
  return fieldPath.split(".").at(-1) ?? fieldPath;
}

function sortMetrics(metrics) {
  return [...metrics].sort((left, right) => (
    (left.order ?? 100) - (right.order ?? 100)
    || String(left.label).localeCompare(String(right.label))
  ));
}

function mergeBySensorId(previous = [], incoming = []) {
  const merged = new Map();
  previous.forEach((item) => {
    if (item?.sensorId) {
      merged.set(item.sensorId, item);
    }
  });
  incoming.forEach((item) => {
    if (item?.sensorId) {
      merged.set(item.sensorId, item);
    }
  });
  return Array.from(merged.values());
}

function latestTimestamp(left, right) {
  if (!left) {
    return right;
  }
  if (!right) {
    return left;
  }
  return Date.parse(right) >= Date.parse(left) ? right : left;
}

function inferSensorType(sensorId) {
  const normalizedSensorId = String(sensorId).toLowerCase();
  if (normalizedSensorId.includes("position")) {
    return "position";
  }
  if (normalizedSensorId.includes("imu")) {
    return "imu";
  }
  if (normalizedSensorId.includes("odometry")) {
    return "odometry";
  }
  if (normalizedSensorId.includes("battery")) {
    return "battery";
  }
  if (normalizedSensorId.includes("network")) {
    return "network";
  }
  if (normalizedSensorId.includes("environment")) {
    return "environment";
  }
  if (normalizedSensorId.includes("gas")) {
    return "gas";
  }
  return "unknown";
}
