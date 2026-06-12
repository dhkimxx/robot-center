import { describe, expect, it } from "vitest";
import {
  calculateLiveDashboardRowCount,
  compactLiveDashboardWidgets,
  createPresetLiveDashboardLayout,
  liveDashboardLayoutStorageKey,
  moveLiveDashboardWidget,
  normalizeLiveDashboardLayout,
  readLiveDashboardLayout,
  resizeLiveDashboardWidget,
  resolveLiveDashboardCollisions,
  writeLiveDashboardLayout
} from "./liveDashboardLayout.js";

function createMemoryStorage(initial = {}) {
  const values = new Map(Object.entries(initial));
  return {
    getItem: (key) => values.get(key) ?? null,
    setItem: (key, value) => values.set(key, value)
  };
}

describe("liveDashboardLayout", () => {
  it("creates the cockpit preset with Grafana-style coordinates", () => {
    const layout = createPresetLiveDashboardLayout("cockpit");

    expect(layout.presetId).toBe("cockpit");
    expect(layout.widgets).toEqual([
      { h: 8, id: "rgb", w: 16, x: 0, y: 0 },
      { h: 8, id: "event", w: 8, x: 16, y: 0 },
      { h: 5, id: "thermal", w: 8, x: 0, y: 8 },
      { h: 5, id: "map", w: 8, x: 8, y: 8 },
      { h: 5, id: "sensor", w: 8, x: 16, y: 8 }
    ]);
  });

  it("normalizes unknown widgets and clamps widget bounds", () => {
    const layout = normalizeLiveDashboardLayout({
      presetId: "missing",
      widgets: [
        { h: 99, id: "rgb", w: 99, x: 99, y: -1 },
        { h: 0, id: "event", w: 0, x: -1, y: 0 },
        { h: 2, id: "unknown", w: 2, x: 0, y: 0 }
      ]
    });

    expect(layout.presetId).toBe("cockpit");
    expect(layout.widgets.find((widget) => widget.id === "event")).toMatchObject({ h: 4, w: 6, x: 0, y: 0 });
    expect(layout.widgets.find((widget) => widget.id === "rgb")).toMatchObject({ h: 16, w: 24, x: 0, y: 4 });
    expect(layout.widgets.some((widget) => widget.id === "unknown")).toBe(false);
  });

  it("moves an active widget and pushes overlapping widgets down", () => {
    const moved = moveLiveDashboardWidget(createPresetLiveDashboardLayout(), "event", { x: 0, y: 0 });

    expect(moved.widgets.find((widget) => widget.id === "event")).toMatchObject({ x: 0, y: 0 });
    expect(moved.widgets.find((widget) => widget.id === "rgb")).toMatchObject({ x: 0, y: 8 });
  });

  it("keeps the active widget at the dropped y position during compact", () => {
    const moved = moveLiveDashboardWidget(createPresetLiveDashboardLayout(), "event", { x: 16, y: 5 });

    expect(moved.widgets.find((widget) => widget.id === "event")).toMatchObject({ x: 16, y: 5 });
  });

  it("resizes an active widget and keeps widgets non-overlapping", () => {
    const resized = resizeLiveDashboardWidget(createPresetLiveDashboardLayout(), "event", { h: 12, w: 10, x: 14, y: 0 });

    expect(resized.widgets.find((widget) => widget.id === "event")).toMatchObject({ h: 12, w: 10, x: 14, y: 0 });
    expect(resized.widgets.find((widget) => widget.id === "sensor").y).toBeGreaterThanOrEqual(12);
  });

  it("keeps a widget anchored when resizing against the right edge", () => {
    const resized = resizeLiveDashboardWidget(createPresetLiveDashboardLayout(), "event", { w: 12 });

    expect(resized.widgets.find((widget) => widget.id === "event")).toMatchObject({ w: 8, x: 16 });
  });

  it("resolves collisions and compacts empty vertical gaps", () => {
    const resolved = resolveLiveDashboardCollisions([
      { h: 3, id: "rgb", w: 12, x: 0, y: 5 },
      { h: 3, id: "event", w: 12, x: 0, y: 5 }
    ]);
    const compacted = compactLiveDashboardWidgets(resolved);

    expect(resolved.map((widget) => [widget.id, widget.y])).toEqual([
      ["event", 5],
      ["rgb", 8]
    ]);
    expect(compacted.map((widget) => [widget.id, widget.y])).toEqual([
      ["event", 0],
      ["rgb", 3]
    ]);
  });

  it("calculates row count with a minimum dashboard height", () => {
    expect(calculateLiveDashboardRowCount(createPresetLiveDashboardLayout())).toBe(13);
    expect(calculateLiveDashboardRowCount({
      widgets: [{ h: 4, id: "event", w: 8, x: 0, y: 20 }]
    })).toBe(24);
  });

  it("persists valid versioned layout and falls back on invalid storage", () => {
    const storage = createMemoryStorage();
    const layout = writeLiveDashboardLayout(resizeLiveDashboardWidget(createPresetLiveDashboardLayout(), "rgb", { w: 20 }), storage);

    expect(JSON.parse(storage.getItem(liveDashboardLayoutStorageKey))).toMatchObject({ version: 3 });
    expect(readLiveDashboardLayout(storage)).toEqual(layout);

    const invalidStorage = createMemoryStorage({
      [liveDashboardLayoutStorageKey]: JSON.stringify({ version: 0, widgets: [] })
    });
    expect(readLiveDashboardLayout(invalidStorage).widgets.map((widget) => widget.id)).toEqual([
      "rgb",
      "event",
      "thermal",
      "map",
      "sensor"
    ]);
  });

  it("uses cockpit when a previous layout version exists in storage", () => {
    const storage = createMemoryStorage({
      [liveDashboardLayoutStorageKey]: JSON.stringify({
        presetId: "classic",
        version: 2,
        widgets: [
          { h: 5, id: "rgb", w: 6 },
          { h: 5, id: "thermal", w: 6 },
          { h: 3, id: "map", w: 6 },
          { h: 3, id: "sensor", w: 6 },
          { h: 3, id: "event", w: 12 }
        ]
      })
    });

    expect(readLiveDashboardLayout(storage).presetId).toBe("cockpit");
  });

  it("keeps the default layout when browser storage fails", () => {
    const brokenStorage = {
      getItem: () => {
        throw new Error("blocked");
      },
      setItem: () => {
        throw new Error("blocked");
      }
    };

    expect(readLiveDashboardLayout(brokenStorage).presetId).toBe("cockpit");
    expect(writeLiveDashboardLayout(createPresetLiveDashboardLayout("classic"), brokenStorage).presetId).toBe("classic");
  });
});
