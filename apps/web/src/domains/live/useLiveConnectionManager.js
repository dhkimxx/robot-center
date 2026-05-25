import { useCallback, useEffect } from "react";
import { fetchRtcConfig } from "../../api/liveApi.js";
import {
  createEmptyLiveSession,
  makeLiveTargetKey
} from "./liveHelpers.js";
import {
  createLiveConnectionCloseHandler,
  createLiveConnectionHandlers
} from "./liveConnectionHandlers.js";
import { mapLiveDataChannelPayload } from "./livePayloadMapper.js";
import { useLiveAutoConnect } from "./useLiveAutoConnect.js";
import { useLiveConnectionRegistry } from "./useLiveConnectionRegistry.js";
import { useLiveSessionStore } from "./useLiveSessionStore.js";
import { useLiveTargetSelection } from "./useLiveTargetSelection.js";
import {
  getMissionCodeFromRobotKey,
  makeMissionConnectionKey,
  makeMissionRobotKey,
  makeMissionRoomId
} from "../missions/missionHelpers.js";
import { createLiveConnectionClient } from "./liveConnectionClient.js";
import {
  activeLiveConnectionStatuses,
  LiveCloseReason,
  LiveSessionStatus
} from "./liveConnectionStates.js";

export function useLiveConnectionManager({
  activeSection,
  liveTargets,
  missionControlMission,
  setMissionControlCode,
  showNotification
}) {
  const {
    selectedLiveTarget,
    selectedLiveTargetKey,
    setSelectedLiveTargetKey
  } = useLiveTargetSelection(liveTargets);
  const {
    closeAllConnections,
    closeMissionConnection,
    getConnection,
    registerConnection,
    removeConnection,
    startConnectionAttempt
  } = useLiveConnectionRegistry();
  const {
    appendLiveEvent,
    applyMappedDataChannelPayload,
    applyTrackStream,
    liveSessions,
    resetMissionStreams,
    selectedSessionForTarget,
    setLiveSessionStatus
  } = useLiveSessionStore();

  const disconnectLiveTarget = useCallback((targetKey) => {
    const missionCode = getMissionCodeFromRobotKey(targetKey);
    closeMissionConnection(missionCode, LiveCloseReason.NAVIGATION);
    resetMissionStreams(missionCode, liveTargets, targetKey);
  }, [closeMissionConnection, liveTargets, resetMissionStreams]);

  const disconnectMissionByCode = useCallback((missionCode) => {
    disconnectLiveTarget(makeMissionRobotKey(missionCode, ""));
  }, [disconnectLiveTarget]);

  useEffect(() => () => {
    closeAllConnections(LiveCloseReason.DISCONNECTED);
  }, [closeAllConnections]);

  const persistDataChannelMessage = useCallback((targetKey, label, message, options = {}) => {
    const mappedPayload = mapLiveDataChannelPayload(label, message);
    applyMappedDataChannelPayload(targetKey, label, mappedPayload, options);
  }, [applyMappedDataChannelPayload]);

  const connectLiveTarget = useCallback(async (target, options = {}) => {
    if (!target) {
      showNotification("선택한 임무에 연결할 로봇이 없습니다.", "warning");
      return;
    }

    const missionCode = target.mission.missionCode;
    const missionRoomId = target.missionRoomId || makeMissionRoomId(target.mission);
    const missionTargetsForRoom = liveTargets.filter((candidate) => candidate.mission.missionCode === missionCode);
    const missionTargets = missionTargetsForRoom.length > 0 ? missionTargetsForRoom : [target];
    const targetKey = makeLiveTargetKey(target);
    const connectionKey = makeMissionConnectionKey(missionCode);
    const currentConnection = getConnection(connectionKey);
    const currentSession = liveSessions[targetKey] ?? createEmptyLiveSession();
    if (!options.force && currentConnection?.targetKey === targetKey && activeLiveConnectionStatuses.has(currentSession.status)) {
      return;
    }
    const attempt = startConnectionAttempt(connectionKey, targetKey, {
      closeReason: LiveCloseReason.SWITCHING_TARGET
    });
    resetMissionStreams(missionCode, liveTargets, targetKey, {
      attemptId: attempt.attemptId,
      replaceAttempt: true
    });
    setMissionControlCode(missionCode);
    setSelectedLiveTargetKey(targetKey);
    setLiveSessionStatus(targetKey, LiveSessionStatus.CONNECTING, {
      attemptId: attempt.attemptId,
      replaceAttempt: true,
      resetStreams: true
    });

    try {
      const rtcConfig = await fetchRtcConfig();
      let client;
      const handlers = createLiveConnectionHandlers({
        appendLiveEvent,
        applyTrackStream,
        attempt,
        missionCode,
        missionTargets,
        persistDataChannelMessage,
        setLiveSessionStatus,
        target,
        targetKey
      });
      client = createLiveConnectionClient({
        missionRoomId,
        robotCode: target.robotCode,
        rtcConfig,
        ...handlers
      });
      client.targetKey = targetKey;
      client.onClose(createLiveConnectionCloseHandler({
        appendLiveEvent,
        attempt,
        client,
        connectionKey,
        removeConnection,
        setLiveSessionStatus,
        targetKey
      }));
      registerConnection(connectionKey, client, attempt);
    } catch (error) {
      if (!attempt.isCurrent()) {
        return;
      }
      setLiveSessionStatus(targetKey, LiveSessionStatus.FAILED, { attemptId: attempt.attemptId });
      showNotification(error instanceof Error ? error.message : "관제 연결 실패", "danger");
      appendLiveEvent(
        targetKey,
        `관제 연결 실패: ${error instanceof Error ? error.message : "알 수 없음"}`,
        { attemptId: attempt.attemptId }
      );
    }
  }, [appendLiveEvent, applyTrackStream, getConnection, liveSessions, liveTargets, persistDataChannelMessage, registerConnection, removeConnection, resetMissionStreams, setLiveSessionStatus, setMissionControlCode, setSelectedLiveTargetKey, showNotification, startConnectionAttempt]);

  const selectedLiveSession = selectedSessionForTarget(selectedLiveTarget);
  const currentMissionConnection = missionControlMission
    ? getConnection(makeMissionConnectionKey(missionControlMission.missionCode))
    : null;

  const reconnectLive = useCallback(() => {
    void connectLiveTarget(selectedLiveTarget, { force: true });
  }, [connectLiveTarget, selectedLiveTarget]);

  useLiveAutoConnect({
    activeSection,
    connectLiveTarget,
    currentConnection: currentMissionConnection,
    missionControlMission,
    selectedLiveSession,
    selectedLiveTarget
  });

  return {
    appendLiveEvent,
    disconnectMissionByCode,
    liveSessions,
    reconnectLive,
    selectedLiveSession,
    selectedLiveTarget,
    selectedLiveTargetKey,
    setSelectedLiveTargetKey
  };
}
