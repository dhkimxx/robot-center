import Button from "../../../components/ui/Button.jsx";
import Surface from "../../../components/ui/Surface.jsx";
import { MissionRobotDropdown } from "./MissionRobotDropdown.jsx";

export function MissionControlToolbar({
  canReconnectSelectedRobot,
  liveSessions,
  mission,
  missionConnectionLabel,
  missionStartLabel,
  missionTargets,
  onOpenMissionReplay,
  onReconnectSelectedMissionTarget,
  selectedMissionTargetKey,
  selectedRecordingTimingLabel,
  selectedStreamTimingLabel,
  setSelectedMissionTargetKey
}) {
  return (
    <Surface className="grid min-h-[76px] min-w-0 grid-cols-[minmax(120px,0.55fr)_minmax(220px,1fr)_minmax(170px,0.7fr)_auto] items-center gap-3 overflow-visible px-3 py-2 max-[900px]:grid-cols-[minmax(120px,0.6fr)_minmax(220px,1fr)] max-[760px]:grid-cols-1">
      <div className="min-w-0">
        <strong className="block truncate text-sm font-extrabold text-slate-50">{mission.missionCode}</strong>
        <span className="mt-0.5 block truncate text-xs font-bold text-slate-500">
          {missionTargets.length > 0 ? `${missionStartLabel} · ${missionConnectionLabel}` : missionConnectionLabel}
        </span>
      </div>
      <div className="min-w-0">
        <div className="flex min-w-0 items-center gap-3">
          <span className="shrink-0 text-xs font-bold text-slate-500">관제 로봇</span>
          <div className="min-w-0 flex-1">
            <MissionRobotDropdown
              liveSessions={liveSessions}
              missionTargets={missionTargets}
              selectedMissionTargetKey={selectedMissionTargetKey}
              setSelectedMissionTargetKey={setSelectedMissionTargetKey}
            />
          </div>
        </div>
      </div>
      <span className="min-w-0 truncate text-xs font-bold text-slate-500">
        {selectedStreamTimingLabel ? `${selectedStreamTimingLabel} · ${selectedRecordingTimingLabel}` : "송출 대기"}
      </span>
      <div className="flex shrink-0 items-center justify-end gap-2 max-[760px]:justify-start">
        <Button size="sm" onClick={() => onOpenMissionReplay?.(mission)}>
          리플레이 보기
        </Button>
        {canReconnectSelectedRobot ? (
          <Button size="sm" onClick={onReconnectSelectedMissionTarget}>재연결</Button>
        ) : null}
      </div>
    </Surface>
  );
}
