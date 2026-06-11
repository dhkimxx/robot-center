import { describe, expect, it } from "vitest";
import {
  countRoomPublishedTracks,
  countRoomRobotPublishers,
  createDatabaseUsageCategories,
  createDatabaseTopTables,
  createRoomPeerSummaries,
  createSystemClearActions,
  formatStorageByteCount,
  makeObjectStorageDisabledReason,
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
        { tableName: "sensor_samples", rowCount: null, totalBytes: "1048576" },
        { tableName: "spatial_ref_sys", rowCount: "8500", totalBytes: "7340032" }
      ]
    });

    expect(usage).toEqual({
      categories: [
        { id: "events", label: "이벤트 데이터", rowCount: 100036, sortOrder: 20, tableCount: 1, totalBytes: 7340032 },
        { id: "sensors", label: "센서 데이터", rowCount: 0, sortOrder: 10, tableCount: 1, totalBytes: 1048576 },
        { id: "internal", label: "시스템 내부 데이터", rowCount: 8500, sortOrder: 90, tableCount: 1, totalBytes: 7340032 }
      ],
      databaseName: "robot_center",
      databaseSizeBytes: 10485760,
      status: "ok",
      tables: [
        { tableName: "events", rowCount: 100036, totalBytes: 7340032 },
        { tableName: "sensor_samples", rowCount: 0, totalBytes: 1048576 },
        { tableName: "spatial_ref_sys", rowCount: 8500, totalBytes: 7340032 }
      ],
      topTables: [
        { label: "이벤트 로그", tableName: "events", rowCount: 100036, totalBytes: 7340032 },
        { label: "센서 샘플", tableName: "sensor_samples", rowCount: 0, totalBytes: 1048576 }
      ],
      trackedTableBytes: 8388608
    });
  });

  it("groups raw database tables into user-facing categories", () => {
    const categories = createDatabaseUsageCategories([
      { tableName: "sensor_samples", rowCount: 20, totalBytes: 500 },
      { tableName: "sensor_descriptors", rowCount: 2, totalBytes: 50 },
      { tableName: "recording_chunks", rowCount: 10, totalBytes: 300 },
      { tableName: "spatial_ref_sys", rowCount: 8500, totalBytes: 700 },
      { tableName: "custom_table", rowCount: 1, totalBytes: 100 }
    ]);

    expect(categories.map((category) => category.label)).toEqual([
      "센서 데이터",
      "녹화 데이터",
      "기타 관제 데이터",
      "시스템 내부 데이터"
    ]);
    expect(categories.find((category) => category.id === "sensors")).toMatchObject({
      rowCount: 22,
      tableCount: 2,
      totalBytes: 550
    });
    expect(categories.some((category) => category.label === "spatial_ref_sys")).toBe(false);
  });

  it("lists top database tables without internal PostGIS tables", () => {
    const topTables = createDatabaseTopTables([
      { tableName: "spatial_ref_sys", rowCount: 8500, totalBytes: 7000 },
      { tableName: "sensor_samples", rowCount: 20, totalBytes: 500 },
      { tableName: "events", rowCount: 100, totalBytes: 900 },
      { tableName: "robots", rowCount: 2, totalBytes: 50 }
    ]);

    expect(topTables).toEqual([
      { label: "이벤트 로그", tableName: "events", rowCount: 100, totalBytes: 900 },
      { label: "센서 샘플", tableName: "sensor_samples", rowCount: 20, totalBytes: 500 },
      { label: "로봇", tableName: "robots", rowCount: 2, totalBytes: 50 }
    ]);
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

  it("blocks object storage clearing while recorder runtime is active", () => {
    expect(makeObjectStorageDisabledReason({
      isProduction: false,
      recorderRuntimeStatus: {
        blockingReason: "active recording target",
        clearable: false,
        status: "ok"
      }
    })).toBe("진행 중인 녹화 대상이 있어 정리를 실행할 수 없습니다.");
  });

  it("creates clear actions with target metrics and blocking reasons", () => {
    const databaseUsage = normalizeDatabaseUsage({
      status: "ok",
      tables: [
        { tableName: "sensor_samples", rowCount: 1200, totalBytes: 3 * 1024 * 1024 },
        { tableName: "sensor_descriptors", rowCount: 4, totalBytes: 4096 },
        { tableName: "events", rowCount: 77, totalBytes: 512 * 1024 },
        { tableName: "storage_objects", rowCount: 9, totalBytes: 4096 }
      ]
    });

    const actions = createSystemClearActions({
      canClearEventData: true,
      canClearObjectStorage: true,
      canClearRecorderRuntime: true,
      canClearSensorData: true,
      canPruneObjectStorage: true,
      canPruneRecorderRuntime: true,
      databaseUsage,
      objectStorageUsage: {
        bucketUsedBytes: 8 * 1024 * 1024,
        objectCount: 9,
        usedBytes: 12 * 1024 * 1024
      },
      recorderRuntimeStatus: {
        clearable: false,
        blockingReason: "active recording target",
        files: 20,
        status: "ok",
        usedBytes: 10 * 1024 * 1024
      }
    });

    expect(actions.find((action) => action.id === "objectStorage")).toMatchObject({
      disabled: true,
      disabledReason: "진행 중인 녹화 대상이 있어 정리를 실행할 수 없습니다.",
      targetMetrics: [
        { label: "삭제 대상", value: "9개 파일" },
        { label: "파일 용량", value: "8.00 MB" },
        { label: "파일 메타데이터", value: "9건" }
      ]
    });
    expect(actions.find((action) => action.id === "objectStoragePrune")).toMatchObject({
      buttonLabel: "운영 중 정리",
      disabled: false,
      targetMetrics: [
        { label: "최대 후보", value: "9개 파일" },
        { label: "파일 용량", value: "8.00 MB" },
        { label: "파일 메타데이터", value: "9건" }
      ]
    });
    expect(actions.find((action) => action.id === "recorderRuntime")).toMatchObject({
      disabled: true,
      disabledReason: "진행 중인 녹화 대상이 있어 정리를 실행할 수 없습니다."
    });
    expect(actions.find((action) => action.id === "recorderRuntimePrune")).toMatchObject({
      buttonLabel: "운영 중 정리",
      disabled: false
    });
    expect(actions.find((action) => action.id === "sensorData").targetMetrics).toEqual([
      { label: "삭제 대상", value: "1,204건" },
      { label: "DB 사용량", value: "3.00 MB" },
      { label: "관련 테이블", value: "2개" }
    ]);
    expect(actions.find((action) => action.id === "eventData").targetMetrics).toEqual([
      { label: "삭제 대상", value: "77건" },
      { label: "DB 사용량", value: "512 KB" },
      { label: "관련 테이블", value: "1개" }
    ]);
  });

  it("blocks every clear action until system status is loaded", () => {
    const actions = createSystemClearActions({
      canClearEventData: true,
      canClearObjectStorage: true,
      canClearRecorderRuntime: true,
      canClearSensorData: true,
      canPruneObjectStorage: true,
      canPruneRecorderRuntime: true,
      statusReady: false
    });

    expect(actions.every((action) => action.disabled)).toBe(true);
    expect(new Set(actions.map((action) => action.disabledReason))).toEqual(new Set([
      "시스템 상태를 확인한 뒤 삭제를 실행할 수 있습니다."
    ]));
  });
});
