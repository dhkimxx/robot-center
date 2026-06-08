import { afterEach, describe, expect, it, vi } from "vitest";
import {
  fetchMissionRecordingChunks,
  fetchMissionRecordingSummary
} from "./recordingsApi.js";

describe("recordingsApi", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("fetches mission recording summary from the mission-specific endpoint", async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ totalChunks: 0, robots: [] }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(fetchMissionRecordingSummary("mission/001")).resolves.toEqual({ totalChunks: 0, robots: [] });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/operator/missions/mission%2F001/recordings/summary",
      expect.any(Object)
    );
  });

  it("fetches mission recording chunks with robot pagination query", async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ recordings: [], page: { total: 0 } }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    await fetchMissionRecordingChunks("mission-054", {
      limit: 80,
      offset: 160,
      robotCode: "robot-041"
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/operator/missions/mission-054/recordings/chunks?robotCode=robot-041&limit=80&offset=160",
      expect.any(Object)
    );
  });
});
