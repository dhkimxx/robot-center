import { getMissionRobotCodes } from "../missions/missionHelpers.js";
import { formatDateTime, makeStatusLabel } from "../../utils/formatters.js";
import Button from "../../components/ui/Button.jsx";
import DefinitionList from "../../components/ui/DefinitionList.jsx";
import EmptyState from "../../components/ui/EmptyState.jsx";
import ListRow from "../../components/ui/ListRow.jsx";
import SectionHeader from "../../components/ui/SectionHeader.jsx";
import StatusBadge from "../../components/ui/StatusBadge.jsx";
import Surface from "../../components/ui/Surface.jsx";

export default function RobotsScreen({
  missions,
  onArchiveRobot,
  onLoadConnectionInfo,
  onOpenEditRobotModal,
  onSelectRobot,
  robots,
  selectedRobot
}) {
  const selectedRobotHasOpenMission = selectedRobot
    ? missions.some((mission) => getMissionRobotCodes(mission).includes(selectedRobot.robotCode) && ["ready", "active"].includes(mission.status))
    : false;

  return (
    <section className="grid h-full min-h-0 grid-cols-[400px_minmax(0,1fr)] items-stretch gap-3 max-[1180px]:grid-cols-1">
      <Surface className="grid min-h-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden">
        <SectionHeader
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
                <ListRow
                  aria-label={`${robot.displayName} ${robot.robotCode} 선택`}
                  aria-pressed={isSelectedRobot}
                  description={`최근 ${formatDateTime(robot.lastSeenAt)}`}
                  key={robot.robotCode}
                  meta={robot.robotCode}
                  onClick={() => onSelectRobot(robot.robotCode)}
                  right={<StatusBadge size="xs" tone={makeRobotStatusTone(robot.status)}>{makeStatusLabel(robot.status)}</StatusBadge>}
                  selected={isSelectedRobot}
                  title={robot.displayName}
                />
              );
            })
          )}
        </div>
      </Surface>

      <Surface className="grid min-h-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden" padding="none">
        <div className="border-b border-slate-700/70 px-4 pt-4">
          <SectionHeader meta={selectedRobot?.robotCode ?? "선택 없음"} title="로봇 상세" />
        </div>
        <div className="min-h-0 overflow-auto p-4">
          {!selectedRobot ? (
            <EmptyState>로봇을 선택하세요.</EmptyState>
          ) : (
            <div className="grid gap-4">
              <div className="flex min-w-0 items-start justify-between gap-4 max-[760px]:grid">
                <div className="min-w-0">
                  <div className="flex min-w-0 flex-wrap items-center gap-2">
                    <strong className="truncate text-lg font-bold leading-tight text-slate-50">{selectedRobot.displayName}</strong>
                    <StatusBadge tone={makeRobotStatusTone(selectedRobot.status)}>{makeStatusLabel(selectedRobot.status)}</StatusBadge>
                  </div>
                  <span className="mt-1 block truncate text-sm font-semibold text-slate-400">{selectedRobot.modelName || "모델 미지정"}</span>
                </div>
                <div className="flex shrink-0 flex-wrap justify-end gap-2 max-[760px]:justify-start">
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

              <Surface className="grid gap-3" padding="sm" variant="section">
                <h3 className="text-xs font-bold uppercase tracking-normal text-slate-500">로봇 개요</h3>
                <DefinitionList
                  items={[
                    { label: "상태", value: makeStatusLabel(selectedRobot.status) },
                    { label: "최근 연결", value: formatDateTime(selectedRobot.lastSeenAt) },
                    { label: "최근 송출", value: formatDateTime(selectedRobot.lastStreamingAt) }
                  ]}
                />
                {selectedRobotHasOpenMission ? (
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

function makeRobotStatusTone(status) {
  if (["online", "streaming", "active"].includes(status)) {
    return "success";
  }
  if (["ready", "idle"].includes(status)) {
    return "info";
  }
  if (["fault", "error", "failed"].includes(status)) {
    return "danger";
  }
  return "neutral";
}
