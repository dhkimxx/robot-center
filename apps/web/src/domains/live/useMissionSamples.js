import { useCallback, useEffect, useState } from "react";
import { fetchSensorLatest } from "../../api/liveApi.js";

export function useMissionSamples({ appendLiveEvent, selectedLiveTarget }) {
  const [serverSensorLatest, setServerSensorLatest] = useState([]);
  const [refreshSequence, setRefreshSequence] = useState(0);
  const [sensorSnapshotState, setSensorSnapshotState] = useState({
    error: "",
    loadedAt: "",
    status: "idle"
  });
  const targetKey = selectedLiveTarget?.key ?? "";
  const missionId = selectedLiveTarget?.mission?.id ?? "";
  const robotCode = selectedLiveTarget?.robotCode ?? "";

  const refreshSensorSnapshot = useCallback(() => {
    if (targetKey && missionId) {
      setRefreshSequence((sequence) => sequence + 1);
    }
  }, [missionId, targetKey]);

  useEffect(() => {
    let cancelled = false;
    async function loadMissionSnapshot() {
      if (!targetKey || !missionId) {
        setServerSensorLatest([]);
        setSensorSnapshotState({
          error: "",
          loadedAt: "",
          status: "idle"
        });
        return;
      }
      setServerSensorLatest([]);
      setSensorSnapshotState({
        error: "",
        loadedAt: "",
        status: "loading"
      });
      try {
        const sensorLatestPayload = await fetchSensorLatest(missionId, robotCode);
        if (!cancelled) {
          setServerSensorLatest(sensorLatestPayload.sensors ?? []);
          setSensorSnapshotState({
            error: "",
            loadedAt: new Date().toISOString(),
            status: "ready"
          });
        }
      } catch (error) {
        if (!cancelled) {
          const message = error instanceof Error ? error.message : "unknown";
          setSensorSnapshotState({
            error: message,
            loadedAt: "",
            status: "error"
          });
          appendLiveEvent(targetKey, `최근 저장 센서값 조회 실패: ${message}`);
        }
      }
    }
    loadMissionSnapshot();
    return () => {
      cancelled = true;
    };
  }, [appendLiveEvent, missionId, refreshSequence, robotCode, targetKey]);

  return {
    refreshSensorSnapshot,
    sensorSnapshotState,
    serverSensorLatest
  };
}
