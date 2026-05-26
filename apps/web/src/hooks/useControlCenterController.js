import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { getTelemetryPositionState } from "../utils/formatters.js";
import { createMissionRobotTargets } from "../domains/missions/missionHelpers.js";
import { useMissionManagementController } from "../domains/missions/useMissionManagementController.js";
import { useRecordingsController } from "../domains/recordings/useRecordingsController.js";
import { useRobotManagementController } from "../domains/robots/useRobotManagementController.js";
import { useLiveConnectionManager } from "../domains/live/useLiveConnectionManager.js";
import { resolveStoredLiveTargetKey } from "../domains/live/useLiveTargetSelection.js";
import { useMissionSamples } from "../domains/live/useMissionSamples.js";
import {
  createSensorPanelSnapshot,
  createTelemetryFromSensorLatest
} from "../domains/live/sensorLatestMapper.js";
import { useOperationStatuses } from "../domains/live/useOperationStatuses.js";
import { useControlCenterData } from "./useControlCenterData.js";
import { useNotifications } from "./useNotifications.js";

export function useControlCenterController({
  activeSection = "missions",
  missionControlCode: routeMissionControlCode = "",
  missionReplayCode: routeMissionReplayCode = "",
  selectedMissionCode: routeSelectedMissionCode = "",
  navigateToPath = null
} = {}) {
  const {
    systemStatus,
    robots,
    missions,
    missionLiveStatuses,
    observedStreams,
    recordings,
    statusError,
    loadAll,
    loadMissionLiveStatus
  } = useControlCenterData();
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
  const primaryRobot = useMemo(() => robots[0] ?? null, [robots]);
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
      .flatMap((mission) => createMissionRobotTargets(mission, robots, observedStreams)),
    [activeMissions, observedStreams, robots]
  );
  const missionControlTargets = useMemo(() => {
    if (!missionControlMission) {
      return [];
    }
    return createMissionRobotTargets(missionControlMission, robots, observedStreams, missionControlLiveStatus);
  }, [missionControlLiveStatus, missionControlMission, observedStreams, robots]);
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
    serverSensorLatest
  } = useMissionSamples({ appendLiveEvent, selectedLiveTarget });
  const activeObservedStream = useMemo(() => {
    if (!selectedLiveTarget) {
      return null;
    }
    return selectedLiveTarget.observedPublisher;
  }, [selectedLiveTarget]);
  const selectedRobotCode = selectedLiveTarget?.robotCode ?? "";
  const latestServerTelemetryFromSensors = useMemo(
    () => createTelemetryFromSensorLatest(serverSensorLatest, selectedRobotCode),
    [selectedRobotCode, serverSensorLatest]
  );
  const latestServerSensorPanel = useMemo(
    () => createSensorPanelSnapshot(serverSensorLatest, selectedRobotCode),
    [selectedRobotCode, serverSensorLatest]
  );
  const latestTelemetry = selectedLiveSession.telemetry ?? latestServerTelemetryFromSensors;
  const latestSensor = selectedLiveSession.sensor ?? latestServerSensorPanel;
  const latestPositionState = getTelemetryPositionState(latestTelemetry);
  const recordingsController = useRecordingsController();

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
    observedStreams
  });

  const closeMissionControl = useCallback((options = {}) => {
    missionController.closeMissionControl(missionControlCode, options);
  }, [missionControlCode, missionController]);

  useEffect(() => {
    if (missionControlCode && !missions.some((mission) => mission.missionCode === missionControlCode)) {
      setMissionControlCode("");
    }
  }, [missionControlCode, missions]);

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


  const operationStatuses = useOperationStatuses({
    activeObservedStream,
    latestPositionState,
    primaryRobot,
    selectedLiveSession,
    selectedLiveTarget,
    statusError
  });

  return {
    statusError,
    notifications,
    dismissNotification,
    missionRouteProps: {
      controlMission: missionControlMission,
      latestSensor,
      latestTelemetry,
      liveEvents: selectedLiveSession.events,
      liveStatus: missionControlLiveStatus,
      liveSessions,
      missionTargets: missionControlTargets,
      missions,
      observedStreams,
      onBackToMissionList: closeMissionControl,
      onEndMission: missionController.endMission,
      onOpenCreateMissionModal: missionController.openMissionCreateModal,
      onOpenMissionControl: missionController.openMissionControl,
      onOpenMissionReplay: missionController.openMissionReplay,
      onOpenPlaybackFile: recordingsController.setRecordingPlaybackFile,
      onReconnectSelectedMissionTarget: reconnectLive,
      onSelectMission: missionController.setSelectedMissionManagementCode,
      onStartMission: missionController.startMission,
      operationStatuses,
      recordings,
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
      observedStreams,
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
      onArchiveRobot: robotController.archiveRobot,
      onLoadConnectionInfo: robotController.loadConnectionInfo,
      onOpenCreateRobotModal: robotController.openRobotCreateModal,
      onOpenEditRobotModal: robotController.openRobotEditModal,
      onSelectRobot: robotController.setSelectedRobotManagementCode,
      robots,
      selectedRobot: robotController.selectedRobot
    },
    systemRouteProps: {
      statusError,
      systemStatus
    }
  };
}
