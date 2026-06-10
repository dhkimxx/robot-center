import StatusBadge from "../../components/ui/StatusBadge.jsx";
import { cn } from "../../utils/cn.js";
import {
  getRobotDisplayName,
  makeFileAvailabilityLabel,
  makeRobotReplayMeta,
  makeRobotSummaryStatusLabel,
  makeRobotSummaryTone
} from "./missionReplayHelpers.js";
import { formatDateTime } from "../../utils/formatters.js";

export function MissionReplayRobotList({
  robotDisplayNamesByCode,
  robotSummaries,
  selectedRobotCode,
  onSelectRobot
}) {
  return (
    <aside className="grid min-h-0 content-start gap-2 overflow-auto pr-1">
      {robotSummaries.map((robotSummary) => {
        const displayName = getRobotDisplayName(robotDisplayNamesByCode, robotSummary.robotCode);
        const isSelected = selectedRobotCode === robotSummary.robotCode;
        return (
          <button
            className={cn(
              "grid min-h-24 w-full gap-2 rounded-lg border border-slate-500/20 bg-white/[0.045] px-3 py-2.5 text-left transition hover:border-sapphire-500/[0.45] hover:bg-sapphire-500/[0.12]",
              isSelected && "border-sapphire-500/55 bg-sapphire-500/[0.10] shadow-[inset_3px_0_0_var(--color-sapphire)]"
            )}
            data-robot-code={robotSummary.robotCode}
            data-testid="mission-replay-robot-option"
            key={robotSummary.robotCode}
            type="button"
            onClick={() => onSelectRobot(robotSummary.robotCode)}
          >
            <div className="flex min-w-0 items-center justify-between gap-2">
              <strong className="truncate text-sm font-bold text-slate-50">{displayName}</strong>
              <StatusBadge size="xs" tone={makeRobotSummaryTone(robotSummary)}>
                {makeRobotSummaryStatusLabel(robotSummary)}
              </StatusBadge>
            </div>
            <span className="truncate text-xs font-semibold text-slate-400">{makeRobotReplayMeta(robotSummary, displayName)}</span>
            <div className="grid grid-cols-2 gap-1 text-[11px] font-semibold text-slate-500">
              <span className="truncate">{makeFileAvailabilityLabel(robotSummary, "rgb_audio_mp4", "RGB")}</span>
              <span className="truncate">{makeFileAvailabilityLabel(robotSummary, "thermal_mp4", "Thermal")}</span>
              <span className="truncate">{makeFileAvailabilityLabel(robotSummary, "sensor_jsonl", "Sensor")}</span>
              <span className="truncate">{makeFileAvailabilityLabel(robotSummary, "telemetry_jsonl", "Telemetry")}</span>
            </div>
            <small className="truncate text-xs font-semibold text-slate-500">최근 {formatDateTime(robotSummary.lastEndedAt)}</small>
          </button>
        );
      })}
    </aside>
  );
}
