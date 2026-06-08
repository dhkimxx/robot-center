import { cn } from "../../utils/cn.js";

const statusBadgeTones = {
  danger: "border-red-400/30 bg-red-400/[0.12] text-red-100",
  info: "border-sapphire-500/30 bg-sapphire-500/[0.14] text-sapphire-100",
  neutral: "border-slate-500/30 bg-white/[0.06] text-slate-300",
  success: "border-emerald-400/30 bg-emerald-400/[0.12] text-emerald-100",
  warning: "border-amber-300/30 bg-amber-300/[0.12] text-amber-100"
};

const statusBadgeSizes = {
  sm: "min-h-7 px-2.5 text-xs",
  xs: "min-h-6 px-2 text-[11px]"
};

export default function StatusBadge({ children, className, size = "sm", tone = "neutral" }) {
  return (
    <span className={cn("inline-flex shrink-0 items-center justify-center whitespace-nowrap rounded-full border font-semibold leading-none", statusBadgeTones[tone], statusBadgeSizes[size], className)}>
      {children}
    </span>
  );
}
