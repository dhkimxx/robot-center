import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { getTelemetryPositionState } from "../utils/formatters.js";
import { createMissionRobotTargets } from "../domains/missions/missionHelpers.js";
import {
  createRecordingPlaybackFile,
  findLatestRecordingForTarget
} from "../domains/recordings/recordingHelpers.js";
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
    streamingStatuses,
    recordings,
    statusError,
    loadAll
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
  const missionReplayMission = useMemo(
    () => missions.find((mission) => mission.missionCode === routeMissionReplayCode) ?? null,
    [missions, routeMissionReplayCode]
  );
  const activeLiveTargets = useMemo(
    () => activeMissions
      .flatMap((mission) => createMissionRobotTargets(mission, robots, streamingStatuses)),
    [activeMissions, robots, streamingStatuses]
  );
  const missionControlTargets = useMemo(() => {
    if (!missionControlMission) {
      return [];
    }
    return createMissionRobotTargets(missionControlMission, robots, streamingStatuses);
  }, [missionControlMission, robots, streamingStatuses]);
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
  const activeStreamingStatus = useMemo(() => {
    if (!selectedLiveTarget) {
      return null;
    }
    return selectedLiveTarget.streamingStatus;
  }, [selectedLiveTarget]);
  const selectedMissionCode = selectedLiveTarget?.mission?.missionCode ?? "";
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
  const latestRecording = useMemo(
    () => findLatestRecordingForTarget(recordings, selectedMissionCode, selectedRobotCode),
    [recordings, selectedMissionCode, selectedRobotCode]
  );
  const latestPlayableRecording = useMemo(
    () => findLatestRecordingForTarget(
      recordings,
      selectedMissionCode,
      selectedRobotCode,
      (recording) => Boolean(createRecordingPlaybackFile(recording))
    ),
    [recordings, selectedMissionCode, selectedRobotCode]
  );
  const recordingsController = useRecordingsController({
    latestPlayableRecording,
    showNotification
  });

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
    streamingStatuses
  });

  const closeMissionControl = useCallback((options = {}) => {
    missionController.closeMissionControl(missionControlCode, options);
  }, [missionControlCode, missionController]);

  useEffect(() => {
    if (missionControlCode && !missions.some((mission) => mission.missionCode === missionControlCode)) {
      setMissionControlCode("");
    }
  }, [missionControlCode, missions]);

  const operationStatuses = useOperationStatuses({
    activeStreamingStatus,
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
      latestRecording,
      latestSensor,
      latestTelemetry,
      liveEvents: selectedLiveSession.events,
      liveSessions,
      missionTargets: missionControlTargets,
      missions,
      onBackToMissionList: closeMissionControl,
      onEndMission: missionController.endMission,
      onOpenCreateMissionModal: missionController.openMissionCreateModal,
      onOpenMissionControl: missionController.openMissionControl,
      onOpenMissionReplay: missionController.openMissionReplay,
      onOpenPlaybackFile: recordingsController.setRecordingPlaybackFile,
      onPlayLatestRecording: recordingsController.playLatestRecording,
      onReconnectSelectedMissionTarget: reconnectLive,
      onSelectMission: missionController.setSelectedMissionManagementCode,
      onStartMission: missionController.startMission,
      operationStatuses,
      playbackRecording: latestPlayableRecording,
      recordings,
      replayMission: missionReplayMission,
      replayMissionCode: routeMissionReplayCode,
      robots,
      selectedMission: missionController.selectedMission,
      selectedMissionTargetKey: selectedLiveTargetKey,
      setSelectedMissionTargetKey: setSelectedLiveTargetKey,
      streamingStatuses
    },
    missionModalProps: {
      createMission: missionController.createMission,
      missionForm: missionController.missionForm,
      missionModal: missionController.missionModal,
      missions,
      onClose: missionController.closeMissionModal,
      robots,
      setMissionForm: missionController.setMissionForm,
      streamingStatuses
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
