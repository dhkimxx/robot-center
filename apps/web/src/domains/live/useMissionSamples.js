import { useEffect, useState } from "react";
import { fetchSensorLatest } from "../../api/liveApi.js";

export function useMissionSamples({ appendLiveEvent, selectedLiveTarget }) {
  const [serverSensorLatest, setServerSensorLatest] = useState([]);

  useEffect(() => {
    let cancelled = false;
    async function loadMissionSamples() {
      if (!selectedLiveTarget) {
        setServerSensorLatest([]);
        return;
      }
      try {
        const sensorLatestPayload = await fetchSensorLatest(selectedLiveTarget.mission.id, selectedLiveTarget.robotCode);
        if (!cancelled) {
          setServerSensorLatest(sensorLatestPayload.sensors ?? []);
        }
      } catch (error) {
        if (!cancelled) {
          appendLiveEvent(selectedLiveTarget.key, `sample polling failed: ${error instanceof Error ? error.message : "unknown"}`);
        }
      }
    }
    loadMissionSamples();
    const timer = window.setInterval(loadMissionSamples, 3000);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [selectedLiveTarget, appendLiveEvent]);

  return {
    serverSensorLatest
  };
}
