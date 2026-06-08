import { MissionControlView } from "../live/components/MissionControlView.jsx";
import { MissionDetailPanel } from "./MissionDetailPanel.jsx";
import { MissionListPanel } from "./MissionListPanel.jsx";
import { MissionReplayScreen } from "./MissionReplayScreen.jsx";
import { getMissionRobotDetails } from "./missionHelpers.js";
import Surface from "../../components/ui/Surface.jsx";
import { PanelSkeleton } from "../../components/ui/Skeleton.jsx";

export default function MissionsScreen({
  controlMission,
  controlMissionCode,
  dataLoadState,
  isSensorSnapshotRefreshing,
  latestSensor,
  latestSensorSourceLabel,
  latestTelemetry,
  liveEvents,
  liveSessions,
  liveStatuses,
  missionTargets,
  missions,
  onBackToMissionList,
  onEndMission,
  onOpenMissionControl,
  onOpenMissionReplay,
  onOpenPlaybackFile,
  onReconnectSelectedMissionTarget,
  onRefreshSensorSnapshot,
  onSelectMission,
  onStartMission,
  operationStatuses,
  replayMission,
  replayMissionCode,
  robots,
  selectedMission,
  selectedMissionTargetKey,
  setSelectedMissionTargetKey
}) {
  const isInitialLoading = Boolean(dataLoadState?.isInitialLoading);

  if (replayMissionCode) {
    return (
      <MissionReplayScreen
        isLoading={isInitialLoading}
        missingMissionCode={replayMissionCode}
        mission={replayMission}
        onBackToMissionList={onBackToMissionList}
        onOpenPlaybackFile={onOpenPlaybackFile}
        robots={robots}
      />
    );
  }

  if (controlMissionCode && !controlMission) {
    return isInitialLoading ? (
      <MissionRouteLoading title="관제 정보를 불러오는 중" />
    ) : (
      <MissionRouteNotFound
        missionCode={controlMissionCode}
        onBackToMissionList={onBackToMissionList}
        title="관제할 임무를 찾을 수 없습니다."
      />
    );
  }

  if (controlMission) {
    return (
      <MissionControlView
        latestSensor={latestSensor}
        latestSensorSourceLabel={latestSensorSourceLabel}
        latestTelemetry={latestTelemetry}
        liveEvents={liveEvents}
        liveSessions={liveSessions}
        mission={controlMission}
        missionTargets={missionTargets}
        onOpenMissionReplay={onOpenMissionReplay}
        onReconnectSelectedMissionTarget={onReconnectSelectedMissionTarget}
        onRefreshSensorSnapshot={onRefreshSensorSnapshot}
        operationStatuses={operationStatuses}
        isSensorSnapshotRefreshing={isSensorSnapshotRefreshing}
        selectedMissionTargetKey={selectedMissionTargetKey}
        setSelectedMissionTargetKey={setSelectedMissionTargetKey}
      />
    );
  }

  return (
    <section className="grid h-full min-h-0 grid-cols-[400px_minmax(0,1fr)] items-stretch gap-3 max-[1180px]:grid-cols-1">
      <MissionListPanel
        isLoading={isInitialLoading}
        liveStatuses={liveStatuses}
        missions={missions}
        onSelectMission={onSelectMission}
        robots={robots}
        selectedMission={selectedMission}
      />

      <Surface className="grid min-h-0 min-w-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden" padding="none">
        <div className="flex min-h-14 min-w-0 items-center justify-between gap-4 border-b border-slate-700/70 px-4">
          <h2 className="text-base font-bold text-slate-50">임무 상세</h2>
          <span className="truncate text-sm font-bold text-slate-400">{selectedMission?.missionCode ?? "선택 없음"}</span>
        </div>
        <div className="min-h-0 overflow-auto p-4">
          {isInitialLoading ? (
            <PanelSkeleton rows={5} />
          ) : !selectedMission ? (
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
              robotDetails={getMissionRobotDetails(selectedMission, robots, liveStatuses?.[selectedMission.missionCode] ?? null)}
            />
          )}
        </div>
      </Surface>
    </section>
  );
}

function MissionRouteLoading({ title }) {
  return (
    <Surface className="grid h-full min-h-0 content-start gap-4">
      <div>
        <h2 className="text-lg font-black text-slate-50">{title}</h2>
        <p className="mt-1 text-sm font-semibold text-slate-500">서버에서 임무와 실시간 상태를 확인하고 있습니다.</p>
      </div>
      <PanelSkeleton rows={6} />
    </Surface>
  );
}

function MissionRouteNotFound({ missionCode, onBackToMissionList, title }) {
  return (
    <Surface className="grid content-start gap-4">
      <div className="flex min-w-0 items-center justify-between gap-3">
        <div className="min-w-0">
          <h2 className="truncate text-lg font-black text-slate-50">{title}</h2>
          <p className="mt-1 truncate text-sm font-semibold text-slate-500">{missionCode}</p>
        </div>
        <button
          className="rounded-lg border border-slate-700/70 bg-white/[0.06] px-3 py-2 text-sm font-bold text-slate-200 hover:bg-white/[0.10]"
          type="button"
          onClick={() => onBackToMissionList()}
        >
          임무 목록
        </button>
      </div>
      <p className="rounded-xl border border-amber-300/20 bg-amber-300/10 px-4 py-3 text-sm font-semibold text-amber-200">
        임무 데이터 로딩은 완료됐지만 해당 임무를 찾을 수 없습니다.
      </p>
    </Surface>
  );
}
