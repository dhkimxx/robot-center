import { cn } from "../../utils/cn.js";
import Button from "./Button.jsx";

const pageSizeOptions = [10, 20, 50];

export default function PaginationControls({
  className,
  onPageChange,
  onPageSizeChange,
  page,
  pageSizeOptions: options = pageSizeOptions
}) {
  return (
    <div className={cn("flex min-w-0 flex-wrap items-center justify-between gap-2 border-t border-slate-700/60 pt-2", className)}>
      <span className="text-xs font-bold text-slate-500">
        {page.totalItems === 0 ? "0건" : `${page.startItem}-${page.endItem} / ${page.totalItems}건`}
      </span>
      <div className="flex min-w-0 items-center gap-2">
        <label className="flex min-w-0 items-center gap-2 text-xs font-bold text-slate-500">
          <span>페이지</span>
          <select
            className="h-8 rounded-lg border border-slate-700/70 bg-command-950 px-2 text-xs font-bold text-slate-200 outline-none focus:border-sapphire-500"
            value={page.pageSize}
            onChange={(event) => onPageSizeChange(Number(event.target.value))}
          >
            {options.map((option) => (
              <option key={option} value={option}>{option}</option>
            ))}
          </select>
        </label>
        <Button
          disabled={page.pageIndex <= 0}
          size="sm"
          variant="secondary"
          onClick={() => onPageChange(page.pageIndex - 1)}
        >
          이전
        </Button>
        <span className="min-w-12 text-center text-xs font-black text-slate-400">
          {page.pageIndex + 1}/{page.totalPages}
        </span>
        <Button
          disabled={page.pageIndex >= page.totalPages - 1}
          size="sm"
          variant="secondary"
          onClick={() => onPageChange(page.pageIndex + 1)}
        >
          다음
        </Button>
      </div>
    </div>
  );
}
