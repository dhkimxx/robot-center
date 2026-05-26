package recording

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media/oggwriter"

	"robot-center/apps/server/internal/domain"
)

const minH264SnapshotBytes int64 = 64 * 1024
const minOggSnapshotBytes int64 = 4 * 1024
const minDataChannelSnapshotBytes int64 = 1

type activeAudioWriter struct {
	chunkID string
	path    string
	writer  *oggwriter.OggWriter
}

type h264ParameterSets struct {
	sps []byte
	pps []byte
}

type h264TrackTiming struct {
	haveTimestamp  bool
	firstTimestamp uint32
	lastTimestamp  uint32
	frameCount     int
}

type h264Snapshot struct {
	path string
	fps  float64
}

type recordingChunkFinalization struct {
	mediaKey string
	chunk    domain.RecordingChunk
}

type MediaUploader interface {
	UploadMediaSnapshots(ctx context.Context, roomID string, chunk domain.RecordingChunk) error
}

type mediaSnapshotter interface {
	createH264Snapshot(roomID string, chunkID string, label string) (h264Snapshot, error)
	createOggSnapshot(chunkID string) (string, error)
	createDataChannelSnapshot(chunkID string, label string) (string, error)
}

type recordingMediaUploader struct {
	appServerClient AppServerClient
	objectStorage   ObjectStorage
	snapshotter     mediaSnapshotter
}

func NewMediaUploader(appServerClient AppServerClient, objectStorage ObjectStorage, snapshotter mediaSnapshotter) MediaUploader {
	return &recordingMediaUploader{
		appServerClient: appServerClient,
		objectStorage:   objectStorage,
		snapshotter:     snapshotter,
	}
}

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
	return safePathToken(mediaKey) + "/" + safePathToken(chunkID)
}

func (w *Worker) recordH264Track(ctx context.Context, roomID string, label string, track *webrtc.TrackRemote) {
	depacketizer := &codecs.H264Packet{}
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		packet, _, err := track.ReadRTP()
		if err != nil {
			return
		}
		payload, err := depacketizer.Unmarshal(packet.Payload)
		if err != nil {
			w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
				status.lastError = err.Error()
			})
			continue
		}
		if len(payload) == 0 {
			continue
		}
		if err := w.appendH264Payload(ctx, roomID, label, payload, packet.Timestamp); err != nil {
			w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
				status.lastError = err.Error()
			})
			log.Printf("recorder-worker h264 append failed room=%s label=%s: %v", roomID, label, err)
		}
	}
}

func (w *Worker) recordOpusTrack(ctx context.Context, roomID string, label string, track *webrtc.TrackRemote) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		packet, _, err := track.ReadRTP()
		if err != nil {
			return
		}
		if err := w.appendOpusPacket(ctx, roomID, label, packet); err != nil {
			w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
				status.lastError = err.Error()
			})
			log.Printf("recorder-worker opus append failed room=%s label=%s: %v", roomID, label, err)
		}
	}
}

