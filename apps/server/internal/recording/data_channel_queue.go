package recording

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"robot-center/apps/server/internal/monitorlog"
)

const (
	recorderDataAppendQueueCapacity = 10000
	recorderDataPostQueueCapacity   = 10000
	recorderDataPostWorkerCount     = 4
	recorderDataPostRetryCount      = 2
	recorderDataPostRetryDelay      = 100 * time.Millisecond
)

type RecorderDataQueueStatus struct {
	AppendQueueDepth   int    `json:"appendQueueDepth"`
	PostQueueDepth     int    `json:"postQueueDepth"`
	AppendDroppedCount uint64 `json:"appendDroppedCount"`
	PostDroppedCount   uint64 `json:"postDroppedCount"`
	AppendFailedCount  uint64 `json:"appendFailedCount"`
	PostFailedCount    uint64 `json:"postFailedCount"`
	PostRetryCount     uint64 `json:"postRetryCount"`
	LastAppendError    string `json:"lastAppendError,omitempty"`
	LastPostError      string `json:"lastPostError,omitempty"`
}

type recorderDataQueueRuntime struct {
	appendDroppedCount uint64
	postDroppedCount   uint64
	appendFailedCount  uint64
	postFailedCount    uint64
	postRetryCount     uint64
	lastAppendError    string
	lastPostError      string
}

type recorderDataChannelMessage struct {
	roomID       string
	robotCode    string
	storageLabel string
	fileLabel    string
	payload      []byte
}

type recorderDataAppendJob struct {
	roomID    string
	robotCode string
	fileLabel string
	payload   []byte
}

type recorderDataPostJob struct {
	roomID       string
	robotCode    string
	storageLabel string
	payload      []byte
}

func (w *Worker) startRecorderDataQueueWorkers(ctx context.Context) {
	w.dataQueueStartOnce.Do(func() {
		go w.runRecorderDataAppendWorker(ctx)
		for workerIndex := 0; workerIndex < recorderDataPostWorkerCount; workerIndex++ {
			go w.runRecorderDataPostWorker(ctx)
		}
	})
}

func (w *Worker) RecorderDataQueueStatus() RecorderDataQueueStatus {
	w.dataQueueMu.RLock()
	runtime := w.dataQueueRuntime
	w.dataQueueMu.RUnlock()

	return RecorderDataQueueStatus{
		AppendQueueDepth:   len(w.dataAppendQueue),
		PostQueueDepth:     len(w.dataPostQueue),
		AppendDroppedCount: runtime.appendDroppedCount,
		PostDroppedCount:   runtime.postDroppedCount,
		AppendFailedCount:  runtime.appendFailedCount,
		PostFailedCount:    runtime.postFailedCount,
		PostRetryCount:     runtime.postRetryCount,
		LastAppendError:    runtime.lastAppendError,
		LastPostError:      runtime.lastPostError,
	}
}

func (w *Worker) enqueueRecorderDataChannelMessage(ctx context.Context, roomID string, label string, payload []byte) {
	message, ok := w.normalizeRecorderDataChannelMessage(roomID, label, payload)
	if !ok {
		return
	}

	if message.fileLabel != "" {
		enqueued := w.enqueueRecorderDataAppendJob(ctx, recorderDataAppendJob{
			roomID:    message.roomID,
			robotCode: message.robotCode,
			fileLabel: message.fileLabel,
			payload:   append([]byte(nil), message.payload...),
		})
		if !enqueued {
			w.recordRecorderDataAppendDropped(roomID, fmt.Sprintf("%s append queue full", message.storageLabel))
		}
	}

	if recorderShouldPostDataChannelLabel(message.storageLabel) {
		enqueued := w.enqueueRecorderDataPostJob(ctx, recorderDataPostJob{
			roomID:       message.roomID,
			robotCode:    message.robotCode,
			storageLabel: message.storageLabel,
			payload:      append([]byte(nil), message.payload...),
		})
		if !enqueued {
			w.recordRecorderDataPostDropped(roomID, fmt.Sprintf("%s post queue full", message.storageLabel))
		}
	}

	if w.markRecorderDataObserved(message.roomID, message.robotCode, message.storageLabel) {
		monitorlog.Event("recorder-worker", "datachannel_first_message", "room", message.roomID, "robot", message.robotCode, "label", message.storageLabel, "bytes", len(message.payload))
	}
}

func (w *Worker) normalizeRecorderDataChannelMessage(roomID string, label string, payload []byte) (recorderDataChannelMessage, bool) {
	storageLabel := recorderStorageDataChannelLabel(label)
	if storageLabel == "" {
		return recorderDataChannelMessage{}, false
	}
	if !recorderShouldPostDataChannelLabel(storageLabel) && recorderDataChannelFileLabel(storageLabel) == "" {
		return recorderDataChannelMessage{}, false
	}
	if !json.Valid(payload) {
		w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
			status.lastError = fmt.Sprintf("invalid %s JSON payload", storageLabel)
		})
		return recorderDataChannelMessage{}, false
	}

	robotCode := robotCodeFromDataPayload(payload)
	if robotCode == "" {
		robotCode = w.singleSubscriberRobotCode(roomID)
	}
	if recorderShouldRequireRobotCode(storageLabel) && robotCode == "" {
		w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
			status.lastError = fmt.Sprintf("%s payload missing robotCode", storageLabel)
		})
		return recorderDataChannelMessage{}, false
	}
	if robotCode != "" {
		payload = recorderDataChannelPayloadWithContext(roomID, robotCode, storageLabel, payload)
	}

	return recorderDataChannelMessage{
		roomID:       roomID,
		robotCode:    robotCode,
		storageLabel: storageLabel,
		fileLabel:    recorderDataChannelFileLabel(storageLabel),
		payload:      payload,
	}, true
}

