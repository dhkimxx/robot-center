export const LiveSessionStatus = Object.freeze({
  CHECKING: "checking",
  CLOSED: "closed",
  COMPLETED: "completed",
  CONNECTED: "connected",
  CONNECTING: "connecting",
  DISCONNECTED: "disconnected",
  FAILED: "failed",
  IDLE: "idle",
  SIGNALING_CLOSED: "signaling closed",
  SIGNALING_CONNECTED: "signaling connected",
  SIGNALING_ERROR: "signaling error"
});

export const LiveCloseReason = Object.freeze({
  CONNECTION_FAILED: "operator connection failed",
  DISCONNECTED: "operator disconnected",
  NAVIGATION: "operator navigation",
  SWITCHING_TARGET: "operator switching target"
});

export const activeLiveConnectionStatuses = new Set([
  LiveSessionStatus.CHECKING,
  LiveSessionStatus.COMPLETED,
  LiveSessionStatus.CONNECTED,
  LiveSessionStatus.CONNECTING,
  LiveSessionStatus.SIGNALING_CONNECTED
]);

export const connectedLiveConnectionStatuses = new Set([
  LiveSessionStatus.COMPLETED,
  LiveSessionStatus.CONNECTED
]);

export const reconnectableLiveStatuses = new Set([
  LiveSessionStatus.CLOSED,
  LiveSessionStatus.DISCONNECTED,
  LiveSessionStatus.FAILED,
  LiveSessionStatus.SIGNALING_CLOSED,
  LiveSessionStatus.SIGNALING_ERROR
]);

export const intentionalLiveCloseReasons = new Set([
  LiveCloseReason.DISCONNECTED,
  LiveCloseReason.NAVIGATION,
  LiveCloseReason.SWITCHING_TARGET
]);

export function isIntentionalLiveCloseReason(reason) {
  return intentionalLiveCloseReasons.has(reason);
}
