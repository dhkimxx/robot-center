import { requestJson } from "./controlCenterApi.js";

export function fetchRtcConfig() {
  return requestJson("/api/rtc-config");
}

export function fetchStreamingStatuses() {
  return requestJson("/api/streaming-statuses");
}

export function fetchTelemetrySamples(missionId) {
  return requestJson(`/api/telemetry?missionId=${encodeURIComponent(missionId)}`);
}

export function fetchSensorReadings(missionId) {
  return requestJson(`/api/sensor-readings?missionId=${encodeURIComponent(missionId)}`);
}
