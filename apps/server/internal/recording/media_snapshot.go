package recording

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"robot-center/apps/server/internal/utils"
	"strconv"
	"strings"
)

func (w *Worker) createH264Snapshot(roomID string, chunkID string, label string) (h264Snapshot, error) {
	sourcePath := h264TrackPath(chunkID, label)
	snapshotPath := filepath.Join(recordingChunkDirectory(chunkID), label+"_snapshot.h264")

	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()

	stat, err := os.Stat(sourcePath)
	if errors.Is(err, os.ErrNotExist) {
		return h264Snapshot{}, nil
	}
	if err != nil {
		return h264Snapshot{}, err
	}
	if stat.Size() < minH264SnapshotBytes {
		return h264Snapshot{}, nil
	}

	source, err := os.Open(sourcePath)
	if err != nil {
		return h264Snapshot{}, err
	}
	defer source.Close()

	snapshot, err := os.Create(snapshotPath)
	if err != nil {
		return h264Snapshot{}, err
	}
	parameterSets := w.h264ParameterSets[h264ParameterSetKey(roomID, label)]
	observedFPS := w.observedH264FPSLocked(chunkID, label)
	if len(parameterSets.sps) > 0 {
		if _, err := snapshot.Write(parameterSets.sps); err != nil {
			_ = snapshot.Close()
			return h264Snapshot{}, err
		}
	}
	if len(parameterSets.pps) > 0 {
		if _, err := snapshot.Write(parameterSets.pps); err != nil {
			_ = snapshot.Close()
			return h264Snapshot{}, err
		}
	}
	if _, err := io.Copy(snapshot, source); err != nil {
		_ = snapshot.Close()
		return h264Snapshot{}, err
	}
	if err := snapshot.Close(); err != nil {
		return h264Snapshot{}, err
	}
	return h264Snapshot{path: snapshotPath, fps: observedFPS}, nil
}

func (w *Worker) createDataChannelSnapshot(chunkID string, label string) (string, error) {
	sourcePath := dataChannelPath(chunkID, label)
	snapshotPath := filepath.Join(recordingChunkDirectory(chunkID), label+"_snapshot.jsonl")

	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()

	stat, err := os.Stat(sourcePath)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if stat.Size() < minDataChannelSnapshotBytes {
		return "", nil
	}

	source, err := os.Open(sourcePath)
	if err != nil {
		return "", err
	}
	defer source.Close()

	snapshot, err := os.Create(snapshotPath)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(snapshot, source); err != nil {
		_ = snapshot.Close()
		return "", err
	}
	if err := snapshot.Close(); err != nil {
		return "", err
	}
	return snapshotPath, nil
}

func (w *Worker) createOggSnapshot(chunkID string) (string, error) {
	sourcePath := opusTrackPath(chunkID)
	snapshotPath := filepath.Join(recordingChunkDirectory(chunkID), "audio_snapshot.ogg")

	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()

	stat, err := os.Stat(sourcePath)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if stat.Size() < minOggSnapshotBytes {
		return "", nil
	}

	source, err := os.Open(sourcePath)
	if err != nil {
		return "", err
	}
	defer source.Close()

	snapshot, err := os.Create(snapshotPath)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(snapshot, source); err != nil {
		_ = snapshot.Close()
		return "", err
	}
	if err := snapshot.Close(); err != nil {
		return "", err
	}
	return snapshotPath, nil
}

func muxH264ToMP4(ctx context.Context, inputPath string, audioPath string, outputPath string, inputFPS float64) error {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return err
	}
	temporaryOutputPath := outputPath + ".tmp.mp4"
	_ = os.Remove(temporaryOutputPath)
	args := []string{
		"-y",
		"-hide_banner",
		"-loglevel", "error",
		"-fflags", "+genpts",
		"-r", formatH264InputFPS(inputFPS),
		"-i", inputPath,
	}
	if audioPath != "" {
		args = append(args, "-i", audioPath, "-map", "0:v:0", "-map", "1:a:0", "-c:v", "copy", "-c:a", "copy")
	} else {
		args = append(args, "-c:v", "copy")
	}
	args = append(args,
		"-movflags", "+faststart",
		temporaryOutputPath,
	)
	command := exec.CommandContext(ctx, ffmpegPath, args...)
	output, err := command.CombinedOutput()
	if err != nil {
		_ = os.Remove(temporaryOutputPath)
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	stat, err := os.Stat(temporaryOutputPath)
	if err != nil {
		return err
	}
	if stat.Size() == 0 {
		_ = os.Remove(temporaryOutputPath)
		return fmt.Errorf("empty mp4 output")
	}
	return os.Rename(temporaryOutputPath, outputPath)
}

