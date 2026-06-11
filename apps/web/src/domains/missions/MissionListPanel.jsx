import { useEffect, useMemo, useState } from "react";
import DataTable from "../../components/ui/DataTable.jsx";
import EmptyState from "../../components/ui/EmptyState.jsx";
import ListFilterInput from "../../components/ui/ListFilterInput.jsx";
import PaginationControls from "../../components/ui/PaginationControls.jsx";
import SegmentedControl from "../../components/ui/SegmentedControl.jsx";
import SectionHeader from "../../components/ui/SectionHeader.jsx";
import StatusBadge from "../../components/ui/StatusBadge.jsx";
import Surface from "../../components/ui/Surface.jsx";
import { ListSkeleton } from "../../components/ui/Skeleton.jsx";
import {
  formatDateTimeFull,
  makeStatusLabel,
  missionTypeLabel
} from "../../utils/formatters.js";
import { createListView, createNextSortState } from "../../utils/listView.js";
import {
  formatMissionRobotCount,
  getMissionRobotDetails,
  isClosedMission
} from "./missionHelpers.js";

const missionPageSizeOptions = [10, 20, 50];
const missionTableGrid = "grid-cols-[minmax(220px,1.4fr)_110px_minmax(180px,1fr)_minmax(170px,1fr)] max-[760px]:grid-cols-[minmax(0,1fr)_auto]";

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
  const [filterText, setFilterText] = useState("");
  const [pageIndex, setPageIndex] = useState(0);
  const [pageSize, setPageSize] = useState(missionPageSizeOptions[0]);
  const [sortDirection, setSortDirection] = useState("desc");
  const [sortKey, setSortKey] = useState("createdAt");
  const listView = useMemo(() => createListView(
    visibleMissions,
    { filterText, pageIndex, pageSize, sortDirection, sortKey },
    (mission) => getMissionFilterValues(mission, robots, liveStatuses?.[mission.missionCode] ?? null),
    getMissionSortValue
  ), [filterText, liveStatuses, pageIndex, pageSize, robots, sortDirection, sortKey, visibleMissions]);

  useEffect(() => {
    setPageIndex(0);
  }, [activeGroup, filterText, pageSize, sortDirection, sortKey]);

  useEffect(() => {
    if (listView.page.pageIndex !== pageIndex) {
      setPageIndex(listView.page.pageIndex);
    }
  }, [listView.page.pageIndex, pageIndex]);

  function handleSortChange(nextSortKey) {
    const nextSort = createNextSortState({
      currentDirection: sortDirection,
      currentKey: sortKey,
      nextKey: nextSortKey
    });
    setSortKey(nextSort.sortKey);
    setSortDirection(nextSort.sortDirection);
  }

  return (
    <Surface className="grid min-h-0 grid-rows-[auto_auto_minmax(0,1fr)_auto] overflow-hidden p-3">
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
        meta={`${listView.filteredItems.length}/${visibleMissions.length}건`}
        title="임무 목록"
      />
      <ListFilterInput
        className="mb-2 w-full"
        placeholder="임무명, 코드, 로봇 검색"
        value={filterText}
        onChange={setFilterText}
      />
      <div className="grid min-h-0 auto-rows-max content-start overflow-auto pr-1">
        {isLoading ? (
          <ListSkeleton count={6} />
        ) : missions.length === 0 ? (
          <EmptyState>생성된 임무가 없습니다.</EmptyState>
        ) : (
          <MissionRows
            emptyLabel={emptyLabel}
            liveStatuses={liveStatuses}
            missions={listView.page.pageItems}
            onSelectMission={onSelectMission}
            onSortChange={handleSortChange}
            robots={robots}
            selectedMission={selectedMission}
            sortDirection={sortDirection}
            sortKey={sortKey}
          />
        )}
      </div>
      {!isLoading && missions.length > 0 ? (
        <PaginationControls
          className="mt-2"
          page={listView.page}
          pageSizeOptions={missionPageSizeOptions}
          onPageChange={setPageIndex}
          onPageSizeChange={setPageSize}
        />
      ) : null}
    </Surface>
  );
}

function MissionRows({
  emptyLabel,
  liveStatuses,
  missions,
  onSelectMission,
  onSortChange,
  robots,
  selectedMission,
  sortDirection,
  sortKey
}) {
  if (missions.length === 0) {
    return <EmptyState>{emptyLabel}</EmptyState>;
  }

  return (
    <DataTable
      columns={createMissionColumns(robots, liveStatuses)}
      emptyLabel={emptyLabel}
      getRowKey={(mission) => mission.missionCode}
      gridTemplateClass={missionTableGrid}
      rowAriaLabel={(mission) => `${mission.name} ${mission.missionCode} 선택`}
      rows={missions}
      selectedRowKey={selectedMission?.missionCode}
      sortDirection={sortDirection}
      sortKey={sortKey}
      onRowClick={(mission) => onSelectMission(mission.missionCode)}
      onSortChange={onSortChange}
    />
  );
}

function createMissionColumns(robots, liveStatuses) {
  return [
    {
      key: "mission",
      label: "임무",
      sortKey: "name",
      render: (mission) => (
        <div className="min-w-0">
          <strong className="block truncate text-sm font-bold text-slate-100">{mission.name}</strong>
          <span className="mt-1 block truncate text-xs font-semibold text-slate-500">
            {mission.missionCode} · {missionTypeLabel(mission.missionType)}
          </span>
        </div>
      )
    },
    {
      key: "status",
      label: "상태",
      sortKey: "status",
      render: (mission) => {
        const isClosed = isClosedMission(mission);
        return (
          <StatusBadge size="xs" tone={isClosed ? "neutral" : mission.status === "active" ? "success" : "warning"}>
            {makeStatusLabel(mission.status)}
          </StatusBadge>
        );
      }
    },
    {
      className: "max-[760px]:hidden",
      key: "robots",
      label: "배정 로봇",
      render: (mission) => {
        const robotDetails = getMissionRobotDetails(mission, robots, liveStatuses?.[mission.missionCode] ?? null);
        const robotSummary = robotDetails.map((robot) => `${robot.robotCode} ${robot.liveLabel}`).join(" · ");
        return (
          <div className="min-w-0">
            <span className="block truncate text-xs font-bold text-slate-300">{formatMissionRobotCount(robotDetails)}</span>
            <span className="mt-1 block truncate text-xs font-semibold text-slate-500">{robotSummary || "-"}</span>
          </div>
        );
      }
    },
    {
      className: "max-[760px]:hidden",
      key: "time",
      label: "시간",
      sortKey: "createdAt",
      render: (mission) => (
        <div className="min-w-0 text-xs font-semibold text-slate-500">
          <span className="block truncate">생성 {formatDateTimeFull(mission.createdAt)}</span>
          <span className="mt-1 block truncate text-slate-400">시작 {formatDateTimeFull(mission.startedAt)}</span>
        </div>
      )
    }
  ];
}

function getMissionFilterValues(mission, robots, liveStatus) {
  const robotDetails = getMissionRobotDetails(mission, robots, liveStatus);
  return [
    mission.name,
    mission.missionCode,
    mission.missionType,
    mission.status,
    mission.siteNote,
    ...robotDetails.flatMap((robot) => [robot.robotCode, robot.displayName, robot.liveLabel])
  ];
}

function getMissionSortValue(mission, sortKey) {
  switch (sortKey) {
    case "name":
      return mission.name;
    case "status":
      return mission.status;
    case "createdAt":
      return mission.createdAt;
    default:
      return mission[sortKey];
  }
}
