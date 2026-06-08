export const detectionOverlayTtlMs = 1200;

const videoTrackSlots = {
  "track.video_1": "rgb",
  "track.video_2": "thermal"
};

const detectionPalette = [
  "#38bdf8",
  "#fb7185",
  "#facc15",
  "#34d399",
  "#a78bfa",
  "#fb923c",
  "#22d3ee",
  "#f472b6"
];

function normalizeEventList(payload) {
  return Array.isArray(payload?.events) ? payload.events : [];
}

function normalizeSeverity(severity) {
  return ["info", "notice", "warning", "critical"].includes(severity) ? severity : "info";
}

function eventTimestamp(event, fallbackTimestamp) {
  return event?.occurredAt || fallbackTimestamp || new Date().toISOString();
}

function isFiniteUnitNumber(value) {
  return Number.isFinite(value) && value >= 0 && value <= 1;
}

function normalizeBbox(bbox) {
  if (!bbox || bbox.format !== "normalized_xywh") {
    return null;
  }
  const normalized = {
    x: Number(bbox.x),
    y: Number(bbox.y),
    width: Number(bbox.width),
    height: Number(bbox.height)
  };
  const hasValidShape = Object.values(normalized).every(isFiniteUnitNumber)
    && normalized.width > 0
    && normalized.height > 0
    && normalized.x + normalized.width <= 1
    && normalized.y + normalized.height <= 1;
  return hasValidShape ? normalized : null;
}

function normalizeDetection(detection, index) {
  const className = String(detection?.className ?? "").trim();
  const confidence = Number(detection?.confidence);
  const bbox = normalizeBbox(detection?.bbox);
  if (!className || !Number.isFinite(confidence) || confidence < 0 || confidence > 1 || !bbox) {
    return null;
  }
  return {
    id: String(detection?.trackingId ?? detection?.classId ?? `${className}-${index}`),
    bbox,
    className,
    confidence
  };
}

const detectionObjectStrategy = {
  eventType: "detection.object",
  createProjection(event, context) {
    const trackId = String(event?.media?.trackId ?? "").trim();
    const trackSlot = videoTrackSlots[trackId];
    if (!trackSlot) {
      return null;
    }
    const detections = (Array.isArray(event?.payload?.detections) ? event.payload.detections : [])
      .map(normalizeDetection)
      .filter(Boolean);
    if (detections.length === 0) {
      return null;
    }
    return {
      detectionOverlay: {
        id: String(event?.eventId ?? `${trackId}-${eventTimestamp(event, context.receivedAt)}`),
        detections,
        occurredAt: eventTimestamp(event, context.receivedAt),
        receivedAt: context.receivedAt,
        trackId,
        trackSlot
      }
    };
  }
};

const missionEventStrategy = {
  eventType: "mission.event",
  createProjection(event, context) {
    const payload = event?.payload ?? {};
    const title = String(event?.title ?? payload.code ?? event?.eventType ?? "").trim();
    if (!title) {
      return null;
    }
    const description = String(event?.description ?? "").trim();
    return {
      liveEvent: {
        at: eventTimestamp(event, context.receivedAt),
        description,
        id: String(event?.eventId ?? `${title}-${eventTimestamp(event, context.receivedAt)}`),
        message: title,
        severity: normalizeSeverity(event?.severity)
      }
    };
  }
};

const eventStrategies = new Map([
  [detectionObjectStrategy.eventType, detectionObjectStrategy],
  [missionEventStrategy.eventType, missionEventStrategy]
]);

export function createEventLiveProjection(payload, options = {}) {
  const receivedAt = payload?.receivedAt || options.receivedAt || new Date().toISOString();
  return normalizeEventList(payload).reduce((projection, event) => {
    const strategy = eventStrategies.get(event?.eventType);
    const strategyProjection = strategy?.createProjection(event, { receivedAt });
    if (strategyProjection?.detectionOverlay) {
      projection.detectionOverlays.push(strategyProjection.detectionOverlay);
    }
    if (strategyProjection?.liveEvent) {
      projection.liveEvents.push(strategyProjection.liveEvent);
    }
    return projection;
  }, {
    detectionOverlays: [],
    liveEvents: []
  });
}

export function createEmptyDetectionOverlays() {
  return {
    rgb: null,
    thermal: null
  };
}

export function isDetectionOverlayFresh(overlay, nowMs = Date.now()) {
  const timestamp = Date.parse(overlay?.receivedAt || overlay?.occurredAt || "");
  return Number.isFinite(timestamp) && nowMs - timestamp <= detectionOverlayTtlMs;
}

export function getDetectionColor(className) {
  const normalized = String(className || "unknown").trim().toLowerCase();
  let hash = 0;
  for (let index = 0; index < normalized.length; index += 1) {
    hash = ((hash << 5) - hash + normalized.charCodeAt(index)) | 0;
  }
  return detectionPalette[Math.abs(hash) % detectionPalette.length];
}
