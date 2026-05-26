import { requestJson } from "./controlCenterApi.js";

export function fetchRtcConfig() {
  return requestJson("/api/rtc-config");
}

export function fetchObservedStreams() {
  return requestJson("/api/observed-streams");
}

export function fetchMissionLiveStatus(missionCode) {
  return requestJson(`/api/missions/${encodeURIComponent(missionCode)}/live-status`);
}

export function fetchSensorLatest(missionId, robotCode = "") {
  const params = new URLSearchParams({ missionId });
  if (robotCode) {
    params.set("robotCode", robotCode);
  }
  return requestJson(`/api/sensor-latest?${params.toString()}`);
}
