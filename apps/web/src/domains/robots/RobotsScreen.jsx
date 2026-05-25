import { getMissionRobotCodes } from "../missions/missionHelpers.js";
import { formatDateTime, makeStatusLabel } from "../../utils/formatters.js";
import Button from "../../components/ui/Button.jsx";
import EmptyState from "../../components/ui/EmptyState.jsx";
import SectionHeader from "../../components/ui/SectionHeader.jsx";
import Surface from "../../components/ui/Surface.jsx";
import { cn } from "../../utils/cn.js";

export default function RobotsScreen({
  missions,
  onArchiveRobot,
  onLoadConnectionInfo,
  onOpenCreateRobotModal,
  onOpenEditRobotModal,
  onSelectRobot,
  robots,
  selectedRobot
}) {
  const selectedRobotHasOpenMission = selectedRobot
    ? missions.some((mission) => getMissionRobotCodes(mission).includes(selectedRobot.robotCode) && ["ready", "active"].includes(mission.status))
    : false;

  return (
    <section className="grid h-full min-h-0 grid-cols-[minmax(0,1.36fr)_minmax(340px,0.82fr)] items-stretch gap-3.5 max-[1180px]:grid-cols-1">
      <Surface className="grid min-h-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden">
        <SectionHeader
          action={<Button size="sm" variant="primary" onClick={onOpenCreateRobotModal}>로봇 등록</Button>}
          meta={`${robots.length}대`}
          title="등록 로봇"
        />
        <div className="grid min-h-0 auto-rows-max gap-2 overflow-auto pr-1">
          {robots.length === 0 ? (
            <EmptyState>등록된 로봇이 없습니다.</EmptyState>
          ) : (
            robots.map((robot) => {
              const isSelectedRobot = selectedRobot?.robotCode === robot.robotCode;
              return (
                <button
                  aria-label={`${robot.displayName} ${robot.robotCode} 선택`}
                  aria-pressed={isSelectedRobot}
                  className={cn(
                    "grid min-h-[52px] w-full gap-1 rounded-lg border border-slate-500/20 bg-white/[0.045] px-3 py-2 text-left transition hover:border-sapphire-500/[0.45] hover:bg-sapphire-500/[0.12]",
                    isSelectedRobot && "border-sapphire-500/55 bg-sapphire-500/[0.10] shadow-[inset_3px_0_0_var(--color-sapphire)]"
                  )}
                  key={robot.robotCode}
                  type="button"
                  onClick={() => onSelectRobot(robot.robotCode)}
                >
                  <strong className="truncate text-sm font-bold leading-tight text-slate-50">{robot.displayName}</strong>
                  <span className="truncate text-xs font-semibold text-slate-400">
                    {robot.robotCode} / {makeStatusLabel(robot.status)} / 최근 {formatDateTime(robot.lastSeenAt)}
                  </span>
                </button>
              );
            })
          )}
        </div>
      </Surface>

      <section className="grid min-h-0 content-start overflow-auto">
        <Surface>
          <SectionHeader meta={selectedRobot?.robotCode ?? "선택 없음"} title="로봇 상세" />
          {!selectedRobot ? (
            <EmptyState>로봇을 선택하세요.</EmptyState>
          ) : (
            <div className="grid gap-5">
              <div className="min-w-0">
                <strong className="block truncate text-xl font-bold leading-tight text-slate-50">{selectedRobot.displayName}</strong>
                <span className="mt-2 block text-sm font-semibold text-slate-400">{selectedRobot.modelName || "모델 미지정"}</span>
              </div>
              <div className="grid gap-3 rounded-xl border border-slate-500/20 bg-white/[0.045] p-4">
                <div className="grid grid-cols-[76px_minmax(0,1fr)] gap-3 text-sm">
                  <span className="font-semibold text-slate-400">상태</span>
                  <span className="font-semibold text-slate-200">{makeStatusLabel(selectedRobot.status)}</span>
                </div>
                <div className="grid grid-cols-[76px_minmax(0,1fr)] gap-3 text-sm">
                  <span className="font-semibold text-slate-400">최근 연결</span>
                  <span className="font-semibold text-slate-200">{formatDateTime(selectedRobot.lastSeenAt)}</span>
                </div>
                <div className="grid grid-cols-[76px_minmax(0,1fr)] gap-3 text-sm">
                  <span className="font-semibold text-slate-400">최근 송출</span>
                  <span className="font-semibold text-slate-200">{formatDateTime(selectedRobot.lastStreamingAt)}</span>
                </div>
                {selectedRobotHasOpenMission ? (
                  <span className="rounded-lg border border-amber-300/20 bg-amber-300/10 px-3 py-2 text-sm font-semibold text-amber-200">
                    삭제 불가: 진행/대기 임무 배정
                  </span>
                ) : null}
              </div>
              <div className="flex flex-wrap justify-end gap-3 pt-1">
                <Button variant="primary" onClick={onOpenEditRobotModal}>수정</Button>
                <Button onClick={() => onLoadConnectionInfo(selectedRobot.robotCode)}>연결 정보</Button>
                <Button
                  variant="danger"
                  disabled={selectedRobotHasOpenMission}
                  onClick={() => onArchiveRobot(selectedRobot.robotCode)}
                >
                  삭제
                </Button>
              </div>
            </div>
          )}
        </Surface>
      </section>
    </section>
  );
}
