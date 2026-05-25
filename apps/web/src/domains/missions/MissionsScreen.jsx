import { MissionControlView } from "../live/components/MissionControlView.jsx";
import { MissionDetailPanel } from "./MissionDetailPanel.jsx";
import { MissionListPanel } from "./MissionListPanel.jsx";
import { getMissionRobotDetails } from "./missionHelpers.js";

export default function MissionsScreen({
  controlMission,
  latestRecording,
  latestSensor,
  latestTelemetry,
  liveEvents,
  liveSessions,
  missionTargets,
  missions,
  onBackToMissionList,
  onEndMission,
  onOpenCreateMissionModal,
  onOpenMissionControl,
  onOpenRecordings,
  onPlayLatestRecording,
  onReconnectSelectedMissionTarget,
  onSelectMission,
  onStartMission,
  operationStatuses,
  playbackRecording,
  robots,
  selectedMission,
  selectedMissionTargetKey,
  setSelectedMissionTargetKey
}) {
  if (controlMission) {
    return (
      <MissionControlView
        latestRecording={latestRecording}
        latestSensor={latestSensor}
        latestTelemetry={latestTelemetry}
        liveEvents={liveEvents}
        liveSessions={liveSessions}
        mission={controlMission}
        missionTargets={missionTargets}
        onBackToMissionList={onBackToMissionList}
        onEndMission={onEndMission}
        onOpenRecordings={onOpenRecordings}
        onPlayLatestRecording={onPlayLatestRecording}
        onReconnectSelectedMissionTarget={onReconnectSelectedMissionTarget}
        onStartMission={onStartMission}
        operationStatuses={operationStatuses}
        playbackRecording={playbackRecording}
        selectedMissionTargetKey={selectedMissionTargetKey}
        setSelectedMissionTargetKey={setSelectedMissionTargetKey}
      />
    );
  }

  return (
    <section className="mission-management-layout">
      <MissionListPanel
        missions={missions}
        onOpenCreateMissionModal={onOpenCreateMissionModal}
        onSelectMission={onSelectMission}
        robots={robots}
        selectedMission={selectedMission}
      />

      <article className="surface">
        <div className="section-heading">
          <h2>임무 상세</h2>
          <span>{selectedMission?.missionCode ?? "선택 없음"}</span>
        </div>
        {!selectedMission ? (
          <p className="empty-state">임무를 선택하세요.</p>
        ) : (
          <MissionDetailPanel
            mission={selectedMission}
            onEndMission={onEndMission}
            onOpenMissionControl={onOpenMissionControl}
            onStartMission={onStartMission}
            robotDetails={getMissionRobotDetails(selectedMission, robots)}
          />
        )}
      </article>
    </section>
  );
}
