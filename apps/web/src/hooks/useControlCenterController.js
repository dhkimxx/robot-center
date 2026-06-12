import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { createMissionRobotTargets } from "../domains/missions/missionHelpers.js";
import {
  createMissionControlPageChrome,
  createMissionListPageChrome,
  createMissionReplayPageChrome
} from "../domains/missions/missionPageChrome.jsx";
import { useMissionManagementController } from "../domains/missions/useMissionManagementController.js";
import { useRecordingsController } from "../domains/recordings/useRecordingsController.js";
import { useRobotManagementController } from "../domains/robots/useRobotManagementController.js";
import { createRobotPageChrome } from "../domains/robots/robotPageChrome.jsx";
import { createSystemPageChrome } from "../domains/system/systemPageChrome.js";
import { useLiveConnectionManager } from "../domains/live/useLiveConnectionManager.js";
import { resolveStoredLiveTargetKey } from "../domains/live/useLiveTargetSelection.js";
import { useMissionSamples } from "../domains/live/useMissionSamples.js";
import {
  createSensorPanelState,
  createSensorPanelSnapshot,
  createTelemetryFromSensorLatest
} from "../domains/live/sensorLatestMapper.js";
import {
  clearEventData,
  clearObjectStorage,
  clearRecorderRuntime,
  clearSensorData,
  pruneObjectStorage,
  pruneRecorderRuntime
} from "../api/systemApi.js";
import { useControlCenterData } from "./useControlCenterData.js";
import { useNotifications } from "./useNotifications.js";

