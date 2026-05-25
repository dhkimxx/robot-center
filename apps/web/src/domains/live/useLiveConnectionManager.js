import { useCallback, useEffect, useRef, useState } from "react";
import { fetchRtcConfig } from "../../api/liveApi.js";
import { makeLiveChannelLabel } from "../../utils/formatters.js";
import {
  createEmptyLiveSession,
  findRobotCodeForRemoteTrack,
  findRobotCodeFromDataMessage,
  findTrackSlot,
  makeLiveTargetKey
} from "./liveHelpers.js";
import { mapLiveDataChannelPayload } from "./livePayloadMapper.js";
import {
  getMissionCodeFromRobotKey,
  makeMissionConnectionKey,
  makeMissionRobotKey,
  makeMissionRoomId
} from "../missions/missionHelpers.js";
import { createLiveConnectionClient } from "./liveConnectionClient.js";

const selectedLiveTargetStorageKey = "robot-center.selectedLiveTargetKey";
const reconnectableLiveStatuses = new Set([
  "closed",
  "disconnected",
  "failed",
  "signaling closed",
  "signaling error"
]);
const activeLiveConnectionStatuses = new Set([
  "checking",
  "completed",
  "connected",
  "connecting",
  "signaling connected"
]);
const intentionalCloseReasons = new Set([
  "operator disconnected",
  "operator navigation",
  "operator switching target"
]);

export function readSelectedLiveTargetKey() {
  try {
    return window.localStorage.getItem(selectedLiveTargetStorageKey) ?? "";
  } catch {
    return "";
  }
}

function writeSelectedLiveTargetKey(targetKey) {
  try {
    if (targetKey) {
      window.localStorage.setItem(selectedLiveTargetStorageKey, targetKey);
      return;
    }
    window.localStorage.removeItem(selectedLiveTargetStorageKey);
  } catch {
    // Local selection persistence is optional; the in-memory state remains authoritative.
  }
}

