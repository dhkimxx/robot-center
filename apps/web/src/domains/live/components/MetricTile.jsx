import { cn } from "../../../utils/cn.js";

const alarmStyles = {
  critical: "border-rose-400/70 bg-rose-500/15 text-rose-50 shadow-[0_0_0_1px_rgba(244,63,94,0.12)]",
  normal: "border-slate-500/20 bg-white/[0.045]",
  warning: "border-amber-300/70 bg-amber-400/15 text-amber-50 shadow-[0_0_0_1px_rgba(251,191,36,0.12)]"
};

export function MetricTile({ alarmLevel = "normal", compact = false, label, unit, value }) {
  const level = alarmStyles[alarmLevel] ? alarmLevel : "normal";
  return (
    <div className={cn(
      "grid content-between rounded-lg border",
      alarmStyles[level],
      compact ? "min-h-[86px] p-2.5" : "p-3"
    )}>
      <span className="truncate text-xs font-bold text-slate-400" title={label}>{label}</span>
      <strong
        className={cn(
          "mt-2 truncate font-extrabold leading-tight text-slate-50",
          compact ? "text-lg" : "text-2xl"
        )}
        title={String(value)}
      >
        {value}
      </strong>
      <small className="mt-1 truncate text-xs font-semibold text-slate-500" title={unit}>{unit}</small>
    </div>
  );
}
