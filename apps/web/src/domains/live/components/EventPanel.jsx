import EmptyState from "../../../components/ui/EmptyState.jsx";
import SectionHeader from "../../../components/ui/SectionHeader.jsx";
import Surface from "../../../components/ui/Surface.jsx";
import { formatDateTime } from "../../../utils/formatters.js";
import { cn } from "../../../utils/cn.js";

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
                !["critical", "warning"].includes(event.severity) && "border-slate-500/20"
              )}
              key={event.id}
            >
              <span className="text-xs font-semibold text-slate-500">{formatDateTime(event.at)}</span>
              <strong className="text-sm font-bold leading-snug text-slate-50">{event.message}</strong>
              {event.description ? (
                <span className="text-xs font-semibold leading-snug text-slate-400">{event.description}</span>
              ) : null}
            </div>
          ))
        )}
      </div>
    </Surface>
  );
}
