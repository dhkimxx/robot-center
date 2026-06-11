import { useEffect, useMemo, useState } from "react";
import { formatDateTime, makeStatusLabel } from "../../utils/formatters.js";
import Button from "../../components/ui/Button.jsx";
import DefinitionList from "../../components/ui/DefinitionList.jsx";
import EmptyState from "../../components/ui/EmptyState.jsx";
import SegmentedControl from "../../components/ui/SegmentedControl.jsx";
import SectionHeader from "../../components/ui/SectionHeader.jsx";
import StatusBadge from "../../components/ui/StatusBadge.jsx";
import Surface from "../../components/ui/Surface.jsx";
import { ListSkeleton, PanelSkeleton } from "../../components/ui/Skeleton.jsx";
import {
  findRobotOpenMission,
  groupRobotsByAvailability,
  isOnlineRobot,
  makeRobotStatusTone
} from "./robotHelpers.js";

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
  const selectedAvailability = selectedRobot && isOnlineRobot(selectedRobot) ? "online" : "offline";
  const [activeAvailability, setActiveAvailability] = useState(selectedAvailability);
  const visibleRobots = activeAvailability === "online" ? onlineRobots : offlineRobots;
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

  return (
    <section className="grid h-full min-h-0 grid-cols-[minmax(0,1fr)_420px] items-stretch gap-3 max-[1180px]:grid-cols-1">
      <Surface className="grid min-h-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden">
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
          meta={`${visibleRobots.length}/${robots.length}대`}
          title="등록 로봇"
        />
        <div className="min-h-0 overflow-auto pr-1">
          {isInitialLoading ? (
            <ListSkeleton count={5} />
          ) : robots.length === 0 ? (
            <EmptyState>등록된 로봇이 없습니다.</EmptyState>
          ) : (
            <RobotManagementRows
              missions={missions}
              robots={visibleRobots}
              selectedRobot={selectedRobot}
              onSelectRobot={onSelectRobot}
            />
          )}
        </div>
      </Surface>

      <Surface className="grid min-h-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden" padding="none">
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
    </section>
  );
}

function RobotManagementRows({ missions, onSelectRobot, robots, selectedRobot }) {
  if (robots.length === 0) {
    return (
      <div className="rounded-lg border border-slate-500/20 bg-white/[0.03] px-3 py-4 text-sm font-semibold text-slate-500">
        해당 상태의 로봇이 없습니다.
      </div>
    );
  }

  return (
    <div className="grid gap-1.5">
      <div className="grid min-h-9 grid-cols-[minmax(160px,1.3fr)_120px_120px_140px_auto] items-center gap-3 rounded-lg border border-slate-700/50 bg-command-950/50 px-3 text-xs font-black text-slate-500 max-[760px]:hidden">
        <span>로봇</span>
        <span>상태</span>
        <span>현재 임무</span>
        <span>최근 연결</span>
        <span className="text-right">관리</span>
      </div>
      {robots.map((robot) => {
        const openMission = findRobotOpenMission(robot.robotCode, missions);
        const isSelectedRobot = selectedRobot?.robotCode === robot.robotCode;
        return (
          <button
            aria-label={`${robot.displayName} ${robot.robotCode} 선택`}
            aria-pressed={isSelectedRobot}
            className={[
              "grid min-h-[72px] w-full grid-cols-[minmax(160px,1.3fr)_120px_120px_140px_auto] items-center gap-3 rounded-lg border border-slate-700/70 bg-white/[0.035] px-3 py-2 text-left transition",
              "hover:border-sapphire-400/35 hover:bg-sapphire-500/[0.08] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sapphire-500",
              isSelectedRobot ? "border-sapphire-400/45 bg-sapphire-500/[0.10] shadow-[inset_3px_0_0_var(--color-sapphire)]" : "",
              "max-[760px]:grid-cols-[minmax(0,1fr)_auto]"
            ].join(" ")}
            key={robot.robotCode}
            type="button"
            onClick={() => onSelectRobot(robot.robotCode)}
          >
            <div className="min-w-0">
              <strong className="block truncate text-sm font-bold text-slate-100">{robot.displayName}</strong>
              <span className="mt-1 block truncate text-xs font-semibold text-slate-500">{robot.robotCode} · {robot.modelName || "모델 미지정"}</span>
            </div>
            <div className="max-[760px]:hidden">
              <StatusBadge size="xs" tone={makeRobotStatusTone(robot.status)}>{makeStatusLabel(robot.status)}</StatusBadge>
            </div>
            <span className="truncate text-xs font-bold text-slate-400 max-[760px]:hidden">{openMission?.missionCode ?? "-"}</span>
            <span className="truncate text-xs font-bold text-slate-500 max-[760px]:hidden">{formatDateTime(robot.lastSeenAt)}</span>
            <StatusBadge className="justify-self-end" size="xs" tone={makeRobotStatusTone(robot.status)}>{makeStatusLabel(robot.status)}</StatusBadge>
          </button>
        );
      })}
    </div>
  );
}
