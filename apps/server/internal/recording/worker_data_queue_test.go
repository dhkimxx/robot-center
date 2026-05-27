package recording

import (
	"context"
	"encoding/json"
	"robot-center/apps/server/internal/config"
	"testing"
	"time"
)

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
