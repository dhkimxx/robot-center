import { useCallback, useState } from "react";
import { makeLiveChannelLabel } from "../../utils/formatters.js";
import { createEmptyLiveSession, makeLiveTargetKey } from "./liveHelpers.js";
import { LiveSessionStatus } from "./liveConnectionStates.js";
import { applyLiveAttemptUpdate } from "./liveSessionAttempts.js";
import { createEmptyDetectionOverlays } from "./liveEventStrategies.js";
import { replaceVideoStreamSlot, resetVideoStreams } from "./liveMediaCleanup.js";
import { mergeSensorSnapshots } from "./sensorDisplayMetrics.js";

export function useLiveSessionStore() {
  const [liveSessions, setLiveSessions] = useState({});

  const updateLiveSession = useCallback((targetKey, updater, options = {}) => {
    setLiveSessions((current) => {
      const previous = current[targetKey] ?? createEmptyLiveSession();
      const next = applyLiveAttemptUpdate(previous, options.attemptId, updater, {
        replaceAttempt: options.replaceAttempt
      });
      if (next === previous) {
        return current;
      }
      return { ...current, [targetKey]: next };
    });
  }, []);

  const appendLiveEvent = useCallback((targetKey, event, options = {}) => {
    if (!targetKey) {
      return;
    }
    const nextEvent = typeof event === "string"
      ? { id: `${Date.now()}-${Math.random()}`, message: event, at: new Date().toISOString() }
      : {
        id: event.id ?? `${Date.now()}-${Math.random()}`,
        message: event.message,
        description: event.description ?? "",
        severity: event.severity ?? "info",
        at: event.at ?? new Date().toISOString()
      };
    updateLiveSession(targetKey, (session) => ({
      ...session,
      events: [
        nextEvent,
        ...session.events
      ].slice(0, 40)
    }), options);
  }, [updateLiveSession]);

  const setLiveSessionStatus = useCallback((targetKey, status, options = {}) => {
    updateLiveSession(targetKey, (session) => ({
      ...session,
      status,
      ...(options.resetStreams ? {
        detectionOverlays: createEmptyDetectionOverlays(),
        videoStreams: resetVideoStreams(session.videoStreams)
      } : {})
    }), options);
  }, [updateLiveSession]);

  const resetMissionStreams = useCallback((missionCode, liveTargets, fallbackTargetKey, options = {}) => {
    const targetKeys = liveTargets
      .filter((candidate) => candidate.mission.missionCode === missionCode)
      .map((candidate) => candidate.key);
    (targetKeys.length > 0 ? targetKeys : [fallbackTargetKey]).forEach((targetKey) => {
      setLiveSessionStatus(targetKey, LiveSessionStatus.DISCONNECTED, { ...options, resetStreams: true });
    });
  }, [setLiveSessionStatus]);

  const applyTrackStream = useCallback((targetKey, slot, stream, options = {}) => {
    updateLiveSession(targetKey, (session) => ({
      ...session,
      videoStreams: replaceVideoStreamSlot(session.videoStreams, slot, stream)
    }), options);
  }, [updateLiveSession]);

  const applyMappedDataChannelPayload = useCallback((targetKey, label, mappedPayload, options = {}) => {
    if (!mappedPayload.ok) {
      appendLiveEvent(targetKey, `${makeLiveChannelLabel(label)} ${mappedPayload.eventMessage}`, options);
      return;
    }

    if (mappedPayload.telemetry || mappedPayload.sensor) {
      updateLiveSession(targetKey, (session) => ({
        ...session,
        ...(mappedPayload.telemetry ? { telemetry: mappedPayload.telemetry } : {}),
        ...(mappedPayload.sensor ? { sensor: mergeSensorSnapshots(session.sensor, mappedPayload.sensor) } : {})
      }), options);
    }

    if (mappedPayload.detectionOverlays?.length > 0) {
      updateLiveSession(targetKey, (session) => {
        const detectionOverlays = {
          ...session.detectionOverlays
        };
        mappedPayload.detectionOverlays.forEach((overlay) => {
          detectionOverlays[overlay.trackSlot] = overlay;
        });
        return {
          ...session,
          detectionOverlays
        };
      }, options);
    }

    if (mappedPayload.liveEvents?.length > 0) {
      mappedPayload.liveEvents.forEach((event) => appendLiveEvent(targetKey, event, options));
    }

    if (mappedPayload.eventMessage) {
      appendLiveEvent(targetKey, mappedPayload.eventMessage, options);
    }
  }, [appendLiveEvent, updateLiveSession]);

  const selectedSessionForTarget = useCallback((target) => (
    liveSessions[makeLiveTargetKey(target)] ?? createEmptyLiveSession()
  ), [liveSessions]);

  return {
    appendLiveEvent,
    applyMappedDataChannelPayload,
    applyTrackStream,
    liveSessions,
    resetMissionStreams,
    selectedSessionForTarget,
    setLiveSessionStatus,
    updateLiveSession
  };
}
