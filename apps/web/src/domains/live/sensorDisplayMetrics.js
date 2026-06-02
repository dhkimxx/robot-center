import { interpretSensorSampleValue } from "./sensorValueInterpreter.js";

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

  appendSampleMetrics(metrics, seenMetricKeys, {
    label: "Payload",
    sensorId: "payload",
    sensorType: "unknown",
    unit: ""
  }, {
    receivedAt: sensor.receivedAt,
    sensorId: "payload",
    values: sensor.payload
  });

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
    if (Array.isArray(item.readings) && item.readings.length > 0) {
      appendReadingMetrics(metrics, seenMetricKeys, descriptor, item.readings, item.latestSample);
      continue;
    }
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
  interpretSensorSampleValue(descriptor, sample).forEach((reading) => {
    appendMetric(metrics, seenMetricKeys, reading);
  });
}

function appendReadingMetrics(metrics, seenMetricKeys, descriptor, readings, sample) {
  const receivedAt = sample?.timestamp ?? sample?.receivedAt;
  readings.forEach((reading, index) => {
    appendMetric(metrics, seenMetricKeys, {
      alarmLevel: reading.alarmLevel ?? "normal",
      key: `${descriptor.sensorId ?? "sensor"}:${reading.fieldPath ?? index}`,
      label: reading.label,
      order: reading.order,
      receivedAt,
      unit: reading.unit ?? "",
      value: reading.value
    });
  });
}

function appendMetric(metrics, seenMetricKeys, reading) {
  const {
    alarmLevel = "normal",
    key,
    label,
    order,
    receivedAt,
    unit,
    value
  } = reading;
  if (value === null || value === undefined || value === "") {
    return;
  }
  const metricKey = label ?? key;
  if (seenMetricKeys.has(metricKey)) {
    return;
  }
  seenMetricKeys.add(metricKey);

  metrics.push({
    key,
    alarmLevel,
    label,
    order,
    receivedAt,
    unit,
    value
  });
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
