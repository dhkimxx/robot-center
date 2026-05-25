import { makeMissionConnectionKey } from "../missions/missionHelpers.js";
import { LiveCloseReason } from "./liveConnectionStates.js";

export function createLiveConnectionRegistryState() {
  const liveConnections = new Map();
  const currentAttempts = new Map();
  const attemptConnectionKeys = new Map();
  let nextAttemptId = 0;

  const getConnection = (connectionKey) => liveConnections.get(connectionKey) ?? null;

  const isCurrentConnection = (connectionKey, client) => liveConnections.get(connectionKey) === client;

  const isCurrentAttempt = (attemptId) => {
    const connectionKey = attemptConnectionKeys.get(attemptId);
    return Boolean(connectionKey)
      && currentAttempts.get(connectionKey)?.attemptId === attemptId;
  };

  const startConnectionAttempt = (connectionKey, targetKey, options = {}) => {
    const attemptId = nextAttemptId + 1;
    nextAttemptId = attemptId;
    const attempt = {
      attemptId,
      connectionKey,
      targetKey,
      isCurrent: () => isCurrentAttempt(attemptId)
    };
    const previousAttempt = currentAttempts.get(connectionKey);
    if (previousAttempt) {
      attemptConnectionKeys.delete(previousAttempt.attemptId);
    }
    currentAttempts.set(connectionKey, attempt);
    attemptConnectionKeys.set(attemptId, connectionKey);

    const existingConnection = liveConnections.get(connectionKey);
    if (existingConnection) {
      liveConnections.delete(connectionKey);
      existingConnection.close?.(options.closeReason ?? LiveCloseReason.SWITCHING_TARGET);
    }

    return attempt;
  };

  const registerConnection = (connectionKey, client, attempt = null) => {
    if (attempt && !attempt.isCurrent()) {
      client?.close?.(LiveCloseReason.CONNECTION_FAILED);
      return false;
    }
    liveConnections.set(connectionKey, client);
    return true;
  };

  const removeConnection = (connectionKey, client, attempt = null) => {
    if (isCurrentConnection(connectionKey, client)) {
      liveConnections.delete(connectionKey);
      if (!attempt || currentAttempts.get(connectionKey)?.attemptId === attempt.attemptId) {
        currentAttempts.delete(connectionKey);
        if (attempt) {
          attemptConnectionKeys.delete(attempt.attemptId);
        }
      }
    }
  };

  const closeLiveConnection = (connectionKey, reason = LiveCloseReason.DISCONNECTED) => {
    const connection = liveConnections.get(connectionKey);
    connection?.close?.(reason);
  };

  const closeMissionConnection = (missionCode, reason = LiveCloseReason.DISCONNECTED) => {
    closeLiveConnection(makeMissionConnectionKey(missionCode), reason);
  };

  const closeAllConnections = (reason = LiveCloseReason.DISCONNECTED) => {
    liveConnections.forEach((connection) => connection?.close?.(reason));
    liveConnections.clear();
    currentAttempts.clear();
    attemptConnectionKeys.clear();
  };

  return {
    closeAllConnections,
    closeMissionConnection,
    getConnection,
    registerConnection,
    removeConnection,
    startConnectionAttempt
  };
}
