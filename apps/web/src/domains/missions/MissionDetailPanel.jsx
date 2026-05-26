import {
  makeStatusLabel,
  missionTypeLabel
} from "../../utils/formatters.js";
import Button from "../../components/ui/Button.jsx";
import { formatMissionRobotCount } from "./missionHelpers.js";

export function MissionDetailPanel({ mission, onEndMission, onOpenMissionControl, onStartMission, robotDetails }) {
  const metadataItems = [
    ["시나리오", missionTypeLabel(mission.missionType)],
    ["상태", makeStatusLabel(mission.status)],
    ["배정 로봇", formatMissionRobotCount(robotDetails)],
    ["현장 메모", mission.siteNote || "-"]
  ];

  return (
    <div className="grid gap-5">
      <div className="min-w-0">
        <strong className="block truncate text-xl font-bold leading-tight text-slate-50">{mission.name}</strong>
        <span className="mt-2 block text-sm font-semibold text-slate-400">{mission.missionCode}</span>
      </div>

      <div className="grid gap-3 rounded-xl border border-slate-500/20 bg-white/[0.045] p-4">
        {metadataItems.map(([label, value]) => (
          <div className="grid grid-cols-[76px_minmax(0,1fr)] gap-3 text-sm" key={label}>
            <span className="font-semibold text-slate-400">{label}</span>
            <span className="min-w-0 break-words font-semibold text-slate-200">{value}</span>
          </div>
        ))}
      </div>

      <div className="flex flex-wrap gap-2">
        {robotDetails.length === 0 ? (
          <span className="inline-flex min-h-8 items-center rounded-full border border-slate-500/20 bg-white/[0.04] px-3 text-sm font-semibold text-slate-400">
            미배정
          </span>
        ) : (
          robotDetails.map((robot) => (
            <span
              className="inline-flex min-h-8 max-w-full items-center rounded-full border border-slate-500/30 bg-white/[0.06] px-3 text-sm font-semibold text-slate-100"
              key={robot.robotCode}
            >
              {robot.robotCode} · {robot.liveLabel}
            </span>
          ))
        )}
      </div>

      <div className="flex flex-wrap justify-end gap-3 pt-1">
        <Button variant="primary" onClick={() => onOpenMissionControl(mission)}>
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
      </div>
    </div>
  );
}
