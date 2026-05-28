import { SensorType, normalizeSensorType } from "./sensorTypes.js";

const fieldDefinitions = {
  accuracyMeter: { label: "위치 오차", unit: "m", order: 46 },
  altitudeMeter: { label: "고도", unit: "m", order: 43 },
  batteryPercent: { label: "배터리", unit: "%", order: 30 },
  headingDegree: { label: "방향", unit: "deg", order: 44 },
  networkState: { label: "네트워크", unit: "", order: 35 },
  speedMeterPerSecond: { label: "속도", unit: "m/s", order: 45 },
  yawDegree: { label: "Yaw", unit: "deg", order: 62 }
};

const segmentLabels = {
  angularVelocity: "각속도",
  linearAcceleration: "선가속도",
  x: "X",
  y: "Y",
  z: "Z"
};

const hiddenFieldNames = new Set([
  "alarm",
  "alarm_code",
  "alarmCode",
  "channel",
  "frameId",
  "high_alarm",
  "highAlarm",
  "latitude",
  "longitude",
  "low_alarm",
  "lowAlarm",
  "positionAvailable",
  "scale_code",
  "scaleCode",
  "valid"
]);

const sensorTypeOrders = {
  [SensorType.GAS]: 10,
  [SensorType.BATTERY]: 30,
  network: 35,
  [SensorType.POSITION]: 40,
  [SensorType.IMU]: 50,
  [SensorType.ODOMETRY]: 60,
  [SensorType.POINT_CLOUD]: 70,
  spatial: 75,
  [SensorType.UNKNOWN]: 90
};

const gasLabelOrders = {
  CO: 0.01,
  H2S: 0.02,
  O2: 0.03,
  CH4: 0.04,
  TEMP: 0.05,
  HUM: 0.06
};

const sensorMetricStrategies = {
  [SensorType.GAS]: appendGasSampleMetrics,
  default: appendDefaultSampleMetrics
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
    label: "Payload",
    sensorId: "payload",
    sensorType: SensorType.UNKNOWN,
    unit: ""
  }, sensor.payload, "", sensor.receivedAt);

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
      label: item.label ?? item.descriptor?.label,
      sensorId: item.sensorId ?? item.descriptor?.sensorId,
      sensorType: item.sensorType ?? item.descriptor?.sensorType,
      unit: item.unit ?? item.descriptor?.unit
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
    samples: mergeBySensorId(previous.samples, incoming.samples)
  };
}

function normalizeExistingMetrics(metrics) {
  return sortMetrics(metrics.map((metric, index) => ({
    key: metric.key ?? `${metric.label}-${metric.unit ?? ""}-${index}`,
    alarmLevel: metric.alarmLevel ?? "normal",
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
  const strategy = sensorMetricStrategies[normalizedDescriptor.sensorType] ?? sensorMetricStrategies.default;
  strategy(metrics, seenMetricKeys, normalizedDescriptor, sample);
}

function appendDefaultSampleMetrics(metrics, seenMetricKeys, descriptor, sample) {
  const value = getSampleValue(sample);
  const receivedAt = sample.timestamp ?? sample.receivedAt;
  if (value && typeof value === "object" && !Array.isArray(value)) {
    appendObjectMetrics(metrics, seenMetricKeys, descriptor, value, "", receivedAt);
    return;
  }

  appendMetric(metrics, seenMetricKeys, descriptor, {
    fieldPath: descriptor.sensorId,
    receivedAt,
    value
  });
}

function appendGasSampleMetrics(metrics, seenMetricKeys, descriptor, sample) {
  appendSingleReadingMetric(metrics, seenMetricKeys, descriptor, sample, {
    normalizedField: "concentration"
  });
}

function appendSingleReadingMetric(metrics, seenMetricKeys, descriptor, sample, { normalizedField }) {
  const value = getSampleValue(sample);
  const receivedAt = sample.timestamp ?? sample.receivedAt;
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    appendDefaultSampleMetrics(metrics, seenMetricKeys, descriptor, sample);
    return;
  }
  const reading = numberOrNull(value[normalizedField] ?? value.concentration);
  if (reading === null) {
    appendDefaultSampleMetrics(metrics, seenMetricKeys, descriptor, sample);
    return;
  }
  appendMetric(metrics, seenMetricKeys, descriptor, {
    fieldPath: normalizedField,
    label: descriptor.label,
    receivedAt,
    unit: stringOrEmpty(value.unit) || descriptor.unit,
    value: reading
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

function appendMetric(metrics, seenMetricKeys, descriptor, { alarmLevel = "normal", fieldPath, label, receivedAt, unit, value }) {
  if (value === null || value === undefined || value === "") {
    return;
  }
  const fieldName = getLastPathSegment(fieldPath);
  const fieldDefinition = fieldDefinitions[fieldName];
  const metricKey = label ?? fieldDefinition?.label ?? fieldPath;
  if (seenMetricKeys.has(metricKey)) {
    return;
  }
  seenMetricKeys.add(metricKey);

  metrics.push({
    key: `${descriptor.sensorId}:${fieldPath}`,
    alarmLevel,
    label: label ?? makeMetricLabel(descriptor, fieldPath, fieldDefinition),
    order: metricOrder(descriptor, fieldDefinition),
    receivedAt,
    unit: unit ?? fieldDefinition?.unit ?? descriptor.unit ?? "",
    value
  });
}

function normalizeDescriptor(descriptor, sample) {
  const sensorId = descriptor?.sensorId ?? sample?.sensorId ?? "sensor.unknown";
  return {
    label: descriptor?.label ?? sensorId,
    sensorId,
    sensorType: normalizeSensorType(descriptor?.sensorType ?? inferSensorType(sensorId)),
    unit: descriptor?.unit ?? ""
  };
}

function metricOrder(descriptor, fieldDefinition) {
  if (descriptor.sensorType === SensorType.GAS) {
    return sensorTypeOrders[SensorType.GAS] + gasLabelOrder(descriptor.label);
  }
  return fieldDefinition?.order ?? (sensorTypeOrders[descriptor.sensorType] ?? sensorTypeOrders.unknown);
}

function gasLabelOrder(label) {
  return gasLabelOrders[normalizeLabelKey(label)] ?? 0.9;
}

function normalizeLabelKey(label) {
  return String(label ?? "").trim().toUpperCase();
}

function getSampleValue(sample) {
  if (!sample) {
    return null;
  }
  if (sample.values !== null && sample.values !== undefined) {
    return sample.values;
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
    return `${descriptor.label} ${readableSegments.join(" ")}`;
  }
  if (fieldPath === descriptor.sensorId) {
    return descriptor.label;
  }
  return `${descriptor.label} ${segmentLabels[fieldPath] ?? fieldPath}`;
}

function getLastPathSegment(fieldPath) {
  return fieldPath.split(".").at(-1) ?? fieldPath;
}

function numberOrNull(value) {
  return typeof value === "number" && Number.isFinite(value) ? value : null;
}

function stringOrEmpty(value) {
  return typeof value === "string" ? value.trim() : "";
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
    return SensorType.POSITION;
  }
  if (normalizedSensorId.includes("imu")) {
    return SensorType.IMU;
  }
  if (normalizedSensorId.includes("odometry")) {
    return SensorType.ODOMETRY;
  }
  if (normalizedSensorId.includes("battery")) {
    return SensorType.BATTERY;
  }
  if (normalizedSensorId.includes("network")) {
    return "network";
  }
  if (normalizedSensorId.includes("gas")) {
    return SensorType.GAS;
  }
  return SensorType.UNKNOWN;
}
