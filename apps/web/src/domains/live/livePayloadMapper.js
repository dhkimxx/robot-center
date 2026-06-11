import { createEventLiveProjection } from "./liveEventStrategies.js";

const telemetryChannelRoles = new Set(["channel.telemetry"]);
const eventChannelRoles = new Set(["channel.event"]);
const spatialChannelRoles = new Set(["channel.spatial"]);
const controlChannelRoles = new Set(["channel.control"]);

function normalizeChannelRole(label, payload) {
  return payload?.channelRole ?? label ?? "";
}

function findSampleValues(payload, kind) {
  const sample = payload?.samples?.find((candidate) => {
    const sensorId = candidate?.sensorId ?? "";
    return sensorId.includes(`.${kind}_`) || sensorId.includes(`.${kind}`);
  });
  return sample?.values ?? null;
}

function findLatestSampleTimestamp(payload) {
  return (payload?.samples ?? [])
    .map((sample) => sample?.timestamp)
    .filter(Boolean)
    .sort((left, right) => Date.parse(right) - Date.parse(left))[0] ?? "";
}

function withReceivedAt(payload) {
  return {
    ...payload,
    receivedAt: payload?.receivedAt ?? findLatestSampleTimestamp(payload)
  };
}

function hasTelemetrySampleValues(payload) {
  return (payload?.samples ?? []).some((sample) => {
    if (!sample?.sensorId) {
      return false;
    }
    return (sample.values !== null && sample.values !== undefined)
      || Boolean(sample.objectKey);
  });
}

function createTelemetryProjection(payload) {
  const positionValues = findSampleValues(payload, "position");
  if (!positionValues) {
    return payload;
  }
  return {
    ...payload,
    payload: {
      position: positionValues,
      positionAvailable: true
    }
  };
}

function createTelemetrySensorProjection(payload) {
  const positionValues = findSampleValues(payload, "position");
  const batteryValues = findSampleValues(payload, "battery");
  const projectedPayload = {
    ...(positionValues ? { position: positionValues, positionAvailable: true } : {}),
    ...(batteryValues ?? {})
  };
  const hasProjectedPayload = Object.keys(projectedPayload).length > 0;

  if (!hasProjectedPayload && !hasTelemetrySampleValues(payload)) {
    return null;
  }

  return {
    ...payload,
    ...(hasProjectedPayload ? { payload: projectedPayload } : {})
  };
}

function createChannelEventMessage(channelRole, payload) {
  if (eventChannelRoles.has(channelRole)) {
    return payload?.event?.message ? `이벤트: ${payload.event.message}` : "";
  }
  if (spatialChannelRoles.has(channelRole)) {
    return "공간 데이터 수신";
  }
  if (controlChannelRoles.has(channelRole)) {
    return payload?.ack?.state ? `제어 응답: ${payload.ack.state}` : "제어 데이터 수신";
  }
  return "데이터 수신";
}

export function mapLiveDataChannelPayload(label, message) {
  let payload;
  try {
    payload = JSON.parse(message);
  } catch {
    return {
      eventMessage: "데이터 해석 실패",
      ok: false
    };
  }

  const channelRole = normalizeChannelRole(label, payload);
  if (telemetryChannelRoles.has(channelRole)) {
    const contextualPayload = withReceivedAt(payload);
    return {
      channelRole,
      ok: true,
      sensor: createTelemetrySensorProjection(contextualPayload),
      telemetry: createTelemetryProjection(contextualPayload)
    };
  }

  if (spatialChannelRoles.has(channelRole)) {
    const contextualPayload = withReceivedAt(payload);
    return {
      channelRole,
      eventMessage: createChannelEventMessage(channelRole, payload),
      ok: true,
      sensor: contextualPayload
    };
  }

  if (
    eventChannelRoles.has(channelRole)
    || controlChannelRoles.has(channelRole)
  ) {
    const contextualPayload = withReceivedAt(payload);
    const eventProjection = eventChannelRoles.has(channelRole)
      ? createEventLiveProjection(contextualPayload)
      : { detectionOverlays: [], liveEvents: [] };
    return {
      channelRole,
      eventMessage: createChannelEventMessage(channelRole, payload),
      detectionOverlays: eventProjection.detectionOverlays,
      liveEvents: eventProjection.liveEvents,
      ok: true
    };
  }

  return {
    channelRole,
    eventMessage: "데이터 수신",
    ok: true
  };
}
