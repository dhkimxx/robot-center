import { useCallback, useEffect, useRef, useState } from "react";
import { fetchMissionLiveStatus } from "../api/liveApi.js";
import { fetchMissions } from "../api/missionsApi.js";
import { fetchRobots } from "../api/robotsApi.js";
import { fetchSystemStatus } from "../api/systemApi.js";

export function useControlCenterData() {
  const [systemStatus, setSystemStatus] = useState(null);
  const [robots, setRobots] = useState([]);
  const [missions, setMissions] = useState([]);
  const [missionLiveStatuses, setMissionLiveStatuses] = useState({});
  const [statusError, setStatusError] = useState("");
  const [dataLoadState, setDataLoadState] = useState({
    hasLoaded: false,
    isLoading: true
  });
  const requestSequenceRef = useRef(0);

  const loadAll = useCallback(async (options = {}) => {
    const requestSequence = requestSequenceRef.current + 1;
    requestSequenceRef.current = requestSequence;
    setDataLoadState((current) => ({
      ...current,
      isLoading: true
    }));
    let payloads;
    try {
      payloads = await Promise.all([
        fetchSystemStatus(),
        fetchRobots(),
        fetchMissions()
      ]);
    } catch (error) {
      if (requestSequence !== requestSequenceRef.current || options.isCancelled?.()) {
        return false;
      }
      setDataLoadState((current) => ({
        ...current,
        isLoading: false
      }));
      throw error;
    }
    if (requestSequence !== requestSequenceRef.current || options.isCancelled?.()) {
      return false;
    }
    const [statusPayload, robotPayload, missionPayload] = payloads;
    const nextMissions = missionPayload.missions ?? [];
    const liveStatusResults = await Promise.allSettled(
      nextMissions
        .filter((mission) => mission.status === "active")
        .map(async (mission) => [mission.missionCode, await fetchMissionLiveStatus(mission.missionCode)])
    );
    if (requestSequence !== requestSequenceRef.current || options.isCancelled?.()) {
      return false;
    }
    const nextLiveStatuses = {};
    liveStatusResults.forEach((result) => {
      if (result.status === "fulfilled") {
        const [missionCode, liveStatus] = result.value;
        nextLiveStatuses[missionCode] = liveStatus;
      }
    });
    setSystemStatus(statusPayload);
    setRobots(robotPayload.robots ?? []);
    setMissions(nextMissions);
    setMissionLiveStatuses(nextLiveStatuses);
    setStatusError("");
    setDataLoadState({
      hasLoaded: true,
      isLoading: false
    });
    return true;
  }, []);

  const loadMissionLiveStatus = useCallback(async (missionCode, options = {}) => {
    if (!missionCode) {
      return null;
    }
    const payload = await fetchMissionLiveStatus(missionCode);
    if (options.isCancelled?.()) {
      return null;
    }
    setMissionLiveStatuses((current) => ({
      ...current,
      [missionCode]: payload
    }));
    return payload;
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
    missionLiveStatuses,
    setMissionLiveStatuses,
    statusError,
    setStatusError,
    dataLoadState: {
      ...dataLoadState,
      isInitialLoading: dataLoadState.isLoading && !dataLoadState.hasLoaded,
      isRefreshing: dataLoadState.isLoading && dataLoadState.hasLoaded
    },
    loadAll,
    loadMissionLiveStatus
  };
}
