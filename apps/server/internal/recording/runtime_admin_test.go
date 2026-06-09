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

func writeRuntimeTestFile(t *testing.T, path string, payload []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		t.Fatal(err)
	}
}
