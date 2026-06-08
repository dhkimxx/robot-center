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
          media: { trackId: "track.video_9" },
          payload: {
            detections: [
              {
                className: "person",
                confidence: 0.9,
                bbox: { format: "normalized_xywh", x: 0.1, y: 0.1, width: 0.2, height: 0.2 }
              }
            ]
          }
        },
        {
          eventType: "detection.object",
          media: { trackId: "track.video_1" },
          payload: {
            detections: [
              {
                className: "person",
                confidence: 0.9,
                bbox: { format: "normalized_xywh", x: 0.9, y: 0.1, width: 0.2, height: 0.2 }
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

  it("expires detection overlays by TTL", () => {
    const overlay = { receivedAt: "2026-06-08T01:00:00.000Z" };

    expect(isDetectionOverlayFresh(overlay, Date.parse("2026-06-08T01:00:00.500Z"))).toBe(true);
    expect(isDetectionOverlayFresh(overlay, Date.parse("2026-06-08T01:00:02.000Z"))).toBe(false);
  });
});
