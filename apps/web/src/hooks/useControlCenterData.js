import { useCallback, useEffect, useState } from "react";
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

  const loadAll = useCallback(async () => {
    const [statusPayload, robotPayload, missionPayload, streamingPayload, recordingPayload] = await Promise.all([
      fetchSystemStatus(),
      fetchRobots(),
      fetchMissions(),
      fetchStreamingStatuses(),
      fetchRecordings()
    ]);
    setSystemStatus(statusPayload);
    setRobots(robotPayload.robots ?? []);
    setMissions(missionPayload.missions ?? []);
    setStreamingStatuses(streamingPayload.streamingStatuses ?? []);
    setRecordings(recordingPayload.recordings ?? []);
    setStatusError("");
  }, []);

  useEffect(() => {
    let cancelled = false;
    async function loadInitial() {
      try {
        await loadAll();
      } catch (error) {
        if (!cancelled) {
          setStatusError(error instanceof Error ? error.message : "status load failed");
        }
      }
    }
    loadInitial();
    const timer = window.setInterval(loadInitial, 5000);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
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
