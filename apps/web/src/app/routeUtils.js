import { navigationItems } from "../config/controlCenterConfig.js";

export function getActiveSection(pathname) {
  return navigationItems.find((item) => pathname === item.path || pathname.startsWith(`${item.path}/`))?.key ?? "missions";
}

export function getRouteMissionControlCode(pathname) {
  const match = pathname.match(/^\/missions\/([^/]+)\/control\/?$/);
  return match ? decodeURIComponent(match[1]) : "";
}
