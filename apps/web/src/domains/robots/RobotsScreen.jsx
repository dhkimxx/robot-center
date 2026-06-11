import { useEffect, useMemo, useState } from "react";
import { formatDateTime, makeStatusLabel } from "../../utils/formatters.js";
import Button from "../../components/ui/Button.jsx";
import DataTable from "../../components/ui/DataTable.jsx";
import DefinitionList from "../../components/ui/DefinitionList.jsx";
import EmptyState from "../../components/ui/EmptyState.jsx";
import ListFilterInput from "../../components/ui/ListFilterInput.jsx";
import PaginationControls from "../../components/ui/PaginationControls.jsx";
import ResizableSplitPane from "../../components/ui/ResizableSplitPane.jsx";
import SegmentedControl from "../../components/ui/SegmentedControl.jsx";
import SectionHeader from "../../components/ui/SectionHeader.jsx";
import StatusBadge from "../../components/ui/StatusBadge.jsx";
import Surface from "../../components/ui/Surface.jsx";
import { ListSkeleton, PanelSkeleton } from "../../components/ui/Skeleton.jsx";
import { createListView, createNextSortState } from "../../utils/listView.js";
import {
  findRobotOpenMission,
  getRobotAvailabilityTab,
  groupRobotsByAvailability,
  isOnlineRobot,
  makeRobotStatusTone
} from "./robotHelpers.js";

const robotPageSize = 10;
const robotTableGrid = "grid-cols-[minmax(190px,1.5fr)_110px_120px_150px] max-[760px]:grid-cols-[minmax(0,1fr)_auto]";

export default function RobotsScreen({
  dataLoadState,
  missions,
  onArchiveRobot,
  onLoadConnectionInfo,
  onOpenEditRobotModal,
  onSelectRobot,
  robots,
  selectedRobot
}) {
  const isInitialLoading = Boolean(dataLoadState?.isInitialLoading);
  const { offlineRobots, onlineRobots } = useMemo(() => groupRobotsByAvailability(robots), [robots]);
  const selectedAvailability = getRobotAvailabilityTab(selectedRobot);
  const [activeAvailability, setActiveAvailability] = useState(selectedAvailability);
  const [filterText, setFilterText] = useState("");
  const [pageIndex, setPageIndex] = useState(0);
  const [sortDirection, setSortDirection] = useState("desc");
  const [sortKey, setSortKey] = useState("lastSeenAt");
  const visibleRobots = activeAvailability === "online" ? onlineRobots : offlineRobots;
  const listView = useMemo(() => createListView(
    visibleRobots,
    { filterText, pageIndex, pageSize: robotPageSize, sortDirection, sortKey },
    (robot) => getRobotFilterValues(robot, missions),
    (robot, nextSortKey) => getRobotSortValue(robot, nextSortKey, missions)
  ), [filterText, missions, pageIndex, sortDirection, sortKey, visibleRobots]);
  const selectedRobotOpenMission = selectedRobot ? findRobotOpenMission(selectedRobot.robotCode, missions) : null;
  const selectedRobotHasOpenMission = Boolean(selectedRobotOpenMission);
  const selectedRobotInActiveAvailability = Boolean(
    selectedRobot
    && (activeAvailability === "online" ? isOnlineRobot(selectedRobot) : !isOnlineRobot(selectedRobot))
  );
  const detailRobot = selectedRobotInActiveAvailability ? selectedRobot : null;
  const detailRobotOpenMission = detailRobot ? selectedRobotOpenMission : null;
  const detailRobotHasOpenMission = Boolean(detailRobotOpenMission);

  useEffect(() => {
    setActiveAvailability(selectedAvailability);
  }, [selectedAvailability]);

  useEffect(() => {
    setPageIndex(0);
  }, [activeAvailability, filterText, sortDirection, sortKey]);

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
    <ResizableSplitPane
      left={(
        <Surface className="grid min-h-0 min-w-0 grid-rows-[auto_auto_minmax(0,1fr)_auto] overflow-hidden">
          <SectionHeader
            action={(
              <SegmentedControl
                onChange={setActiveAvailability}
                options={[
                  { count: onlineRobots.length, label: "온라인", value: "online" },
                  { count: offlineRobots.length, label: "오프라인", value: "offline" }
                ]}
                value={activeAvailability}
              />
            )}
            meta={`${listView.filteredItems.length}/${visibleRobots.length}대`}
            title="등록 로봇"
          />
          <ListFilterInput
            className="mb-2 w-full"
            placeholder="로봇명, 코드, 모델 검색"
            value={filterText}
            onChange={setFilterText}
          />
          <div className="min-h-0 overflow-auto pr-1">
            {isInitialLoading ? (
              <ListSkeleton count={5} />
            ) : robots.length === 0 ? (
              <EmptyState>등록된 로봇이 없습니다.</EmptyState>
            ) : (
              <RobotManagementRows
                missions={missions}
                robots={listView.page.pageItems}
                selectedRobot={selectedRobot}
                sortDirection={sortDirection}
                sortKey={sortKey}
                onSelectRobot={onSelectRobot}
                onSortChange={handleSortChange}
              />
            )}
          </div>
          {!isInitialLoading && robots.length > 0 ? (
            <PaginationControls
              className="mt-2"
              page={listView.page}
              onPageChange={setPageIndex}
            />
          ) : null}
        </Surface>
      )}
      leftMinWidth={560}
      right={(
        <Surface className="grid min-h-0 min-w-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden" padding="none">
          <div className="border-b border-slate-700/70 px-4 pt-4">
            <SectionHeader meta={detailRobot?.robotCode ?? "선택 없음"} title="로봇 상세" />
          </div>
          <div className="min-h-0 overflow-auto p-4">
            {isInitialLoading ? (
              <PanelSkeleton rows={5} />
            ) : !detailRobot ? (
              <EmptyState>로봇을 선택하세요.</EmptyState>
            ) : (
              <div className="grid gap-4">
                <div className="flex min-w-0 items-start justify-between gap-4 max-[760px]:grid">
                  <div className="min-w-0">
                    <div className="flex min-w-0 flex-wrap items-center gap-2">
                      <strong className="truncate text-lg font-bold leading-tight text-slate-50">{detailRobot.displayName}</strong>
                      <StatusBadge tone={makeRobotStatusTone(detailRobot.status)}>{makeStatusLabel(detailRobot.status)}</StatusBadge>
                    </div>
                    <span className="mt-1 block truncate text-sm font-semibold text-slate-400">{detailRobot.modelName || "모델 미지정"}</span>
                  </div>
                  <div className="flex shrink-0 flex-wrap justify-end gap-2 max-[760px]:justify-start">
                    <Button variant="primary" onClick={onOpenEditRobotModal}>수정</Button>
                    <Button onClick={() => onLoadConnectionInfo(detailRobot.robotCode)}>연결 정보</Button>
                    <Button
                      variant="danger"
                      disabled={detailRobotHasOpenMission}
                      onClick={() => onArchiveRobot(detailRobot.robotCode)}
                    >
                      삭제
                    </Button>
                  </div>
                </div>

                <Surface className="grid gap-3" padding="sm" variant="section">
                  <h3 className="text-xs font-bold uppercase tracking-normal text-slate-500">로봇 개요</h3>
                  <DefinitionList
                    items={[
                      { label: "상태", value: makeStatusLabel(detailRobot.status) },
                      { label: "최근 연결", value: formatDateTime(detailRobot.lastSeenAt) },
                      { label: "현재 임무", value: detailRobotOpenMission?.missionCode ?? "-" }
                    ]}
                  />
                  {detailRobotHasOpenMission ? (
                    <span className="rounded-lg border border-amber-300/20 bg-amber-300/10 px-3 py-2 text-sm font-semibold text-amber-200">
                      삭제 불가: 진행/대기 임무 배정
                    </span>
                  ) : null}
                </Surface>
              </div>
            )}
          </div>
        </Surface>
      )}
      rightMinWidth={340}
      storageKey="robot-management-split"
    />
  );
}

