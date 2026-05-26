import { cn } from "../../utils/cn.js";

export default function ListRow({
  as: Component = "button",
  children,
  className,
  description,
  disabled = false,
  meta,
  muted = false,
  onClick,
  right,
  selected = false,
  title,
  type = "button",
  ...props
}) {
  const isButton = Component === "button";

  return (
    <Component
      className={cn(
        "grid min-h-[64px] w-full grid-cols-[minmax(0,1fr)_auto] items-center gap-3 rounded-lg border border-slate-700/70 bg-white/[0.035] px-3 py-2 text-left transition",
        isButton && "hover:border-sapphire-400/35 hover:bg-sapphire-500/[0.08] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sapphire-500",
        selected && "border-sapphire-400/45 bg-sapphire-500/[0.10] shadow-[inset_3px_0_0_var(--color-sapphire)]",
        muted && "border-slate-800/80 bg-slate-800/35 text-slate-500",
        disabled && "pointer-events-none opacity-45",
        className
      )}
      disabled={isButton ? disabled : undefined}
      onClick={onClick}
      type={isButton ? type : undefined}
      {...props}
    >
      <div className="grid min-w-0 gap-1">
        {title ? <strong className={cn("truncate text-sm font-bold text-slate-100", muted && "text-slate-400")}>{title}</strong> : null}
        {meta ? <span className={cn("truncate text-xs font-semibold text-slate-400", muted && "text-slate-500")}>{meta}</span> : null}
        {description ? <span className={cn("truncate text-xs font-semibold text-slate-500", muted && "text-slate-600")}>{description}</span> : null}
        {children}
      </div>
      {right ? <div className="flex shrink-0 items-center justify-end gap-2">{right}</div> : null}
    </Component>
  );
}
