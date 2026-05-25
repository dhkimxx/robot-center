import { useCallback, useRef } from "react";
import { createLiveConnectionRegistryState } from "./liveConnectionRegistry.js";

export function useLiveConnectionRegistry() {
  const registryRef = useRef(null);
  if (!registryRef.current) {
    registryRef.current = createLiveConnectionRegistryState();
  }
  const registry = registryRef.current;

  return {
    closeAllConnections: useCallback((reason) => registry.closeAllConnections(reason), [registry]),
    closeMissionConnection: useCallback((missionCode, reason) => registry.closeMissionConnection(missionCode, reason), [registry]),
    getConnection: useCallback((connectionKey) => registry.getConnection(connectionKey), [registry]),
    registerConnection: useCallback((connectionKey, client, attempt) => registry.registerConnection(connectionKey, client, attempt), [registry]),
    removeConnection: useCallback((connectionKey, client, attempt) => registry.removeConnection(connectionKey, client, attempt), [registry]),
    startConnectionAttempt: useCallback((connectionKey, targetKey, options) => registry.startConnectionAttempt(connectionKey, targetKey, options), [registry])
  };
}
