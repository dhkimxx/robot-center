import { describe, expect, it } from "vitest";
import { createMissionRobotTargets } from "./missionHelpers.js";

describe("missionHelpers", () => {
  it("creates mission robot targets with missionCode room id", () => {
    const mission = {
      id: "mission-id-001",
      missionCode: "mission-001",
      robotCodes: ["robot-001"]
    };
    const robots = [
      {
        displayName: "Robot 1",
        robotCode: "robot-001"
      }
    ];
    const streamingStatuses = [
      {
        missionId: "mission-id-001",
        robotCode: "robot-001",
        roomId: "mission-001__robot-001"
      }
    ];

    const targets = createMissionRobotTargets(mission, robots, streamingStatuses);

    expect(targets).toHaveLength(1);
    expect(targets[0]).toMatchObject({
      key: "mission-001:robot-001",
      missionRoomId: "mission-001",
      roomId: "mission-001",
      robotCode: "robot-001"
    });
  });
});
