import { navigationItems } from "../config/controlCenterConfig.js";

const defaultNavigationKey = "missions";

export function normalizePathname(pathname) {
  const normalizedPathname = pathname.replace(/\/+$/, "");
  return normalizedPathname || "/";
}

export function findNavigationItemByPath(pathname) {
  const normalizedPathname = normalizePathname(pathname);
  return navigationItems.find((item) => item.path === normalizedPathname) ?? null;
}

export function getNavigationKeyFromPath(pathname) {
  return findNavigationItemByPath(pathname)?.key ?? defaultNavigationKey;
}

export function getNavigationPath(tabKey) {
  return navigationItems.find((item) => item.key === tabKey)?.path ?? "/missions";
}
