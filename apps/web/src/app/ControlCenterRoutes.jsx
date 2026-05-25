import { Navigate, Route, Routes } from "react-router-dom";
import MissionsScreen from "../domains/missions/MissionsScreen.jsx";
import RecordingsScreen from "../domains/recordings/RecordingsScreen.jsx";
import RobotsScreen from "../domains/robots/RobotsScreen.jsx";
import SystemScreen from "../domains/system/SystemScreen.jsx";

export function ControlCenterRoutes({ controller, navigateToPath }) {
  const selectedLiveSession = controller.selectedLiveSession;
  const missionScreenProps = {
    latestRecording: controller.latestRecording,
    latestSensor: controller.latestSensor,
    latestTelemetry: controller.latestTelemetry,
    liveEvents: selectedLiveSession.events,
    liveSessions: controller.liveSessions,
    missionTargets: controller.missionControlTargets,
    missions: controller.missions,
    onBackToMissionList: controller.closeMissionControl,
    onEndMission: controller.endMission,
    onOpenCreateMissionModal: controller.openMissionCreateModal,
    onOpenMissionControl: controller.openMissionControl,
    onOpenRecordings: () => navigateToPath("/recordings"),
    onPlayLatestRecording: controller.playLatestRecording,
    onReconnectSelectedMissionTarget: controller.reconnectLive,
    onSelectMission: controller.setSelectedMissionManagementCode,
    onStartMission: controller.startMission,
    operationStatuses: controller.operationStatuses,
    playbackRecording: controller.latestPlayableRecording,
    robots: controller.robots,
    selectedMission: controller.selectedMission,
    selectedMissionTargetKey: controller.selectedLiveTargetKey,
    setSelectedMissionTargetKey: controller.setSelectedLiveTargetKey
  };

  return (
    <Routes>
      <Route path="/" element={<Navigate replace to="/missions" />} />
      <Route path="/missions" element={<MissionsScreen {...missionScreenProps} controlMission={null} />} />
      <Route path="/missions/:missionCode/control" element={<MissionsScreen {...missionScreenProps} controlMission={controller.missionControlMission} />} />
      <Route
        path="/robots"
        element={(
          <RobotsScreen
            missions={controller.missions}
            onArchiveRobot={controller.archiveRobot}
            onLoadConnectionInfo={controller.loadConnectionInfo}
            onOpenCreateRobotModal={controller.openRobotCreateModal}
            onOpenEditRobotModal={controller.openRobotEditModal}
            onRotateRobotToken={controller.rotateRobotToken}
            onSelectRobot={controller.setSelectedRobotManagementCode}
            robots={controller.robots}
            selectedRobot={controller.selectedRobot}
          />
        )}
      />
      <Route
        path="/recordings"
        element={<RecordingsScreen onOpenPlaybackFile={controller.setRecordingPlaybackFile} recordings={controller.recordings} />}
      />
      <Route
        path="/system"
        element={<SystemScreen statusError={controller.statusError} systemStatus={controller.systemStatus} />}
      />
      <Route path="*" element={<Navigate replace to="/missions" />} />
    </Routes>
  );
}
