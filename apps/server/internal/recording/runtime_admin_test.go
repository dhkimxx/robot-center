package recording

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"robot-center/apps/server/internal/config"
	"robot-center/apps/server/internal/domain"
)

func TestWorkerClearRuntimeRecordingsRemovesRuntimeFiles(t *testing.T) {
	runtimeDirectory := t.TempDir()
	t.Setenv("RECORDER_RUNTIME_DIR", runtimeDirectory)
	writeRuntimeTestFile(t, filepath.Join(runtimeDirectory, "chunk-001", "rgb.h264"), []byte("rgb"))
	writeRuntimeTestFile(t, filepath.Join(runtimeDirectory, "chunk-001", "telemetry.jsonl"), []byte("telemetry"))
	writeRuntimeTestFile(t, filepath.Join(runtimeDirectory, "chunk-002", "thermal.h264"), []byte("thermal"))

	worker := NewWorker(config.RecorderWorkerConfig{Environment: "development"})
	result, err := worker.ClearRuntimeRecordings(context.Background(), domain.ClearRecorderRuntimeConfirmation)
	if err != nil {
		t.Fatalf("ClearRuntimeRecordings returned error: %v", err)
	}
	if result.RecordingDirectoriesDeleted != 2 || result.FilesDeleted != 3 || result.DeletedBytes != int64(len("rgbtelemetrythermal")) {
		t.Fatalf("unexpected clear result: %#v", result)
	}
	if entries, err := os.ReadDir(runtimeDirectory); err != nil || len(entries) != 0 {
		t.Fatalf("expected empty runtime directory, entries=%#v err=%v", entries, err)
	}
}

func TestWorkerClearRuntimeRecordingsRequiresConfirmation(t *testing.T) {
	worker := NewWorker(config.RecorderWorkerConfig{Environment: "development"})
	_, err := worker.ClearRuntimeRecordings(context.Background(), "wrong")
	if err != ErrRecorderRuntimeClearConfirmationRequired {
		t.Fatalf("expected confirmation error, got %v", err)
	}
}

func TestWorkerClearRuntimeRecordingsRejectsProduction(t *testing.T) {
	worker := NewWorker(config.RecorderWorkerConfig{Environment: "production"})
	_, err := worker.ClearRuntimeRecordings(context.Background(), domain.ClearRecorderRuntimeConfirmation)
	if err != ErrRecorderRuntimeClearForbidden {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestWorkerClearRuntimeRecordingsRejectsActiveState(t *testing.T) {
	runtimeDirectory := t.TempDir()
	t.Setenv("RECORDER_RUNTIME_DIR", runtimeDirectory)
	writeRuntimeTestFile(t, filepath.Join(runtimeDirectory, "chunk-001", "rgb.h264"), []byte("rgb"))

	worker := NewWorker(config.RecorderWorkerConfig{Environment: "development"})
	worker.activeTargets["mission-001/robot-001"] = domain.Mission{MissionCode: "mission-001", RobotCode: "robot-001"}
	_, err := worker.ClearRuntimeRecordings(context.Background(), domain.ClearRecorderRuntimeConfirmation)
	if err != ErrRecorderRuntimeClearActive {
		t.Fatalf("expected active state error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(runtimeDirectory, "chunk-001", "rgb.h264")); err != nil {
		t.Fatalf("expected runtime file to remain, got %v", err)
	}
}

func TestWorkerPruneRuntimeRecordingsKeepsActiveAndPendingChunks(t *testing.T) {
	runtimeDirectory := t.TempDir()
	t.Setenv("RECORDER_RUNTIME_DIR", runtimeDirectory)
	writeRuntimeTestFile(t, filepath.Join(runtimeDirectory, "chunk-active", "rgb.h264"), []byte("active"))
	writeRuntimeTestFile(t, filepath.Join(runtimeDirectory, "chunk-pending", "thermal.h264"), []byte("pending"))
	writeRuntimeTestFile(t, filepath.Join(runtimeDirectory, "chunk-old", "rgb.h264"), []byte("old"))
	writeRuntimeTestFile(t, filepath.Join(runtimeDirectory, "chunk-old", "telemetry.jsonl"), []byte("telemetry"))

	worker := NewWorker(config.RecorderWorkerConfig{Environment: "development"})
	worker.activeTargets["mission-001/robot-001"] = domain.Mission{MissionCode: "mission-001", RobotCode: "robot-001"}
	worker.activeChunks["mission-001/robot-001"] = domain.RecordingChunk{ID: "chunk-active"}
	worker.pendingFinalizations["mission-001/robot-001/chunk-pending"] = recordingChunkFinalization{
		chunk: domain.RecordingChunk{ID: "chunk-pending"},
	}

	result, err := worker.PruneRuntimeRecordings(context.Background(), domain.PruneRecorderRuntimeConfirmation)
	if err != nil {
		t.Fatalf("PruneRuntimeRecordings returned error: %v", err)
	}
	if result.RecordingDirectoriesDeleted != 1 || result.FilesDeleted != 2 || result.DeletedBytes != int64(len("oldtelemetry")) {
		t.Fatalf("unexpected prune result: %#v", result)
	}
	if _, err := os.Stat(filepath.Join(runtimeDirectory, "chunk-active", "rgb.h264")); err != nil {
		t.Fatalf("expected active runtime file to remain, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(runtimeDirectory, "chunk-pending", "thermal.h264")); err != nil {
		t.Fatalf("expected pending runtime file to remain, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(runtimeDirectory, "chunk-old")); !os.IsNotExist(err) {
		t.Fatalf("expected old chunk directory to be removed, got %v", err)
	}
}

func TestWorkerRuntimeStatusReportsUsageAndClearableState(t *testing.T) {
	runtimeDirectory := t.TempDir()
	t.Setenv("RECORDER_RUNTIME_DIR", runtimeDirectory)
	writeRuntimeTestFile(t, filepath.Join(runtimeDirectory, "chunk-001", "rgb.h264"), []byte("rgb"))
	writeRuntimeTestFile(t, filepath.Join(runtimeDirectory, "chunk-001", "telemetry.jsonl"), []byte("telemetry"))

	worker := NewWorker(config.RecorderWorkerConfig{Environment: "development"})
	status := worker.RuntimeStatus(context.Background())
	if status.Status != "ok" || status.RecordingDirectories != 1 || status.Files != 2 || status.UsedBytes != int64(len("rgbtelemetry")) {
		t.Fatalf("unexpected runtime status: %#v", status)
	}
	if !status.Clearable || status.BlockingReason != "" {
		t.Fatalf("expected runtime to be clearable, got %#v", status)
	}
	if status.TotalBytes <= 0 || status.AvailableBytes <= 0 {
		t.Fatalf("expected filesystem capacity in status, got %#v", status)
	}
}

func TestWorkerRuntimeStatusReportsActiveBlockingState(t *testing.T) {
	runtimeDirectory := t.TempDir()
	t.Setenv("RECORDER_RUNTIME_DIR", runtimeDirectory)
	worker := NewWorker(config.RecorderWorkerConfig{Environment: "development"})
	worker.activeTargets["mission-001/robot-001"] = domain.Mission{MissionCode: "mission-001", RobotCode: "robot-001"}

	status := worker.RuntimeStatus(context.Background())
	if status.Clearable || status.BlockingReason != "active recording target" {
		t.Fatalf("expected active blocking status, got %#v", status)
	}
}

func writeRuntimeTestFile(t *testing.T, path string, payload []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		t.Fatal(err)
	}
}
