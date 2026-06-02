package recording

import (
	"context"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media/oggwriter"
	"log"
	"os"
	"strings"
	"time"
)

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
		w.markRecorderRobotTrackActivity(roomID, label, time.Now().UTC())
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
		w.markRecorderRobotTrackActivity(roomID, label, time.Now().UTC())
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
		return ""
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
	case "track.audio_1":
		return "audio"
	case "track.audio_2":
		return ""
	default:
		return ""
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
