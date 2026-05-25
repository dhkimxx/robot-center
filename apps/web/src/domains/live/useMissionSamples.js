import { useEffect, useState } from "react";
import {
  fetchSensorReadings,
  fetchTelemetrySamples
} from "../../api/liveApi.js";

export function useMissionSamples({ appendLiveEvent, selectedLiveTarget }) {
  const [serverTelemetry, setServerTelemetry] = useState([]);
  const [serverSensors, setServerSensors] = useState([]);

  useEffect(() => {
    let cancelled = false;
    async function loadMissionSamples() {
      if (!selectedLiveTarget) {
        setServerTelemetry([]);
        setServerSensors([]);
        return;
      }
      try {
        const [telemetryPayload, sensorPayload] = await Promise.all([
          fetchTelemetrySamples(selectedLiveTarget.mission.id),
          fetchSensorReadings(selectedLiveTarget.mission.id)
        ]);
        if (!cancelled) {
          setServerTelemetry(telemetryPayload.telemetry ?? []);
          setServerSensors(sensorPayload.sensorReadings ?? []);
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
    serverTelemetry,
    serverSensors
  };
}
