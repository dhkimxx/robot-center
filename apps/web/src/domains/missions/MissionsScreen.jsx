import { MissionControlView } from "../live/components/MissionControlView.jsx";
import { MissionDetailPanel } from "./MissionDetailPanel.jsx";
import { MissionListPanel } from "./MissionListPanel.jsx";
import { MissionReplayScreen } from "./MissionReplayScreen.jsx";
import { getMissionRobotDetails } from "./missionHelpers.js";
import Surface from "../../components/ui/Surface.jsx";

export default function MissionsScreen({
  controlMission,
  latestSensor,
  latestTelemetry,
  liveEvents,
  liveSessions,
  missionTargets,
  missions,
  observedStreams,
  onBackToMissionList,
  onEndMission,
  onOpenMissionControl,
  onOpenMissionReplay,
  onOpenPlaybackFile,
  onReconnectSelectedMissionTarget,
  onSelectMission,
  onStartMission,
  operationStatuses,
  recordings,
  replayMission,
  replayMissionCode,
  robots,
  selectedMission,
  selectedMissionTargetKey,
  setSelectedMissionTargetKey
}) {
  if (replayMissionCode) {
    return (
      <MissionReplayScreen
        missingMissionCode={replayMissionCode}
        mission={replayMission}
        onBackToMissionList={onBackToMissionList}
        onOpenPlaybackFile={onOpenPlaybackFile}
        observedStreams={observedStreams}
        recordings={recordings}
        robots={robots}
      />
    );
  }

  if (controlMission) {
    return (
      <MissionControlView
        latestSensor={latestSensor}
        latestTelemetry={latestTelemetry}
        liveEvents={liveEvents}
        liveSessions={liveSessions}
        mission={controlMission}
        missionTargets={missionTargets}
        onReconnectSelectedMissionTarget={onReconnectSelectedMissionTarget}
        operationStatuses={operationStatuses}
        selectedMissionTargetKey={selectedMissionTargetKey}
        setSelectedMissionTargetKey={setSelectedMissionTargetKey}
      />
    );
  }

  return (
    <section className="grid h-full min-h-0 grid-cols-[400px_minmax(0,1fr)] items-stretch gap-3 max-[1180px]:grid-cols-1">
      <MissionListPanel
        missions={missions}
        onSelectMission={onSelectMission}
        robots={robots}
        selectedMission={selectedMission}
        observedStreams={observedStreams}
      />

      <Surface className="grid min-h-0 min-w-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden" padding="none">
        <div className="flex min-h-14 min-w-0 items-center justify-between gap-4 border-b border-slate-700/70 px-4">
          <h2 className="text-base font-bold text-slate-50">임무 상세</h2>
          <span className="truncate text-sm font-bold text-slate-400">{selectedMission?.missionCode ?? "선택 없음"}</span>
        </div>
        <div className="min-h-0 overflow-auto p-4">
          {!selectedMission ? (
            <p className="rounded-xl border border-amber-300/20 bg-amber-300/10 px-4 py-3 text-sm font-semibold text-amber-200">
              임무를 선택하세요.
            </p>
          ) : (
            <MissionDetailPanel
              mission={selectedMission}
              onEndMission={onEndMission}
              onOpenMissionControl={onOpenMissionControl}
              onOpenMissionReplay={onOpenMissionReplay}
              onStartMission={onStartMission}
              robotDetails={getMissionRobotDetails(selectedMission, robots, observedStreams)}
            />
          )}
        </div>
      </Surface>
    </section>
  );
}
