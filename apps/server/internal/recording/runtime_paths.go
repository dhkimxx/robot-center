package recording

import (
	"os"
	"path/filepath"
	"strings"

	"robot-center/apps/server/internal/utils"
)

func recordingRuntimeDirectory() string {
	if path := strings.TrimSpace(os.Getenv("RECORDER_RUNTIME_DIR")); path != "" {
		return path
	}
	return filepath.Join(".runtime", "recordings")
}

func recordingChunkDirectory(chunkID string) string {
	return filepath.Join(recordingRuntimeDirectory(), utils.SafePathToken(chunkID))
}

func h264TrackPath(chunkID string, label string) string {
	return filepath.Join(recordingChunkDirectory(chunkID), utils.SafePathToken(label)+".h264")
}

func opusTrackPath(chunkID string) string {
	return filepath.Join(recordingChunkDirectory(chunkID), "audio.ogg")
}

func dataChannelPath(chunkID string, label string) string {
	return filepath.Join(recordingChunkDirectory(chunkID), utils.SafePathToken(label)+".jsonl")
}
