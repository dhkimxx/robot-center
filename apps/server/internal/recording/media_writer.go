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
	"strings"

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

type MediaUploader interface {
	UploadMediaSnapshots(ctx context.Context, roomID string, chunk domain.RecordingChunk) error
}

type mediaSnapshotter interface {
	createH264Snapshot(roomID string, chunkID string, label string) (string, error)
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

func (w *Worker) setActiveRecordingChunk(roomID string, chunk domain.RecordingChunk) {
	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()
	w.activeChunks[roomID] = chunk
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
		if err := w.appendH264Payload(roomID, label, payload); err != nil {
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
		if err := w.appendOpusPacket(roomID, label, packet); err != nil {
			w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
				status.lastError = err.Error()
			})
			log.Printf("recorder-worker opus append failed room=%s label=%s: %v", roomID, label, err)
		}
	}
}

func (w *Worker) appendH264Payload(roomID string, label string, payload []byte) error {
	storageLabel := recordingStorageMediaLabel(label)
	if storageLabel != "rgb" && storageLabel != "thermal" {
		return nil
	}

	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()

	chunk, ok := w.activeChunks[roomID]
	if !ok {
		return nil
	}
	w.updateH264ParameterSetsLocked(roomID, storageLabel, payload)
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

func (w *Worker) appendOpusPacket(roomID string, label string, packet *rtp.Packet) error {
	if recordingStorageAudioLabel(label) != "audio" {
		return nil
	}

	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()

	chunk, ok := w.activeChunks[roomID]
	if !ok {
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
	snapshotPath, err := u.snapshotter.createH264Snapshot(roomID, chunk.ID, label)
	if err != nil {
		return err
	}
	if snapshotPath == "" {
		return nil
	}
	defer os.Remove(snapshotPath)

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
	if err := muxH264ToMP4(ctx, snapshotPath, audioSnapshotPath, outputPath); err != nil {
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

func (w *Worker) createH264Snapshot(roomID string, chunkID string, label string) (string, error) {
	sourcePath := h264TrackPath(chunkID, label)
	snapshotPath := filepath.Join(recordingChunkDirectory(chunkID), label+"_snapshot.h264")

	w.mediaMu.Lock()
	defer w.mediaMu.Unlock()

	stat, err := os.Stat(sourcePath)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if stat.Size() < minH264SnapshotBytes {
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
	parameterSets := w.h264ParameterSets[h264ParameterSetKey(roomID, label)]
	if len(parameterSets.sps) > 0 {
		if _, err := snapshot.Write(parameterSets.sps); err != nil {
			_ = snapshot.Close()
			return "", err
		}
	}
	if len(parameterSets.pps) > 0 {
		if _, err := snapshot.Write(parameterSets.pps); err != nil {
			_ = snapshot.Close()
			return "", err
		}
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

func muxH264ToMP4(ctx context.Context, inputPath string, audioPath string, outputPath string) error {
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
		"-r", "30",
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
