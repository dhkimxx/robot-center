import { useMemo } from "react";
import EmptyState from "../../components/ui/EmptyState.jsx";
import ListRow from "../../components/ui/ListRow.jsx";
import SectionHeader from "../../components/ui/SectionHeader.jsx";
import StatusBadge from "../../components/ui/StatusBadge.jsx";
import Surface from "../../components/ui/Surface.jsx";
import { ListSkeleton } from "../../components/ui/Skeleton.jsx";
import {
  makeStatusLabel,
  missionTypeLabel
} from "../../utils/formatters.js";
import {
  formatMissionRobotCount,
  getMissionRobotDetails,
  groupMissionsByLifecycle,
  isClosedMission
} from "./missionHelpers.js";

export function MissionListPanel({
  isLoading = false,
  liveStatuses,
  missions,
  onSelectMission,
  robots,
  selectedMission
}) {
  const { closedMissions, openMissions } = useMemo(() => groupMissionsByLifecycle(missions), [missions]);

  return (
    <Surface className="grid min-h-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden p-3">
      <SectionHeader
        className="mb-3"
        title="임무 목록"
      />
      <div className="grid min-h-0 auto-rows-max content-start gap-3 overflow-auto pr-1">
        {isLoading ? (
          <ListSkeleton count={6} />
        ) : missions.length === 0 ? (
          <EmptyState>생성된 임무가 없습니다.</EmptyState>
        ) : (
          <>
            <MissionListGroup
              emptyLabel="진행 중이거나 시작 가능한 임무가 없습니다."
              liveStatuses={liveStatuses}
              missions={openMissions}
              onSelectMission={onSelectMission}
              robots={robots}
              selectedMission={selectedMission}
              title="진행 임무"
            />
            <MissionListGroup
              emptyLabel="종료된 임무가 없습니다."
              liveStatuses={liveStatuses}
              missions={closedMissions}
              onSelectMission={onSelectMission}
              robots={robots}
              selectedMission={selectedMission}
              title="종료 임무"
            />
          </>
        )}
      </div>
    </Surface>
  );
}

function MissionListGroup({
  emptyLabel,
  liveStatuses,
  missions,
  onSelectMission,
  robots,
  selectedMission,
  title
}) {
  return (
    <section className="grid gap-1.5">
      <div className="flex items-center justify-between gap-3 px-1">
        <h3 className="text-xs font-black uppercase tracking-normal text-slate-400">{title}</h3>
        <span className="text-xs font-bold text-slate-500">{missions.length}건</span>
      </div>
      {missions.length === 0 ? (
        <div className="rounded-lg border border-slate-500/20 bg-white/[0.03] px-3 py-3 text-sm font-semibold text-slate-500">
          {emptyLabel}
        </div>
      ) : (
        <div className="grid gap-1.5">
          {missions.map((mission) => {
            const isSelectedMission = selectedMission?.missionCode === mission.missionCode;
            const isClosed = isClosedMission(mission);
            const robotDetails = getMissionRobotDetails(mission, robots, liveStatuses?.[mission.missionCode] ?? null);
            const missionMeta = `${mission.missionCode} · ${missionTypeLabel(mission.missionType)} · ${formatMissionRobotCount(robotDetails)}`;
            const robotSummary = robotDetails.map((robot) => `${robot.robotCode} ${robot.liveLabel}`).join(" · ");
            return (
              <ListRow
                aria-label={`${mission.name} ${mission.missionCode} 선택`}
                aria-pressed={isSelectedMission}
                description={robotDetails.length > 0 ? robotSummary : null}
                key={mission.missionCode}
                meta={missionMeta}
                muted={isClosed}
                onClick={() => onSelectMission(mission.missionCode)}
                right={
                  <StatusBadge size="xs" tone={isClosed ? "neutral" : mission.status === "active" ? "success" : "warning"}>
                    {makeStatusLabel(mission.status)}
                  </StatusBadge>
                }
                selected={isSelectedMission}
                title={mission.name}
              >
              </ListRow>
            );
          })}
        </div>
      )}
    </section>
  );
}
