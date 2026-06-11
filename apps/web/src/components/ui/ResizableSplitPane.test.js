import { describe, expect, it } from "vitest";
import {
  clampSplitRatio,
  getSplitRatioBounds,
  parseStoredSplitRatio,
  stringifyStoredSplitRatio
} from "./ResizableSplitPane.jsx";

describe("ResizableSplitPane", () => {
  it("keeps the default 3:2 split when both panes fit", () => {
    expect(clampSplitRatio(0.6, {
      containerWidth: 1600,
      leftMinWidth: 640,
      rightMinWidth: 360
    })).toBe(0.6);
  });

  it("clamps the left pane to its minimum width", () => {
    const ratio = clampSplitRatio(0.25, {
      containerWidth: 1600,
      leftMinWidth: 640,
      rightMinWidth: 360
    });

    expect(ratio).toBeCloseTo(640 / 1590);
  });

  it("clamps the right pane to its minimum width", () => {
    const ratio = clampSplitRatio(0.9, {
      containerWidth: 1600,
      leftMinWidth: 640,
      rightMinWidth: 480
    });

    expect(ratio).toBeCloseTo(1 - (480 / 1590));
  });

  it("falls back to the default ratio for invalid input", () => {
    expect(clampSplitRatio(Number.NaN, {
      containerWidth: 1600,
      fallbackRatio: 0.6,
      leftMinWidth: 640,
      rightMinWidth: 360
    })).toBe(0.6);
  });

  it("uses conservative bounds when the container is not measurable yet", () => {
    expect(getSplitRatioBounds(0, 640, 360)).toEqual({
      maxRatio: 0.8,
      minRatio: 0.2
    });
  });

  it("ignores legacy numeric storage values", () => {
    expect(parseStoredSplitRatio("0.2", 0.6)).toBe(0.6);
  });

  it("restores versioned storage values", () => {
    expect(parseStoredSplitRatio(stringifyStoredSplitRatio(0.65), 0.6)).toBe(0.65);
  });
});
