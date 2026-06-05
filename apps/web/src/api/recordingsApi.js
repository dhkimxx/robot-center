import { requestJson } from "./controlCenterApi.js";

export function fetchRecordings(options = {}) {
  const missionCode = String(options.missionCode ?? "").trim();
  const query = new URLSearchParams();
  if (missionCode) {
    query.set("missionCode", missionCode);
  }
  const queryString = query.toString();
  return requestJson(`/api/v1/operator/recordings${queryString ? `?${queryString}` : ""}`);
}