function RobotManagementRows({
  missions,
  onSelectRobot,
  onSortChange,
  robots,
  selectedRobot,
  sortDirection,
  sortKey
}) {
  if (robots.length === 0) {
    return <EmptyState>해당 조건의 로봇이 없습니다.</EmptyState>;
  }

  return (
    <DataTable
      columns={createRobotColumns(missions)}
      emptyLabel="해당 조건의 로봇이 없습니다."
      getRowKey={(robot) => robot.robotCode}
      gridTemplateClass={robotTableGrid}
      rowAriaLabel={(robot) => `${robot.displayName} ${robot.robotCode} 선택`}
      rows={robots}
      selectedRowKey={selectedRobot?.robotCode}
      sortDirection={sortDirection}
      sortKey={sortKey}
      onRowClick={(robot) => onSelectRobot(robot.robotCode)}
      onSortChange={onSortChange}
    />
  );
}

function createRobotColumns(missions) {
  return [
    {
      key: "robot",
      label: "로봇",
      sortKey: "displayName",
      render: (robot) => (
        <div className="min-w-0">
          <strong className="block truncate text-sm font-bold text-slate-100">{robot.displayName}</strong>
          <span className="mt-1 block truncate text-xs font-semibold text-slate-500">{robot.robotCode} · {robot.modelName || "모델 미지정"}</span>
        </div>
      )
    },
    {
      key: "status",
      label: "상태",
      sortKey: "status",
      render: (robot) => (
        <StatusBadge className="justify-self-end max-[760px]:justify-self-end" size="xs" tone={makeRobotStatusTone(robot.status)}>
          {makeStatusLabel(robot.status)}
        </StatusBadge>
      )
    },
    {
      className: "max-[760px]:hidden",
      key: "mission",
      label: "현재 임무",
      sortKey: "missionCode",
      render: (robot) => {
        const openMission = findRobotOpenMission(robot.robotCode, missions);
        return <span className="truncate text-xs font-bold text-slate-400">{openMission?.missionCode ?? "-"}</span>;
      }
    },
    {
      className: "max-[760px]:hidden",
      key: "lastSeenAt",
      label: "최근 연결",
      sortKey: "lastSeenAt",
      render: (robot) => (
        <span className="truncate text-xs font-bold text-slate-500">{formatDateTime(robot.lastSeenAt)}</span>
      )
    }
  ];
}

function getRobotFilterValues(robot, missions) {
  const openMission = findRobotOpenMission(robot.robotCode, missions);
  return [
    robot.robotCode,
    robot.displayName,
    robot.modelName,
    robot.status,
    openMission?.missionCode
  ];
}

function getRobotSortValue(robot, sortKey, missions) {
  switch (sortKey) {
    case "displayName":
      return robot.displayName;
    case "status":
      return robot.status;
    case "missionCode":
      return findRobotOpenMission(robot.robotCode, missions)?.missionCode ?? "";
    case "lastSeenAt":
      return robot.lastSeenAt ?? "";
    default:
      return robot[sortKey];
  }
}
