package recording

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"robot-center/apps/server/internal/config"
	"robot-center/apps/server/internal/domain"
)

type fakeAppServerClient struct {
	mu                  sync.Mutex
	targets             []domain.Mission
	tickTarget          domain.Mission
	tickTargets         []domain.Mission
	tickDuration        time.Duration
	tickAt              time.Time
	tickResult          domain.RecordingTickResult
	tickResults         []domain.RecordingTickResult
	markedChunkID       string
	markedChunkSize     int64
	markedChunkContext  RecordingUploadContext
	claimedJobs         []domain.RecordingFinalizationJob
	completedJobID      string
	completedJobContext RecordingUploadContext
	partialJobID        string
	partialJobContext   RecordingUploadContext
	failedJobID         string
	failedJobContext    RecordingUploadContext
	postErr             error
	postedLabels        []string
	postedPayloads      [][]byte
	postedPayloadCh     chan []byte
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

func (c *fakeAppServerClient) MarkRecordingFileUploaded(_ context.Context, _ string, _ string, _ int64, _ RecordingUploadContext) error {
	return nil
}

func (c *fakeAppServerClient) ClaimRecordingFinalizationJobs(_ context.Context, _ string, _ int, _ time.Duration) ([]domain.RecordingFinalizationJob, error) {
	jobs := c.claimedJobs
	c.claimedJobs = nil
	return jobs, nil
}

func (c *fakeAppServerClient) MarkRecordingFinalizationJobCompleted(_ context.Context, jobID string, uploadContext RecordingUploadContext) error {
	c.completedJobID = jobID
	c.completedJobContext = uploadContext
	return nil
}

func (c *fakeAppServerClient) MarkRecordingFinalizationJobPartial(_ context.Context, jobID string, uploadContext RecordingUploadContext, _ string) error {
	c.partialJobID = jobID
	c.partialJobContext = uploadContext
	return nil
}

func (c *fakeAppServerClient) MarkRecordingFinalizationJobFailed(_ context.Context, jobID string, uploadContext RecordingUploadContext, _ string) error {
	c.failedJobID = jobID
	c.failedJobContext = uploadContext
	return nil
}

func (c *fakeAppServerClient) MarkRecordingChunkUploaded(_ context.Context, chunkID string, sizeBytes int64, uploadContext RecordingUploadContext) error {
	c.markedChunkID = chunkID
	c.markedChunkSize = sizeBytes
	c.markedChunkContext = uploadContext
	return nil
}

func (c *fakeAppServerClient) PostDataChannelPayload(_ context.Context, label string, payload []byte) error {
	if c.postErr != nil {
		return c.postErr
	}
	payloadCopy := append([]byte(nil), payload...)
	c.mu.Lock()
	c.postedLabels = append(c.postedLabels, label)
	c.postedPayloads = append(c.postedPayloads, payloadCopy)
	c.mu.Unlock()
	if c.postedPayloadCh != nil {
		select {
		case c.postedPayloadCh <- payloadCopy:
		default:
		}
	}
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
	uploadedTypes   []string
	noMedia         bool
}

func (u *fakeMediaUploader) UploadMediaSnapshots(_ context.Context, mediaKey string, chunk domain.RecordingChunk, _ RecordingUploadContext) (RecordingMediaUploadResult, error) {
	u.finalizedKeys = append(u.finalizedKeys, mediaKey)
	u.finalizedChunks = append(u.finalizedChunks, chunk)
	if u.err != nil {
		return RecordingMediaUploadResult{}, u.err
	}
	if u.noMedia {
		return RecordingMediaUploadResult{}, nil
	}
	uploadedTypes := u.uploadedTypes
	if len(uploadedTypes) == 0 {
		uploadedTypes = []string{"rgb_audio_mp4"}
	}
	return RecordingMediaUploadResult{UploadedFileTypes: uploadedTypes}, nil
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

func TestWorkerQueuesRecorderDataChannelMessageForAppServerPost(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	appServerClient := &fakeAppServerClient{postedPayloadCh: make(chan []byte, 1)}
	worker := newWorkerWithCollaborators(config.RecorderWorkerConfig{}, appServerClient, &fakeObjectStorage{})
	worker.updateSubscriberStatus("mission-001", func(status *recorderSessionStatus) {
		status.robotCode = "robot-001"
		status.robotCodes = map[string]struct{}{"robot-001": {}}
	})
	worker.startRecorderDataQueueWorkers(ctx)

	worker.enqueueRecorderDataChannelMessage(ctx, "mission-001", "channel.telemetry", []byte(`{"missionId":"mission-001","samples":[]}`))

	var payload []byte
	select {
	case payload = <-appServerClient.postedPayloadCh:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for queued datachannel post")
	}
	var body map[string]any
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatalf("posted payload is not JSON: %v", err)
	}
	if body["robotCode"] != "robot-001" {
		t.Fatalf("posted robotCode = %#v, want robot-001", body["robotCode"])
	}
	waitForCondition(t, time.Second, func() bool {
		status := worker.SubscriberStatus()
		return len(status.Rooms) == 1 && status.Rooms[0].TelemetryStoredCount == 1
	})
	queueStatus := worker.RecorderDataQueueStatus()
	if queueStatus.PostDroppedCount != 0 || queueStatus.PostFailedCount != 0 {
		t.Fatalf("queue status = %#v, want no drops or failures", queueStatus)
	}
}

func TestWorkerDropsRecorderDataChannelPostWhenQueueIsFull(t *testing.T) {
	ctx := context.Background()
	worker := newWorkerWithCollaborators(config.RecorderWorkerConfig{}, &fakeAppServerClient{}, &fakeObjectStorage{})
	worker.dataPostQueue = make(chan recorderDataPostJob, 1)
	worker.dataPostQueue <- recorderDataPostJob{roomID: "mission-001", robotCode: "robot-001", storageLabel: "channel.spatial"}
	worker.updateSubscriberStatus("mission-001", func(status *recorderSessionStatus) {
		status.robotCode = "robot-001"
		status.robotCodes = map[string]struct{}{"robot-001": {}}
	})

	worker.enqueueRecorderDataChannelMessage(ctx, "mission-001", "channel.spatial", []byte(`{"missionId":"mission-001","robotCode":"robot-001","samples":[]}`))

	queueStatus := worker.RecorderDataQueueStatus()
	if queueStatus.PostDroppedCount != 1 {
		t.Fatalf("post dropped count = %d, want 1", queueStatus.PostDroppedCount)
	}
	if queueStatus.PostQueueDepth != 1 {
		t.Fatalf("post queue depth = %d, want 1", queueStatus.PostQueueDepth)
	}
	if queueStatus.LastPostError == "" {
		t.Fatal("last post error is empty, want queue full error")
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

func TestWorkerTickFinalizesClaimedJobWhenMissionAlreadyEnded(t *testing.T) {
	chunk := domain.RecordingChunk{
		ID:                "chunk-001",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-001_manifest.json",
		MediaObjectKeys:   map[string]string{},
	}
	appServerClient := &fakeAppServerClient{
		claimedJobs: []domain.RecordingFinalizationJob{
			{ID: "job-001", RecordingChunkID: chunk.ID, Attempts: 2, Chunk: chunk},
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

	if len(mediaUploader.finalizedChunks) != 1 {
		t.Fatalf("finalized chunks = %d, want 1", len(mediaUploader.finalizedChunks))
	}
	if appServerClient.completedJobID != "job-001" {
		t.Fatalf("completed job id = %q, want job-001", appServerClient.completedJobID)
	}
	if appServerClient.markedChunkID != chunk.ID {
		t.Fatalf("marked chunk id = %q, want %q", appServerClient.markedChunkID, chunk.ID)
	}
	if appServerClient.markedChunkContext.WorkerID != worker.workerID || appServerClient.markedChunkContext.Attempt != 2 {
		t.Fatalf("marked chunk context = %#v, want worker=%q attempt=2", appServerClient.markedChunkContext, worker.workerID)
	}
	if appServerClient.completedJobContext.WorkerID != worker.workerID || appServerClient.completedJobContext.Attempt != 2 {
		t.Fatalf("completed job context = %#v, want worker=%q attempt=2", appServerClient.completedJobContext, worker.workerID)
	}
}

func TestWorkerMarksClaimedJobPartialWhenNoMediaExists(t *testing.T) {
	chunk := domain.RecordingChunk{
		ID:                "chunk-001",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-001_manifest.json",
		MediaObjectKeys:   map[string]string{},
	}
	appServerClient := &fakeAppServerClient{
		claimedJobs: []domain.RecordingFinalizationJob{
			{ID: "job-001", RecordingChunkID: chunk.ID, Attempts: 3, Chunk: chunk},
		},
	}
	objectStorage := &fakeObjectStorage{manifestSize: 42}
	mediaUploader := &fakeMediaUploader{noMedia: true}
	worker := newWorkerWithCollaborators(
		config.RecorderWorkerConfig{RecordingChunkDuration: 10 * time.Minute},
		appServerClient,
		objectStorage,
	)
	worker.mediaUploader = mediaUploader

	worker.tick(context.Background())

	if appServerClient.partialJobID != "job-001" {
		t.Fatalf("partial job id = %q, want job-001", appServerClient.partialJobID)
	}
	if appServerClient.partialJobContext.WorkerID != worker.workerID || appServerClient.partialJobContext.Attempt != 3 {
		t.Fatalf("partial job context = %#v, want worker=%q attempt=3", appServerClient.partialJobContext, worker.workerID)
	}
	if appServerClient.markedChunkID != "" {
		t.Fatalf("marked chunk id = %q, want empty when no media exists", appServerClient.markedChunkID)
	}
}

func TestWorkerMarksClaimedJobFailedWithClaimContext(t *testing.T) {
	chunk := domain.RecordingChunk{
		ID:                "chunk-001",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/chunk-001_manifest.json",
		MediaObjectKeys:   map[string]string{},
	}
	appServerClient := &fakeAppServerClient{
		claimedJobs: []domain.RecordingFinalizationJob{
			{ID: "job-001", RecordingChunkID: chunk.ID, Attempts: 4, Chunk: chunk},
		},
	}
	objectStorage := &fakeObjectStorage{manifestSize: 42}
	mediaUploader := &fakeMediaUploader{err: errors.New("mux failed")}
	worker := newWorkerWithCollaborators(
		config.RecorderWorkerConfig{RecordingChunkDuration: 10 * time.Minute},
		appServerClient,
		objectStorage,
	)
	worker.mediaUploader = mediaUploader

	worker.tick(context.Background())

	if appServerClient.failedJobID != "job-001" {
		t.Fatalf("failed job id = %q, want job-001", appServerClient.failedJobID)
	}
	if appServerClient.failedJobContext.WorkerID != worker.workerID || appServerClient.failedJobContext.Attempt != 4 {
		t.Fatalf("failed job context = %#v, want worker=%q attempt=4", appServerClient.failedJobContext, worker.workerID)
	}
	if appServerClient.markedChunkID != "" {
		t.Fatalf("marked chunk id = %q, want empty after upload failure", appServerClient.markedChunkID)
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

func waitForCondition(t *testing.T, timeout time.Duration, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition was not met before timeout")
}
