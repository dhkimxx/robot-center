import { describe, expect, it } from "vitest";
import { getRouteMissionControlCode } from "./routeUtils.js";

describe("routeUtils", () => {
  it("extracts mission code from mission control route", () => {
    expect(getRouteMissionControlCode("/missions/mission-001/control")).toBe("mission-001");
  });

  it("decodes encoded mission code", () => {
    expect(getRouteMissionControlCode("/missions/mission%2F001/control")).toBe("mission/001");
  });
});
