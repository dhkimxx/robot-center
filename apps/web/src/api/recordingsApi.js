import { requestJson } from "./controlCenterApi.js";

export function fetchMissionRecordingSummary(missionCode) {
  return requestJson(`/api/v1/operator/missions/${encodeURIComponent(missionCode)}/recordings/summary`);
}

export function fetchMissionRecordingChunks(missionCode, options = {}) {
  const query = new URLSearchParams();
  const robotCode = String(options.robotCode ?? "").trim();
  if (robotCode) {
    query.set("robotCode", robotCode);
  }
  if (Number.isFinite(Number(options.limit))) {
    query.set("limit", String(Number(options.limit)));
  }
  if (Number.isFinite(Number(options.offset))) {
    query.set("offset", String(Number(options.offset)));
  }
  const queryString = query.toString();
  return requestJson(`/api/v1/operator/missions/${encodeURIComponent(missionCode)}/recordings/chunks${queryString ? `?${queryString}` : ""}`);
}