export function useControlCenterController({
  activeSection = "missions",
  missionControlCode: routeMissionControlCode = "",
  missionReplayCode: routeMissionReplayCode = "",
  selectedMissionCode: routeSelectedMissionCode = "",
  navigateToPath = null
} = {}) {
  const shouldPollActiveMissionLiveStatuses = !routeMissionControlCode && !routeMissionReplayCode;
  const {
    systemStatus,
    robots,
    missions,
    missionLiveStatuses,
    statusError,
    dataLoadState,
    loadAll,
    loadMissionLiveStatus
  } = useControlCenterData({
    includeActiveMissionLiveStatuses: shouldPollActiveMissionLiveStatuses
  });
  const {
    notifications,
    showNotification,
    dismissNotification
  } = useNotifications();
  const [connectionInfo, setConnectionInfo] = useState(null);
  const [missionControlCode, setMissionControlCode] = useState("");
  const previousRouteMissionControlCodeRef = useRef("");

  const activeMissions = useMemo(
    () => missions.filter((mission) => mission.status === "active"),
    [missions]
  );
  const robotController = useRobotManagementController({
    connectionInfo,
    loadAll,
    robots,
    setConnectionInfo,
    showNotification
  });
  const missionControlMission = useMemo(
    () => missions.find((mission) => mission.missionCode === missionControlCode) ?? null,
    [missionControlCode, missions]
  );
  const missionControlLiveStatus = missionControlCode ? missionLiveStatuses[missionControlCode] ?? null : null;
  const missionReplayMission = useMemo(
    () => missions.find((mission) => mission.missionCode === routeMissionReplayCode) ?? null,
    [missions, routeMissionReplayCode]
  );
  const activeLiveTargets = useMemo(
    () => activeMissions
      .flatMap((mission) => createMissionRobotTargets(mission, robots, missionLiveStatuses[mission.missionCode] ?? null)),
    [activeMissions, missionLiveStatuses, robots]
  );
  const missionControlTargets = useMemo(() => {
    if (!missionControlMission) {
      return [];
    }
    return createMissionRobotTargets(missionControlMission, robots, missionControlLiveStatus);
  }, [missionControlLiveStatus, missionControlMission, robots]);
  const liveTargets = useMemo(
    () => (missionControlMission ? missionControlTargets : activeLiveTargets),
    [activeLiveTargets, missionControlMission, missionControlTargets]
  );
  const liveConnectionManager = useLiveConnectionManager({
    activeSection,
    liveTargets,
    missionControlMission,
    setMissionControlCode,
    showNotification
  });
  const {
    appendLiveEvent,
    disconnectMissionByCode,
    liveSessions,
    reconnectLive,
    selectedLiveSession,
    selectedLiveTarget,
    selectedLiveTargetKey,
    setSelectedLiveTargetKey
  } = liveConnectionManager;
  const {
    refreshSensorSnapshot,
    sensorSnapshotState,
    serverSensorLatest
  } = useMissionSamples({ appendLiveEvent, selectedLiveTarget });
  const selectedRobotCode = selectedLiveTarget?.robotCode ?? "";
  const latestServerTelemetryFromSensors = useMemo(
    () => createTelemetryFromSensorLatest(serverSensorLatest, selectedRobotCode),
    [selectedRobotCode, serverSensorLatest]
  );
  const latestServerSensorPanel = useMemo(
    () => createSensorPanelSnapshot(serverSensorLatest, selectedRobotCode),
    [selectedRobotCode, serverSensorLatest]
  );
  const sensorPanelState = useMemo(
    () => createSensorPanelState({
      liveSensor: selectedLiveSession.sensor,
      snapshotSensor: latestServerSensorPanel,
      snapshotState: sensorSnapshotState
    }),
    [latestServerSensorPanel, selectedLiveSession.sensor, sensorSnapshotState]
  );
  const latestTelemetry = selectedLiveSession.telemetry ?? latestServerTelemetryFromSensors;
  const latestSensor = sensorPanelState.sensor;
  const recordingsController = useRecordingsController();

  const clearSystemObjectStorage = useCallback(async () => {
    try {
      const payload = await clearObjectStorage();
      const result = payload.objectStorage ?? {};
      showNotification(
        `객체 스토리지 파일 ${result.deletedObjectCount ?? 0}개, ${formatByteCount(result.deletedBytes ?? 0)} 삭제 완료`,
        "success"
      );
      await loadAll();
      return result;
    } catch (error) {
      showNotification(error instanceof Error ? error.message : "Object storage 정리 실패", "danger");
      throw error;
    }
  }, [loadAll, showNotification]);

  const pruneSystemObjectStorage = useCallback(async () => {
    try {
      const payload = await pruneObjectStorage();
      const result = payload.objectStorage ?? {};
      showNotification(
        `운영 중 객체 스토리지 파일 ${result.deletedObjectCount ?? 0}개, ${formatByteCount(result.deletedBytes ?? 0)} 정리 완료`,
        "success"
      );
      await loadAll();
      return result;
    } catch (error) {
      showNotification(error instanceof Error ? error.message : "Object storage 운영 중 정리 실패", "danger");
      throw error;
    }
  }, [loadAll, showNotification]);

  const clearSystemSensorData = useCallback(async () => {
    try {
      const payload = await clearSensorData();
      const result = payload.sensorData ?? {};
      showNotification(
        `센서 데이터 샘플 ${result.sensorSamplesDeleted ?? 0}개, 정의 ${result.sensorDescriptorsDeleted ?? 0}개 삭제 완료`,
        "success"
      );
      await loadAll();
      refreshSensorSnapshot();
      return result;
    } catch (error) {
      showNotification(error instanceof Error ? error.message : "Sensor 데이터 정리 실패", "danger");
      throw error;
    }
  }, [loadAll, refreshSensorSnapshot, showNotification]);

  const clearSystemEventData = useCallback(async () => {
    try {
      const payload = await clearEventData();
      const result = payload.eventData ?? {};
      showNotification(
        `이벤트 데이터 ${result.eventsDeleted ?? 0}건 삭제 완료`,
        "success"
      );
      await loadAll();
      return result;
    } catch (error) {
      showNotification(error instanceof Error ? error.message : "Event 데이터 정리 실패", "danger");
      throw error;
    }
  }, [loadAll, showNotification]);

  const clearSystemRecorderRuntime = useCallback(async () => {
    try {
      const payload = await clearRecorderRuntime();
      const result = payload.recorderRuntime ?? {};
      showNotification(
        `녹화 런타임 파일 ${result.filesDeleted ?? 0}개, ${formatByteCount(result.deletedBytes ?? 0)} 삭제 완료`,
        "success"
      );
      await loadAll();
      return result;
    } catch (error) {
      showNotification(error instanceof Error ? error.message : "녹화 런타임 정리 실패", "danger");
      throw error;
    }
  }, [loadAll, showNotification]);

  const pruneSystemRecorderRuntime = useCallback(async () => {
    try {
      const payload = await pruneRecorderRuntime();
      const result = payload.recorderRuntime ?? {};
      showNotification(
        `운영 중 녹화 런타임 파일 ${result.filesDeleted ?? 0}개, ${formatByteCount(result.deletedBytes ?? 0)} 정리 완료`,
        "success"
      );
      await loadAll();
      return result;
    } catch (error) {
      showNotification(error instanceof Error ? error.message : "녹화 런타임 운영 중 정리 실패", "danger");
      throw error;
    }
  }, [loadAll, showNotification]);

  useEffect(() => {
    const previousMissionCode = previousRouteMissionControlCodeRef.current;
    if (previousMissionCode && previousMissionCode !== routeMissionControlCode) {
      disconnectMissionByCode(previousMissionCode);
    }
    previousRouteMissionControlCodeRef.current = routeMissionControlCode;
    setMissionControlCode(routeMissionControlCode || "");
  }, [disconnectMissionByCode, routeMissionControlCode]);

  const missionController = useMissionManagementController({
    disconnectMissionByCode,
    loadAll,
    missions,
    navigateToPath,
    resolveStoredLiveTargetKey,
    robots,
    routeSelectedMissionCode,
    setMissionControlCode,
    setSelectedLiveTargetKey,
    showNotification,
  });

  const closeMissionControl = useCallback((options = {}) => {
    missionController.closeMissionControl(missionControlCode, options);
  }, [missionControlCode, missionController]);

  useEffect(() => {
    if (dataLoadState.hasLoaded && missionControlCode && !missions.some((mission) => mission.missionCode === missionControlCode)) {
      setMissionControlCode("");
    }
  }, [dataLoadState.hasLoaded, missionControlCode, missions]);

  useEffect(() => {
    if (!missionControlCode) {
      return undefined;
    }
    let cancelled = false;
    let timer = null;

    async function loadLiveStatus() {
      try {
        await loadMissionLiveStatus(missionControlCode, { isCancelled: () => cancelled });
      } catch {
        // The live screen can still render SFU/session state if the aggregate
        // endpoint is briefly unavailable. Polling will retry shortly.
      }
      if (!cancelled) {
        timer = window.setTimeout(loadLiveStatus, 3000);
      }
    }

    loadLiveStatus();
    return () => {
      cancelled = true;
      if (timer) {
        window.clearTimeout(timer);
      }
    };
  }, [loadMissionLiveStatus, missionControlCode]);

  const pageChrome = useMemo(() => {
    if (routeMissionControlCode) {
      return createMissionControlPageChrome({
        controlMission: missionControlMission,
        isLoading: dataLoadState.isInitialLoading,
        missionTargets: missionControlTargets,
        onBackToMissionList: closeMissionControl,
        onEndMission: missionController.endMission,
        onStartMission: missionController.startMission,
        routeMissionControlCode
      });
    }

    if (routeMissionReplayCode) {
      return createMissionReplayPageChrome({
        isLoading: dataLoadState.isInitialLoading,
        replayMission: missionReplayMission,
        routeMissionReplayCode
      });
    }

    if (activeSection === "robots") {
      return createRobotPageChrome({
        isLoading: dataLoadState.isInitialLoading,
        liveStatuses: missionLiveStatuses,
        onOpenCreateRobotModal: robotController.openRobotCreateModal,
        robots
      });
    }

    if (activeSection === "system") {
      return createSystemPageChrome({
        isLoading: dataLoadState.isInitialLoading,
        liveStatuses: missionLiveStatuses,
        systemStatus
      });
    }

    return createMissionListPageChrome({
      isLoading: dataLoadState.isInitialLoading,
      liveStatuses: missionLiveStatuses,
      missions,
      onOpenCreateMissionModal: missionController.openMissionCreateModal
    });
  }, [
    activeSection,
    closeMissionControl,
    dataLoadState.isInitialLoading,
    missionControlMission,
    missionControlTargets,
    missionController.endMission,
    missionController.openMissionCreateModal,
    missionController.startMission,
    missionLiveStatuses,
    missionReplayMission,
    missions,
    robotController.openRobotCreateModal,
    robots,
    routeMissionControlCode,
    routeMissionReplayCode,
    systemStatus
  ]);

  return {
    statusError,
    dataLoadState,
    notifications,
    dismissNotification,
    pageChrome,
    missionRouteProps: {
      controlMission: missionControlMission,
      controlMissionCode: routeMissionControlCode,
      dataLoadState,
      isSensorSnapshotRefreshing: sensorSnapshotState.status === "loading",
      latestSensor,
      latestSensorSourceLabel: sensorPanelState.sourceLabel,
      latestTelemetry,
      liveEvents: selectedLiveSession.events,
      liveStatus: missionControlLiveStatus,
      liveStatuses: missionLiveStatuses,
      liveSessions,
      missionTargets: missionControlTargets,
      missions,
      onBackToMissionList: closeMissionControl,
      onEndMission: missionController.endMission,
      onOpenCreateMissionModal: missionController.openMissionCreateModal,
      onOpenMissionControl: missionController.openMissionControl,
      onOpenMissionReplay: missionController.openMissionReplay,
      onOpenPlaybackFile: recordingsController.setRecordingPlaybackFile,
      onReconnectSelectedMissionTarget: reconnectLive,
      onRefreshSensorSnapshot: refreshSensorSnapshot,
      onSelectMission: missionController.setSelectedMissionManagementCode,
      onStartMission: missionController.startMission,
      replayMission: missionReplayMission,
      replayMissionCode: routeMissionReplayCode,
      robots,
      selectedMission: missionController.selectedMission,
      selectedMissionTargetKey: selectedLiveTargetKey,
      setSelectedMissionTargetKey: setSelectedLiveTargetKey
    },
    missionModalProps: {
      createMission: missionController.createMission,
      missionForm: missionController.missionForm,
      missionModal: missionController.missionModal,
      missions,
      onClose: missionController.closeMissionModal,
      robots,
      setMissionForm: missionController.setMissionForm
    },
    playbackModalProps: {
      recordingPlaybackFile: recordingsController.recordingPlaybackFile,
      setRecordingPlaybackFile: recordingsController.setRecordingPlaybackFile
    },
    robotModalProps: {
      closeRobotModal: robotController.closeRobotModal,
      confirmArchiveRobot: robotController.confirmArchiveRobot,
      connectionInfo,
      createRobot: robotController.createRobot,
      pendingArchiveRobot: robotController.pendingArchiveRobot,
      pendingArchiveRobotCode: robotController.pendingArchiveRobotCode,
      robotEditForm: robotController.robotEditForm,
      robotForm: robotController.robotForm,
      robotModal: robotController.robotModal,
      rotateRobotToken: robotController.rotateRobotToken,
      selectedRobot: robotController.selectedRobot,
      setPendingArchiveRobotCode: robotController.setPendingArchiveRobotCode,
      setRobotEditForm: robotController.setRobotEditForm,
      setRobotForm: robotController.setRobotForm,
      updateRobot: robotController.updateRobot
    },
    robotRouteProps: {
      missions,
      dataLoadState,
      onArchiveRobot: robotController.archiveRobot,
      onLoadConnectionInfo: robotController.loadConnectionInfo,
      onOpenCreateRobotModal: robotController.openRobotCreateModal,
      onOpenEditRobotModal: robotController.openRobotEditModal,
      onSelectRobot: robotController.setSelectedRobotManagementCode,
      robots,
      selectedRobot: robotController.selectedRobot
    },
    systemRouteProps: {
      onClearEventData: clearSystemEventData,
      onClearObjectStorage: clearSystemObjectStorage,
      onClearRecorderRuntime: clearSystemRecorderRuntime,
      onClearSensorData: clearSystemSensorData,
      onPruneObjectStorage: pruneSystemObjectStorage,
      onPruneRecorderRuntime: pruneSystemRecorderRuntime,
      dataLoadState,
      statusError,
      systemStatus
    }
  };
}

function formatByteCount(bytes) {
  if (!Number.isFinite(bytes) || bytes <= 0) {
    return "0B";
  }
  const units = ["B", "KB", "MB", "GB", "TB"];
  let value = bytes;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }
  const fractionDigits = value >= 10 || unitIndex === 0 ? 0 : 1;
  return `${value.toFixed(fractionDigits)}${units[unitIndex]}`;
}
