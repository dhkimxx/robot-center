import { cn } from "../../../utils/cn.js";

export function MetricTile({ compact = false, label, unit, value }) {
  return (
    <div className={cn(
      "grid content-between rounded-lg border border-slate-500/20 bg-white/[0.045]",
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
