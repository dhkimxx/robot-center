import { useCallback, useEffect, useMemo, useState } from "react";
import {
  fetchMissionRecordingChunks,
  fetchMissionRecordingSummary
} from "../../api/recordingsApi.js";
import { makeRecordingSessionGroups } from "../recordings/recordingHelpers.js";
import {
  createEmptyReplayChunkState,
  createEmptyReplaySummaryState,
  replayChunkPageSize
} from "./missionReplayHelpers.js";

export function useMissionReplayRecordings(missionCode) {
  const [selectedRobotCode, setSelectedRobotCode] = useState("");
  const [summaryReloadKey, setSummaryReloadKey] = useState(0);
  const [chunkReloadKey, setChunkReloadKey] = useState(0);
  const [summaryState, setSummaryState] = useState(() => createEmptyReplaySummaryState());
  const [chunkState, setChunkState] = useState(() => createEmptyReplayChunkState());

  useEffect(() => {
    if (!missionCode) {
      setSummaryState(createEmptyReplaySummaryState());
      setSelectedRobotCode("");
      return undefined;
    }

    let cancelled = false;
    setSummaryState({ error: "", status: "loading", summary: null });
    fetchMissionRecordingSummary(missionCode)
      .then((payload) => {
        if (cancelled) {
          return;
        }
        const robots = payload.robots ?? [];
        setSummaryState({
          error: "",
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
      })
      .catch((error) => {
        if (cancelled) {
          return;
        }
        setSummaryState({
          error: error instanceof Error ? error.message : "recording summary load failed",
          status: "error",
          summary: null
        });
        setSelectedRobotCode("");
      });

    return () => {
      cancelled = true;
    };
  }, [missionCode, summaryReloadKey]);

  useEffect(() => {
    if (!missionCode || !selectedRobotCode) {
      setChunkState(createEmptyReplayChunkState());
      return undefined;
    }

    let cancelled = false;
    setChunkState({
      chunks: [],
      error: "",
      page: null,
      robotCode: selectedRobotCode,
      status: "loading"
    });
    fetchMissionRecordingChunks(missionCode, {
      limit: replayChunkPageSize,
      offset: 0,
      robotCode: selectedRobotCode
    })
      .then((payload) => {
        if (cancelled) {
          return;
        }
        setChunkState({
          chunks: payload.recordings ?? [],
          error: "",
          page: payload.page ?? null,
          robotCode: selectedRobotCode,
          status: "success"
        });
      })
      .catch((error) => {
        if (cancelled) {
          return;
        }
        setChunkState({
          chunks: [],
          error: error instanceof Error ? error.message : "recording chunks load failed",
          page: null,
          robotCode: selectedRobotCode,
          status: "error"
        });
      });

    return () => {
      cancelled = true;
    };
  }, [chunkReloadKey, missionCode, selectedRobotCode]);

  const robotSummaries = summaryState.summary?.robots ?? [];
  const selectedRobotSummary = robotSummaries.find((robot) => robot.robotCode === selectedRobotCode) ?? robotSummaries[0] ?? null;
  const sessionGroups = useMemo(() => makeRecordingSessionGroups(chunkState.chunks), [chunkState.chunks]);
  const isSummaryLoading = summaryState.status === "loading";
  const isChunkLoading = chunkState.status === "loading";
  const isLoadingMore = chunkState.status === "loadingMore";

  const reloadSummary = useCallback(() => {
    setSummaryReloadKey((value) => value + 1);
  }, []);

  const reloadChunks = useCallback(() => {
    setChunkReloadKey((value) => value + 1);
  }, []);

  const loadMoreChunks = useCallback(async () => {
    if (!missionCode || !selectedRobotSummary || !chunkState.page?.hasMore || isLoadingMore) {
      return;
    }
    const targetRobotCode = selectedRobotSummary.robotCode;
    const nextOffset = chunkState.page.nextOffset ?? chunkState.chunks.length;
    setChunkState((current) => ({
      ...current,
      error: "",
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
          page: payload.page ?? null,
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
          status: "error"
        };
      });
    }
  }, [chunkState.chunks.length, chunkState.page, isLoadingMore, missionCode, selectedRobotSummary]);

  return {
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
