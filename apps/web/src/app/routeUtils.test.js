import { describe, expect, it } from "vitest";
import { getRouteMissionControlCode, getRouteMissionReplayCode, getRouteSelectedMissionCode } from "./routeUtils.js";

describe("routeUtils", () => {
  it("extracts mission code from mission control route", () => {
    expect(getRouteMissionControlCode("/missions/mission-001/control")).toBe("mission-001");
  });

  it("decodes encoded mission code", () => {
    expect(getRouteMissionControlCode("/missions/mission%2F001/control")).toBe("mission/001");
  });

  it("extracts mission code from mission replay route", () => {
    expect(getRouteMissionReplayCode("/missions/mission-001/replay")).toBe("mission-001");
  });

  it("extracts selected mission code from query", () => {
    expect(getRouteSelectedMissionCode("?selected=mission-001")).toBe("mission-001");
  });

  it("decodes selected mission code from query", () => {
    expect(getRouteSelectedMissionCode("?selected=mission%2F001")).toBe("mission/001");
  });
});
