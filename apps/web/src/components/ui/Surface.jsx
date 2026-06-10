import { cn } from "../../utils/cn.js";

const surfaceVariants = {
  panel: "border-slate-700/70 bg-command-850/90",
  section: "border-slate-700/60 bg-command-800/65",
  subtle: "border-slate-800/80 bg-slate-950/20"
};

const surfacePadding = {
  md: "p-4",
  none: "p-0",
  sm: "p-3"
};

export default function Surface({
  as: Component = "article",
  children,
  className,
  padding = "md",
  variant = "panel",
  ...props
}) {
  return (
    <Component className={cn("rounded-xl border shadow-none", surfaceVariants[variant], surfacePadding[padding], className)} {...props}>
      {children}
    </Component>
  );
}
