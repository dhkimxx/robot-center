import { cn } from "../../utils/cn.js";

export default function SegmentedControl({ className = "", onChange, options, value }) {
  return (
    <div className={cn("inline-flex min-h-9 rounded-lg border border-slate-600/45 bg-command-950/70 p-1", className)}>
      {options.map((option) => {
        const active = option.value === value;
        return (
          <button
            className={cn(
              "inline-flex min-h-7 items-center justify-center gap-1.5 rounded-md px-3 text-xs font-black transition",
              active ? "bg-sapphire-500 text-white shadow-sm shadow-sapphire-950/30" : "text-slate-400 hover:bg-white/[0.06] hover:text-slate-100"
            )}
            key={option.value}
            type="button"
            onClick={() => onChange(option.value)}
          >
            <span>{option.label}</span>
            {option.count || option.count === 0 ? (
              <span className={cn("tabular-nums", active ? "text-sapphire-100" : "text-slate-500")}>{option.count}</span>
            ) : null}
          </button>
        );
      })}
    </div>
  );
}
