import { cn } from "../../utils/cn.js";

export default function ListFilterInput({
  className,
  onChange,
  placeholder = "검색",
  value
}) {
  return (
    <input
      className={cn(
        "h-9 min-w-0 rounded-lg border border-slate-700/70 bg-command-950/70 px-3 text-sm font-semibold text-slate-100 outline-none placeholder:text-slate-600 focus:border-sapphire-500",
        className
      )}
      placeholder={placeholder}
      type="search"
      value={value}
      onChange={(event) => onChange(event.target.value)}
    />
  );
}