func recordingChunkDirectory(chunkID string) string {
	return filepath.Join(".runtime", "recordings", utils.SafePathToken(chunkID))
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

func h264ParameterSetKey(roomID string, label string) string {
	return utils.SafePathToken(roomID) + "/" + utils.SafePathToken(label)
}

func h264TrackTimingKey(chunkID string, label string) string {
	return utils.SafePathToken(chunkID) + "/" + utils.SafePathToken(label)
}

func (w *Worker) updateH264ParameterSetsLocked(roomID string, label string, payload []byte) {
	parameterSets := w.h264ParameterSets[h264ParameterSetKey(roomID, label)]
	changed := false
	for _, nalu := range splitAnnexBNALUs(payload) {
		naluType, ok := h264NALUType(nalu)
		if !ok {
			continue
		}
		switch naluType {
		case 7:
			parameterSets.sps = append([]byte(nil), nalu...)
			changed = true
		case 8:
			parameterSets.pps = append([]byte(nil), nalu...)
			changed = true
		}
	}
	if changed {
		w.h264ParameterSets[h264ParameterSetKey(roomID, label)] = parameterSets
	}
}

func (w *Worker) updateH264TrackTimingLocked(chunkID string, label string, timestamp uint32) {
	timing := w.h264Timings[h264TrackTimingKey(chunkID, label)]
	if !timing.haveTimestamp {
		timing.haveTimestamp = true
		timing.firstTimestamp = timestamp
		timing.lastTimestamp = timestamp
		timing.frameCount = 1
		w.h264Timings[h264TrackTimingKey(chunkID, label)] = timing
		return
	}
	if timestamp == timing.lastTimestamp {
		return
	}
	timing.lastTimestamp = timestamp
	timing.frameCount++
	w.h264Timings[h264TrackTimingKey(chunkID, label)] = timing
}

func (w *Worker) observedH264FPSLocked(chunkID string, label string) float64 {
	return w.h264Timings[h264TrackTimingKey(chunkID, label)].observedFPS()
}

func (t h264TrackTiming) observedFPS() float64 {
	if !t.haveTimestamp || t.frameCount < 2 {
		return 0
	}
	elapsedSeconds := float64(uint32(t.lastTimestamp-t.firstTimestamp)) / 90000
	if elapsedSeconds <= 0 {
		return 0
	}
	return float64(t.frameCount-1) / elapsedSeconds
}

func formatH264InputFPS(fps float64) string {
	if fps < 1 || fps > 120 {
		fps = 30
	}
	return strconv.FormatFloat(fps, 'f', 3, 64)
}

func splitAnnexBNALUs(payload []byte) [][]byte {
	var nalus [][]byte
	for cursor := 0; cursor < len(payload); {
		start, startCodeLength := findAnnexBStartCode(payload, cursor)
		if start < 0 {
			break
		}
		next, _ := findAnnexBStartCode(payload, start+startCodeLength)
		if next < 0 {
			next = len(payload)
		}
		nalu := payload[start:next]
		if len(nalu) > startCodeLength {
			nalus = append(nalus, nalu)
		}
		cursor = next
	}
	return nalus
}

func findAnnexBStartCode(payload []byte, offset int) (int, int) {
	for index := offset; index+3 <= len(payload); index++ {
		if index+4 <= len(payload) && payload[index] == 0x00 && payload[index+1] == 0x00 && payload[index+2] == 0x00 && payload[index+3] == 0x01 {
			return index, 4
		}
		if payload[index] == 0x00 && payload[index+1] == 0x00 && payload[index+2] == 0x01 {
			return index, 3
		}
	}
	return -1, 0
}

func h264NALUType(nalu []byte) (byte, bool) {
	_, startCodeLength := findAnnexBStartCode(nalu, 0)
	if startCodeLength == 0 || len(nalu) <= startCodeLength {
		return 0, false
	}
	return nalu[startCodeLength] & 0x1f, true
}
