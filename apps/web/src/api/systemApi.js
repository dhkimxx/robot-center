import { requestJson } from "./controlCenterApi.js";

export function fetchSystemStatus() {
  return requestJson("/api/system/status");
}

export function clearObjectStorage() {
  return requestJson("/api/system/object-storage/clear", {
    body: JSON.stringify({ confirmation: "CLEAR_OBJECT_STORAGE" }),
    method: "POST",
    timeoutMs: 60000
  });
}