func (w *Worker) appendH264Payload(ctx context.Context, roomID string, label string, payload []byte, rtpTimestamp uint32) error {
	storageLabel := recordingStorageMediaLabel(label)
	if storageLabel != "rgb" && storageLabel != "thermal" {
		return nil
	}

	chunk, ok, err := w.ensureActiveRecordingChunk(ctx, roomID, time.Now().UTC())
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()

	activeChunk, ok := w.activeChunks[roomID]
	if !ok || activeChunk.ID != chunk.ID {
		return nil
	}
	w.updateH264ParameterSetsLocked(roomID, storageLabel, payload)
	w.updateH264TrackTimingLocked(chunk.ID, storageLabel, rtpTimestamp)
	directory := recordingChunkDirectory(chunk.ID)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(h264TrackPath(chunk.ID, storageLabel), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(payload)
	return err
}

func recordingStorageMediaLabel(label string) string {
	switch strings.TrimSpace(label) {
	case "track.video_1":
		return "rgb"
	case "track.video_2":
		return "thermal"
	default:
		return strings.TrimSpace(label)
	}
}

func (w *Worker) appendOpusPacket(ctx context.Context, roomID string, label string, packet *rtp.Packet) error {
	if recordingStorageAudioLabel(label) != "audio" {
		return nil
	}

	chunk, ok, err := w.ensureActiveRecordingChunk(ctx, roomID, time.Now().UTC())
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()

	activeChunk, ok := w.activeChunks[roomID]
	if !ok || activeChunk.ID != chunk.ID {
		return nil
	}
	writer, err := w.getAudioWriterLocked(roomID, chunk.ID)
	if err != nil {
		return err
	}
	return writer.WriteRTP(packet)
}

func recordingStorageAudioLabel(label string) string {
	switch strings.TrimSpace(label) {
	case "track.audio_1", "audio":
		return "audio"
	case "track.audio_2":
		return ""
	default:
		return strings.TrimSpace(label)
	}
}

func (w *Worker) appendDataChannelPayload(roomID string, label string, payload []byte) error {
	if label != "sensor" && label != "telemetry" {
		return nil
	}

	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()

	chunk, ok := w.activeChunks[roomID]
	if !ok {
		return nil
	}
	directory := recordingChunkDirectory(chunk.ID)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(dataChannelPath(chunk.ID, label), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(payload); err != nil {
		return err
	}
	if len(payload) == 0 || payload[len(payload)-1] != '\n' {
		_, err = file.Write([]byte("\n"))
	}
	return err
}

func (w *Worker) getAudioWriterLocked(roomID string, chunkID string) (*oggwriter.OggWriter, error) {
	activeWriter := w.audioWriters[roomID]
	if activeWriter != nil && activeWriter.chunkID == chunkID {
		return activeWriter.writer, nil
	}
	if activeWriter != nil && activeWriter.writer != nil {
		_ = activeWriter.writer.Close()
		delete(w.audioWriters, roomID)
	}

	directory := recordingChunkDirectory(chunkID)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, err
	}
	path := opusTrackPath(chunkID)
	writer, err := oggwriter.New(path, 48000, 2)
	if err != nil {
		return nil, err
	}
	w.audioWriters[roomID] = &activeAudioWriter{
		chunkID: chunkID,
		path:    path,
		writer:  writer,
	}
	return writer, nil
}

func (w *Worker) closeAudioWriter(roomID string) {
	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()
	activeWriter := w.audioWriters[roomID]
	if activeWriter == nil || activeWriter.writer == nil {
		return
	}
	_ = activeWriter.writer.Close()
	delete(w.audioWriters, roomID)
}

func (w *Worker) closeAudioWriterForChunk(roomID string, chunkID string) {
	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()
	activeWriter := w.audioWriters[roomID]
	if activeWriter == nil || activeWriter.writer == nil || activeWriter.chunkID != chunkID {
		return
	}
	_ = activeWriter.writer.Close()
	delete(w.audioWriters, roomID)
}

func (u *recordingMediaUploader) UploadMediaSnapshots(ctx context.Context, roomID string, chunk domain.RecordingChunk) error {
	var uploadErrors []error
	if err := u.muxAndUploadH264Snapshot(ctx, roomID, chunk, "rgb", "rgb_audio_mp4", chunk.MediaObjectKeys["rgbMp4"]); err != nil {
		uploadErrors = append(uploadErrors, err)
	}
	if err := u.muxAndUploadH264Snapshot(ctx, roomID, chunk, "thermal", "thermal_mp4", chunk.MediaObjectKeys["thermal"]); err != nil {
		uploadErrors = append(uploadErrors, err)
	}
	if err := u.uploadDataChannelSnapshot(ctx, chunk, "sensor", "sensor_jsonl", chunk.MediaObjectKeys["sensor"]); err != nil {
		uploadErrors = append(uploadErrors, err)
	}
	if err := u.uploadDataChannelSnapshot(ctx, chunk, "telemetry", "telemetry_jsonl", chunk.MediaObjectKeys["telemetry"]); err != nil {
		uploadErrors = append(uploadErrors, err)
	}
	return errors.Join(uploadErrors...)
}

func (u *recordingMediaUploader) muxAndUploadH264Snapshot(ctx context.Context, roomID string, chunk domain.RecordingChunk, label string, fileType string, objectKey string) error {
	if strings.TrimSpace(objectKey) == "" {
		return nil
	}
	snapshot, err := u.snapshotter.createH264Snapshot(roomID, chunk.ID, label)
	if err != nil {
		return err
	}
	if snapshot.path == "" {
		return nil
	}
	defer os.Remove(snapshot.path)

	audioSnapshotPath := ""
	if label == "rgb" {
		audioSnapshotPath, err = u.snapshotter.createOggSnapshot(chunk.ID)
		if err != nil {
			log.Printf("recorder-worker audio snapshot skipped chunk=%s: %v", chunk.ID, err)
		}
		if audioSnapshotPath != "" {
			defer os.Remove(audioSnapshotPath)
		}
	}
	outputPath := filepath.Join(recordingChunkDirectory(chunk.ID), label+".mp4")
	if err := muxH264ToMP4(ctx, snapshot.path, audioSnapshotPath, outputPath, snapshot.fps); err != nil {
		return fmt.Errorf("%s mp4 mux failed: %w", label, err)
	}
	sizeBytes, err := u.objectStorage.UploadFile(ctx, objectKey, outputPath, "video/mp4")
	if err != nil {
		return fmt.Errorf("%s mp4 upload failed: %w", label, err)
	}
	if err := u.appServerClient.MarkRecordingFileUploaded(ctx, chunk.ID, fileType, sizeBytes); err != nil {
		return fmt.Errorf("%s file status update failed: %w", label, err)
	}
	log.Printf("recorder-worker mp4 uploaded chunk=%s label=%s key=%s", chunk.ID, label, objectKey)
	return nil
}

func (u *recordingMediaUploader) uploadDataChannelSnapshot(ctx context.Context, chunk domain.RecordingChunk, label string, fileType string, objectKey string) error {
	if strings.TrimSpace(objectKey) == "" {
		return nil
	}
	snapshotPath, err := u.snapshotter.createDataChannelSnapshot(chunk.ID, label)
	if err != nil {
		return err
	}
	if snapshotPath == "" {
		return nil
	}
	defer os.Remove(snapshotPath)

	sizeBytes, err := u.objectStorage.UploadFile(ctx, objectKey, snapshotPath, "application/x-ndjson")
	if err != nil {
		return fmt.Errorf("%s jsonl upload failed: %w", label, err)
	}
	if err := u.appServerClient.MarkRecordingFileUploaded(ctx, chunk.ID, fileType, sizeBytes); err != nil {
		return fmt.Errorf("%s file status update failed: %w", label, err)
	}
	log.Printf("recorder-worker jsonl uploaded chunk=%s label=%s key=%s", chunk.ID, label, objectKey)
	return nil
}

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
	return filepath.Join(".runtime", "recordings", safePathToken(chunkID))
}

func h264TrackPath(chunkID string, label string) string {
	return filepath.Join(recordingChunkDirectory(chunkID), safePathToken(label)+".h264")
}

func opusTrackPath(chunkID string) string {
	return filepath.Join(recordingChunkDirectory(chunkID), "audio.ogg")
}

func dataChannelPath(chunkID string, label string) string {
	return filepath.Join(recordingChunkDirectory(chunkID), safePathToken(label)+".jsonl")
}

func h264ParameterSetKey(roomID string, label string) string {
	return safePathToken(roomID) + "/" + safePathToken(label)
}

func h264TrackTimingKey(chunkID string, label string) string {
	return safePathToken(chunkID) + "/" + safePathToken(label)
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

func safePathToken(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "..", "_")
	return replacer.Replace(value)
}
