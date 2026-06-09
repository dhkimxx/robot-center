export const detectionOverlaySettingsStorageKey = "robot-center.detectionOverlaySettings";

export const defaultDetectionOverlaySettings = {
  maxDetections: 10,
  ttlSeconds: 3
};

export const detectionOverlaySettingLimits = {
  maxDetections: { max: 50, min: 1 },
  ttlSeconds: { max: 30, min: 1 }
};

export function normalizeDetectionOverlaySettings(settings = {}) {
  return {
    maxDetections: clampInteger(
      settings.maxDetections,
      detectionOverlaySettingLimits.maxDetections.min,
      detectionOverlaySettingLimits.maxDetections.max,
      defaultDetectionOverlaySettings.maxDetections
    ),
    ttlSeconds: clampInteger(
      settings.ttlSeconds,
      detectionOverlaySettingLimits.ttlSeconds.min,
      detectionOverlaySettingLimits.ttlSeconds.max,
      defaultDetectionOverlaySettings.ttlSeconds
    )
  };
}

export function readDetectionOverlaySettings(storage = globalThis.localStorage) {
  if (!storage) {
    return defaultDetectionOverlaySettings;
  }
  try {
    return normalizeDetectionOverlaySettings(JSON.parse(storage.getItem(detectionOverlaySettingsStorageKey) || "{}"));
  } catch {
    return defaultDetectionOverlaySettings;
  }
}

export function writeDetectionOverlaySettings(settings, storage = globalThis.localStorage) {
  const normalized = normalizeDetectionOverlaySettings(settings);
  if (!storage) {
    return normalized;
  }
  try {
    storage.setItem(detectionOverlaySettingsStorageKey, JSON.stringify(normalized));
  } catch {
    // Storage failures should not block live control.
  }
  return normalized;
}

function clampInteger(value, min, max, fallback) {
  const numberValue = Number(value);
  if (!Number.isFinite(numberValue)) {
    return fallback;
  }
  return Math.min(max, Math.max(min, Math.round(numberValue)));
}
