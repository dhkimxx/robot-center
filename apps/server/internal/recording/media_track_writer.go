package recording

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media/oggwriter"
	"github.com/pion/webrtc/v4/pkg/media/samplebuilder"
)

func (w *Worker) recordH264Track(ctx context.Context, roomID string, label string, track *webrtc.TrackRemote, peerConnection *webrtc.PeerConnection) {
	builder := samplebuilder.New(200, &codecs.H264Packet{}, track.Codec().ClockRate, samplebuilder.WithMaxTimeDelay(2*time.Second))
	flushSamples := func() {
		for sample := builder.Pop(); sample != nil; sample = builder.Pop() {
			if len(sample.Data) == 0 {
				continue
			}
			if err := w.appendH264AccessUnit(ctx, roomID, label, sample.Data, sample.PacketTimestamp, time.Now().UTC(), track, peerConnection); err != nil {
				w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
					status.lastError = err.Error()
				})
				log.Printf("recorder-worker h264 append failed room=%s label=%s: %v", roomID, label, err)
			}
		}
	}
	flushRemainingSamples := func() {
		for sample := builder.Pop(); sample != nil; sample = builder.Pop() {
			if len(sample.Data) == 0 {
				continue
			}
			if err := w.appendH264AccessUnit(ctx, roomID, label, sample.Data, sample.PacketTimestamp, time.Now().UTC(), track, peerConnection); err != nil {
				w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
					status.lastError = err.Error()
				})
				log.Printf("recorder-worker h264 append failed room=%s label=%s: %v", roomID, label, err)
			}
		}
	}
	for {
		select {
		case <-ctx.Done():
			flushRemainingSamples()
			return
		default:
		}

		packet, _, err := track.ReadRTP()
		if err != nil {
			flushRemainingSamples()
			return
		}
		w.markRecorderRobotTrackActivity(roomID, label, time.Now().UTC())
		builder.Push(packet.Clone())
		flushSamples()
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

func (w *Worker) appendH264AccessUnit(ctx context.Context, roomID string, label string, payload []byte, rtpTimestamp uint32, observedAt time.Time, track *webrtc.TrackRemote, peerConnection *webrtc.PeerConnection) error {
	storageLabel := recordingStorageMediaLabel(label)
	if storageLabel != "rgb" && storageLabel != "thermal" {
		return nil
	}

	payload = removeH264AccessUnitDelimiters(payload)
	if len(payload) == 0 {
		return nil
	}
	payloadContainsVCL := h264PayloadContainsVCL(payload)
	payloadContainsIDR := h264PayloadContainsIDR(payload)
	if !payloadContainsVCL {
		w.mediaMu.Lock()
		w.updateH264ParameterSetsLocked(roomID, storageLabel, payload)
		w.mediaMu.Unlock()
		return nil
	}
	expiredChunk, isExpired := w.expiredActiveRecordingChunk(roomID, observedAt)
	if isExpired && !payloadContainsIDR {
		w.requestRecorderKeyFrame(roomID, storageLabel, expiredChunk.ID, track, peerConnection, observedAt)
	}
	chunk, ok, err := w.ensureActiveRecordingChunkForWrite(ctx, roomID, observedAt, payloadContainsIDR)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	w.mediaMu.Lock()

	activeChunk, ok := w.activeChunks[roomID]
	if !ok || activeChunk.ID != chunk.ID {
		w.mediaMu.Unlock()
		return nil
	}

	rolledOverOnKeyframe := isExpired && payloadContainsIDR && chunk.ID != expiredChunk.ID
	if rolledOverOnKeyframe {
		w.markH264ChunkTracksWaitingForKeyframeLocked(roomID, chunk.ID)
	}
	keyframeWaitKey := h264ChunkKeyframeWaitKey(roomID, chunk.ID, storageLabel)
	trackHasStarted := w.h264Timings[h264TrackTimingKey(chunk.ID, storageLabel)].haveTimestamp
	if !trackHasStarted || w.h264ChunkKeyframeWaits[keyframeWaitKey] {
		if !payloadContainsIDR {
			w.updateH264ParameterSetsLocked(roomID, storageLabel, payload)
			w.mediaMu.Unlock()
			w.requestRecorderKeyFrame(roomID, storageLabel, chunk.ID, track, peerConnection, observedAt)
			return nil
		}
		payload = trimH264PayloadToRandomAccessPoint(payload)
		delete(w.h264ChunkKeyframeWaits, keyframeWaitKey)
	}
	w.updateH264ParameterSetsLocked(roomID, storageLabel, payload)
	w.updateH264TrackTimingLocked(chunk.ID, storageLabel, rtpTimestamp, observedAt)
	directory := recordingChunkDirectory(chunk.ID)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		w.mediaMu.Unlock()
		return err
	}
	file, err := os.OpenFile(h264TrackPath(chunk.ID, storageLabel), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		w.mediaMu.Unlock()
		return err
	}
	defer file.Close()
	_, err = file.Write(payload)
	w.mediaMu.Unlock()
	return err
}

func (w *Worker) markH264ChunkTracksWaitingForKeyframeLocked(roomID string, chunkID string) {
	for _, label := range []string{"rgb", "thermal"} {
		w.h264ChunkKeyframeWaits[h264ChunkKeyframeWaitKey(roomID, chunkID, label)] = true
	}
}

func (w *Worker) requestRecorderKeyFrame(roomID string, label string, chunkID string, track *webrtc.TrackRemote, peerConnection *webrtc.PeerConnection, requestedAt time.Time) {
	if track == nil || peerConnection == nil || track.SSRC() == 0 {
		return
	}
	requestKey := h264KeyframeRequestKey(roomID, label)
	w.mediaMu.Lock()
	previousRequestAt := w.h264KeyframeRequests[requestKey]
	if !previousRequestAt.IsZero() && requestedAt.Sub(previousRequestAt) < time.Second {
		w.mediaMu.Unlock()
		return
	}
	w.h264KeyframeRequests[requestKey] = requestedAt
	w.mediaMu.Unlock()

	if err := peerConnection.WriteRTCP([]rtcp.Packet{
		&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())},
	}); err != nil {
		log.Printf("recorder-worker keyframe request failed room=%s label=%s chunk=%s: %v", roomID, label, chunkID, err)
		return
	}
	log.Printf("recorder-worker keyframe requested room=%s label=%s chunk=%s", roomID, label, chunkID)
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

	chunk, ok, err := w.ensureActiveRecordingChunkForWrite(ctx, roomID, time.Now().UTC(), false)
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
