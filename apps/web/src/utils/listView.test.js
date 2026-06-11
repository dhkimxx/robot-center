import { describe, expect, it } from "vitest";
import {
  createListView,
  createNextSortState,
  filterListItems,
  paginateListItems,
  sortListItems
} from "./listView.js";

describe("listView", () => {
  const rows = [
    { code: "robot-001", label: "Alpha", status: "offline" },
    { code: "robot-002", label: "Beta", status: "online" },
    { code: "robot-003", label: "Gamma", status: "online" }
  ];

  it("filters rows from shared searchable values", () => {
    const filteredRows = filterListItems(rows, "beta", (row) => [row.code, row.label, row.status]);

    expect(filteredRows.map((row) => row.code)).toEqual(["robot-002"]);
  });

  it("sorts rows by the selected key and direction", () => {
    const sortedRows = sortListItems(rows, { sortDirection: "desc", sortKey: "label" }, (row, sortKey) => row[sortKey]);

    expect(sortedRows.map((row) => row.label)).toEqual(["Gamma", "Beta", "Alpha"]);
  });

  it("paginates rows and clamps out-of-range pages", () => {
    const page = paginateListItems(rows, { pageIndex: 5, pageSize: 2 });

    expect(page.pageIndex).toBe(1);
    expect(page.totalPages).toBe(2);
    expect(page.startItem).toBe(3);
    expect(page.endItem).toBe(3);
    expect(page.pageItems.map((row) => row.code)).toEqual(["robot-003"]);
  });

  it("creates a full list view from filter, sort and page state", () => {
    const listView = createListView(
      rows,
      { filterText: "online", pageIndex: 0, pageSize: 1, sortDirection: "desc", sortKey: "code" },
      (row) => [row.code, row.label, row.status],
      (row, sortKey) => row[sortKey]
    );

    expect(listView.filteredItems).toHaveLength(2);
    expect(listView.page.totalItems).toBe(2);
    expect(listView.page.pageItems.map((row) => row.code)).toEqual(["robot-003"]);
  });

  it("toggles sort direction for the same key", () => {
    expect(createNextSortState({
      currentDirection: "asc",
      currentKey: "code",
      nextKey: "code"
    })).toEqual({
      sortDirection: "desc",
      sortKey: "code"
    });
    expect(createNextSortState({
      currentDirection: "desc",
      currentKey: "code",
      nextKey: "label"
    })).toEqual({
      sortDirection: "asc",
      sortKey: "label"
    });
  });
});
