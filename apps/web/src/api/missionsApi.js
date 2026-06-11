import { requestJson } from "./controlCenterApi.js";
import { createListQueryPath } from "./listQueryApi.js";

function missionPath(missionCode) {
  return encodeURIComponent(missionCode);
}

export function fetchMissions(query) {
  return requestJson(createListQueryPath("/api/v1/operator/missions", query));
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
