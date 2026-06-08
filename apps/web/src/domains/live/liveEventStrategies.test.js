import { describe, expect, it } from "vitest";
import {
  createEventLiveProjection,
  getDetectionColor,
  isDetectionOverlayFresh
} from "./liveEventStrategies.js";

describe("createEventLiveProjection", () => {
  it("drops detection objects with invalid track id or bbox", () => {
    const projection = createEventLiveProjection({
      receivedAt: "2026-06-08T01:00:00.000Z",
      events: [
        {
          eventType: "detection.object",
          values: {
            trackId: "track.video_9",
            detections: [
              {
                className: "person",
                confidence: 0.9,
                bbox: { x: 0.1, y: 0.1, width: 0.2, height: 0.2 }
              }
            ]
          }
        },
        {
          eventType: "detection.object",
          values: {
            trackId: "track.video_1",
            detections: [
              {
                className: "person",
                confidence: 0.9,
                bbox: { x: 0.9, y: 0.1, width: 0.2, height: 0.2 }
              }
            ]
          }
        }
      ]
    });

    expect(projection.detectionOverlays).toEqual([]);
    expect(projection.liveEvents).toEqual([]);
  });

  it("keeps class colors deterministic without payload color", () => {
    expect(getDetectionColor("person")).toBe(getDetectionColor("person"));
  });

  it("keeps empty detection snapshots so the matching overlay can be cleared", () => {
    const projection = createEventLiveProjection({
      receivedAt: "2026-06-08T01:00:00.000Z",
      events: [
        {
          eventType: "detection.object",
          values: {
            trackId: "track.video_1",
            detections: []
          }
        }
      ]
    });

    expect(projection.detectionOverlays).toEqual([
      expect.objectContaining({
        detections: [],
        trackId: "track.video_1",
        trackSlot: "rgb"
      })
    ]);
  });

  it("expires detection overlays by TTL", () => {
    const overlay = { timestamp: "2026-06-08T01:00:00.000Z" };

    expect(isDetectionOverlayFresh(overlay, Date.parse("2026-06-08T01:00:02.500Z"))).toBe(true);
    expect(isDetectionOverlayFresh(overlay, Date.parse("2026-06-08T01:00:04.000Z"))).toBe(false);
  });
});
