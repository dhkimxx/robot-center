import SectionHeader from "../../../components/ui/SectionHeader.jsx";
import Surface from "../../../components/ui/Surface.jsx";
import { cn } from "../../../utils/cn.js";

const statusCellToneClasses = {
  danger: "border-l-red-400/70",
  ok: "border-l-emerald-400/70",
  waiting: "border-l-amber-300/70"
};

export function ConnectionStatusPanel({ compact = false, statuses }) {
  return (
    <Surface
      as="article"
      className={cn(
        "grid gap-3",
        compact && "rounded-xl p-3 shadow-none"
      )}
    >
      <SectionHeader
        className="mb-0"
        title="연결 상태"
        meta={`${statuses.filter((status) => status.tone === "ok").length}/${statuses.length}`}
      />
      <div className="grid grid-cols-2 gap-2 max-[540px]:grid-cols-1">
        {statuses.map((status) => (
          <div
            className={cn(
              "grid min-h-[92px] gap-1 rounded-lg border border-l-4 border-slate-500/20 border-l-slate-400/50 bg-white/[0.045] p-3",
              statusCellToneClasses[status.tone]
            )}
            key={status.label}
          >
            <span className="text-xs font-bold text-slate-400">{status.label}</span>
            <strong className="text-base font-extrabold text-slate-50">{status.value}</strong>
            <small className="text-xs font-semibold leading-snug text-slate-500">{status.detail}</small>
          </div>
        ))}
      </div>
    </Surface>
  );
}
