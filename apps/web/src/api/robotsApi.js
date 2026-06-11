import { requestJson } from "./controlCenterApi.js";
import { createListQueryPath } from "./listQueryApi.js";

function robotPath(robotCode) {
  return encodeURIComponent(robotCode);
}

export function fetchRobots(query) {
  return requestJson(createListQueryPath("/api/v1/operator/robots", query));
}

export function createRobotRequest(robotForm) {
  return requestJson("/api/v1/operator/robots", {
    method: "POST",
    body: JSON.stringify(robotForm)
  });
}

export function fetchRobotConnectionInfo(robotCode) {
  return requestJson(`/api/v1/operator/robots/${robotPath(robotCode)}/connection-info`);
}

export function updateRobotRequest(robotCode, robotEditForm) {
  return requestJson(`/api/v1/operator/robots/${robotPath(robotCode)}`, {
    method: "PATCH",
    body: JSON.stringify(robotEditForm)
  });
}

export function rotateRobotConnectionToken(robotCode) {
  return requestJson(`/api/v1/operator/robots/${robotPath(robotCode)}/connection-token`, { method: "POST" });
}

export function archiveRobotRequest(robotCode) {
  return requestJson(`/api/v1/operator/robots/${robotPath(robotCode)}`, { method: "DELETE" });
}
