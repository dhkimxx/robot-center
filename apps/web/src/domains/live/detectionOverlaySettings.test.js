import { describe, expect, it } from "vitest";
import {
  detectionOverlaySettingsStorageKey,
  normalizeDetectionOverlaySettings,
  readDetectionOverlaySettings,
  writeDetectionOverlaySettings
} from "./detectionOverlaySettings.js";

describe("detection overlay settings", () => {
  it("normalizes ttl and count settings", () => {
    expect(normalizeDetectionOverlaySettings({
      maxDetections: "999",
      ttlSeconds: "0"
    })).toEqual({
      maxDetections: 50,
      ttlSeconds: 1
    });

    expect(normalizeDetectionOverlaySettings({
      maxDetections: "abc",
      ttlSeconds: "abc"
    })).toEqual({
      maxDetections: 10,
      ttlSeconds: 3
    });
  });

  it("persists normalized settings when storage is available", () => {
    const storage = new MapStorage();

    const saved = writeDetectionOverlaySettings({
      maxDetections: 12,
      ttlSeconds: 7
    }, storage);

    expect(saved).toEqual({ maxDetections: 12, ttlSeconds: 7 });
    expect(JSON.parse(storage.getItem(detectionOverlaySettingsStorageKey))).toEqual(saved);
    expect(readDetectionOverlaySettings(storage)).toEqual(saved);
  });
});

class MapStorage {
  constructor() {
    this.values = new Map();
  }

  getItem(key) {
    return this.values.get(key) ?? null;
  }

  setItem(key, value) {
    this.values.set(key, value);
  }
}
