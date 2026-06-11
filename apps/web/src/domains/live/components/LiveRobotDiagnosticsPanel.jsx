import SectionHeader from "../../../components/ui/SectionHeader.jsx";
import Surface from "../../../components/ui/Surface.jsx";
import { cn } from "../../../utils/cn.js";
import { makeLiveRobotDiagnostics } from "../liveDiagnostics.js";

const diagnosticToneClasses = {
  danger: "border-l-red-400/70",
  ok: "border-l-emerald-400/70",
  waiting: "border-l-amber-300/70"
};

export function LiveRobotDiagnosticsPanel({ session, target }) {
  const diagnostics = makeLiveRobotDiagnostics({ session, target });
  const okCount = diagnostics.filter((diagnostic) => diagnostic.tone === "ok").length;

  return (
    <Surface as="article" className="grid gap-3">
      <SectionHeader
        className="mb-0"
        title="선택 로봇 진단"
        meta={target ? `${okCount}/${diagnostics.length}` : "대기"}
      />
      <div className="grid gap-2">
        {diagnostics.map((diagnostic) => (
          <div
            className={cn(
              "grid min-h-[68px] gap-1 rounded-lg border border-l-4 border-slate-500/20 border-l-slate-400/50 bg-white/[0.045] p-3",
              diagnosticToneClasses[diagnostic.tone]
            )}
            key={diagnostic.key}
          >
            <div className="flex min-w-0 items-center justify-between gap-2">
              <span className="truncate text-xs font-bold text-slate-400">{diagnostic.label}</span>
              <strong className="shrink-0 text-sm font-extrabold text-slate-50">{diagnostic.value}</strong>
            </div>
            <small className="break-words text-xs font-semibold leading-snug text-slate-500">{diagnostic.detail}</small>
          </div>
        ))}
      </div>
    </Surface>
  );
}
