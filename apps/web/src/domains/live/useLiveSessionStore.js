import { useCallback, useState } from "react";
import { makeLiveChannelLabel } from "../../utils/formatters.js";
import { createEmptyLiveSession, makeLiveTargetKey } from "./liveHelpers.js";
import { LiveSessionStatus } from "./liveConnectionStates.js";
import { applyLiveAttemptUpdate } from "./liveSessionAttempts.js";
import { createEmptyDetectionOverlays } from "./liveEventStrategies.js";
import { replaceVideoStreamSlot, resetVideoStreams } from "./liveMediaCleanup.js";
import { mergeSensorSnapshots } from "./sensorDisplayMetrics.js";

const maxLiveEvents = 40;

function normalizeLiveEventSeverity(severity) {
  const normalized = String(severity ?? "").trim().toLowerCase();
  return ["info", "notice", "warning", "critical"].includes(normalized) ? normalized : "info";
}

function createLiveEventRecord(event) {
  const timestamp = new Date().toISOString();
  if (typeof event === "string") {
    return {
      at: timestamp,
      description: "",
      id: `${Date.now()}-${Math.random()}`,
      message: event,
      severity: "info"
    };
  }
  const eventType = String(event?.eventType ?? "").trim();
  const eventId = String(event?.eventId ?? "").trim();
  const code = String(event?.code ?? "").trim();
  const message = String(event?.message ?? (code || eventType)).trim();
  if (!message) {
    return null;
  }
  return {
    at: event.at ?? timestamp,
    category: String(event?.category ?? "").trim(),
    code,
    description: String(event?.description ?? "").trim(),
    eventId,
    eventType,
    id: event.id ?? (eventId ? `${eventType || "event"}:${eventId}` : `${Date.now()}-${Math.random()}`),
    message,
    receivedAt: event.receivedAt ?? "",
    severity: normalizeLiveEventSeverity(event.severity)
  };
}

function mergeLiveEvent(events, nextEvent) {
  if (!nextEvent) {
    return events;
  }
  return [
    nextEvent,
    ...events.filter((event) => event.id !== nextEvent.id)
  ].slice(0, maxLiveEvents);
}

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
    const nextEvent = createLiveEventRecord(event);
    if (!nextEvent) {
      return;
    }
    updateLiveSession(targetKey, (session) => ({
      ...session,
      events: mergeLiveEvent(session.events, nextEvent)
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
