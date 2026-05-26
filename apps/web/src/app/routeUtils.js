import { navigationItems } from "../config/controlCenterConfig.js";

export function getActiveSection(pathname) {
  return navigationItems.find((item) => pathname === item.path || pathname.startsWith(`${item.path}/`))?.key ?? "missions";
}

export function getRouteMissionControlCode(pathname) {
  const match = pathname.match(/^\/missions\/([^/]+)\/control\/?$/);
  return match ? decodeURIComponent(match[1]) : "";
}

export function getRouteMissionReplayCode(pathname) {
  const match = pathname.match(/^\/missions\/([^/]+)\/replay\/?$/);
  return match ? decodeURIComponent(match[1]) : "";
}

export function getRouteSelectedMissionCode(search) {
  const params = new URLSearchParams(search);
  return params.get("selected") ?? "";
}
