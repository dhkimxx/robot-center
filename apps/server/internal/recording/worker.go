package recording

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"robot-center/apps/server/internal/config"
	"robot-center/apps/server/internal/domain"
)

type Worker struct {
	config                 config.RecorderWorkerConfig
	appServerClient        AppServerClient
	objectStorage          ObjectStorage
	mediaUploader          MediaUploader
	workerID               string
	dataQueueStartOnce     sync.Once
	dataAppendQueue        chan recorderDataAppendJob
	dataPostQueue          chan recorderDataPostJob
	dataQueueMu            sync.RWMutex
	dataQueueRuntime       recorderDataQueueRuntime
	subscriberMu           sync.RWMutex
	subscriberCancels      map[string]context.CancelFunc
	subscriberStatuses     map[string]recorderSessionStatus
	chunkMu                sync.Mutex
	mediaMu                sync.Mutex
	activeTargets          map[string]domain.Mission
	activeChunks           map[string]domain.RecordingChunk
	activeChunkMediaAt     map[string]time.Time
	pendingFinalizations   map[string]recordingChunkFinalization
	audioWriters           map[string]*activeAudioWriter
	h264ParameterSets      map[string]h264ParameterSets
	h264ChunkKeyframeWaits map[string]bool
	h264KeyframeRequests   map[string]time.Time
	h264Timings            map[string]h264TrackTiming
}

func NewWorker(cfg config.RecorderWorkerConfig) *Worker {
	return newWorkerWithCollaborators(
		cfg,
		NewHTTPAppServerClient(cfg.AppServerInternalURL, nil),
		NewMinIOObjectStorage(cfg.MinIOInternalURL, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOBucket),
	)
}

func newWorkerWithCollaborators(cfg config.RecorderWorkerConfig, appServerClient AppServerClient, objectStorage ObjectStorage) *Worker {
	if appServerClient == nil {
		appServerClient = NewHTTPAppServerClient(cfg.AppServerInternalURL, nil)
	}
	if objectStorage == nil {
		objectStorage = NewMinIOObjectStorage(cfg.MinIOInternalURL, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOBucket)
	}
	hostname, _ := os.Hostname()
	if strings.TrimSpace(hostname) == "" {
		hostname = "local"
	}
	worker := &Worker{
		config:                 cfg,
		appServerClient:        appServerClient,
		objectStorage:          objectStorage,
		workerID:               "recorder-" + hostname,
		dataAppendQueue:        make(chan recorderDataAppendJob, recorderDataAppendQueueCapacity),
		dataPostQueue:          make(chan recorderDataPostJob, recorderDataPostQueueCapacity),
		subscriberCancels:      map[string]context.CancelFunc{},
		subscriberStatuses:     map[string]recorderSessionStatus{},
		activeTargets:          map[string]domain.Mission{},
		activeChunks:           map[string]domain.RecordingChunk{},
		activeChunkMediaAt:     map[string]time.Time{},
		pendingFinalizations:   map[string]recordingChunkFinalization{},
		audioWriters:           map[string]*activeAudioWriter{},
		h264ParameterSets:      map[string]h264ParameterSets{},
		h264ChunkKeyframeWaits: map[string]bool{},
		h264KeyframeRequests:   map[string]time.Time{},
		h264Timings:            map[string]h264TrackTiming{},
	}
	worker.mediaUploader = NewMediaUploader(appServerClient, objectStorage, worker)
	return worker
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.config.PollInterval)
	defer ticker.Stop()

	log.Printf("recorder-worker polling app-server=%s interval=%s chunk=%s", w.config.AppServerInternalURL, w.config.PollInterval, w.config.RecordingChunkDuration)
	w.startRecorderDataQueueWorkers(ctx)
	go w.runSubscriberLoop(ctx)
	w.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			log.Println("recorder-worker stopped")
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	targets, err := w.appServerClient.FetchRecordingTargets(ctx)
	if err != nil {
		log.Printf("recorder-worker target fetch failed: %v", err)
		return
	}
	activeTargetKeys := map[string]struct{}{}
	if len(targets) == 0 {
		w.updateActiveRecordingTargets(nil)
		w.finalizeInactiveRecordingChunks(activeTargetKeys)
		w.processPendingRecordingChunkFinalizations(ctx)
		w.processClaimedRecordingFinalizationJobs(ctx)
		log.Println("recorder-worker tick: no active recording targets")
		return
	}

	activeTargetKeys = w.updateActiveRecordingTargets(targets)
	w.finalizeInactiveRecordingChunks(activeTargetKeys)
	w.finalizeIdleRecordingChunks(time.Now().UTC())
	w.processPendingRecordingChunkFinalizations(ctx)
	w.processClaimedRecordingFinalizationJobs(ctx)
}

