const defaultPageSizeOptions = [10, 20, 50];

export function normalizeListViewState({
  pageIndex = 0,
  pageSize = defaultPageSizeOptions[0],
  sortDirection = "asc",
  sortKey = ""
} = {}) {
  const normalizedPageSize = Math.max(1, Number(pageSize) || defaultPageSizeOptions[0]);
  return {
    pageIndex: Math.max(0, Number(pageIndex) || 0),
    pageSize: normalizedPageSize,
    sortDirection: sortDirection === "desc" ? "desc" : "asc",
    sortKey
  };
}

export function createNextSortState({ currentDirection, currentKey, nextKey }) {
  if (currentKey !== nextKey) {
    return {
      sortDirection: "asc",
      sortKey: nextKey
    };
  }
  return {
    sortDirection: currentDirection === "asc" ? "desc" : "asc",
    sortKey: currentKey
  };
}

export function filterListItems(items, filterText, getFilterValues) {
  const normalizedFilter = String(filterText ?? "").trim().toLowerCase();
  if (!normalizedFilter) {
    return [...items];
  }
  return items.filter((item) => (
    getFilterValues(item)
      .filter((value) => value !== null && value !== undefined)
      .some((value) => String(value).toLowerCase().includes(normalizedFilter))
  ));
}

export function sortListItems(items, { sortDirection = "asc", sortKey = "" } = {}, getSortValue) {
  if (!sortKey) {
    return [...items];
  }
  const direction = sortDirection === "desc" ? -1 : 1;
  return [...items].sort((left, right) => {
    const leftValue = normalizeSortValue(getSortValue(left, sortKey));
    const rightValue = normalizeSortValue(getSortValue(right, sortKey));
    if (leftValue < rightValue) {
      return -1 * direction;
    }
    if (leftValue > rightValue) {
      return direction;
    }
    return 0;
  });
}

export function paginateListItems(items, { pageIndex = 0, pageSize = defaultPageSizeOptions[0] } = {}) {
  const totalItems = items.length;
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize));
  const safePageIndex = Math.min(Math.max(0, pageIndex), totalPages - 1);
  const startIndex = safePageIndex * pageSize;
  const pageItems = items.slice(startIndex, startIndex + pageSize);
  return {
    endItem: pageItems.length === 0 ? 0 : startIndex + pageItems.length,
    pageIndex: safePageIndex,
    pageItems,
    pageSize,
    startItem: pageItems.length === 0 ? 0 : startIndex + 1,
    totalItems,
    totalPages
  };
}

export function createListView(items, options, getFilterValues, getSortValue) {
  const normalizedState = normalizeListViewState(options);
  const filteredItems = filterListItems(items, options?.filterText ?? "", getFilterValues);
  const sortedItems = sortListItems(filteredItems, normalizedState, getSortValue);
  const page = paginateListItems(sortedItems, normalizedState);
  return {
    filteredItems,
    page,
    sortedItems,
    state: normalizedState
  };
}

function normalizeSortValue(value) {
  if (value === null || value === undefined) {
    return "";
  }
  if (value instanceof Date) {
    return value.getTime();
  }
  if (typeof value === "number") {
    return value;
  }
  return String(value).toLowerCase();
}
