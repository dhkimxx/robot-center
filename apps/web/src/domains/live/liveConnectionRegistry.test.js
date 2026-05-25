import { describe, expect, it, vi } from "vitest";
import { createLiveConnectionRegistryState } from "./liveConnectionRegistry.js";
import { LiveCloseReason } from "./liveConnectionStates.js";

describe("live connection registry", () => {
  it("removes an existing connection before closing it when a new attempt starts", () => {
    const registry = createLiveConnectionRegistryState();
    const connectionKey = "mission:mission-006";
    let connectionDuringClose = "not-called";
    const previousConnection = {
      close: vi.fn(() => {
        connectionDuringClose = registry.getConnection(connectionKey);
      })
    };
    const previousAttempt = registry.startConnectionAttempt(connectionKey, "mission-006:robot-001");
    registry.registerConnection(connectionKey, previousConnection, previousAttempt);

    const nextAttempt = registry.startConnectionAttempt(connectionKey, "mission-006:robot-002", {
      closeReason: LiveCloseReason.SWITCHING_TARGET
    });

    expect(previousConnection.close).toHaveBeenCalledWith(LiveCloseReason.SWITCHING_TARGET);
    expect(connectionDuringClose).toBeNull();
    expect(registry.getConnection(connectionKey)).toBeNull();
    expect(previousAttempt.isCurrent()).toBe(false);
    expect(nextAttempt.isCurrent()).toBe(true);
  });

  it("does not register a stale client after a newer attempt starts", () => {
    const registry = createLiveConnectionRegistryState();
    const connectionKey = "mission:mission-006";
    const staleAttempt = registry.startConnectionAttempt(connectionKey, "mission-006:robot-001");
    registry.startConnectionAttempt(connectionKey, "mission-006:robot-002");
    const staleClient = { close: vi.fn() };

    const registered = registry.registerConnection(connectionKey, staleClient, staleAttempt);

    expect(registered).toBe(false);
    expect(staleClient.close).toHaveBeenCalledWith(LiveCloseReason.CONNECTION_FAILED);
    expect(registry.getConnection(connectionKey)).toBeNull();
  });
});
