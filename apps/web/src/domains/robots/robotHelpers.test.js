import { describe, expect, it } from "vitest";
import { shouldRefreshRobotEditForm } from "./robotHelpers.js";

describe("robotHelpers", () => {
  it("does not refresh edit form while edit modal is open", () => {
    expect(shouldRefreshRobotEditForm({
      nextRobotCode: "robot-002",
      previousRobotCode: "robot-001",
      robotModal: "edit"
    })).toBe(false);
  });

  it("does not refresh edit form when selected robot code is unchanged", () => {
    expect(shouldRefreshRobotEditForm({
      nextRobotCode: "robot-001",
      previousRobotCode: "robot-001",
      robotModal: null
    })).toBe(false);
  });

  it("refreshes edit form when selected robot changes outside edit mode", () => {
    expect(shouldRefreshRobotEditForm({
      nextRobotCode: "robot-002",
      previousRobotCode: "robot-001",
      robotModal: null
    })).toBe(true);
  });
});
