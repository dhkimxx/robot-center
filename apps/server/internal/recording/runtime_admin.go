package recording

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"robot-center/apps/server/internal/domain"
)

var (
	ErrRecorderRuntimeClearForbidden            = errors.New("recorder runtime clear is disabled in production")
	ErrRecorderRuntimeClearConfirmationRequired = errors.New("recorder runtime clear confirmation is required")
	ErrRecorderRuntimeClearActive               = errors.New("recorder runtime has active recording state")
)

func (w *Worker) ClearRuntimeRecordings(ctx context.Context, confirmation string) (domain.RecorderRuntimeClearResult, error) {
	if strings.EqualFold(strings.TrimSpace(w.config.Environment), "production") {
		return domain.RecorderRuntimeClearResult{}, ErrRecorderRuntimeClearForbidden
	}
	if strings.TrimSpace(confirmation) != domain.ClearRecorderRuntimeConfirmation {
		return domain.RecorderRuntimeClearResult{}, ErrRecorderRuntimeClearConfirmationRequired
	}

	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()
	if len(w.activeTargets) > 0 || len(w.activeChunks) > 0 || len(w.pendingFinalizations) > 0 || len(w.audioWriters) > 0 {
		return domain.RecorderRuntimeClearResult{}, ErrRecorderRuntimeClearActive
	}

	select {
	case <-ctx.Done():
		return domain.RecorderRuntimeClearResult{}, ctx.Err()
	default:
	}

	runtimeDirectory := recordingRuntimeDirectory()
	result, err := summarizeRecordingRuntime(runtimeDirectory)
	if err != nil {
		return domain.RecorderRuntimeClearResult{}, err
	}
	if err := os.RemoveAll(runtimeDirectory); err != nil {
		return domain.RecorderRuntimeClearResult{}, err
	}
	if err := os.MkdirAll(runtimeDirectory, 0o755); err != nil {
		return domain.RecorderRuntimeClearResult{}, err
	}
	return result, nil
}

func summarizeRecordingRuntime(runtimeDirectory string) (domain.RecorderRuntimeClearResult, error) {
	var result domain.RecorderRuntimeClearResult
	stat, err := os.Stat(runtimeDirectory)
	if errors.Is(err, os.ErrNotExist) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	if !stat.IsDir() {
		return result, nil
	}

	err = filepath.WalkDir(runtimeDirectory, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == runtimeDirectory {
			return nil
		}
		if entry.IsDir() {
			result.RecordingDirectoriesDeleted++
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		result.FilesDeleted++
		result.DeletedBytes += info.Size()
		return nil
	})
	if err != nil {
		return domain.RecorderRuntimeClearResult{}, err
	}
	return result, nil
}
