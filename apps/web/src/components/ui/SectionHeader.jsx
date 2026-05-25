import { cn } from "../../utils/cn.js";

export default function SectionHeader({ action, className, meta, title }) {
  return (
    <div className={cn("mb-4 flex min-w-0 items-center justify-between gap-4", className)}>
      <div className="min-w-0">
        <h2 className="truncate text-lg font-bold text-slate-50">{title}</h2>
        {meta ? <span className="mt-1 block truncate text-sm font-semibold text-slate-400">{meta}</span> : null}
      </div>
      {action ? <div className="shrink-0">{action}</div> : null}
    </div>
  );
}
