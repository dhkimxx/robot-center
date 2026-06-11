import { cn } from "../../../utils/cn.js";

export function RobotLiveStatusChips({ className = "", summary, target }) {
  if (!target || !summary) {
    return null;
  }
  return (
    <div className={cn("flex min-w-0 flex-wrap items-center gap-1.5", className)}>
      <StateChip active={target.isStreaming} tone={target.isStreaming ? "streaming" : "idle"}>
        {summary.streamLabel}
      </StateChip>
      <StateChip active={summary.recordingState.isActive} tone={summary.recordingState.tone}>
        {summary.recordingLabel}
      </StateChip>
      <StateChip tone={summary.connectionLabel === "연결됨" ? "connected" : summary.connectionLabel === "장애" ? "danger" : "idle"}>
        {summary.connectionLabel}
      </StateChip>
      {summary.channelLabel ? (
        <span className="truncate text-xs font-semibold text-slate-500">{summary.channelLabel}</span>
      ) : null}
    </div>
  );
}

function StateChip({ active = false, children, tone }) {
  return (
    <span
      className={cn(
        "inline-flex h-6 items-center gap-1.5 rounded-full border px-2 text-[11px] font-bold",
        tone === "streaming" && "border-emerald-400/25 bg-emerald-400/[0.10] text-emerald-100",
        tone === "recording" && "border-amber-300/25 bg-amber-300/[0.10] text-amber-100",
        tone === "available" && "border-sapphire-400/25 bg-sapphire-400/[0.10] text-sapphire-100",
        tone === "connected" && "border-sapphire-400/25 bg-sapphire-400/[0.10] text-sapphire-100",
        tone === "danger" && "border-red-400/25 bg-red-400/[0.10] text-red-100",
        tone === "idle" && "border-slate-500/20 bg-white/[0.04] text-slate-400"
      )}
    >
      <span
        aria-hidden
        className={cn(
          "h-1.5 w-1.5 rounded-full",
          tone === "streaming" && "bg-emerald-300",
          tone === "recording" && "bg-amber-200",
          (tone === "available" || tone === "connected") && "bg-sapphire-300",
          tone === "danger" && "bg-red-300",
          tone === "idle" && "bg-slate-500",
          active && "motion-safe:animate-pulse"
        )}
      />
      {children}
    </span>
  );
}