export function useLiveConnectionManager({
  activeSection,
  liveTargets,
  missionControlMission,
  setMissionControlCode,
  showNotification
}) {
  const [selectedLiveTargetKey, setSelectedLiveTargetKey] = useState(readSelectedLiveTargetKey);
  const [liveSessions, setLiveSessions] = useState({});
  const liveConnectionsRef = useRef(new Map());
  const autoConnectingTargetKeyRef = useRef("");

  const updateLiveSession = useCallback((targetKey, updater) => {
    setLiveSessions((current) => {
      const previous = current[targetKey] ?? createEmptyLiveSession();
      return {
        ...current,
        [targetKey]: updater(previous)
      };
    });
  }, []);

  const appendLiveEvent = useCallback((targetKey, message) => {
    if (!targetKey) {
      return;
    }
    updateLiveSession(targetKey, (session) => ({
      ...session,
      events: [
        { id: `${Date.now()}-${Math.random()}`, message, at: new Date().toISOString() },
        ...session.events
      ].slice(0, 40)
    }));
  }, [updateLiveSession]);

  const closeLiveConnection = useCallback((connectionKey, reason = "operator disconnected") => {
    const connection = liveConnectionsRef.current.get(connectionKey);
    connection?.close?.(reason);
    liveConnectionsRef.current.delete(connectionKey);
  }, []);

  const disconnectLiveTarget = useCallback((targetKey) => {
    const missionCode = getMissionCodeFromRobotKey(targetKey);
    const connectionKey = makeMissionConnectionKey(missionCode);
    closeLiveConnection(connectionKey, "operator navigation");
    liveConnectionsRef.current.delete(targetKey);
    const targetKeys = liveTargets
      .filter((candidate) => candidate.mission.missionCode === missionCode)
      .map((candidate) => candidate.key);
    (targetKeys.length > 0 ? targetKeys : [targetKey]).forEach((candidateKey) => {
      updateLiveSession(candidateKey, (session) => ({
        ...session,
        status: "disconnected",
        videoStreams: { rgb: null, thermal: null, audio: null }
      }));
    });
  }, [closeLiveConnection, liveTargets, updateLiveSession]);

  const disconnectMissionByCode = useCallback((missionCode) => {
    disconnectLiveTarget(makeMissionRobotKey(missionCode, ""));
  }, [disconnectLiveTarget]);

  useEffect(() => () => {
    liveConnectionsRef.current.forEach((connection) => connection?.close?.("operator disconnected"));
    liveConnectionsRef.current.clear();
  }, []);

  const persistDataChannelMessage = useCallback((targetKey, label, message) => {
    const mappedPayload = mapLiveDataChannelPayload(label, message);
    if (!mappedPayload.ok) {
      appendLiveEvent(targetKey, `${makeLiveChannelLabel(label)} ${mappedPayload.eventMessage}`);
      return;
    }

    if (mappedPayload.telemetry || mappedPayload.sensor) {
      updateLiveSession(targetKey, (session) => ({
        ...session,
        ...(mappedPayload.telemetry ? { telemetry: mappedPayload.telemetry } : {}),
        ...(mappedPayload.sensor ? { sensor: mappedPayload.sensor } : {})
      }));
    }
    if (mappedPayload.eventMessage) {
      appendLiveEvent(targetKey, mappedPayload.eventMessage);
      return;
    }
  }, [appendLiveEvent, updateLiveSession]);

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
    const currentConnection = liveConnectionsRef.current.get(connectionKey);
    const currentSession = liveSessions[targetKey] ?? createEmptyLiveSession();
    if (!options.force && currentConnection?.targetKey === targetKey && activeLiveConnectionStatuses.has(currentSession.status)) {
      return;
    }
    disconnectLiveTarget(targetKey);
    setMissionControlCode(missionCode);
    setSelectedLiveTargetKey(targetKey);
    updateLiveSession(targetKey, (session) => ({
      ...session,
      status: "connecting",
      videoStreams: { rgb: null, thermal: null, audio: null }
    }));

    try {
      const rtcConfig = await fetchRtcConfig();
      let videoTrackOrder = 0;
      const client = createLiveConnectionClient({
        missionRoomId,
        robotCode: target.robotCode,
        rtcConfig,
        onDataChannelMessage: (label, message) => {
          const robotCode = findRobotCodeFromDataMessage(message) || target.robotCode;
          persistDataChannelMessage(makeMissionRobotKey(missionCode, robotCode), label, message);
        },
        onEvent: (message, robotCode = target.robotCode) => {
          appendLiveEvent(makeMissionRobotKey(missionCode, robotCode), message);
        },
        onStatusChange: (status) => {
          updateLiveSession(targetKey, (session) => ({ ...session, status }));
        },
        onTrack: (event) => {
          const robotCode = findRobotCodeForRemoteTrack(event, missionTargets) || target.robotCode;
          const routedTargetKey = makeMissionRobotKey(missionCode, robotCode);
          const stream = new MediaStream([event.track]);
          const slot = findTrackSlot(event, videoTrackOrder);
          if (slot !== "audio") {
            videoTrackOrder += 1;
          }
          if (slot === "audio") {
            updateLiveSession(routedTargetKey, (session) => ({
              ...session,
              videoStreams: { ...session.videoStreams, audio: stream }
            }));
            appendLiveEvent(routedTargetKey, "오디오 수신");
            return;
          }

          updateLiveSession(routedTargetKey, (session) => ({
            ...session,
            videoStreams: { ...session.videoStreams, [slot]: stream }
          }));
          appendLiveEvent(routedTargetKey, `${makeLiveChannelLabel(slot)} 영상 수신`);
        }
      });
      client.targetKey = targetKey;
      client.onClose(({ reason } = {}) => {
        const isIntentionalClose = intentionalCloseReasons.has(reason);
        updateLiveSession(targetKey, (session) => ({
          ...session,
          status: isIntentionalClose ? "idle" : "signaling closed",
          ...(isIntentionalClose ? { videoStreams: { rgb: null, thermal: null, audio: null } } : {})
        }));
        if (!isIntentionalClose) {
          appendLiveEvent(targetKey, "관제 연결 종료");
        }
        const currentConnection = liveConnectionsRef.current.get(connectionKey);
        if (currentConnection?.targetKey === targetKey) {
          liveConnectionsRef.current.delete(connectionKey);
        }
      });
      liveConnectionsRef.current.set(connectionKey, client);
    } catch (error) {
      closeLiveConnection(connectionKey, "operator connection failed");
      updateLiveSession(targetKey, (session) => ({ ...session, status: "failed" }));
      showNotification(error instanceof Error ? error.message : "관제 연결 실패", "danger");
      appendLiveEvent(targetKey, `관제 연결 실패: ${error instanceof Error ? error.message : "알 수 없음"}`);
    }
  }, [appendLiveEvent, closeLiveConnection, disconnectLiveTarget, liveSessions, liveTargets, persistDataChannelMessage, setMissionControlCode, showNotification, updateLiveSession]);

  const selectedLiveTarget = liveTargets.find((target) => target.key === selectedLiveTargetKey) ?? liveTargets[0] ?? null;
  const selectedLiveSession = liveSessions[makeLiveTargetKey(selectedLiveTarget)] ?? createEmptyLiveSession();

  const reconnectLive = useCallback(() => {
    void connectLiveTarget(selectedLiveTarget, { force: true });
  }, [connectLiveTarget, selectedLiveTarget]);

  useEffect(() => {
    if (liveTargets.length === 0) {
      setSelectedLiveTargetKey("");
      return;
    }
    if (!selectedLiveTargetKey || !liveTargets.some((target) => target.key === selectedLiveTargetKey)) {
      setSelectedLiveTargetKey(liveTargets[0].key);
    }
  }, [liveTargets, selectedLiveTargetKey]);

  useEffect(() => {
    writeSelectedLiveTargetKey(selectedLiveTargetKey);
  }, [selectedLiveTargetKey]);

  useEffect(() => {
    if (activeSection !== "missions" || !missionControlMission || missionControlMission.status !== "active" || !selectedLiveTarget) {
      return;
    }

    const targetKey = makeLiveTargetKey(selectedLiveTarget);
    const missionConnectionKey = makeMissionConnectionKey(missionControlMission.missionCode);
    const currentConnection = liveConnectionsRef.current.get(missionConnectionKey);
    const selectedSession = liveSessions[targetKey] ?? createEmptyLiveSession();

    if (
      reconnectableLiveStatuses.has(selectedSession.status)
      && selectedSession.events.length > 0
      && (!currentConnection || currentConnection.targetKey === targetKey)
    ) {
      return;
    }
    if (currentConnection?.targetKey === targetKey && activeLiveConnectionStatuses.has(selectedSession.status)) {
      return;
    }
    if (autoConnectingTargetKeyRef.current === targetKey) {
      return;
    }

    autoConnectingTargetKeyRef.current = targetKey;
    void connectLiveTarget(selectedLiveTarget).finally(() => {
      if (autoConnectingTargetKeyRef.current === targetKey) {
        autoConnectingTargetKeyRef.current = "";
      }
    });
  }, [activeSection, connectLiveTarget, liveSessions, missionControlMission, selectedLiveTarget]);

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
