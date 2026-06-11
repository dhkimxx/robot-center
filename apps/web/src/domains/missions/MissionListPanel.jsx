import EmptyState from "../../components/ui/EmptyState.jsx";
import ListRow from "../../components/ui/ListRow.jsx";
import SegmentedControl from "../../components/ui/SegmentedControl.jsx";
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
  isClosedMission
} from "./missionHelpers.js";

export function MissionListPanel({
  activeGroup,
  closedMissions,
  isLoading = false,
  liveStatuses,
  missions,
  onSelectMission,
  onSetActiveGroup,
  openMissions,
  robots,
  selectedMission,
  visibleMissions
}) {
  const emptyLabel = activeGroup === "closed" ? "종료된 임무가 없습니다." : "진행 중이거나 시작 가능한 임무가 없습니다.";

  return (
    <Surface className="grid min-h-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden p-3">
      <SectionHeader
        action={(
          <SegmentedControl
            onChange={onSetActiveGroup}
            options={[
              { count: openMissions.length, label: "진행 중", value: "open" },
              { count: closedMissions.length, label: "종료됨", value: "closed" }
            ]}
            value={activeGroup}
          />
        )}
        className="mb-3"
        meta={`${visibleMissions.length}건`}
        title="임무 목록"
      />
      <div className="grid min-h-0 auto-rows-max content-start overflow-auto pr-1">
        {isLoading ? (
          <ListSkeleton count={6} />
        ) : missions.length === 0 ? (
          <EmptyState>생성된 임무가 없습니다.</EmptyState>
        ) : (
          <MissionGrid
            emptyLabel={emptyLabel}
            liveStatuses={liveStatuses}
            missions={visibleMissions}
            onSelectMission={onSelectMission}
            robots={robots}
            selectedMission={selectedMission}
          />
        )}
      </div>
    </Surface>
  );
}

function MissionGrid({
  emptyLabel,
  liveStatuses,
  missions,
  onSelectMission,
  robots,
  selectedMission
}) {
  if (missions.length === 0) {
    return (
      <div className="rounded-lg border border-slate-500/20 bg-white/[0.03] px-3 py-4 text-sm font-semibold text-slate-500">
        {emptyLabel}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-[repeat(auto-fit,minmax(280px,1fr))] gap-2">
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
            className="min-h-[94px]"
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
          />
        );
      })}
    </div>
  );
}
