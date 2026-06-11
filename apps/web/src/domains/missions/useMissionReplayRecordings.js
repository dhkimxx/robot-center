import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  fetchMissionRecordingChunks,
  fetchMissionRecordingSummary
} from "../../api/recordingsApi.js";
import { makeRecordingSessionGroups } from "../recordings/recordingHelpers.js";
import {
  createEmptyReplayChunkState,
  createEmptyReplaySummaryState,
  replayAutoRefreshIntervalMs,
  replayChunkPageSize,
  shouldAutoRefreshReplayRecordings
} from "./missionReplayHelpers.js";

export function useMissionReplayRecordings(missionCode, missionStatus = "") {
  const [selectedRobotCode, setSelectedRobotCode] = useState("");
  const [summaryState, setSummaryState] = useState(() => createEmptyReplaySummaryState());
  const [chunkState, setChunkState] = useState(() => createEmptyReplayChunkState());
  const summaryRequestIDRef = useRef(0);
  const chunkRequestIDRef = useRef(0);

  const loadSummary = useCallback(async ({ silent = false } = {}) => {
    if (!missionCode) {
      setSummaryState(createEmptyReplaySummaryState());
      setSelectedRobotCode("");
      return;
    }

    const requestID = summaryRequestIDRef.current + 1;
    summaryRequestIDRef.current = requestID;
    setSummaryState((current) => (
      silent
        ? { ...current, error: "", refreshing: true }
        : { ...createEmptyReplaySummaryState(), status: "loading" }
    ));
    try {
      const payload = await fetchMissionRecordingSummary(missionCode);
      if (summaryRequestIDRef.current !== requestID) {
        return;
      }
      const robots = payload.robots ?? [];
      setSummaryState({
        error: "",
        loadedAt: new Date().toISOString(),
        refreshing: false,
        status: "success",
        summary: {
          ...payload,
          robots
        }
      });
      setSelectedRobotCode((currentRobotCode) => (
        robots.some((robot) => robot.robotCode === currentRobotCode)
          ? currentRobotCode
          : robots[0]?.robotCode ?? ""
      ));
    } catch (error) {
      if (summaryRequestIDRef.current !== requestID) {
        return;
      }
      setSummaryState((current) => ({
        error: error instanceof Error ? error.message : "recording summary load failed",
        loadedAt: current.loadedAt,
        refreshing: false,
        status: silent && current.summary ? "success" : "error",
        summary: silent ? current.summary : null
      }));
      if (!silent) {
        setSelectedRobotCode("");
      }
    }
  }, [missionCode]);

  const loadChunks = useCallback(async ({ limit = replayChunkPageSize, robotCode = selectedRobotCode, silent = false } = {}) => {
    if (!missionCode || !robotCode) {
      setChunkState(createEmptyReplayChunkState());
      return;
    }

    const requestID = chunkRequestIDRef.current + 1;
    chunkRequestIDRef.current = requestID;
    setChunkState((current) => (
      silent
        ? { ...current, error: "", refreshing: true }
        : {
          chunks: [],
          error: "",
          loadedAt: "",
          page: null,
          refreshing: false,
          robotCode,
          status: "loading"
        }
    ));
    try {
      const payload = await fetchMissionRecordingChunks(missionCode, {
        limit: Math.max(1, limit),
        offset: 0,
        robotCode
      });
      if (chunkRequestIDRef.current !== requestID) {
        return;
      }
      setChunkState({
        chunks: payload.recordings ?? [],
        error: "",
        loadedAt: new Date().toISOString(),
        page: payload.page ?? null,
        refreshing: false,
        robotCode,
        status: "success"
      });
    } catch (error) {
      if (chunkRequestIDRef.current !== requestID) {
        return;
      }
      setChunkState((current) => ({
        chunks: silent ? current.chunks : [],
        error: error instanceof Error ? error.message : "recording chunks load failed",
        loadedAt: current.loadedAt,
        page: silent ? current.page : null,
        refreshing: false,
        robotCode,
        status: silent && current.chunks.length > 0 ? "success" : "error"
      }));
    }
  }, [missionCode, selectedRobotCode]);

  useEffect(() => {
    void loadSummary({ silent: false });
  }, [loadSummary]);

  useEffect(() => {
    void loadChunks({ silent: false });
  }, [loadChunks]);

  const robotSummaries = summaryState.summary?.robots ?? [];
  const selectedRobotSummary = robotSummaries.find((robot) => robot.robotCode === selectedRobotCode) ?? robotSummaries[0] ?? null;
  const sessionGroups = useMemo(() => makeRecordingSessionGroups(chunkState.chunks), [chunkState.chunks]);
  const isSummaryLoading = summaryState.status === "loading";
  const isChunkLoading = chunkState.status === "loading";
  const isLoadingMore = chunkState.status === "loadingMore";
  const autoRefreshEnabled = shouldAutoRefreshReplayRecordings({ missionStatus, selectedRobotSummary });

  const reloadSummary = useCallback(() => {
    void loadSummary({ silent: false });
  }, [loadSummary]);

  const reloadChunks = useCallback(() => {
    void loadChunks({ silent: false });
  }, [loadChunks]);

  useEffect(() => {
    if (!missionCode || !autoRefreshEnabled || typeof window === "undefined") {
      return undefined;
    }
    const intervalID = window.setInterval(() => {
      void loadSummary({ silent: true });
      if (selectedRobotCode) {
        void loadChunks({
          limit: Math.max(replayChunkPageSize, chunkState.chunks.length),
          robotCode: selectedRobotCode,
          silent: true
        });
      }
    }, replayAutoRefreshIntervalMs);
    return () => {
      window.clearInterval(intervalID);
    };
  }, [autoRefreshEnabled, chunkState.chunks.length, loadChunks, loadSummary, missionCode, selectedRobotCode]);

  const loadMoreChunks = useCallback(async () => {
    if (!missionCode || !selectedRobotSummary || !chunkState.page?.hasMore || isLoadingMore) {
      return;
    }
    const targetRobotCode = selectedRobotSummary.robotCode;
    const nextOffset = chunkState.page.nextOffset ?? chunkState.chunks.length;
    setChunkState((current) => ({
      ...current,
      error: "",
      refreshing: false,
      status: "loadingMore"
    }));
    try {
      const payload = await fetchMissionRecordingChunks(missionCode, {
        limit: replayChunkPageSize,
        offset: nextOffset,
        robotCode: targetRobotCode
      });
      setChunkState((current) => {
        if (current.robotCode !== targetRobotCode) {
          return current;
        }
        return {
          chunks: [...current.chunks, ...(payload.recordings ?? [])],
          error: "",
          loadedAt: new Date().toISOString(),
          page: payload.page ?? null,
          refreshing: false,
          robotCode: targetRobotCode,
          status: "success"
        };
      });
    } catch (error) {
      setChunkState((current) => {
        if (current.robotCode !== targetRobotCode) {
          return current;
        }
        return {
          ...current,
          error: error instanceof Error ? error.message : "recording chunks load failed",
          refreshing: false,
          status: "error"
        };
      });
    }
  }, [chunkState.chunks.length, chunkState.page, isLoadingMore, missionCode, selectedRobotSummary]);

  return {
    autoRefreshEnabled,
    chunkState,
    isChunkLoading,
    isLoadingMore,
    isSummaryLoading,
    loadMoreChunks,
    reloadChunks,
    reloadSummary,
    robotSummaries,
    selectedRobotCode,
    selectedRobotSummary,
    sessionGroups,
    setSelectedRobotCode,
    summaryState
  };
}
