package recording

import (
	"context"
	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/utils"
	"strings"
	"time"
)

func (w *Worker) setActiveRecordingChunk(roomID string, chunk domain.RecordingChunk) (domain.RecordingChunk, bool) {
	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()
	previousChunk, hadPreviousChunk := w.activeChunks[roomID]
	w.activeChunks[roomID] = chunk
	if !hadPreviousChunk || previousChunk.ID == "" || previousChunk.ID == chunk.ID {
		return domain.RecordingChunk{}, false
	}
	return previousChunk, true
}

func (w *Worker) currentRecordingChunk(mediaKey string, observedAt time.Time) (domain.RecordingChunk, bool) {
	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()
	chunk, ok := w.activeChunks[mediaKey]
	if !ok || strings.TrimSpace(chunk.ID) == "" {
		return domain.RecordingChunk{}, false
	}
	if chunk.EndedAt.IsZero() || observedAt.Before(chunk.EndedAt) {
		return chunk, true
	}
	return domain.RecordingChunk{}, false
}

func (w *Worker) recordingTarget(mediaKey string) (domain.Mission, bool) {
	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()
	target, ok := w.activeTargets[mediaKey]
	return target, ok
}

func (w *Worker) ensureActiveRecordingChunk(ctx context.Context, mediaKey string, observedAt time.Time) (domain.RecordingChunk, bool, error) {
	if observedAt.IsZero() {
		observedAt = time.Now().UTC()
	}
	if chunk, ok := w.currentRecordingChunk(mediaKey, observedAt); ok {
		return chunk, true, nil
	}

	w.chunkMu.Lock()
	defer w.chunkMu.Unlock()
	if chunk, ok := w.currentRecordingChunk(mediaKey, observedAt); ok {
		return chunk, true, nil
	}

	target, ok := w.recordingTarget(mediaKey)
	if !ok {
		return domain.RecordingChunk{}, false, nil
	}
	result, err := w.appServerClient.CreateRecordingTick(ctx, target, w.config.RecordingChunkDuration, observedAt)
	if err != nil {
		return domain.RecordingChunk{}, false, err
	}
	previousChunk, shouldFinalizePreviousChunk := w.setActiveRecordingChunk(mediaKey, result.Chunk)
	if shouldFinalizePreviousChunk {
		w.queueRecordingChunkFinalization(mediaKey, previousChunk)
	}
	return result.Chunk, true, nil
}

func (w *Worker) queueInactiveRecordingChunks(activeTargetKeys map[string]struct{}) {
	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()

	for mediaKey, chunk := range w.activeChunks {
		if _, ok := activeTargetKeys[mediaKey]; ok {
			continue
		}
		delete(w.activeChunks, mediaKey)
		w.pendingFinalizations[recordingChunkFinalizationKey(mediaKey, chunk.ID)] = recordingChunkFinalization{
			mediaKey: mediaKey,
			chunk:    chunk,
		}
	}
}

func (w *Worker) queueRecordingChunkFinalization(mediaKey string, chunk domain.RecordingChunk) {
	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()
	w.pendingFinalizations[recordingChunkFinalizationKey(mediaKey, chunk.ID)] = recordingChunkFinalization{
		mediaKey: mediaKey,
		chunk:    chunk,
	}
}

func (w *Worker) pendingRecordingChunkFinalizations() []recordingChunkFinalization {
	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()

	pendingFinalizations := make([]recordingChunkFinalization, 0, len(w.pendingFinalizations))
	for _, pendingFinalization := range w.pendingFinalizations {
		pendingFinalizations = append(pendingFinalizations, pendingFinalization)
	}
	return pendingFinalizations
}

func (w *Worker) removePendingRecordingChunkFinalization(mediaKey string, chunkID string) {
	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()
	delete(w.pendingFinalizations, recordingChunkFinalizationKey(mediaKey, chunkID))
}

func recordingChunkFinalizationKey(mediaKey string, chunkID string) string {
	return utils.SafePathToken(mediaKey) + "/" + utils.SafePathToken(chunkID)
}
