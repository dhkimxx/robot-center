package recording

import (
	"context"
	"os"
	"robot-center/apps/server/internal/config"
	"robot-center/apps/server/internal/domain"
	"testing"
	"time"
)

func TestWorkerTickCachesTargetsWithoutOpeningChunk(t *testing.T) {
	target := domain.Mission{
		MissionCode: "mission-001",
		RobotCode:   "robot-001",
	}
	appServerClient := &fakeAppServerClient{
		targets: []domain.Mission{target},
	}
	objectStorage := &fakeObjectStorage{manifestSize: 42}
	mediaUploader := &fakeMediaUploader{}
	worker := newWorkerWithCollaborators(
		config.RecorderWorkerConfig{RecordingChunkDuration: 10 * time.Minute},
		appServerClient,
		objectStorage,
	)
	worker.mediaUploader = mediaUploader

	worker.tick(context.Background())

	if len(appServerClient.tickTargets) != 0 {
		t.Fatalf("tick opened recording chunks before media arrived: %#v", appServerClient.tickTargets)
	}
	if objectStorage.manifestUploads != 0 {
		t.Fatalf("manifest uploads = %d, want 0", objectStorage.manifestUploads)
	}
	if len(mediaUploader.finalizedChunks) != 0 {
		t.Fatalf("finalized chunks = %d, want 0", len(mediaUploader.finalizedChunks))
	}
	if appServerClient.markedChunkID != "" {
		t.Fatalf("marked chunk id = %q, want empty", appServerClient.markedChunkID)
	}
	mediaKey := recorderMediaKey(target.MissionCode, target.RobotCode)
	if _, ok := worker.activeChunks[mediaKey]; ok {
		t.Fatal("active chunk was created before media arrived")
	}
	if cachedTarget, ok := worker.activeTargets[mediaKey]; !ok || cachedTarget.MissionCode != target.MissionCode {
		t.Fatalf("active target was not cached: %#v", worker.activeTargets)
	}
}

func TestWorkerOpensChunkWhenMediaArrives(t *testing.T) {
	target := domain.Mission{
		MissionCode: "mission-001",
		RobotCode:   "robot-001",
	}
	observedAt := time.Date(2026, 5, 26, 1, 0, 0, 0, time.UTC)
	chunk1 := domain.RecordingChunk{
		ID:                "chunk-001",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-001_manifest.json",
		MediaObjectKeys:   map[string]string{},
		EndedAt:           observedAt.Add(10 * time.Minute),
	}
	appServerClient := &fakeAppServerClient{
		targets:    []domain.Mission{target},
		tickResult: domain.RecordingTickResult{Chunk: chunk1},
	}
	objectStorage := &fakeObjectStorage{manifestSize: 42}
	worker := newWorkerWithCollaborators(
		config.RecorderWorkerConfig{RecordingChunkDuration: 10 * time.Minute},
		appServerClient,
		objectStorage,
	)

	worker.tick(context.Background())
	chunk, ok, err := worker.ensureActiveRecordingChunk(context.Background(), recorderMediaKey(target.MissionCode, target.RobotCode), observedAt)
	if err != nil {
		t.Fatalf("ensureActiveRecordingChunk returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected media arrival to open a recording chunk")
	}
	if chunk.ID != chunk1.ID {
		t.Fatalf("active chunk id = %q, want %q", chunk.ID, chunk1.ID)
	}
	if appServerClient.tickTarget.MissionCode != target.MissionCode || appServerClient.tickTarget.RobotCode != target.RobotCode {
		t.Fatalf("tick target = %#v, want %#v", appServerClient.tickTarget, target)
	}
	if appServerClient.tickDuration != 10*time.Minute {
		t.Fatalf("tick duration = %s, want 10m", appServerClient.tickDuration)
	}
	if !appServerClient.tickAt.Equal(observedAt) {
		t.Fatalf("tickAt = %s, want %s", appServerClient.tickAt, observedAt)
	}
}

