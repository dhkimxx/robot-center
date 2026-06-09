import { requestJson } from "./controlCenterApi.js";

export function fetchSystemStatus() {
  return requestJson("/api/v1/system/status");
}

export function clearObjectStorage() {
  return requestJson("/api/v1/system/object-storage/clear", {
    body: JSON.stringify({ confirmation: "CLEAR_OBJECT_STORAGE" }),
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
