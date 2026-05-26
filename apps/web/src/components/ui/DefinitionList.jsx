import { cn } from "../../utils/cn.js";

export default function DefinitionList({ className, items }) {
  return (
    <dl className={cn("grid gap-3", className)}>
      {items.map(({ className: itemClassName, label, value, wrap = false }) => (
        <div className={cn("grid grid-cols-[96px_minmax(0,1fr)] gap-4 text-sm max-[560px]:grid-cols-1 max-[560px]:gap-1", itemClassName)} key={label}>
          <dt className="font-semibold text-slate-500">{label}</dt>
          <dd className={cn("min-w-0 font-semibold text-slate-200", wrap ? "break-words leading-6" : "truncate whitespace-nowrap")}>
            {value}
          </dd>
        </div>
      ))}
    </dl>
  );
}