func TestWorkerMarksRecorderRobotTrackActivityOnMediaPacket(t *testing.T) {
	worker := NewWorker(config.RecorderWorkerConfig{})
	observedAt := time.Date(2026, 5, 26, 9, 10, 0, 0, time.UTC)
	mediaKey := recorderMediaKey("mission-001", "robot-001")

	worker.markRecorderRobotTrackActivity(mediaKey, "track.video_1", observedAt)

	status := worker.SubscriberStatus()
	if len(status.Rooms) != 1 {
		t.Fatalf("room statuses = %d, want 1", len(status.Rooms))
	}
	room := status.Rooms[0]
	if room.RoomID != "mission-001" {
		t.Fatalf("room id = %q, want mission-001", room.RoomID)
	}
	if room.TrackCount != 1 {
		t.Fatalf("room track count = %d, want 1", room.TrackCount)
	}
	if len(room.Robots) != 1 {
		t.Fatalf("robot statuses = %d, want 1", len(room.Robots))
	}
	robot := room.Robots[0]
	if robot.RobotCode != "robot-001" {
		t.Fatalf("robot code = %q, want robot-001", robot.RobotCode)
	}
	if robot.TrackCount != 1 {
		t.Fatalf("robot track count = %d, want 1", robot.TrackCount)
	}
	if !robot.LastTrackAt.Equal(observedAt) {
		t.Fatalf("lastTrackAt = %s, want %s", robot.LastTrackAt, observedAt)
	}
}

func TestWorkerFinalizesPreviousChunkOnRollover(t *testing.T) {
	target := domain.Mission{
		MissionCode: "mission-001",
		RobotCode:   "robot-001",
	}
	observedAt := time.Date(2026, 5, 26, 1, 0, 0, 0, time.UTC)
	chunk1 := domain.RecordingChunk{
		ID:                "chunk-001",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-001_manifest.json",
		MediaObjectKeys:   map[string]string{},
		EndedAt:           observedAt.Add(10 * time.Minute),
	}
	chunk2 := domain.RecordingChunk{
		ID:                "chunk-002",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-002_manifest.json",
		MediaObjectKeys:   map[string]string{},
		EndedAt:           observedAt.Add(20 * time.Minute),
	}
	appServerClient := &fakeAppServerClient{
		targets: []domain.Mission{target},
		tickResults: []domain.RecordingTickResult{
			{Chunk: chunk1},
			{Chunk: chunk2},
		},
	}
	objectStorage := &fakeObjectStorage{manifestSize: 42}
	mediaUploader := &fakeMediaUploader{}
	worker := newWorkerWithCollaborators(
		config.RecorderWorkerConfig{RecordingChunkDuration: 10 * time.Minute},
		appServerClient,
		objectStorage,
	)
	worker.mediaUploader = mediaUploader

	worker.tick(context.Background())
	if _, _, err := worker.ensureActiveRecordingChunk(context.Background(), recorderMediaKey(target.MissionCode, target.RobotCode), observedAt); err != nil {
		t.Fatalf("first chunk open failed: %v", err)
	}
	if _, _, err := worker.ensureActiveRecordingChunk(context.Background(), recorderMediaKey(target.MissionCode, target.RobotCode), observedAt.Add(11*time.Minute)); err != nil {
		t.Fatalf("second chunk open failed: %v", err)
	}
	worker.processPendingRecordingChunkFinalizations(context.Background())

	if len(mediaUploader.finalizedChunks) != 1 {
		t.Fatalf("finalized chunks = %d, want 1", len(mediaUploader.finalizedChunks))
	}
	if mediaUploader.finalizedChunks[0].ID != chunk1.ID {
		t.Fatalf("finalized chunk = %q, want %q", mediaUploader.finalizedChunks[0].ID, chunk1.ID)
	}
	if objectStorage.manifestKey != chunk1.ManifestObjectKey {
		t.Fatalf("manifest key = %q, want %q", objectStorage.manifestKey, chunk1.ManifestObjectKey)
	}
	if objectStorage.manifestBody["status"] != "uploaded" {
		t.Fatalf("manifest status = %#v, want uploaded", objectStorage.manifestBody["status"])
	}
	if appServerClient.markedChunkID != chunk1.ID {
		t.Fatalf("marked chunk id = %q, want %q", appServerClient.markedChunkID, chunk1.ID)
	}
	activeChunk := worker.activeChunks[recorderMediaKey(target.MissionCode, target.RobotCode)]
	if activeChunk.ID != chunk2.ID {
		t.Fatalf("active chunk id = %q, want %q", activeChunk.ID, chunk2.ID)
	}
}

