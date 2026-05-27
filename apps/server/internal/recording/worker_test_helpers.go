package recording

import (
	"context"
	"robot-center/apps/server/internal/domain"
	"sync"
	"testing"
	"time"
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
