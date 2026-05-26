import { requestJson } from "./controlCenterApi.js";

export function fetchRtcConfig() {
  return requestJson("/api/rtc-config");
}

export function fetchStreamingStatuses() {
  return requestJson("/api/streaming-statuses");
}

export function fetchSensorLatest(missionId, robotCode = "") {
  const params = new URLSearchParams({ missionId });
  if (robotCode) {
    params.set("robotCode", robotCode);
  }
  return requestJson(`/api/sensor-latest?${params.toString()}`);
}
