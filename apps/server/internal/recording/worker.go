package recording

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"robot-center/apps/server/internal/config"
	"robot-center/apps/server/internal/domain"
)

type Worker struct {
	config               config.RecorderWorkerConfig
	appServerClient      AppServerClient
	objectStorage        ObjectStorage
	mediaUploader        MediaUploader
	subscriberMu         sync.RWMutex
	subscriberCancels    map[string]context.CancelFunc
	subscriberStatuses   map[string]recorderSessionStatus
	mediaMu              sync.Mutex
	activeChunks         map[string]domain.RecordingChunk
	pendingFinalizations map[string]recordingChunkFinalization
	audioWriters         map[string]*activeAudioWriter
	h264ParameterSets    map[string]h264ParameterSets
}

func NewWorker(cfg config.RecorderWorkerConfig) *Worker {
	return newWorkerWithCollaborators(
		cfg,
		NewHTTPAppServerClient(cfg.AppServerURL, nil),
		NewMinIOObjectStorage(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOBucket),
	)
}

func newWorkerWithCollaborators(cfg config.RecorderWorkerConfig, appServerClient AppServerClient, objectStorage ObjectStorage) *Worker {
	if appServerClient == nil {
		appServerClient = NewHTTPAppServerClient(cfg.AppServerURL, nil)
	}
	if objectStorage == nil {
		objectStorage = NewMinIOObjectStorage(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOBucket)
	}
	worker := &Worker{
		config:               cfg,
		appServerClient:      appServerClient,
		objectStorage:        objectStorage,
		subscriberCancels:    map[string]context.CancelFunc{},
		subscriberStatuses:   map[string]recorderSessionStatus{},
		activeChunks:         map[string]domain.RecordingChunk{},
		pendingFinalizations: map[string]recordingChunkFinalization{},
		audioWriters:         map[string]*activeAudioWriter{},
		h264ParameterSets:    map[string]h264ParameterSets{},
	}
	worker.mediaUploader = NewMediaUploader(appServerClient, objectStorage, worker)
	return worker
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.config.PollInterval)
	defer ticker.Stop()

	log.Printf("recorder-worker polling app-server=%s interval=%s chunk=%s", w.config.AppServerURL, w.config.PollInterval, w.config.RecordingChunkDuration)
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
		w.finalizeInactiveRecordingChunks(activeTargetKeys)
		w.processPendingRecordingChunkFinalizations(ctx)
		log.Println("recorder-worker tick: no active recording targets")
		return
	}

	for _, target := range targets {
		mediaKey := recorderMediaKey(target.MissionCode, target.RobotCode)
		activeTargetKeys[mediaKey] = struct{}{}
		result, err := w.appServerClient.CreateRecordingTick(ctx, target, w.config.RecordingChunkDuration, time.Now().UTC())
		if err != nil {
			log.Printf("recorder-worker tick failed mission=%s robot=%s: %v", target.MissionCode, target.RobotCode, err)
			continue
		}
		previousChunk, shouldFinalizePreviousChunk := w.setActiveRecordingChunk(mediaKey, result.Chunk)
		if shouldFinalizePreviousChunk {
			w.queueRecordingChunkFinalization(mediaKey, previousChunk)
		}
	}
	w.finalizeInactiveRecordingChunks(activeTargetKeys)
	w.processPendingRecordingChunkFinalizations(ctx)
}

func (w *Worker) finalizeInactiveRecordingChunks(activeTargetKeys map[string]struct{}) {
	w.queueInactiveRecordingChunks(activeTargetKeys)
}

func (w *Worker) processPendingRecordingChunkFinalizations(ctx context.Context) {
	pendingFinalizations := w.pendingRecordingChunkFinalizations()
	for _, pendingFinalization := range pendingFinalizations {
		if err := w.finalizeRecordingChunk(ctx, pendingFinalization.mediaKey, pendingFinalization.chunk); err != nil {
			log.Printf("recorder-worker chunk finalize failed chunk=%s: %v", pendingFinalization.chunk.ID, err)
			continue
		}
		w.removePendingRecordingChunkFinalization(pendingFinalization.mediaKey, pendingFinalization.chunk.ID)
	}
}

func (w *Worker) finalizeRecordingChunk(ctx context.Context, mediaKey string, chunk domain.RecordingChunk) error {
	if strings.TrimSpace(chunk.ID) == "" {
		return nil
	}
	w.closeAudioWriterForChunk(mediaKey, chunk.ID)
	if err := w.mediaUploader.UploadMediaSnapshots(ctx, mediaKey, chunk); err != nil {
		return err
	}
	finalizedChunk := chunk
	finalizedChunk.Status = "uploaded"
	manifestSizeBytes, err := w.objectStorage.UploadManifest(ctx, finalizedChunk.ManifestObjectKey, domain.NewRecordingManifest(finalizedChunk))
	if err != nil {
		return err
	}
	if err := w.appServerClient.MarkRecordingChunkUploaded(ctx, finalizedChunk.ID, manifestSizeBytes); err != nil {
		return err
	}
	log.Printf("recorder-worker chunk finalized mission=%s robot=%s chunk=%s key=%s", finalizedChunk.MissionCode, finalizedChunk.RobotCode, finalizedChunk.ID, finalizedChunk.ManifestObjectKey)
	return nil
}
