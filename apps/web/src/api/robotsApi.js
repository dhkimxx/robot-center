import { requestJson } from "./controlCenterApi.js";

function robotPath(robotCode) {
  return encodeURIComponent(robotCode);
}

export function fetchRobots() {
  return requestJson("/api/robots");
}

export function createRobotRequest(robotForm) {
  return requestJson("/api/robots", {
    method: "POST",
    body: JSON.stringify(robotForm)
  });
}

export function fetchRobotConnectionInfo(robotCode) {
  return requestJson(`/api/robots/${robotPath(robotCode)}/connection-info`);
}

export function updateRobotRequest(robotCode, robotEditForm) {
  return requestJson(`/api/robots/${robotPath(robotCode)}`, {
    method: "PATCH",
    body: JSON.stringify(robotEditForm)
  });
}

export function rotateRobotConnectionToken(robotCode) {
  return requestJson(`/api/robots/${robotPath(robotCode)}/connection-token`, { method: "POST" });
}

export function archiveRobotRequest(robotCode) {
  return requestJson(`/api/robots/${robotPath(robotCode)}`, { method: "DELETE" });
}
