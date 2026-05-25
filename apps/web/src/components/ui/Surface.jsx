import { cn } from "../../utils/cn.js";

export default function Surface({ as: Component = "article", children, className }) {
  return (
    <Component className={cn("rounded-[14px] border border-slate-500/25 bg-command-800/95 p-4 shadow-command", className)}>
      {children}
    </Component>
  );
}