func TestWorkerDefersRolloverUntilVideoKeyframe(t *testing.T) {
	target := domain.Mission{
		MissionCode: "mission-001",
		RobotCode:   "robot-001",
	}
	observedAt := time.Date(2026, 5, 26, 1, 0, 0, 0, time.UTC)
	chunk1 := domain.RecordingChunk{
		ID:                "chunk-001",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-001_manifest.json",
		MediaObjectKeys:   map[string]string{},
		EndedAt:           observedAt.Add(10 * time.Minute),
	}
	chunk2 := domain.RecordingChunk{
		ID:                "chunk-002",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-002_manifest.json",
		MediaObjectKeys:   map[string]string{},
		EndedAt:           observedAt.Add(20 * time.Minute),
	}
	appServerClient := &fakeAppServerClient{
		targets: []domain.Mission{target},
		tickResults: []domain.RecordingTickResult{
			{Chunk: chunk1},
			{Chunk: chunk2},
		},
	}
	worker := newWorkerWithCollaborators(
		config.RecorderWorkerConfig{RecordingChunkDuration: 10 * time.Minute},
		appServerClient,
		&fakeObjectStorage{manifestSize: 42},
	)
	mediaKey := recorderMediaKey(target.MissionCode, target.RobotCode)

	worker.tick(context.Background())
	if _, _, err := worker.ensureActiveRecordingChunkForWrite(context.Background(), mediaKey, observedAt, true); err != nil {
		t.Fatalf("first chunk open failed: %v", err)
	}
	expiredAt := observedAt.Add(11 * time.Minute)
	audioChunk, ok, err := worker.ensureActiveRecordingChunkForWrite(context.Background(), mediaKey, expiredAt, false)
	if err != nil {
		t.Fatalf("audio write chunk lookup failed: %v", err)
	}
	if !ok || audioChunk.ID != chunk1.ID {
		t.Fatalf("audio write chunk = %#v/%t, want chunk-001", audioChunk, ok)
	}
	nonKeyframeChunk, ok, err := worker.ensureActiveRecordingChunkForWrite(context.Background(), mediaKey, expiredAt, false)
	if err != nil {
		t.Fatalf("non-keyframe video chunk lookup failed: %v", err)
	}
	if !ok || nonKeyframeChunk.ID != chunk1.ID {
		t.Fatalf("non-keyframe video chunk = %#v/%t, want chunk-001", nonKeyframeChunk, ok)
	}

	keyframeChunk, ok, err := worker.ensureActiveRecordingChunkForWrite(context.Background(), mediaKey, expiredAt, true)
	if err != nil {
		t.Fatalf("keyframe rollover failed: %v", err)
	}
	if !ok || keyframeChunk.ID != chunk2.ID {
		t.Fatalf("keyframe chunk = %#v/%t, want chunk-002", keyframeChunk, ok)
	}
}

func TestH264ChunkKeyframeWaitsArePerChunkAndTrack(t *testing.T) {
	worker := NewWorker(config.RecorderWorkerConfig{})
	worker.mediaMu.Lock()
	defer worker.mediaMu.Unlock()
	worker.markH264ChunkTracksWaitingForKeyframeLocked("mission-001:robot-001", "chunk-002")
	if !worker.h264ChunkKeyframeWaits[h264ChunkKeyframeWaitKey("mission-001:robot-001", "chunk-002", "rgb")] {
		t.Fatal("expected rgb to wait for keyframe")
	}
	if !worker.h264ChunkKeyframeWaits[h264ChunkKeyframeWaitKey("mission-001:robot-001", "chunk-002", "thermal")] {
		t.Fatal("expected thermal to wait for keyframe")
	}
	delete(worker.h264ChunkKeyframeWaits, h264ChunkKeyframeWaitKey("mission-001:robot-001", "chunk-002", "rgb"))
	if worker.h264ChunkKeyframeWaits[h264ChunkKeyframeWaitKey("mission-001:robot-001", "chunk-002", "rgb")] {
		t.Fatal("expected rgb keyframe wait to clear independently")
	}
	if !worker.h264ChunkKeyframeWaits[h264ChunkKeyframeWaitKey("mission-001:robot-001", "chunk-002", "thermal")] {
		t.Fatal("expected thermal keyframe wait to remain")
	}
}

