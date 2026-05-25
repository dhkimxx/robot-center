import { useMemo } from "react";
import Button from "../../components/ui/Button.jsx";
import EmptyState from "../../components/ui/EmptyState.jsx";
import SectionHeader from "../../components/ui/SectionHeader.jsx";
import Surface from "../../components/ui/Surface.jsx";
import { cn } from "../../utils/cn.js";
import {
  makeStatusLabel,
  missionTypeLabel
} from "../../utils/formatters.js";
import {
  formatMissionRobotCount,
  getMissionRobotDetails
} from "./missionHelpers.js";

const closedMissionStatuses = new Set(["completed", "ended", "cancelled"]);

export function MissionListPanel({
  missions,
  onOpenCreateMissionModal,
  onSelectMission,
  robots,
  selectedMission
}) {
  const orderedMissions = useMemo(() => {
    const statusOrder = { active: 0, ready: 1, completed: 2, ended: 2, cancelled: 3 };
    return [...missions].sort((left, right) => {
      const leftOrder = statusOrder[left.status] ?? 9;
      const rightOrder = statusOrder[right.status] ?? 9;
      if (leftOrder !== rightOrder) {
        return leftOrder - rightOrder;
      }
      return (right.startedAt ?? right.createdAt ?? "").localeCompare(left.startedAt ?? left.createdAt ?? "");
    });
  }, [missions]);
  const activeMissionCount = missions.filter((mission) => mission.status === "active").length;

  return (
    <Surface className="grid min-h-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden">
      <SectionHeader
        action={<Button size="sm" variant="primary" onClick={onOpenCreateMissionModal}>임무 생성</Button>}
        title="진행 임무"
        meta={`진행 ${activeMissionCount}건 / 전체 ${missions.length}건`}
      />
      <div className="grid min-h-0 auto-rows-max content-start gap-2 overflow-auto pr-1">
        {missions.length === 0 ? (
          <EmptyState>생성된 임무가 없습니다.</EmptyState>
        ) : (
          orderedMissions.map((mission) => {
            const isSelectedMission = selectedMission?.missionCode === mission.missionCode;
            const isClosedMission = closedMissionStatuses.has(mission.status);
            const robotDetails = getMissionRobotDetails(mission, robots);
            return (
              <button
                aria-label={`${mission.name} ${mission.missionCode} 선택`}
                aria-pressed={isSelectedMission}
                className={cn(
                  "grid min-h-[58px] gap-1 rounded-lg border border-slate-500/20 bg-white/[0.045] px-3 py-2 text-left transition hover:border-sapphire-500/[0.34] hover:bg-sapphire-500/[0.09] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sapphire-500",
                  isSelectedMission && "border-sapphire-500/50 bg-sapphire-500/[0.09] shadow-[inset_3px_0_0_var(--color-sapphire)]",
                  isClosedMission && "bg-slate-400/[0.12] opacity-75 hover:bg-slate-400/[0.16]",
                  isClosedMission && isSelectedMission && "bg-slate-400/[0.20]"
                )}
                key={mission.missionCode}
                type="button"
                onClick={() => onSelectMission(mission.missionCode)}
              >
                <strong className="truncate text-sm font-bold text-slate-50">{mission.name}</strong>
                <span className="truncate text-xs font-semibold text-slate-400">
                  {mission.missionCode} / {missionTypeLabel(mission.missionType)} / {makeStatusLabel(mission.status)} / {formatMissionRobotCount(robotDetails)}
                </span>
                {robotDetails.length > 0 ? (
                  <span className="truncate text-xs font-semibold text-slate-500">
                    {robotDetails.map((robot) => `${robot.robotCode} ${makeStatusLabel(robot.status)}`).join(" / ")}
                  </span>
                ) : null}
              </button>
            );
          })
        )}
      </div>
    </Surface>
  );
}
