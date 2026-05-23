package recording

import (
	"context"
	"log"
	"sync"
	"time"

	"robot-center/apps/server/internal/config"
	"robot-center/apps/server/internal/domain"
)

type Worker struct {
	config             config.RecorderWorkerConfig
	appServerClient    AppServerClient
	objectStorage      ObjectStorage
	mediaUploader      MediaUploader
	subscriberMu       sync.RWMutex
	subscriberCancels  map[string]context.CancelFunc
	subscriberStatuses map[string]recorderSessionStatus
	mediaMu            sync.Mutex
	activeChunks       map[string]domain.RecordingChunk
	audioWriters       map[string]*activeAudioWriter
	h264ParameterSets  map[string]h264ParameterSets
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
		config:             cfg,
		appServerClient:    appServerClient,
		objectStorage:      objectStorage,
		subscriberCancels:  map[string]context.CancelFunc{},
		subscriberStatuses: map[string]recorderSessionStatus{},
		activeChunks:       map[string]domain.RecordingChunk{},
		audioWriters:       map[string]*activeAudioWriter{},
		h264ParameterSets:  map[string]h264ParameterSets{},
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
	if len(targets) == 0 {
		log.Println("recorder-worker tick: no active recording targets")
		return
	}

	for _, target := range targets {
		mediaKey := recorderMediaKey(target.MissionCode, target.RobotCode)
		result, err := w.appServerClient.CreateRecordingTick(ctx, target, w.config.RecordingChunkDuration, time.Now().UTC())
		if err != nil {
			log.Printf("recorder-worker tick failed mission=%s robot=%s: %v", target.MissionCode, target.RobotCode, err)
			continue
		}
		w.setActiveRecordingChunk(mediaKey, result.Chunk)
		manifestSizeBytes, err := w.objectStorage.UploadManifest(ctx, result.Chunk.ManifestObjectKey, result.Manifest)
		if err != nil {
			log.Printf("recorder-worker manifest upload failed key=%s: %v", result.Chunk.ManifestObjectKey, err)
			continue
		}
		if err := w.mediaUploader.UploadMediaSnapshots(ctx, mediaKey, result.Chunk); err != nil {
			log.Printf("recorder-worker media snapshot upload failed chunk=%s: %v", result.Chunk.ID, err)
		}
		if err := w.appServerClient.MarkRecordingChunkUploaded(ctx, result.Chunk.ID, manifestSizeBytes); err != nil {
			log.Printf("recorder-worker upload status update failed chunk=%s: %v", result.Chunk.ID, err)
			continue
		}
		log.Printf("recorder-worker manifest uploaded mission=%s robot=%s key=%s", target.MissionCode, target.RobotCode, result.Chunk.ManifestObjectKey)
	}
}