func TestAppendH264PayloadStartsChunkAtIDR(t *testing.T) {
	previousWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory failed: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("change working directory failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousWorkingDirectory)
	})

	target := domain.Mission{
		MissionCode: "mission-001",
		RobotCode:   "robot-001",
	}
	chunk := domain.RecordingChunk{
		ID:                "chunk-001",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-001_manifest.json",
		MediaObjectKeys:   map[string]string{},
		EndedAt:           time.Now().UTC().Add(10 * time.Minute),
	}
	worker := newWorkerWithCollaborators(
		config.RecorderWorkerConfig{RecordingChunkDuration: 10 * time.Minute},
		&fakeAppServerClient{
			targets:    []domain.Mission{target},
			tickResult: domain.RecordingTickResult{Chunk: chunk},
		},
		&fakeObjectStorage{manifestSize: 42},
	)
	mediaKey := recorderMediaKey(target.MissionCode, target.RobotCode)
	worker.tick(context.Background())

	nonIDRPayload := []byte{
		0x00, 0x00, 0x00, 0x01, 0x67, 0x11,
		0x00, 0x00, 0x00, 0x01, 0x68, 0x22,
		0x00, 0x00, 0x00, 0x01, 0x41, 0x33,
	}
	if err := worker.appendH264AccessUnit(context.Background(), mediaKey, "track.video_1", nonIDRPayload, 0, time.Now().UTC(), nil, nil); err != nil {
		t.Fatalf("append non-IDR payload failed: %v", err)
	}
	if _, err := os.Stat(h264TrackPath(chunk.ID, "rgb")); !os.IsNotExist(err) {
		t.Fatalf("expected non-IDR payload to be dropped before file creation, stat err=%v", err)
	}

	idrPayload := []byte{
		0x00, 0x00, 0x00, 0x01, 0x67, 0x44,
		0x00, 0x00, 0x00, 0x01, 0x68, 0x55,
		0x00, 0x00, 0x00, 0x01, 0x65, 0x66,
	}
	if err := worker.appendH264AccessUnit(context.Background(), mediaKey, "track.video_1", idrPayload, 3000, time.Now().UTC(), nil, nil); err != nil {
		t.Fatalf("append IDR payload failed: %v", err)
	}
	writtenPayload, err := os.ReadFile(h264TrackPath(chunk.ID, "rgb"))
	if err != nil {
		t.Fatalf("read H264 track failed: %v", err)
	}
	nalus := splitAnnexBNALUs(writtenPayload)
	if len(nalus) == 0 {
		t.Fatal("expected IDR payload to be written")
	}
	firstType, ok := h264NALUType(nalus[0])
	if !ok || firstType != 7 {
		t.Fatalf("first written NALU type = %d/%t, want SPS", firstType, ok)
	}
}

func TestWorkerTickFinalizesActiveChunkWhenTargetDisappears(t *testing.T) {
	target := domain.Mission{
		MissionCode: "mission-001",
		RobotCode:   "robot-001",
	}
	observedAt := time.Date(2026, 5, 26, 1, 0, 0, 0, time.UTC)
	chunk := domain.RecordingChunk{
		ID:                "chunk-001",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-001_manifest.json",
		MediaObjectKeys:   map[string]string{},
		EndedAt:           observedAt.Add(10 * time.Minute),
	}
	appServerClient := &fakeAppServerClient{
		targets:    []domain.Mission{target},
		tickResult: domain.RecordingTickResult{Chunk: chunk},
	}
	objectStorage := &fakeObjectStorage{manifestSize: 42}
	mediaUploader := &fakeMediaUploader{}
	worker := newWorkerWithCollaborators(
		config.RecorderWorkerConfig{RecordingChunkDuration: 10 * time.Minute},
		appServerClient,
		objectStorage,
	)
	worker.mediaUploader = mediaUploader

	worker.tick(context.Background())
	if _, _, err := worker.ensureActiveRecordingChunk(context.Background(), recorderMediaKey(target.MissionCode, target.RobotCode), observedAt); err != nil {
		t.Fatalf("chunk open failed: %v", err)
	}
	appServerClient.targets = nil
	worker.tick(context.Background())

	if len(mediaUploader.finalizedChunks) != 1 {
		t.Fatalf("finalized chunks = %d, want 1", len(mediaUploader.finalizedChunks))
	}
	if mediaUploader.finalizedChunks[0].ID != chunk.ID {
		t.Fatalf("finalized chunk = %q, want %q", mediaUploader.finalizedChunks[0].ID, chunk.ID)
	}
	if appServerClient.markedChunkID != chunk.ID {
		t.Fatalf("marked chunk id = %q, want %q", appServerClient.markedChunkID, chunk.ID)
	}
	if _, ok := worker.activeChunks[recorderMediaKey(target.MissionCode, target.RobotCode)]; ok {
		t.Fatal("active chunk remained after target disappeared")
	}
}

