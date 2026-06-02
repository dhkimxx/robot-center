import { describe, expect, it } from "vitest";
import {
  countRoomPublishedTracks,
  countRoomRobotPublishers,
  createRoomPeerSummaries,
  makeRoomStreamingState
} from "./SystemScreen.jsx";

describe("SystemScreen room summaries", () => {
  it("deduplicates robot peers by robot code and prefers publishers", () => {
    const room = {
      peers: [
        { peerId: "old-robot-1", role: "robot", robotCode: "robot-001" },
        { peerId: "old-robot-2", role: "robot", robotCode: "robot-001" },
        { peerId: "operator-1", role: "operator", selectedRobotCode: "robot-001" },
        { peerId: "recorder-1", role: "recorder" }
      ],
      publishers: [
        { publisherPeerId: "publisher-1", robotCode: "robot-001", state: "publishing", trackCount: 3 }
      ]
    };

    const peers = createRoomPeerSummaries(room);

    expect(countRoomRobotPublishers(room)).toBe(1);
    expect(peers.map((peer) => `${peer.role}:${peer.robotCode ?? peer.selectedRobotCode ?? peer.peerId}`)).toEqual([
      "robot:robot-001",
      "operator:robot-001",
      "recorder:recorder-1"
    ]);
  });

  it("marks a room as publishing only when publisher media tracks exist", () => {
    expect(makeRoomStreamingState({
      peers: [{ peerId: "robot-1", role: "robot", robotCode: "robot-001" }],
      publishers: []
    })).toBe("연결됨");

    expect(makeRoomStreamingState({
      publishers: [{ robotCode: "robot-001", state: "connected", trackCount: 0 }]
    })).toBe("연결됨");

    expect(makeRoomStreamingState({
      publishers: [{ robotCode: "robot-001", state: "publishing", trackCount: 3 }]
    })).toBe("송출 중");
  });

  it("counts publisher canonical tracks before raw room track list", () => {
    const room = {
      publishedTracks: ["robot-001:track.video_1", "robot-001:unmapped.video"],
      publishers: [{ robotCode: "robot-001", state: "publishing", trackCount: 1 }]
    };

    expect(countRoomPublishedTracks(room)).toBe(1);
  });
});
