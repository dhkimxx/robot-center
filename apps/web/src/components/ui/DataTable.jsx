import { LuArrowDown, LuArrowUp, LuChevronsUpDown } from "react-icons/lu";
import { cn } from "../../utils/cn.js";
import EmptyState from "./EmptyState.jsx";

export default function DataTable({
  columns,
  emptyLabel = "표시할 항목이 없습니다.",
  getRowKey,
  gridTemplateClass,
  onRowClick,
  onSortChange,
  rowAriaLabel,
  rowClassName,
  rows,
  selectedRowKey,
  sortDirection = "asc",
  sortKey = ""
}) {
  if (rows.length === 0) {
    return <EmptyState>{emptyLabel}</EmptyState>;
  }

  return (
    <div className="grid gap-1.5">
      <div className={cn("grid min-h-9 items-center gap-3 rounded-lg border border-slate-700/50 bg-command-950/50 px-3 text-xs font-black text-slate-500 max-[760px]:hidden", gridTemplateClass)}>
        {columns.map((column) => (
          <DataTableHeaderCell
            column={column}
            key={column.key}
            onSortChange={onSortChange}
            sortDirection={sortDirection}
            sortKey={sortKey}
          />
        ))}
      </div>
      {rows.map((row) => {
        const rowKey = getRowKey(row);
        const isSelected = selectedRowKey === rowKey;
        const RowComponent = onRowClick ? "button" : "div";
        return (
          <RowComponent
            aria-label={rowAriaLabel ? rowAriaLabel(row) : undefined}
            aria-pressed={onRowClick ? isSelected : undefined}
            className={cn(
              "grid min-h-[72px] w-full items-center gap-3 rounded-lg border border-slate-700/70 bg-white/[0.035] px-3 py-2 text-left transition",
              onRowClick && "hover:border-sapphire-400/35 hover:bg-sapphire-500/[0.08] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sapphire-500",
              isSelected && "border-sapphire-400/45 bg-sapphire-500/[0.10] shadow-[inset_3px_0_0_var(--color-sapphire)]",
              gridTemplateClass,
              rowClassName
            )}
            key={rowKey}
            type={onRowClick ? "button" : undefined}
            onClick={onRowClick ? () => onRowClick(row) : undefined}
          >
            {columns.map((column) => (
              <div className={cn("min-w-0", column.className)} key={column.key}>
                {column.render(row)}
              </div>
            ))}
          </RowComponent>
        );
      })}
    </div>
  );
}

function DataTableHeaderCell({ column, onSortChange, sortDirection, sortKey }) {
  const isSortable = Boolean(column.sortKey && onSortChange);
  const isActiveSort = sortKey === column.sortKey;
  if (!isSortable) {
    return (
      <span className={cn("truncate", column.headerClassName, column.className)}>{column.label}</span>
    );
  }
  return (
    <button
      aria-label={`${column.label} 정렬${isActiveSort ? `, 현재 ${sortDirection === "asc" ? "오름차순" : "내림차순"}` : ""}`}
      aria-pressed={isActiveSort}
      className={cn(
        "group inline-flex min-w-0 items-center gap-1.5 rounded-md px-1.5 py-1 text-left transition hover:bg-slate-700/35 hover:text-slate-100",
        isActiveSort ? "bg-sapphire-500/10 text-sapphire-100" : "text-slate-500",
        column.headerClassName,
        column.className
      )}
      type="button"
      onClick={() => onSortChange(column.sortKey)}
    >
      <span className="truncate">{column.label}</span>
      <SortIcon active={isActiveSort} direction={sortDirection} />
    </button>
  );
}

function SortIcon({ active, direction }) {
  const Icon = active
    ? direction === "asc"
      ? LuArrowUp
      : LuArrowDown
    : LuChevronsUpDown;

  return (
    <Icon
      aria-hidden="true"
      className={cn(
        "size-3.5 shrink-0 transition",
        active ? "text-sapphire-300" : "text-slate-600 group-hover:text-slate-300"
      )}
    />
  );
}
