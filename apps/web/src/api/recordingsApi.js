import { requestJson } from "./controlCenterApi.js";

export function fetchRecordings() {
  return requestJson("/api/recordings");
}
