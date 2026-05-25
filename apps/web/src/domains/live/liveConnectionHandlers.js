import { makeLiveChannelLabel } from "../../utils/formatters.js";
import { makeMissionRobotKey } from "../missions/missionHelpers.js";
import {
  findRobotCodeForRemoteTrack,
  findRobotCodeFromDataMessage,
  findTrackSlot
} from "./liveHelpers.js";
import {
  isIntentionalLiveCloseReason,
  LiveCloseReason,
  LiveSessionStatus
} from "./liveConnectionStates.js";

export function createLiveConnectionHandlers({
  appendLiveEvent,
  applyTrackStream,
  attempt,
  missionCode,
  missionTargets,
  persistDataChannelMessage,
  setLiveSessionStatus,
  target,
  targetKey
}) {
  let videoTrackOrder = 0;

  return {
    onDataChannelMessage(label, message) {
      if (!attempt.isCurrent()) {
        return;
      }
      const robotCode = findRobotCodeFromDataMessage(message) || target.robotCode;
      persistDataChannelMessage(makeMissionRobotKey(missionCode, robotCode), label, message, {
        attemptId: attempt.attemptId
      });
    },
    onEvent(message, robotCode = target.robotCode) {
      if (!attempt.isCurrent()) {
        return;
      }
      appendLiveEvent(makeMissionRobotKey(missionCode, robotCode), message, {
        attemptId: attempt.attemptId
      });
    },
    onStatusChange(status) {
      if (!attempt.isCurrent()) {
        return;
      }
      setLiveSessionStatus(targetKey, status, { attemptId: attempt.attemptId });
    },
    onTrack(event) {
      if (!attempt.isCurrent()) {
        return;
      }
      const robotCode = findRobotCodeForRemoteTrack(event, missionTargets) || target.robotCode;
      const routedTargetKey = makeMissionRobotKey(missionCode, robotCode);
      const stream = new MediaStream([event.track]);
      const slot = findTrackSlot(event, videoTrackOrder);
      if (slot !== "audio") {
        videoTrackOrder += 1;
      }
      applyTrackStream(routedTargetKey, slot, stream, { attemptId: attempt.attemptId });
      appendLiveEvent(
        routedTargetKey,
        slot === "audio" ? "오디오 수신" : `${makeLiveChannelLabel(slot)} 영상 수신`,
        { attemptId: attempt.attemptId }
      );
    }
  };
}

export function createLiveConnectionCloseHandler({
  appendLiveEvent,
  attempt,
  client,
  connectionKey,
  removeConnection,
  setLiveSessionStatus,
  targetKey
}) {
  return ({ reason } = {}) => {
    if (!attempt.isCurrent()) {
      return;
    }
    const isIntentionalClose = isIntentionalLiveCloseReason(reason);
    const nextStatus = reason === LiveCloseReason.CONNECTION_FAILED
      ? LiveSessionStatus.SIGNALING_ERROR
      : isIntentionalClose ? LiveSessionStatus.IDLE : LiveSessionStatus.SIGNALING_CLOSED;
    setLiveSessionStatus(
      targetKey,
      nextStatus,
      { attemptId: attempt.attemptId, resetStreams: isIntentionalClose }
    );
    if (!isIntentionalClose) {
      appendLiveEvent(targetKey, "관제 연결 종료", { attemptId: attempt.attemptId });
    }
    removeConnection(connectionKey, client, attempt);
  };
}
