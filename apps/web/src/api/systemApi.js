import { requestJson } from "./controlCenterApi.js";

export function fetchSystemStatus({ scope = "full" } = {}) {
  const query = scope ? `?scope=${encodeURIComponent(scope)}` : "";
  return requestJson(`/api/v1/system/status${query}`);
}

export function clearObjectStorage() {
  return requestJson("/api/v1/system/object-storage/clear", {
    body: JSON.stringify({ confirmation: "CLEAR_OBJECT_STORAGE" }),
    method: "POST",
    timeoutMs: 60000
  });
}

export function pruneObjectStorage() {
  return requestJson("/api/v1/system/object-storage/prune", {
    body: JSON.stringify({ confirmation: "PRUNE_OBJECT_STORAGE" }),
    method: "POST",
    timeoutMs: 60000
  });
}

export function clearSensorData() {
  return requestJson("/api/v1/system/sensors/clear", {
    body: JSON.stringify({ confirmation: "CLEAR_SENSOR_DATA" }),
    method: "POST",
    timeoutMs: 60000
  });
}

export function clearEventData() {
  return requestJson("/api/v1/system/events/clear", {
    body: JSON.stringify({ confirmation: "CLEAR_EVENT_DATA" }),
    method: "POST",
    timeoutMs: 60000
  });
}

export function clearRecorderRuntime() {
  return requestJson("/api/v1/system/recorder-runtime/clear", {
    body: JSON.stringify({ confirmation: "CLEAR_RECORDER_RUNTIME" }),
    method: "POST",
    timeoutMs: 60000
  });
}

export function pruneRecorderRuntime() {
  return requestJson("/api/v1/system/recorder-runtime/prune", {
    body: JSON.stringify({ confirmation: "PRUNE_RECORDER_RUNTIME" }),
    method: "POST",
    timeoutMs: 60000
  });
}
