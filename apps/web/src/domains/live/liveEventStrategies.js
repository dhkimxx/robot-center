export const detectionOverlayTtlMs = 3000;

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
  const normalized = String(severity ?? "").trim().toLowerCase();
  return ["info", "notice", "warning", "critical"].includes(normalized) ? normalized : "info";
}

function eventTimestamp(event, fallbackTimestamp) {
  return event?.timestamp || fallbackTimestamp || new Date().toISOString();
}

function isFiniteUnitNumber(value) {
  return Number.isFinite(value) && value >= 0 && value <= 1;
}

function normalizeBbox(bbox) {
  if (!bbox) {
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
    const values = event?.values ?? {};
    const trackId = String(values.trackId ?? "").trim();
    const trackSlot = videoTrackSlots[trackId];
    if (!trackSlot) {
      return null;
    }
    const rawDetections = Array.isArray(values.detections) ? values.detections : null;
    if (!rawDetections) {
      return null;
    }
    const detections = rawDetections
      .map(normalizeDetection)
      .filter(Boolean);
    if (rawDetections.length > 0 && detections.length === 0) {
      return null;
    }
    return {
      detectionOverlay: {
        id: String(event?.eventId ?? `${trackId}-${eventTimestamp(event, context.receivedAt)}`),
        detections,
        receivedAt: context.receivedAt,
        timestamp: eventTimestamp(event, context.receivedAt),
        trackId,
        trackSlot
      }
    };
  }
};

const missionEventStrategy = {
  eventType: "mission.event",
  createProjection(event, context) {
    const values = event?.values ?? {};
    const eventType = String(event?.eventType ?? "mission.event").trim();
    const eventId = String(event?.eventId ?? "").trim();
    const code = String(values.code ?? "").trim();
    const category = String(values.category ?? "").trim();
    const timestamp = eventTimestamp(event, context.receivedAt);
    const title = String(values.title ?? "").trim() || code || eventType;
    if (!title) {
      return null;
    }
    const description = String(values.description ?? "").trim();
    return {
      liveEvent: {
        at: timestamp,
        category,
        code,
        description,
        eventId,
        eventType,
        id: eventId ? `${eventType}:${eventId}` : `${eventType}:${code || title}:${timestamp}`,
        message: title,
        receivedAt: context.receivedAt,
        severity: normalizeSeverity(values.severity)
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

export function isDetectionOverlayFresh(overlay, nowMs = Date.now(), ttlMs = detectionOverlayTtlMs) {
  const measuredAt = Date.parse(overlay?.receivedAt || overlay?.timestamp || "");
  const normalizedTtlMs = Math.max(0, Number(ttlMs) || detectionOverlayTtlMs);
  return Number.isFinite(measuredAt) && nowMs - measuredAt <= normalizedTtlMs;
}

export function getDetectionColor(className) {
  const normalized = String(className || "unknown").trim().toLowerCase();
  let hash = 0;
  for (let index = 0; index < normalized.length; index += 1) {
    hash = ((hash << 5) - hash + normalized.charCodeAt(index)) | 0;
  }
  return detectionPalette[Math.abs(hash) % detectionPalette.length];
}