func TestWorkerFinalizesIdleChunkWhileTargetStaysActive(t *testing.T) {
	target := domain.Mission{
		MissionCode: "mission-001",
		RobotCode:   "robot-001",
	}
	observedAt := time.Date(2026, 5, 26, 1, 0, 0, 0, time.UTC)
	chunk1 := domain.RecordingChunk{
		ID:                "chunk-001",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-001_manifest.json",
		MediaObjectKeys:   map[string]string{},
		StartedAt:         observedAt,
		EndedAt:           observedAt.Add(10 * time.Minute),
	}
	chunk2 := domain.RecordingChunk{
		ID:                "chunk-002",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-002_manifest.json",
		MediaObjectKeys:   map[string]string{},
		StartedAt:         observedAt.Add(3 * time.Minute),
		EndedAt:           observedAt.Add(13 * time.Minute),
	}
	appServerClient := &fakeAppServerClient{
		targets: []domain.Mission{target},
		tickResults: []domain.RecordingTickResult{
			{Chunk: chunk1},
			{Chunk: chunk2},
		},
	}
	mediaUploader := &fakeMediaUploader{}
	worker := newWorkerWithCollaborators(
		config.RecorderWorkerConfig{
			RecordingChunkDuration:    10 * time.Minute,
			RecordingMediaIdleTimeout: 2 * time.Minute,
		},
		appServerClient,
		&fakeObjectStorage{manifestSize: 42},
	)
	worker.mediaUploader = mediaUploader
	mediaKey := recorderMediaKey(target.MissionCode, target.RobotCode)

	worker.tick(context.Background())
	if _, _, err := worker.ensureActiveRecordingChunk(context.Background(), mediaKey, observedAt); err != nil {
		t.Fatalf("chunk open failed: %v", err)
	}
	worker.finalizeIdleRecordingChunks(observedAt.Add(119 * time.Second))
	worker.processPendingRecordingChunkFinalizations(context.Background())
	if len(mediaUploader.finalizedChunks) != 0 {
		t.Fatalf("finalized chunks before timeout = %d, want 0", len(mediaUploader.finalizedChunks))
	}

	worker.finalizeIdleRecordingChunks(observedAt.Add(121 * time.Second))
	worker.processPendingRecordingChunkFinalizations(context.Background())
	if len(mediaUploader.finalizedChunks) != 1 {
		t.Fatalf("finalized chunks after timeout = %d, want 1", len(mediaUploader.finalizedChunks))
	}
	if mediaUploader.finalizedChunks[0].ID != chunk1.ID {
		t.Fatalf("finalized chunk = %q, want %q", mediaUploader.finalizedChunks[0].ID, chunk1.ID)
	}
	if _, ok := worker.activeTargets[mediaKey]; !ok {
		t.Fatal("active target disappeared after idle chunk finalization")
	}
	if _, ok := worker.activeChunks[mediaKey]; ok {
		t.Fatal("idle active chunk remained")
	}

	nextChunk, ok, err := worker.ensureActiveRecordingChunk(context.Background(), mediaKey, observedAt.Add(3*time.Minute))
	if err != nil {
		t.Fatalf("next chunk open failed: %v", err)
	}
	if !ok || nextChunk.ID != chunk2.ID {
		t.Fatalf("next chunk = %#v/%t, want chunk-002", nextChunk, ok)
	}
}
