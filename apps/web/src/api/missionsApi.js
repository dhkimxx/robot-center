import { requestJson } from "./controlCenterApi.js";

function missionPath(missionCode) {
  return encodeURIComponent(missionCode);
}

export function fetchMissions() {
  return requestJson("/api/v1/operator/missions");
}

export function createMissionRequest(missionForm) {
  return requestJson("/api/v1/operator/missions", {
    method: "POST",
    body: JSON.stringify(missionForm)
  });
}

export function startMissionRequest(missionCode) {
  return requestJson(`/api/v1/operator/missions/${missionPath(missionCode)}/start`, { method: "POST" });
}

export function endMissionRequest(missionCode) {
  return requestJson(`/api/v1/operator/missions/${missionPath(missionCode)}/end`, { method: "POST" });
}
