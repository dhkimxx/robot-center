import { afterEach, describe, expect, it, vi } from "vitest";
import { requestJson } from "./controlCenterApi.js";

describe("requestJson", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
    vi.useRealTimers();
  });

  it("returns parsed JSON payloads", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(JSON.stringify({ ok: true }), { status: 200 })));

    await expect(requestJson("/api/example")).resolves.toEqual({ ok: true });
  });

  it("keeps HTTP status when an error response is not JSON", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response("<html>bad gateway</html>", { status: 502 })));

    await expect(requestJson("/api/example")).rejects.toThrow("request failed: 502");
  });

  it("reports invalid JSON for successful malformed responses", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response("not-json", { status: 200 })));

    await expect(requestJson("/api/example")).rejects.toThrow("invalid JSON response: 200");
  });

  it("aborts slow requests with a timeout error", async () => {
    vi.useFakeTimers();
    vi.stubGlobal("fetch", vi.fn((_path, options) => new Promise((_resolve, reject) => {
      options.signal.addEventListener("abort", () => {
        const error = new Error("aborted");
        error.name = "AbortError";
        reject(error);
      });
    })));

    const request = requestJson("/api/slow", { timeoutMs: 50 });
    const expectation = expect(request).rejects.toThrow("request timed out after 50ms");
    await vi.advanceTimersByTimeAsync(50);

    await expectation;
  });
});
