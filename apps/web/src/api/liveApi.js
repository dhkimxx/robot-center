import { requestJson } from "./controlCenterApi.js";

export function fetchRtcConfig() {
  return requestJson("/api/v1/operator/rtc-config");
}

export function fetchMissionLiveStatus(missionCode) {
  return requestJson(`/api/v1/operator/missions/${encodeURIComponent(missionCode)}/live-status`);
}

export function fetchSensorLatest(missionId, robotCode = "") {
  const params = new URLSearchParams({ missionId });
  if (robotCode) {
    params.set("robotCode", robotCode);
  }
  return requestJson(`/api/v1/operator/sensor-latest?${params.toString()}`);
}
