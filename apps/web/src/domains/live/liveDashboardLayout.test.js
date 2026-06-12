import { describe, expect, it } from "vitest";
import {
  createPresetLiveDashboardLayout,
  liveDashboardLayoutStorageKey,
  moveLiveDashboardWidget,
  normalizeLiveDashboardLayout,
  readLiveDashboardLayout,
  resizeLiveDashboardWidget,
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
  it("creates the cockpit preset with required live widgets", () => {
    const layout = createPresetLiveDashboardLayout("cockpit");

    expect(layout.presetId).toBe("cockpit");
    expect(layout.widgets.map((widget) => widget.id)).toEqual([
      "rgb",
      "event",
      "thermal",
      "map",
      "sensor"
    ]);
  });

  it("normalizes unknown widgets and clamps widget size", () => {
    const layout = normalizeLiveDashboardLayout({
      presetId: "missing",
      widgets: [
        { h: 99, id: "rgb", w: 99 },
        { h: 0, id: "event", w: 0 },
        { h: 2, id: "unknown", w: 2 }
      ]
    });

    expect(layout.presetId).toBe("cockpit");
    expect(layout.widgets.find((widget) => widget.id === "rgb")).toMatchObject({ h: 9, w: 12 });
    expect(layout.widgets.find((widget) => widget.id === "event")).toMatchObject({ h: 3, w: 3 });
    expect(layout.widgets.some((widget) => widget.id === "unknown")).toBe(false);
  });

  it("resizes and reorders widgets without changing other widget ids", () => {
    const resized = resizeLiveDashboardWidget(createPresetLiveDashboardLayout(), "event", { h: 8, w: 5 });
    const moved = moveLiveDashboardWidget(resized, "event", "forward");

    expect(resized.widgets.find((widget) => widget.id === "event")).toMatchObject({ h: 8, w: 5 });
    expect(moved.widgets.map((widget) => widget.id)).toEqual([
      "rgb",
      "thermal",
      "event",
      "map",
      "sensor"
    ]);
  });

  it("persists valid versioned layout and falls back on invalid storage", () => {
    const storage = createMemoryStorage();
    const layout = writeLiveDashboardLayout(resizeLiveDashboardWidget(createPresetLiveDashboardLayout(), "rgb", { w: 10 }), storage);

    expect(JSON.parse(storage.getItem(liveDashboardLayoutStorageKey))).toMatchObject({ version: 1 });
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
