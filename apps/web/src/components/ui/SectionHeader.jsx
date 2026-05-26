import { cn } from "../../utils/cn.js";

const sectionHeaderSizes = {
  compact: {
    title: "text-sm",
    meta: "text-xs"
  },
  page: {
    title: "text-lg",
    meta: "text-sm"
  },
  section: {
    title: "text-base",
    meta: "text-xs"
  }
};

export default function SectionHeader({ action, className, meta, size = "section", title }) {
  const sizeClasses = sectionHeaderSizes[size] ?? sectionHeaderSizes.section;

  return (
    <div className={cn("mb-3 flex min-w-0 items-center justify-between gap-4", className)}>
      <div className="min-w-0">
        <h2 className={cn("truncate font-bold text-slate-50", sizeClasses.title)}>{title}</h2>
        {meta ? <span className={cn("mt-1 block truncate font-semibold text-slate-500", sizeClasses.meta)}>{meta}</span> : null}
      </div>
      {action ? <div className="shrink-0">{action}</div> : null}
    </div>
  );
}
