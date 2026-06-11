package recording

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"robot-center/apps/server/internal/domain"
)

var (
	ErrRecorderRuntimeClearForbidden            = errors.New("recorder runtime clear is disabled in production")
	ErrRecorderRuntimeClearConfirmationRequired = errors.New("recorder runtime clear confirmation is required")
	ErrRecorderRuntimeClearActive               = errors.New("recorder runtime has active recording state")
)

const recorderRuntimeUsageCacheTTL = 30 * time.Second

func (w *Worker) RuntimeStatus(ctx context.Context) domain.RecorderRuntimeStatus {
	now := time.Now().UTC()
	status, err := w.cachedRuntimeUsage(ctx, now)
	if err != nil {
		status = domain.RecorderRuntimeStatus{
			Status:    "unavailable",
			Error:     err.Error(),
			UpdatedAt: now,
		}
	}

	clearable, blockingReason := w.runtimeClearability()
	if strings.EqualFold(strings.TrimSpace(w.config.Environment), "production") {
		clearable = false
		blockingReason = "production environment"
	}
	if status.Status == "" {
		status.Status = "ok"
	}
	status.Clearable = clearable && status.Status == "ok"
	status.BlockingReason = blockingReason
	status.UpdatedAt = now
	return status
}

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
	status, err := summarizeRecordingRuntime(ctx, runtimeDirectory)
	if err != nil {
		return domain.RecorderRuntimeClearResult{}, err
	}
	result := domain.RecorderRuntimeClearResult{
		RecordingDirectoriesDeleted: status.RecordingDirectories,
		FilesDeleted:                status.Files,
		DeletedBytes:                status.UsedBytes,
	}
	if err := os.RemoveAll(runtimeDirectory); err != nil {
		return domain.RecorderRuntimeClearResult{}, err
	}
	if err := os.MkdirAll(runtimeDirectory, 0o755); err != nil {
		return domain.RecorderRuntimeClearResult{}, err
	}
	w.invalidateRuntimeUsageCache()
	return result, nil
}

func (w *Worker) PruneRuntimeRecordings(ctx context.Context, confirmation string) (domain.RecorderRuntimeClearResult, error) {
	if strings.EqualFold(strings.TrimSpace(w.config.Environment), "production") {
		return domain.RecorderRuntimeClearResult{}, ErrRecorderRuntimeClearForbidden
	}
	if strings.TrimSpace(confirmation) != domain.PruneRecorderRuntimeConfirmation {
		return domain.RecorderRuntimeClearResult{}, ErrRecorderRuntimeClearConfirmationRequired
	}

	protectedEntries := w.protectedRuntimeEntries()
	runtimeDirectory := recordingRuntimeDirectory()
	entries, err := os.ReadDir(runtimeDirectory)
	if errors.Is(err, os.ErrNotExist) {
		return domain.RecorderRuntimeClearResult{}, nil
	}
	if err != nil {
		return domain.RecorderRuntimeClearResult{}, err
	}

	var result domain.RecorderRuntimeClearResult
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return domain.RecorderRuntimeClearResult{}, ctx.Err()
		default:
		}
		if _, protected := protectedEntries[entry.Name()]; protected {
			continue
		}
		entryPath := filepath.Join(runtimeDirectory, entry.Name())
		entrySummary, err := summarizeRuntimeEntry(ctx, entryPath)
		if err != nil {
			return domain.RecorderRuntimeClearResult{}, err
		}
		if err := os.RemoveAll(entryPath); err != nil {
			return domain.RecorderRuntimeClearResult{}, err
		}
		result.RecordingDirectoriesDeleted += entrySummary.RecordingDirectories
		result.FilesDeleted += entrySummary.Files
		result.DeletedBytes += entrySummary.UsedBytes
	}
	w.invalidateRuntimeUsageCache()
	return result, nil
}

func (w *Worker) runtimeClearability() (bool, string) {
	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()
	switch {
	case len(w.activeTargets) > 0:
		return false, "active recording target"
	case len(w.activeChunks) > 0:
		return false, "active recording chunk"
	case len(w.pendingFinalizations) > 0:
		return false, "pending finalization"
	case len(w.audioWriters) > 0:
		return false, "active audio writer"
	default:
		return true, ""
	}
}

