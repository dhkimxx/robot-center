package recording

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"robot-center/apps/server/internal/utils"
	"strconv"
	"strings"
	"time"
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
	observedDurationSeconds := w.observedH264DurationSecondsLocked(chunkID, label)
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
	return h264Snapshot{path: snapshotPath, fps: observedFPS, durationSeconds: observedDurationSeconds}, nil
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

func muxH264ToMP4(ctx context.Context, inputPath string, audioPath string, outputPath string, inputFPS float64, expectedVideoDurationSeconds float64) error {
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
	if err := validateMP4VideoStream(ctx, temporaryOutputPath, expectedVideoDurationSeconds); err != nil {
		_ = os.Remove(temporaryOutputPath)
		return err
	}
	return os.Rename(temporaryOutputPath, outputPath)
}

func validateMP4VideoStream(ctx context.Context, outputPath string, expectedVideoDurationSeconds float64) error {
	ffprobePath, err := exec.LookPath("ffprobe")
	if err != nil {
		return err
	}
	command := exec.CommandContext(
		ctx,
		ffprobePath,
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=codec_name,width,height,duration",
		"-of", "json",
		outputPath,
	)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		return fmt.Errorf("mp4 ffprobe failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	var probe struct {
		Streams []struct {
			CodecName string `json:"codec_name"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
			Duration  string `json:"duration"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(output, &probe); err != nil {
		return fmt.Errorf("mp4 ffprobe parse failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	if len(probe.Streams) == 0 {
		return fmt.Errorf("mp4 output has no video stream")
	}
	stream := probe.Streams[0]
	if !strings.EqualFold(stream.CodecName, "h264") {
		return fmt.Errorf("mp4 video codec = %q, want h264", stream.CodecName)
	}
	if stream.Width <= 0 || stream.Height <= 0 {
		return fmt.Errorf("mp4 video dimensions are invalid: %dx%d", stream.Width, stream.Height)
	}
	if expectedVideoDurationSeconds >= 10 {
		actualDurationSeconds, err := strconv.ParseFloat(stream.Duration, 64)
		if err != nil {
			return fmt.Errorf("mp4 video duration missing")
		}
		minDurationSeconds := expectedVideoDurationSeconds * 0.80
		if actualDurationSeconds < minDurationSeconds {
			return fmt.Errorf("mp4 video duration %.3fs is shorter than expected %.3fs", actualDurationSeconds, expectedVideoDurationSeconds)
		}
	}
	return nil
}

func h264ParameterSetKey(roomID string, label string) string {
	return utils.SafePathToken(roomID) + "/" + utils.SafePathToken(label)
}

func h264ChunkKeyframeWaitKey(roomID string, chunkID string, label string) string {
	return utils.SafePathToken(roomID) + "/" + utils.SafePathToken(chunkID) + "/" + utils.SafePathToken(label)
}

func h264KeyframeRequestKey(roomID string, label string) string {
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

func (w *Worker) updateH264TrackTimingLocked(chunkID string, label string, timestamp uint32, observedAt time.Time) {
	timing := w.h264Timings[h264TrackTimingKey(chunkID, label)]
	if !timing.haveTimestamp {
		timing.haveTimestamp = true
		timing.lastTimestamp = timestamp
		timing.frameCount = 1
		timing.firstObservedAt = observedAt
		timing.lastObservedAt = observedAt
		w.h264Timings[h264TrackTimingKey(chunkID, label)] = timing
		return
	}
	if timestamp == timing.lastTimestamp {
		return
	}
	delta := uint32(timestamp - timing.lastTimestamp)
	if delta <= maxContinuousH264RTPDelta {
		timing.accumulatedTimestampDelta += uint64(delta)
	}
	timing.lastTimestamp = timestamp
	timing.frameCount++
	if timing.firstObservedAt.IsZero() {
		timing.firstObservedAt = observedAt
	}
	timing.lastObservedAt = observedAt
	w.h264Timings[h264TrackTimingKey(chunkID, label)] = timing
}

func (w *Worker) observedH264FPSLocked(chunkID string, label string) float64 {
	return w.h264Timings[h264TrackTimingKey(chunkID, label)].observedFPS()
}

func (w *Worker) observedH264DurationSecondsLocked(chunkID string, label string) float64 {
	return w.h264Timings[h264TrackTimingKey(chunkID, label)].observedDurationSeconds()
}

func (t h264TrackTiming) observedFPS() float64 {
	if !t.haveTimestamp || t.frameCount < 2 {
		return 0
	}
	elapsedSeconds := t.observedDurationSeconds()
	if elapsedSeconds <= 0 {
		return 0
	}
	return float64(t.frameCount-1) / elapsedSeconds
}

func (t h264TrackTiming) observedDurationSeconds() float64 {
	if !t.haveTimestamp || t.frameCount < 2 {
		return 0
	}
	if !t.firstObservedAt.IsZero() && t.lastObservedAt.After(t.firstObservedAt) {
		return t.lastObservedAt.Sub(t.firstObservedAt).Seconds()
	}
	return float64(t.accumulatedTimestampDelta) / 90000
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

func h264PayloadContainsIDR(payload []byte) bool {
	for _, nalu := range splitAnnexBNALUs(payload) {
		naluType, ok := h264NALUType(nalu)
		if ok && naluType == 5 {
			return true
		}
	}
	return false
}

func h264PayloadContainsVCL(payload []byte) bool {
	for _, nalu := range splitAnnexBNALUs(payload) {
		naluType, ok := h264NALUType(nalu)
		if ok && naluType >= 1 && naluType <= 5 {
			return true
		}
	}
	return false
}

func removeH264AccessUnitDelimiters(payload []byte) []byte {
	nalus := splitAnnexBNALUs(payload)
	if len(nalus) == 0 {
		return payload
	}
	var filtered []byte
	for _, nalu := range nalus {
		naluType, ok := h264NALUType(nalu)
		if ok && naluType == 9 {
			continue
		}
		filtered = append(filtered, nalu...)
	}
	return filtered
}

func trimH264PayloadToRandomAccessPoint(payload []byte) []byte {
	nalus := splitAnnexBNALUs(payload)
	if len(nalus) == 0 {
		return payload
	}
	idrIndex := -1
	for index, nalu := range nalus {
		naluType, ok := h264NALUType(nalu)
		if ok && naluType == 5 {
			idrIndex = index
			break
		}
	}
	if idrIndex < 0 {
		return payload
	}

	startIndex := idrIndex
	for startIndex > 0 {
		naluType, ok := h264NALUType(nalus[startIndex-1])
		if !ok || !h264NALUCanPrefixRandomAccessPoint(naluType) {
			break
		}
		startIndex--
	}

	var trimmed []byte
	for _, nalu := range nalus[startIndex:] {
		trimmed = append(trimmed, nalu...)
	}
	if len(trimmed) == 0 {
		return payload
	}
	return trimmed
}

func h264NALUCanPrefixRandomAccessPoint(naluType byte) bool {
	switch naluType {
	case 6, 7, 8, 9:
		return true
	default:
		return false
	}
}
