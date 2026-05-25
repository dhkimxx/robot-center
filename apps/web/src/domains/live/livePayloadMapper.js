const telemetryChannelRoles = new Set(["channel.telemetry", "telemetry"]);
const eventChannelRoles = new Set(["channel.event", "event"]);
const spatialChannelRoles = new Set(["channel.spatial", "spatial"]);
const controlChannelRoles = new Set(["channel.control", "control"]);

function normalizeChannelRole(label, payload) {
  return payload?.channelRole ?? label ?? "";
}

function findSampleValues(payload, kind) {
  const sample = payload?.samples?.find((candidate) => {
    const sensorId = candidate?.sensorId ?? "";
    return candidate?.kind === kind || sensorId.includes(`.${kind}_`) || sensorId.includes(`.${kind}`);
  });
  return sample?.values ?? null;
}

function createTelemetrySensorProjection(payload) {
  const environmentValues = findSampleValues(payload, "environment");
  const batteryValues = findSampleValues(payload, "battery");
  const projectedPayload = {
    ...(environmentValues ?? {}),
    ...(batteryValues ?? {}),
    ...(payload?.payload ?? {})
  };

  if (Object.keys(projectedPayload).length === 0) {
    return null;
  }

  return {
    ...payload,
    payload: projectedPayload
  };
}

function createChannelEventMessage(channelRole, payload) {
  if (eventChannelRoles.has(channelRole)) {
    return payload?.event?.message ? `이벤트: ${payload.event.message}` : "이벤트 데이터 수신";
  }
  if (spatialChannelRoles.has(channelRole)) {
    return payload?.state ? `공간 상태: ${payload.state}` : "공간 데이터 수신";
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
  const messageType = payload?.messageType ?? "";

  if (telemetryChannelRoles.has(channelRole)) {
    return {
      channelRole,
      ok: true,
      sensor: createTelemetrySensorProjection(payload),
      telemetry: payload
    };
  }

  if (label === "sensor" || messageType === "sensor") {
    return {
      channelRole: "sensor",
      ok: true,
      sensor: payload
    };
  }

  if (
    eventChannelRoles.has(channelRole)
    || spatialChannelRoles.has(channelRole)
    || controlChannelRoles.has(channelRole)
  ) {
    return {
      channelRole,
      eventMessage: createChannelEventMessage(channelRole, payload),
      ok: true
    };
  }

  return {
    channelRole,
    eventMessage: "데이터 수신",
    ok: true
  };
}
