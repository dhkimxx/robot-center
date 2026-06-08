import EmptyState from "../../../components/ui/EmptyState.jsx";
import SectionHeader from "../../../components/ui/SectionHeader.jsx";
import Surface from "../../../components/ui/Surface.jsx";
import { formatDateTime } from "../../../utils/formatters.js";
import { cn } from "../../../utils/cn.js";

const severityLabels = {
  critical: "심각",
  info: "정보",
  notice: "알림",
  warning: "주의"
};

function makeSeverityClass(severity) {
  switch (severity) {
    case "critical":
      return "border-rose-400/55 bg-rose-400/15 text-rose-100";
    case "warning":
      return "border-amber-300/50 bg-amber-300/15 text-amber-100";
    case "notice":
      return "border-sky-300/45 bg-sky-300/10 text-sky-100";
    default:
      return "border-slate-500/25 bg-slate-500/10 text-slate-200";
  }
}

export function EventPanel({ className = "", liveEvents }) {
  return (
    <Surface className={cn("grid min-h-0 grid-rows-[auto_minmax(0,1fr)] gap-3 overflow-hidden", className)}>
      <SectionHeader className="mb-0" title="이벤트" meta={`${liveEvents.length}건`} />
      <div className="grid min-h-0 content-start gap-2 overflow-auto pr-1">
        {liveEvents.length === 0 ? (
          <EmptyState>관제 연결 이벤트가 없습니다.</EmptyState>
        ) : (
          liveEvents.map((event) => (
            <div
              className={cn(
                "grid gap-1 rounded-lg border bg-white/[0.045] px-3 py-2",
                event.severity === "critical" && "border-rose-400/55",
                event.severity === "warning" && "border-amber-300/50",
                event.severity === "notice" && "border-sky-300/45",
                !["critical", "warning", "notice"].includes(event.severity) && "border-slate-500/20"
              )}
              key={event.id}
            >
              <div className="flex min-w-0 items-center gap-2">
                <span
                  className={cn(
                    "shrink-0 rounded-full border px-2 py-0.5 text-[11px] font-black leading-none",
                    makeSeverityClass(event.severity)
                  )}
                >
                  {severityLabels[event.severity] ?? severityLabels.info}
                </span>
                <span className="min-w-0 truncate text-xs font-semibold text-slate-500">{formatDateTime(event.at)}</span>
              </div>
              <strong className="truncate text-sm font-bold leading-snug text-slate-50">{event.message}</strong>
              {event.description ? (
                <span className="text-xs font-semibold leading-snug text-slate-400">{event.description}</span>
              ) : null}
              {event.code ? (
                <span className="truncate text-[11px] font-semibold text-slate-500">{event.code}</span>
              ) : null}
            </div>
          ))
        )}
      </div>
    </Surface>
  );
}
