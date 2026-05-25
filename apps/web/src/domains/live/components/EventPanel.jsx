import EmptyState from "../../../components/ui/EmptyState.jsx";
import SectionHeader from "../../../components/ui/SectionHeader.jsx";
import Surface from "../../../components/ui/Surface.jsx";
import { formatDateTime } from "../../../utils/formatters.js";

export function EventPanel({ liveEvents }) {
  return (
    <Surface className="grid gap-3">
      <SectionHeader className="mb-0" title="이벤트" meta={`${liveEvents.length}건`} />
      <div className="grid max-h-[220px] content-start gap-2 overflow-auto pr-1">
        {liveEvents.length === 0 ? (
          <EmptyState>관제 연결 이벤트가 없습니다.</EmptyState>
        ) : (
          liveEvents.map((event) => (
            <div className="grid gap-1 rounded-lg border border-slate-500/20 bg-white/[0.045] px-3 py-2" key={event.id}>
              <span className="text-xs font-semibold text-slate-500">{formatDateTime(event.at)}</span>
              <strong className="text-sm font-bold leading-snug text-slate-50">{event.message}</strong>
            </div>
          ))
        )}
      </div>
    </Surface>
  );
}
