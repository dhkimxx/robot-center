import { afterEach, describe, expect, it, vi } from "vitest";
import { clearEventData, clearRecorderRuntime, fetchSystemStatus, pruneObjectStorage, pruneRecorderRuntime } from "./systemApi.js";

describe("systemApi", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("fetches lightweight system status with an explicit scope", async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ status: "ok" }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(fetchSystemStatus({ scope: "overview" })).resolves.toEqual({ status: "ok" });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/system/status?scope=overview",
      expect.any(Object)
    );
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

  it("prunes object storage through the system endpoint", async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ objectStorage: { deletedObjectCount: 2 } }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(pruneObjectStorage()).resolves.toEqual({ objectStorage: { deletedObjectCount: 2 } });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/system/object-storage/prune",
      expect.objectContaining({
        body: JSON.stringify({ confirmation: "PRUNE_OBJECT_STORAGE" }),
        method: "POST"
      })
    );
  });

  it("prunes recorder runtime through the system endpoint", async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ recorderRuntime: { filesDeleted: 1 } }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(pruneRecorderRuntime()).resolves.toEqual({ recorderRuntime: { filesDeleted: 1 } });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/system/recorder-runtime/prune",
      expect.objectContaining({
        body: JSON.stringify({ confirmation: "PRUNE_RECORDER_RUNTIME" }),
        method: "POST"
      })
    );
  });
});
