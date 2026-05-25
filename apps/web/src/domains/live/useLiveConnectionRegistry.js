import { useCallback, useRef } from "react";
import { makeMissionConnectionKey } from "../missions/missionHelpers.js";
import { LiveCloseReason } from "./liveConnectionStates.js";

export function useLiveConnectionRegistry() {
  const liveConnectionsRef = useRef(new Map());
  const currentAttemptsRef = useRef(new Map());
  const attemptConnectionKeysRef = useRef(new Map());
  const nextAttemptIdRef = useRef(0);

  const getConnection = useCallback((connectionKey) => (
    liveConnectionsRef.current.get(connectionKey) ?? null
  ), []);

  const isCurrentConnection = useCallback((connectionKey, client) => (
    liveConnectionsRef.current.get(connectionKey) === client
  ), []);

  const isCurrentAttempt = useCallback((attemptId) => {
    const connectionKey = attemptConnectionKeysRef.current.get(attemptId);
    return Boolean(connectionKey)
      && currentAttemptsRef.current.get(connectionKey)?.attemptId === attemptId;
  }, []);

  const startConnectionAttempt = useCallback((connectionKey, targetKey, options = {}) => {
    const attemptId = nextAttemptIdRef.current + 1;
    nextAttemptIdRef.current = attemptId;
    const attempt = {
      attemptId,
      connectionKey,
      targetKey,
      isCurrent: () => isCurrentAttempt(attemptId)
    };
    const previousAttempt = currentAttemptsRef.current.get(connectionKey);
    if (previousAttempt) {
      attemptConnectionKeysRef.current.delete(previousAttempt.attemptId);
    }
    currentAttemptsRef.current.set(connectionKey, attempt);
    attemptConnectionKeysRef.current.set(attemptId, connectionKey);

    const existingConnection = liveConnectionsRef.current.get(connectionKey);
    existingConnection?.close?.(options.closeReason ?? LiveCloseReason.SWITCHING_TARGET);

    return attempt;
  }, [isCurrentAttempt]);

  const registerConnection = useCallback((connectionKey, client, attempt = null) => {
    if (attempt && !attempt.isCurrent()) {
      client?.close?.(LiveCloseReason.CONNECTION_FAILED);
      return false;
    }
    liveConnectionsRef.current.set(connectionKey, client);
    return true;
  }, []);

  const removeConnection = useCallback((connectionKey, client, attempt = null) => {
    if (isCurrentConnection(connectionKey, client)) {
      liveConnectionsRef.current.delete(connectionKey);
      if (!attempt || currentAttemptsRef.current.get(connectionKey)?.attemptId === attempt.attemptId) {
        currentAttemptsRef.current.delete(connectionKey);
        if (attempt) {
          attemptConnectionKeysRef.current.delete(attempt.attemptId);
        }
      }
    }
  }, [isCurrentConnection]);

  const closeLiveConnection = useCallback((connectionKey, reason = LiveCloseReason.DISCONNECTED) => {
    const connection = liveConnectionsRef.current.get(connectionKey);
    connection?.close?.(reason);
  }, []);

  const closeMissionConnection = useCallback((missionCode, reason = LiveCloseReason.DISCONNECTED) => {
    closeLiveConnection(makeMissionConnectionKey(missionCode), reason);
  }, [closeLiveConnection]);

  const closeAllConnections = useCallback((reason = LiveCloseReason.DISCONNECTED) => {
    liveConnectionsRef.current.forEach((connection) => connection?.close?.(reason));
    liveConnectionsRef.current.clear();
    currentAttemptsRef.current.clear();
    attemptConnectionKeysRef.current.clear();
  }, []);

  return {
    closeAllConnections,
    closeLiveConnection,
    closeMissionConnection,
    getConnection,
    isCurrentConnection,
    isCurrentAttempt,
    registerConnection,
    removeConnection,
    startConnectionAttempt
  };
}
