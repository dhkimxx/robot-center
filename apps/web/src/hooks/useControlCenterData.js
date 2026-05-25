import { useCallback, useEffect, useRef, useState } from "react";
import { fetchStreamingStatuses } from "../api/liveApi.js";
import { fetchMissions } from "../api/missionsApi.js";
import { fetchRecordings } from "../api/recordingsApi.js";
import { fetchRobots } from "../api/robotsApi.js";
import { fetchSystemStatus } from "../api/systemApi.js";

export function useControlCenterData() {
  const [systemStatus, setSystemStatus] = useState(null);
  const [robots, setRobots] = useState([]);
  const [missions, setMissions] = useState([]);
  const [streamingStatuses, setStreamingStatuses] = useState([]);
  const [recordings, setRecordings] = useState([]);
  const [statusError, setStatusError] = useState("");
  const requestSequenceRef = useRef(0);

  const loadAll = useCallback(async (options = {}) => {
    const requestSequence = requestSequenceRef.current + 1;
    requestSequenceRef.current = requestSequence;
    let payloads;
    try {
      payloads = await Promise.all([
        fetchSystemStatus(),
        fetchRobots(),
        fetchMissions(),
        fetchStreamingStatuses(),
        fetchRecordings()
      ]);
    } catch (error) {
      if (requestSequence !== requestSequenceRef.current || options.isCancelled?.()) {
        return false;
      }
      throw error;
    }
    if (requestSequence !== requestSequenceRef.current || options.isCancelled?.()) {
      return false;
    }
    const [statusPayload, robotPayload, missionPayload, streamingPayload, recordingPayload] = payloads;
    setSystemStatus(statusPayload);
    setRobots(robotPayload.robots ?? []);
    setMissions(missionPayload.missions ?? []);
    setStreamingStatuses(streamingPayload.streamingStatuses ?? []);
    setRecordings(recordingPayload.recordings ?? []);
    setStatusError("");
    return true;
  }, []);

  useEffect(() => {
    let cancelled = false;
    let timer = null;

    async function scheduleNextLoad() {
      try {
        await loadAll({ isCancelled: () => cancelled });
      } catch (error) {
        if (!cancelled) {
          setStatusError(error instanceof Error ? error.message : "status load failed");
        }
      }
      if (!cancelled) {
        timer = window.setTimeout(scheduleNextLoad, 5000);
      }
    }

    scheduleNextLoad();
    return () => {
      cancelled = true;
      if (timer) {
        window.clearTimeout(timer);
      }
    };
  }, [loadAll]);

  return {
    systemStatus,
    setSystemStatus,
    robots,
    setRobots,
    missions,
    setMissions,
    streamingStatuses,
    setStreamingStatuses,
    recordings,
    setRecordings,
    statusError,
    setStatusError,
    loadAll
  };
}