func (w *Worker) enqueueRecorderDataAppendJob(ctx context.Context, job recorderDataAppendJob) bool {
	select {
	case <-ctx.Done():
		return false
	case w.dataAppendQueue <- job:
		return true
	default:
		return false
	}
}

func (w *Worker) enqueueRecorderDataPostJob(ctx context.Context, job recorderDataPostJob) bool {
	select {
	case <-ctx.Done():
		return false
	case w.dataPostQueue <- job:
		return true
	default:
		return false
	}
}

func (w *Worker) runRecorderDataAppendWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-w.dataAppendQueue:
			mediaKey := recorderMediaKey(job.roomID, job.robotCode)
			if err := w.appendDataChannelPayload(mediaKey, job.fileLabel, job.payload); err != nil {
				w.recordRecorderDataAppendFailed(job.roomID, err.Error())
				log.Printf("recorder-worker datachannel append failed room=%s robot=%s label=%s: %v", job.roomID, job.robotCode, job.fileLabel, err)
			}
		}
	}
}

func (w *Worker) runRecorderDataPostWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-w.dataPostQueue:
			if err := w.postRecorderDataChannelPayloadWithRetry(ctx, job); err != nil {
				w.recordRecorderDataPostFailed(job.roomID, err.Error())
				log.Printf("recorder-worker datachannel persist failed room=%s robot=%s label=%s: %v", job.roomID, job.robotCode, job.storageLabel, err)
				continue
			}
			w.markRecorderDataPersisted(job.roomID, job.robotCode, job.storageLabel)
		}
	}
}

func (w *Worker) postRecorderDataChannelPayloadWithRetry(ctx context.Context, job recorderDataPostJob) error {
	var lastErr error
	for attempt := 0; attempt <= recorderDataPostRetryCount; attempt++ {
		if attempt > 0 {
			w.recordRecorderDataPostRetry()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(recorderDataPostRetryDelay):
			}
		}
		if err := w.appServerClient.PostDataChannelPayload(ctx, job.storageLabel, job.payload); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return lastErr
}

func (w *Worker) markRecorderDataObserved(roomID string, robotCode string, storageLabel string) bool {
	if robotCode == "" {
		return false
	}
	observedAt := time.Now().UTC()
	firstMessage := false
	w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
		status.robotCodes[robotCode] = struct{}{}
		if status.robotCode == "" {
			status.robotCode = robotCode
		}
		robotStatus := ensureRecorderRobotRuntime(status, robotCode)
		if _, ok := robotStatus.dataChannelLabels[storageLabel]; !ok {
			firstMessage = true
		}
		robotStatus.dataChannelLabels[storageLabel] = struct{}{}
		robotStatus.lastDataAt = observedAt
		robotStatus.updatedAt = observedAt
		status.robotStatuses[robotCode] = robotStatus
	})
	return firstMessage
}

func (w *Worker) markRecorderDataPersisted(roomID string, robotCode string, storageLabel string) {
	persistedAt := time.Now().UTC()
	w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
		if storageLabel == "channel.telemetry" {
			status.telemetryStoredCount++
		}
		if robotCode != "" {
			robotStatus := ensureRecorderRobotRuntime(status, robotCode)
			robotStatus.lastPersistedAt = persistedAt
			robotStatus.updatedAt = persistedAt
			status.robotStatuses[robotCode] = robotStatus
		}
		status.lastPersistedLabel = storageLabel
		status.lastPersistedAt = persistedAt
		status.lastError = ""
	})
	w.dataQueueMu.Lock()
	w.dataQueueRuntime.lastPostError = ""
	w.dataQueueMu.Unlock()
}

func (w *Worker) recordRecorderDataAppendDropped(roomID string, message string) {
	w.dataQueueMu.Lock()
	w.dataQueueRuntime.appendDroppedCount++
	w.dataQueueRuntime.lastAppendError = message
	w.dataQueueMu.Unlock()
	w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
		status.lastError = message
	})
}

func (w *Worker) recordRecorderDataPostDropped(roomID string, message string) {
	w.dataQueueMu.Lock()
	w.dataQueueRuntime.postDroppedCount++
	w.dataQueueRuntime.lastPostError = message
	w.dataQueueMu.Unlock()
	w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
		status.lastError = message
	})
}

func (w *Worker) recordRecorderDataAppendFailed(roomID string, message string) {
	w.dataQueueMu.Lock()
	w.dataQueueRuntime.appendFailedCount++
	w.dataQueueRuntime.lastAppendError = message
	w.dataQueueMu.Unlock()
	w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
		status.lastError = message
	})
}

func (w *Worker) recordRecorderDataPostFailed(roomID string, message string) {
	w.dataQueueMu.Lock()
	w.dataQueueRuntime.postFailedCount++
	w.dataQueueRuntime.lastPostError = message
	w.dataQueueMu.Unlock()
	w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
		status.lastError = message
	})
}

func (w *Worker) recordRecorderDataPostRetry() {
	w.dataQueueMu.Lock()
	w.dataQueueRuntime.postRetryCount++
	w.dataQueueMu.Unlock()
}

func recorderShouldPostDataChannelLabel(storageLabel string) bool {
	return storageLabel == "channel.telemetry" || storageLabel == "channel.event"
}

func recorderShouldRequireRobotCode(storageLabel string) bool {
	return storageLabel == "channel.telemetry" || storageLabel == "channel.event"
}
