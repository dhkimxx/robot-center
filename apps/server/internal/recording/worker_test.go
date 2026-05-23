package recording

import (
	"context"
	"testing"
	"time"

	"robot-center/apps/server/internal/config"
	"robot-center/apps/server/internal/domain"
)

type fakeAppServerClient struct {
	targets         []domain.Mission
	tickTarget      domain.Mission
	tickDuration    time.Duration
	tickAt          time.Time
	tickResult      domain.RecordingTickResult
	markedChunkID   string
	markedChunkSize int64
}

func (c *fakeAppServerClient) FetchRecordingTargets(_ context.Context) ([]domain.Mission, error) {
	return c.targets, nil
}

func (c *fakeAppServerClient) CreateRecordingTick(_ context.Context, target domain.Mission, chunkDuration time.Duration, tickAt time.Time) (domain.RecordingTickResult, error) {
	c.tickTarget = target
	c.tickDuration = chunkDuration
	c.tickAt = tickAt
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
	manifestKey  string
	manifestBody map[string]any
	manifestSize int64
}

func (s *fakeObjectStorage) UploadManifest(_ context.Context, objectKey string, manifest map[string]any) (int64, error) {
	s.manifestKey = objectKey
	s.manifestBody = manifest
	return s.manifestSize, nil
}

func (s *fakeObjectStorage) UploadFile(_ context.Context, _ string, _ string, _ string) (int64, error) {
	return 0, nil
}

func TestWorkerTickUsesInjectedCollaborators(t *testing.T) {
	target := domain.Mission{
		MissionCode: "mission-001",
		RobotCode:   "robot-001",
	}
	chunk := domain.RecordingChunk{
		ID:                "chunk-001",
		MissionCode:       "mission-001",
		RobotCode:         "robot-001",
		ManifestObjectKey: "missions/mission-001/robots/robot-001/manifest.json",
		MediaObjectKeys:   map[string]string{},
	}
	appServerClient := &fakeAppServerClient{
		targets: []domain.Mission{target},
		tickResult: domain.RecordingTickResult{
			Chunk:    chunk,
			Manifest: map[string]any{"chunkId": chunk.ID},
		},
	}
	objectStorage := &fakeObjectStorage{manifestSize: 42}
	worker := newWorkerWithCollaborators(
		config.RecorderWorkerConfig{RecordingChunkDuration: 10 * time.Minute},
		appServerClient,
		objectStorage,
	)

	worker.tick(context.Background())

	if appServerClient.tickTarget.MissionCode != target.MissionCode || appServerClient.tickTarget.RobotCode != target.RobotCode {
		t.Fatalf("tick target = %#v, want %#v", appServerClient.tickTarget, target)
	}
	if appServerClient.tickDuration != 10*time.Minute {
		t.Fatalf("tick duration = %s, want 10m", appServerClient.tickDuration)
	}
	if appServerClient.tickAt.IsZero() {
		t.Fatal("tickAt was not populated")
	}
	if objectStorage.manifestKey != chunk.ManifestObjectKey {
		t.Fatalf("manifest key = %q, want %q", objectStorage.manifestKey, chunk.ManifestObjectKey)
	}
	if objectStorage.manifestBody["chunkId"] != chunk.ID {
		t.Fatalf("manifest body = %#v", objectStorage.manifestBody)
	}
	if appServerClient.markedChunkID != chunk.ID {
		t.Fatalf("marked chunk id = %q, want %q", appServerClient.markedChunkID, chunk.ID)
	}
	if appServerClient.markedChunkSize != 42 {
		t.Fatalf("marked chunk size = %d, want 42", appServerClient.markedChunkSize)
	}
}
