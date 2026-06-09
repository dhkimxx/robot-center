import { afterEach, describe, expect, it, vi } from "vitest";
import { clearEventData, clearRecorderRuntime } from "./systemApi.js";

describe("systemApi", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("clears event data through the system endpoint", async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ eventData: { eventsDeleted: 3 } }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(clearEventData()).resolves.toEqual({ eventData: { eventsDeleted: 3 } });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/system/events/clear",
      expect.objectContaining({
        body: JSON.stringify({ confirmation: "CLEAR_EVENT_DATA" }),
        method: "POST"
      })
    );
  });

  it("clears recorder runtime through the system endpoint", async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ recorderRuntime: { filesDeleted: 2 } }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(clearRecorderRuntime()).resolves.toEqual({ recorderRuntime: { filesDeleted: 2 } });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/system/recorder-runtime/clear",
      expect.objectContaining({
        body: JSON.stringify({ confirmation: "CLEAR_RECORDER_RUNTIME" }),
        method: "POST"
      })
    );
  });
});
