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

export function interpretSensorSampleValue(descriptor, sample) {
  if (!sample) {
    return [];
  }
  const normalizedDescriptor = normalizeDescriptor(descriptor, sample);
  if (normalizedDescriptor.sensorType === SensorType.GAS) {
    return interpretGasSampleValue(normalizedDescriptor, sample);
  }
  return interpretDefaultSampleValue(normalizedDescriptor, sample);
}

function interpretDefaultSampleValue(descriptor, sample) {
  const value = getSampleValue(sample);
  const receivedAt = sample.timestamp ?? sample.receivedAt;
  if (value && typeof value === "object" && !Array.isArray(value)) {
    const readings = [];
    appendObjectReadings(readings, descriptor, value, "", receivedAt);
    return readings;
  }

  return [createReading(descriptor, {
    fieldPath: descriptor.sensorId,
    receivedAt,
    value
  })];
}

function interpretGasSampleValue(descriptor, sample) {
  const value = getSampleValue(sample);
  const receivedAt = sample.timestamp ?? sample.receivedAt;
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return interpretDefaultSampleValue(descriptor, sample);
  }
  const reading = numberOrNull(value.concentration);
  if (reading === null) {
    return interpretDefaultSampleValue(descriptor, sample);
  }
  return [createReading(descriptor, {
    fieldPath: "concentration",
    label: descriptor.label,
    receivedAt,
    unit: stringOrEmpty(value.unit) || descriptor.unit,
    value: reading
  })];
}

function appendObjectReadings(readings, descriptor, value, basePath, receivedAt) {
  Object.entries(value).forEach(([fieldName, fieldValue]) => {
    const fieldPath = basePath ? `${basePath}.${fieldName}` : fieldName;
    if (hiddenFieldNames.has(fieldName)) {
      return;
    }
    if (fieldValue && typeof fieldValue === "object" && !Array.isArray(fieldValue)) {
      appendObjectReadings(readings, descriptor, fieldValue, fieldPath, receivedAt);
      return;
    }
    readings.push(createReading(descriptor, {
      fieldPath,
      receivedAt,
      value: fieldValue
    }));
  });
}

function createReading(descriptor, { alarmLevel = "normal", fieldPath, label, receivedAt, unit, value }) {
  const fieldName = getLastPathSegment(fieldPath);
  const fieldDefinition = fieldDefinitions[fieldName];
  return {
    alarmLevel,
    fieldPath,
    key: `${descriptor.sensorId}:${fieldPath}`,
    label: label ?? makeReadingLabel(descriptor, fieldPath, fieldDefinition),
    order: metricOrder(descriptor, fieldDefinition),
    receivedAt,
    unit: unit ?? fieldDefinition?.unit ?? descriptor.unit ?? "",
    value
  };
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

function makeReadingLabel(descriptor, fieldPath, fieldDefinition) {
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