func (w *Worker) updateActiveRecordingTargets(targets []domain.Mission) map[string]struct{} {
	activeTargetKeys := map[string]struct{}{}
	activeTargets := map[string]domain.Mission{}
	for _, target := range targets {
		mediaKey := recorderMediaKey(target.MissionCode, target.RobotCode)
		activeTargetKeys[mediaKey] = struct{}{}
		activeTargets[mediaKey] = target
	}

	w.mediaMu.Lock()
	w.activeTargets = activeTargets
	w.mediaMu.Unlock()
	return activeTargetKeys
}

func (w *Worker) finalizeInactiveRecordingChunks(activeTargetKeys map[string]struct{}) {
	w.queueInactiveRecordingChunks(activeTargetKeys)
}

func (w *Worker) processPendingRecordingChunkFinalizations(ctx context.Context) {
	pendingFinalizations := w.pendingRecordingChunkFinalizations()
	for _, pendingFinalization := range pendingFinalizations {
		if err := w.finalizeRecordingChunk(ctx, pendingFinalization.mediaKey, pendingFinalization.chunk, RecordingUploadContext{}); err != nil {
			log.Printf("recorder-worker chunk finalize failed chunk=%s: %v", pendingFinalization.chunk.ID, err)
			continue
		}
		w.removePendingRecordingChunkFinalization(pendingFinalization.mediaKey, pendingFinalization.chunk.ID)
	}
}

func (w *Worker) processClaimedRecordingFinalizationJobs(ctx context.Context) {
	jobs, err := w.appServerClient.ClaimRecordingFinalizationJobs(ctx, w.workerID, 8, 2*time.Minute)
	if err != nil {
		log.Printf("recorder-worker finalization job claim failed: %v", err)
		return
	}
	for _, job := range jobs {
		mediaKey := recorderMediaKey(job.Chunk.MissionCode, job.Chunk.RobotCode)
		uploadContext := RecordingUploadContext{WorkerID: w.workerID, Attempt: job.Attempts}
		err := w.finalizeRecordingChunk(ctx, mediaKey, job.Chunk, uploadContext)
		switch {
		case err == nil:
			if markErr := w.appServerClient.MarkRecordingFinalizationJobCompleted(ctx, job.ID, uploadContext); markErr != nil {
				log.Printf("recorder-worker finalization job complete callback failed job=%s: %v", job.ID, markErr)
			}
		case errors.Is(err, errNoRecordingMedia):
			if markErr := w.appServerClient.MarkRecordingFinalizationJobPartial(ctx, job.ID, uploadContext, truncateRecordingFinalizationReason(err.Error())); markErr != nil {
				log.Printf("recorder-worker finalization job partial callback failed job=%s: %v", job.ID, markErr)
			}
		default:
			if markErr := w.appServerClient.MarkRecordingFinalizationJobFailed(ctx, job.ID, uploadContext, truncateRecordingFinalizationReason(err.Error())); markErr != nil {
				log.Printf("recorder-worker finalization job failed callback failed job=%s: %v", job.ID, markErr)
			}
		}
	}
}

func truncateRecordingFinalizationReason(reason string) string {
	reason = strings.TrimSpace(reason)
	const maxReasonLength = 800
	if len(reason) <= maxReasonLength {
		return reason
	}
	return reason[:maxReasonLength] + "..."
}

func (w *Worker) finalizeRecordingChunk(ctx context.Context, mediaKey string, chunk domain.RecordingChunk, uploadContext RecordingUploadContext) error {
	if strings.TrimSpace(chunk.ID) == "" {
		return nil
	}
	w.closeAudioWriterForChunk(mediaKey, chunk.ID)
	uploadResult, err := w.mediaUploader.UploadMediaSnapshots(ctx, mediaKey, chunk, uploadContext)
	if err != nil {
		return err
	}
	if len(uploadResult.UploadedFileTypes) == 0 {
		return errNoRecordingMedia
	}
	finalizedChunk := chunk
	finalizedChunk.Status = "uploaded"
	manifestSizeBytes, err := w.objectStorage.UploadManifest(ctx, finalizedChunk.ManifestObjectKey, domain.NewRecordingManifest(finalizedChunk))
	if err != nil {
		return err
	}
	if err := w.appServerClient.MarkRecordingChunkUploaded(ctx, finalizedChunk.ID, manifestSizeBytes, uploadContext); err != nil {
		return err
	}
	log.Printf("recorder-worker chunk finalized mission=%s robot=%s chunk=%s key=%s", finalizedChunk.MissionCode, finalizedChunk.RobotCode, finalizedChunk.ID, finalizedChunk.ManifestObjectKey)
	return nil
}
