import { cn } from "../../utils/cn.js";

export default function MetricStrip({ className, items }) {
  return (
    <div className={cn("flex min-h-9 flex-wrap items-center gap-2 rounded-lg border border-slate-700/70 bg-slate-950/20 px-3 text-xs font-semibold text-slate-400", className)}>
      {items.map(({ label, tone = "neutral", value }) => (
        <span
          className={cn(
            "inline-flex items-center gap-1 rounded-full px-2 py-1",
            tone === "active" ? "bg-sapphire-400/[0.12] text-sapphire-100" : "bg-white/[0.04] text-slate-400"
          )}
          key={label}
        >
          {label}
          <strong className="text-slate-100">{value}</strong>
        </span>
      ))}
    </div>
  );
}
