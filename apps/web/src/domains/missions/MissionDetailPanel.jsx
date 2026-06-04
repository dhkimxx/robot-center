import {
  formatDateTimeFull,
  makeStatusLabel,
  missionTypeLabel
} from "../../utils/formatters.js";
import { Link } from "react-router-dom";
import Button from "../../components/ui/Button.jsx";
import DefinitionList from "../../components/ui/DefinitionList.jsx";
import ListRow from "../../components/ui/ListRow.jsx";
import StatusBadge from "../../components/ui/StatusBadge.jsx";
import Surface from "../../components/ui/Surface.jsx";
import {
  formatMissionRobotCount,
  isClosedMission,
  makeLiveRecordingTimingLabel,
  makeLiveStreamTimingLabel
} from "./missionHelpers.js";

export function MissionDetailPanel({
  mission,
  onEndMission,
  onOpenMissionControl,
  onOpenMissionReplay,
  onStartMission,
  robotDetails
}) {
  const isClosed = isClosedMission(mission);
  const missionOverviewItems = [
    { label: "시나리오", value: missionTypeLabel(mission.missionType) },
    { label: "상태", value: makeStatusLabel(mission.status) },
    { label: "배정 로봇", value: formatMissionRobotCount(robotDetails) },
    { label: "생성 시간", value: formatDateTimeFull(mission.createdAt) },
    { label: "시작 시간", value: mission.startedAt ? formatDateTimeFull(mission.startedAt) : "시작 전" },
    { label: "종료 시간", value: mission.endedAt ? formatDateTimeFull(mission.endedAt) : "-" },
    { className: "min-[981px]:col-span-3", label: "현장 메모", value: mission.siteNote || "-", wrap: true }
  ];

  return (
    <div className="grid min-h-0 content-start gap-4">
      <div className="flex min-w-0 items-start justify-between gap-4 max-[760px]:grid">
        <div className="min-w-0">
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <strong className="truncate text-lg font-bold leading-tight text-slate-50">{mission.name}</strong>
            <StatusBadge tone={isClosed ? "neutral" : mission.status === "active" ? "success" : "warning"}>
              {makeStatusLabel(mission.status)}
            </StatusBadge>
          </div>
          <span className="mt-1 block truncate text-sm font-semibold text-slate-400">
            {mission.missionCode} · {missionTypeLabel(mission.missionType)} · {formatMissionRobotCount(robotDetails)}
          </span>
        </div>
        <div className="flex shrink-0 flex-wrap justify-end gap-2 max-[760px]:justify-start">
          {isClosed ? (
            <Button
              as={Link}
              reloadDocument
              to={`/missions/${encodeURIComponent(mission.missionCode)}/replay`}
              variant="primary"
              onClick={() => onOpenMissionReplay(mission, { navigate: false })}
            >
              리플레이 보기
            </Button>
          ) : (
            <>
              <Button
                as={Link}
                reloadDocument
                to={`/missions/${encodeURIComponent(mission.missionCode)}/control`}
                variant="primary"
                onClick={() => onOpenMissionControl(mission, { navigate: false })}
              >
                관제 진입
              </Button>
              <Button
                disabled={mission.status !== "ready"}
                onClick={() => onStartMission(mission.missionCode)}
              >
                시작
              </Button>
              <Button
                disabled={mission.status !== "active"}
                onClick={() => onEndMission(mission.missionCode)}
              >
                종료
              </Button>
            </>
          )}
        </div>
      </div>

      <div className="grid min-h-0 gap-3">
        <Surface className="grid min-w-0 content-start gap-3" padding="sm" variant="section">
          <h3 className="text-xs font-bold uppercase tracking-normal text-slate-500">임무 개요</h3>
          <DefinitionList className="grid-cols-3 gap-x-6 max-[980px]:grid-cols-1" items={missionOverviewItems} />
        </Surface>

        <Surface className="grid min-w-0 content-start gap-2" padding="sm" variant="section">
          <div className="flex items-center justify-between gap-3">
            <h3 className="text-xs font-bold uppercase tracking-normal text-slate-500">배정 로봇</h3>
            <span className="text-xs font-bold text-slate-500">{robotDetails.length}대</span>
          </div>
          {robotDetails.length === 0 ? (
            <span className="inline-flex min-h-10 items-center rounded-lg border border-slate-500/20 bg-white/[0.04] px-3 text-sm font-semibold text-slate-400">
              미배정
            </span>
          ) : (
            <div className="grid grid-cols-2 gap-2 max-[980px]:grid-cols-1">
              {robotDetails.map((robot) => (
                <ListRow
                  as="div"
                  key={robot.robotCode}
                  meta={robot.robotCode}
                  right={(
                    <StatusBadge size="xs" tone={robot.isStreaming ? "success" : isClosed ? "neutral" : "warning"}>
                      {robot.liveLabel}
                    </StatusBadge>
                  )}
                  title={robot.displayName}
                >
                  <span className="truncate text-xs font-semibold text-slate-500">
                    {makeLiveStreamTimingLabel(robot.liveStatus?.stream)}
                  </span>
                  <span className="truncate text-xs font-semibold text-slate-600">
                    {makeLiveRecordingTimingLabel(robot.liveStatus?.recording)}
                  </span>
                </ListRow>
              ))}
            </div>
          )}
        </Surface>
      </div>
    </div>
  );
}
