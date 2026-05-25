import { cn } from "../../utils/cn.js";

export default function EmptyState({ children, className }) {
  return (
    <p className={cn("self-start rounded-xl border border-amber-300/20 bg-amber-300/10 px-4 py-3 text-sm font-semibold text-amber-200", className)}>
      {children}
    </p>
  );
}
