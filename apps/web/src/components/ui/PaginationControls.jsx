import { LuChevronLeft, LuChevronRight } from "react-icons/lu";
import { cn } from "../../utils/cn.js";

const visiblePageLimit = 7;

export default function PaginationControls({
  className,
  onPageChange,
  page
}) {
  const currentPage = page.pageIndex + 1;
  const visiblePages = createVisiblePages(currentPage, page.totalPages);

  return (
    <nav
      aria-label="페이지 이동"
      className={cn("flex min-w-0 items-center justify-center gap-3 border-t border-slate-700/60 pt-2", className)}
    >
      <PaginationButton
        disabled={page.pageIndex <= 0}
        label="이전 페이지"
        onClick={() => onPageChange(page.pageIndex - 1)}
      >
        <LuChevronLeft aria-hidden="true" className="size-5" />
      </PaginationButton>

      {visiblePages.map((pageNumber, index) => (
        pageNumber === "gap" ? (
          <span
            key={`gap-${index}`}
            aria-hidden="true"
            className="grid h-8 min-w-8 place-items-center text-sm font-bold text-slate-600"
          >
            ...
          </span>
        ) : (
          <PaginationButton
            key={pageNumber}
            active={pageNumber === currentPage}
            label={`${pageNumber} 페이지`}
            onClick={() => onPageChange(pageNumber - 1)}
          >
            {pageNumber}
          </PaginationButton>
        )
      ))}

      <PaginationButton
        disabled={page.pageIndex >= page.totalPages - 1}
        label="다음 페이지"
        onClick={() => onPageChange(page.pageIndex + 1)}
      >
        <LuChevronRight aria-hidden="true" className="size-5" />
      </PaginationButton>
    </nav>
  );
}

function PaginationButton({ active = false, children, disabled = false, label, onClick }) {
  return (
    <button
      aria-current={active ? "page" : undefined}
      aria-label={label}
      className={cn(
        "relative grid h-8 min-w-8 place-items-center border-b-2 border-transparent px-1 text-sm font-black text-slate-400 transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sapphire-500 disabled:pointer-events-none disabled:text-slate-700",
        "hover:text-slate-100",
        active && "border-sapphire-500 text-slate-50"
      )}
      disabled={disabled}
      type="button"
      onClick={onClick}
    >
      {children}
    </button>
  );
}

function createVisiblePages(currentPage, totalPages) {
  if (totalPages <= visiblePageLimit) {
    return Array.from({ length: totalPages }, (_, index) => index + 1);
  }

  const pages = new Set([1, totalPages, currentPage]);
  pages.add(Math.max(1, currentPage - 1));
  pages.add(Math.min(totalPages, currentPage + 1));

  const sortedPages = Array.from(pages).sort((left, right) => left - right);
  return sortedPages.flatMap((pageNumber, index) => {
    const previousPageNumber = sortedPages[index - 1];
    if (!previousPageNumber || pageNumber - previousPageNumber === 1) {
      return [pageNumber];
    }
    return ["gap", pageNumber];
  });
}
