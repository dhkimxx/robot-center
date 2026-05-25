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
    <section className="grid h-full min-h-0 grid-cols-[minmax(0,1.52fr)_minmax(340px,0.88fr)] items-start gap-3 max-[1180px]:grid-cols-1">
      <MissionListPanel
        missions={missions}
        onOpenCreateMissionModal={onOpenCreateMissionModal}
        onSelectMission={onSelectMission}
        robots={robots}
        selectedMission={selectedMission}
      />

      <article className="self-start rounded-[14px] border border-slate-500/25 bg-command-800/95 p-6 shadow-command">
        <div className="mb-6 flex min-w-0 items-start justify-between gap-4">
          <h2 className="text-lg font-bold text-slate-50">임무 상세</h2>
          <span className="truncate text-sm font-bold text-slate-400">{selectedMission?.missionCode ?? "선택 없음"}</span>
        </div>
        {!selectedMission ? (
          <p className="rounded-xl border border-amber-300/20 bg-amber-300/10 px-4 py-3 text-sm font-semibold text-amber-200">
            임무를 선택하세요.
          </p>
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
