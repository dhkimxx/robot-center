package recording

import (
	"context"
	"errors"
	"testing"
	"time"

	"robot-center/apps/server/internal/config"
	"robot-center/apps/server/internal/domain"
)

type fakeAppServerClient struct {
	targets         []domain.Mission
	tickTarget      domain.Mission
	tickTargets     []domain.Mission
	tickDuration    time.Duration
	tickAt          time.Time
	tickResult      domain.RecordingTickResult
	tickResults     []domain.RecordingTickResult
	markedChunkID   string
	markedChunkSize int64
}

func (c *fakeAppServerClient) FetchRecordingTargets(_ context.Context) ([]domain.Mission, error) {
	return c.targets, nil
}

func (c *fakeAppServerClient) CreateRecordingTick(_ context.Context, target domain.Mission, chunkDuration time.Duration, tickAt time.Time) (domain.RecordingTickResult, error) {
	c.tickTarget = target
	c.tickTargets = append(c.tickTargets, target)
	c.tickDuration = chunkDuration
	c.tickAt = tickAt
	if len(c.tickResults) > 0 {
		result := c.tickResults[0]
		c.tickResults = c.tickResults[1:]
		return result, nil
	}
	return c.tickResult, nil
}

func (c *fakeAppServerClient) MarkRecordingFileUploaded(_ context.Context, _ string, _ string, _ int64) error {
	return nil
}

func (c *fakeAppServerClient) MarkRecordingChunkUploaded(_ context.Context, chunkID string, sizeBytes int64) error {
	c.markedChunkID = chunkID
	c.markedChunkSize = sizeBytes
	return nil
}

func (c *fakeAppServerClient) PostDataChannelPayload(_ context.Context, _ string, _ []byte) error {
	return nil
}

type fakeObjectStorage struct {
	manifestKey     string
	manifestBody    map[string]any
	manifestSize    int64
	manifestUploads int
}

func (s *fakeObjectStorage) UploadManifest(_ context.Context, objectKey string, manifest map[string]any) (int64, error) {
	s.manifestKey = objectKey
	s.manifestBody = manifest
	s.manifestUploads++
	return s.manifestSize, nil
}

func (s *fakeObjectStorage) UploadFile(_ context.Context, _ string, _ string, _ string) (int64, error) {
	return 0, nil
}

type fakeMediaUploader struct {
	finalizedChunks []domain.RecordingChunk
	finalizedKeys   []string
	err             error
}

func (u *fakeMediaUploader) UploadMediaSnapshots(_ context.Context, mediaKey string, chunk domain.RecordingChunk) error {
	u.finalizedKeys = append(u.finalizedKeys, mediaKey)
	u.finalizedChunks = append(u.finalizedChunks, chunk)
	if u.err != nil {
		return u.err
	}
	return nil
}

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

func TestWorkerTickRetriesFailedChunkFinalization(t *testing.T) {
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
			{Chunk: chunk2},
		},
	}
	objectStorage := &fakeObjectStorage{manifestSize: 42}
	mediaUploader := &fakeMediaUploader{err: errors.New("temporary upload failure")}
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

	if objectStorage.manifestUploads != 0 {
		t.Fatalf("manifest uploads = %d, want 0 after failed media upload", objectStorage.manifestUploads)
	}
	if len(worker.pendingFinalizations) != 1 {
		t.Fatalf("pending finalizations = %d, want 1", len(worker.pendingFinalizations))
	}

	mediaUploader.err = nil
	worker.processPendingRecordingChunkFinalizations(context.Background())

	if objectStorage.manifestUploads != 1 {
		t.Fatalf("manifest uploads = %d, want 1 after retry", objectStorage.manifestUploads)
	}
	if appServerClient.markedChunkID != chunk1.ID {
		t.Fatalf("marked chunk id = %q, want %q", appServerClient.markedChunkID, chunk1.ID)
	}
	if len(worker.pendingFinalizations) != 0 {
		t.Fatalf("pending finalizations = %d, want 0 after retry", len(worker.pendingFinalizations))
	}
}
