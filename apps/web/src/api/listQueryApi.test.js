import { describe, expect, it } from "vitest";
import { createListQueryPath } from "./listQueryApi.js";

describe("createListQueryPath", () => {
  it("keeps the path unchanged when no list query is provided", () => {
    expect(createListQueryPath("/api/v1/operator/robots")).toBe("/api/v1/operator/robots");
  });

  it("serializes shared list query parameters", () => {
    expect(createListQueryPath("/api/v1/operator/missions", {
      filter: "robot 1",
      limit: 10,
      offset: 20,
      order: "desc",
      sort: "missionCode"
    })).toBe("/api/v1/operator/missions?limit=10&offset=20&sort=missionCode&order=desc&filter=robot+1");
  });
});
