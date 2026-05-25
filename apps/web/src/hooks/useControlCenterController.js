import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { getTelemetryPositionState } from "../utils/formatters.js";
import { findLatestSampleForRobot } from "../domains/live/liveHelpers.js";
import { createMissionRobotTargets } from "../domains/missions/missionHelpers.js";
import {
  createRecordingPlaybackFile,
  findLatestRecordingForTarget
} from "../domains/recordings/recordingHelpers.js";
import { useMissionManagementController } from "../domains/missions/useMissionManagementController.js";
import { useRecordingsController } from "../domains/recordings/useRecordingsController.js";
import { useRobotManagementController } from "../domains/robots/useRobotManagementController.js";
import {
  readSelectedLiveTargetKey,
  useLiveConnectionManager
} from "../domains/live/useLiveConnectionManager.js";
import { useMissionSamples } from "../domains/live/useMissionSamples.js";
import { useOperationStatuses } from "../domains/live/useOperationStatuses.js";
import { useControlCenterData } from "./useControlCenterData.js";
import { useNotifications } from "./useNotifications.js";

export function useControlCenterController({
  activeSection = "missions",
  missionControlCode: routeMissionControlCode = "",
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
    serverTelemetry,
    serverSensors
  } = useMissionSamples({ appendLiveEvent, selectedLiveTarget });
  const activeStreamingStatus = useMemo(() => {
    if (!selectedLiveTarget) {
      return null;
    }
    return selectedLiveTarget.streamingStatus;
  }, [selectedLiveTarget]);
  const selectedMissionCode = selectedLiveTarget?.mission?.missionCode ?? "";
  const selectedRobotCode = selectedLiveTarget?.robotCode ?? "";
  const latestServerTelemetry = useMemo(
    () => findLatestSampleForRobot(serverTelemetry, selectedRobotCode),
    [selectedRobotCode, serverTelemetry]
  );
  const latestServerSensor = useMemo(
    () => findLatestSampleForRobot(serverSensors, selectedRobotCode),
    [selectedRobotCode, serverSensors]
  );
  const latestTelemetry = selectedLiveSession.telemetry ?? latestServerTelemetry;
  const latestSensor = selectedLiveSession.sensor ?? latestServerSensor;
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
    readSelectedLiveTargetKey,
    robots,
    setMissionControlCode,
    setSelectedLiveTargetKey,
    showNotification,
    streamingStatuses
  });

  const closeMissionControl = useCallback(() => {
    missionController.closeMissionControl(missionControlCode);
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
    systemStatus,
    robots,
    missions,
    recordings,
    connectionInfo,
    statusError,
    notifications,
    robotForm: robotController.robotForm,
    setRobotForm: robotController.setRobotForm,
    selectedRobot: robotController.selectedRobot,
    robotEditForm: robotController.robotEditForm,
    setRobotEditForm: robotController.setRobotEditForm,
    missionForm: missionController.missionForm,
    setMissionForm: missionController.setMissionForm,
    recordingPlaybackFile: recordingsController.recordingPlaybackFile,
    setRecordingPlaybackFile: recordingsController.setRecordingPlaybackFile,
    missionControlMission,
    missionControlTargets,
    selectedLiveSession,
    liveSessions,
    latestRecording,
    latestPlayableRecording,
    latestTelemetry,
    latestSensor,
    operationStatuses,
    selectedMission: missionController.selectedMission,
    selectedLiveTargetKey,
    setSelectedLiveTargetKey,
    robotModal: robotController.robotModal,
    missionModal: missionController.missionModal,
    pendingArchiveRobotCode: robotController.pendingArchiveRobotCode,
    pendingArchiveRobot: robotController.pendingArchiveRobot,
    setPendingArchiveRobotCode: robotController.setPendingArchiveRobotCode,
    archiveRobot: robotController.archiveRobot,
    closeMissionModal: missionController.closeMissionModal,
    closeRobotModal: robotController.closeRobotModal,
    confirmArchiveRobot: robotController.confirmArchiveRobot,
    createMission: missionController.createMission,
    createRobot: robotController.createRobot,
    dismissNotification,
    endMission: missionController.endMission,
    loadConnectionInfo: robotController.loadConnectionInfo,
    openMissionControl: missionController.openMissionControl,
    openMissionCreateModal: missionController.openMissionCreateModal,
    openRobotCreateModal: robotController.openRobotCreateModal,
    openRobotEditModal: robotController.openRobotEditModal,
    playLatestRecording: recordingsController.playLatestRecording,
    reconnectLive,
    rotateRobotToken: robotController.rotateRobotToken,
    closeMissionControl,
    setSelectedMissionManagementCode: missionController.setSelectedMissionManagementCode,
    setSelectedRobotManagementCode: robotController.setSelectedRobotManagementCode,
    startMission: missionController.startMission,
    updateRobot: robotController.updateRobot
  };
}
