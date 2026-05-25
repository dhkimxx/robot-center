import { requestJson } from "./controlCenterApi.js";

export function fetchSystemStatus() {
  return requestJson("/api/system/status");
}