func (w *Worker) protectedRuntimeEntries() map[string]struct{} {
	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()
	protectedEntries := make(map[string]struct{}, len(w.activeChunks)+len(w.pendingFinalizations)+len(w.audioWriters))
	for _, chunk := range w.activeChunks {
		if strings.TrimSpace(chunk.ID) != "" {
			protectedEntries[chunk.ID] = struct{}{}
		}
	}
	for _, pendingFinalization := range w.pendingFinalizations {
		if strings.TrimSpace(pendingFinalization.chunk.ID) != "" {
			protectedEntries[pendingFinalization.chunk.ID] = struct{}{}
		}
	}
	for _, writer := range w.audioWriters {
		if writer != nil && strings.TrimSpace(writer.chunkID) != "" {
			protectedEntries[writer.chunkID] = struct{}{}
		}
	}
	return protectedEntries
}

func (w *Worker) cachedRuntimeUsage(ctx context.Context, now time.Time) (domain.RecorderRuntimeStatus, error) {
	w.runtimeUsageMu.Lock()
	if !w.runtimeUsageCachedAt.IsZero() && now.Sub(w.runtimeUsageCachedAt) < recorderRuntimeUsageCacheTTL {
		cached := w.runtimeUsageCache
		w.runtimeUsageMu.Unlock()
		return cached, nil
	}
	w.runtimeUsageMu.Unlock()

	status, err := summarizeRecordingRuntime(ctx, recordingRuntimeDirectory())
	if err != nil {
		return domain.RecorderRuntimeStatus{}, err
	}
	status.Status = "ok"
	status.UpdatedAt = now

	w.runtimeUsageMu.Lock()
	w.runtimeUsageCache = status
	w.runtimeUsageCachedAt = now
	w.runtimeUsageMu.Unlock()
	return status, nil
}

func (w *Worker) invalidateRuntimeUsageCache() {
	w.runtimeUsageMu.Lock()
	w.runtimeUsageCache = domain.RecorderRuntimeStatus{}
	w.runtimeUsageCachedAt = time.Time{}
	w.runtimeUsageMu.Unlock()
}

func summarizeRecordingRuntime(ctx context.Context, runtimeDirectory string) (domain.RecorderRuntimeStatus, error) {
	result := domain.RecorderRuntimeStatus{Status: "ok"}
	stat, err := os.Stat(runtimeDirectory)
	if errors.Is(err, os.ErrNotExist) {
		applyRecordingRuntimeCapacity(&result, runtimeDirectory)
		return result, nil
	}
	if err != nil {
		return result, err
	}
	if !stat.IsDir() {
		applyRecordingRuntimeCapacity(&result, runtimeDirectory)
		return result, nil
	}

	err = filepath.WalkDir(runtimeDirectory, func(path string, entry os.DirEntry, walkErr error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if walkErr != nil {
			return walkErr
		}
		if path == runtimeDirectory {
			return nil
		}
		if entry.IsDir() {
			result.RecordingDirectories++
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		result.Files++
		result.UsedBytes += info.Size()
		return nil
	})
	if err != nil {
		return domain.RecorderRuntimeStatus{}, err
	}
	applyRecordingRuntimeCapacity(&result, runtimeDirectory)
	return result, nil
}

func summarizeRuntimeEntry(ctx context.Context, entryPath string) (domain.RecorderRuntimeStatus, error) {
	result := domain.RecorderRuntimeStatus{Status: "ok"}
	info, err := os.Stat(entryPath)
	if errors.Is(err, os.ErrNotExist) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	if !info.IsDir() {
		result.Files = 1
		result.UsedBytes = info.Size()
		return result, nil
	}
	result.RecordingDirectories = 1
	err = filepath.WalkDir(entryPath, func(path string, entry os.DirEntry, walkErr error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if walkErr != nil {
			return walkErr
		}
		if path == entryPath {
			return nil
		}
		if entry.IsDir() {
			result.RecordingDirectories++
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		result.Files++
		result.UsedBytes += info.Size()
		return nil
	})
	if err != nil {
		return domain.RecorderRuntimeStatus{}, err
	}
	return result, nil
}

func applyRecordingRuntimeCapacity(status *domain.RecorderRuntimeStatus, runtimeDirectory string) {
	statPath := runtimeDirectory
	if _, err := os.Stat(statPath); err != nil {
		statPath = filepath.Dir(runtimeDirectory)
	}
	var fileSystem syscall.Statfs_t
	if err := syscall.Statfs(statPath, &fileSystem); err != nil {
		return
	}
	blockSize := int64(fileSystem.Bsize)
	totalBytes := int64(fileSystem.Blocks) * blockSize
	availableBytes := int64(fileSystem.Bavail) * blockSize
	status.TotalBytes = totalBytes
	status.AvailableBytes = availableBytes
	if totalBytes > 0 {
		status.UsedPercent = (float64(status.UsedBytes) / float64(totalBytes)) * 100
	}
}
