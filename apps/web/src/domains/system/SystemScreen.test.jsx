import { describe, expect, it } from "vitest";
import {
  countRoomPublishedTracks,
  countRoomRobotPublishers,
  createRoomPeerSummaries,
  formatStorageByteCount,
  normalizeDatabaseUsage,
  normalizeObjectStorageUsage,
  normalizeRecorderRuntimeStatus,
  makeRoomStreamingState
} from "./systemViewModel.js";

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

describe("SystemScreen usage summaries", () => {
  it("normalizes database usage for display", () => {
    const usage = normalizeDatabaseUsage({
      databaseName: "robot_center",
      databaseSizeBytes: "10485760",
      trackedTableBytes: "8388608",
      status: "ok",
      tables: [
        { tableName: "events", rowCount: "100036", totalBytes: "7340032" },
        { tableName: "sensor_samples", rowCount: null, totalBytes: "1048576" }
      ]
    });

    expect(usage).toEqual({
      databaseName: "robot_center",
      databaseSizeBytes: 10485760,
      status: "ok",
      tables: [
        { tableName: "events", rowCount: 100036, totalBytes: 7340032 },
        { tableName: "sensor_samples", rowCount: 0, totalBytes: 1048576 }
      ],
      trackedTableBytes: 8388608
    });
  });

  it("normalizes object storage percent from byte counts when needed", () => {
    const usage = normalizeObjectStorageUsage({
      bucket: "robot-center",
      status: "ok",
      totalBytes: 1000,
      usedBytes: 250,
      usedPercent: 0
    });

    expect(usage.usedPercent).toBe(25);
  });

  it("normalizes recorder runtime status for display", () => {
    const status = normalizeRecorderRuntimeStatus({
      blockingReason: "active recording target",
      clearable: false,
      files: "12",
      recordingDirectories: "3",
      status: "ok",
      totalBytes: 2000,
      usedBytes: 500,
      usedPercent: 0
    });

    expect(status).toEqual({
      availableBytes: 0,
      blockingReason: "active recording target",
      clearable: false,
      files: 12,
      recordingDirectories: 3,
      status: "ok",
      totalBytes: 2000,
      usedBytes: 500,
      usedPercent: 25
    });
  });

  it("formats byte counts with stable units", () => {
    expect(formatStorageByteCount(0)).toBe("0 B");
    expect(formatStorageByteCount(1024)).toBe("1.00 KB");
    expect(formatStorageByteCount(1536 * 1024)).toBe("1.50 MB");
  });
});
