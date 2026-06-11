import Surface from "../../../components/ui/Surface.jsx";
import { cn } from "../../../utils/cn.js";

const compactStatusToneClasses = {
  danger: "border-l-red-400/80 bg-red-400/[0.08]",
  ok: "border-l-emerald-400/80 bg-emerald-400/[0.08]",
  waiting: "border-l-amber-300/80 bg-amber-300/[0.08]"
};

export function ControlStatusSummary({ diagnostics, statuses }) {
  const summaryItems = [
    ...diagnostics.map((diagnostic) => ({
      detail: diagnostic.detail,
      key: `diagnostic-${diagnostic.key}`,
      label: diagnostic.label,
      tone: diagnostic.tone,
      value: diagnostic.value
    })),
    ...statuses.map((status) => ({
      detail: status.detail,
      key: `status-${status.label}`,
      label: status.label,
      tone: status.tone,
      value: status.value
    }))
  ];
  const okCount = summaryItems.filter((item) => item.tone === "ok").length;

  return (
    <Surface className="flex min-h-[76px] min-w-0 items-center gap-3 overflow-hidden px-3 py-2.5">
      <div className="grid w-12 shrink-0 gap-0.5">
        <span className="text-xs font-bold text-slate-500">상태</span>
        <strong className="text-sm font-extrabold text-slate-50">{okCount}/{summaryItems.length}</strong>
      </div>
      <div className="flex min-w-0 flex-1 gap-2 overflow-x-auto py-0.5">
        {summaryItems.map((item) => (
          <div
            className={cn(
              "grid min-h-10 min-w-[132px] gap-0.5 rounded-lg border border-l-4 border-slate-500/20 border-l-slate-400/60 px-2.5 py-1.5",
              compactStatusToneClasses[item.tone]
            )}
            key={item.key}
            title={item.detail}
          >
            <span className="truncate text-[11px] font-bold text-slate-500">{item.label}</span>
            <strong className="truncate text-sm font-extrabold leading-tight text-slate-50">{item.value}</strong>
          </div>
        ))}
      </div>
    </Surface>
  );
}
