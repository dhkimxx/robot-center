package recording

import (
	"testing"
	"time"

	"robot-center/apps/server/internal/config"
)

func TestSplitAnnexBNALUs(t *testing.T) {
	payload := []byte{
		0x00, 0x00, 0x00, 0x01, 0x67, 0x01, 0x02,
		0x00, 0x00, 0x01, 0x68, 0x03,
		0x00, 0x00, 0x00, 0x01, 0x65, 0x04,
	}

	nalus := splitAnnexBNALUs(payload)
	if len(nalus) != 3 {
		t.Fatalf("expected 3 NALUs, got %d", len(nalus))
	}

	expectedTypes := []byte{7, 8, 5}
	for index, expectedType := range expectedTypes {
		actualType, ok := h264NALUType(nalus[index])
		if !ok {
			t.Fatalf("expected NALU type at index %d", index)
		}
		if actualType != expectedType {
			t.Fatalf("expected NALU type %d at index %d, got %d", expectedType, index, actualType)
		}
	}
}

func TestUpdateH264ParameterSets(t *testing.T) {
	worker := NewWorker(config.RecorderWorkerConfig{})
	payload := []byte{
		0x00, 0x00, 0x00, 0x01, 0x67, 0x01,
		0x00, 0x00, 0x00, 0x01, 0x68, 0x02,
	}

	worker.updateH264ParameterSetsLocked("mission-001", "rgb", payload)
	parameterSets := worker.h264ParameterSets[h264ParameterSetKey("mission-001", "rgb")]

	if len(parameterSets.sps) == 0 {
		t.Fatal("expected SPS to be stored")
	}
	if len(parameterSets.pps) == 0 {
		t.Fatal("expected PPS to be stored")
	}
}

func TestH264TrackTimingObservedFPS(t *testing.T) {
	worker := NewWorker(config.RecorderWorkerConfig{})
	for index := 0; index < 16; index++ {
		worker.updateH264TrackTimingLocked("chunk-001", "rgb", uint32(index*6000), time.Time{})
	}

	got := worker.observedH264FPSLocked("chunk-001", "rgb")
	if got < 14.99 || got > 15.01 {
		t.Fatalf("observed FPS = %f, want 15", got)
	}
}

func TestH264TrackTimingIgnoresTimestampDiscontinuity(t *testing.T) {
	worker := NewWorker(config.RecorderWorkerConfig{})
	for _, timestamp := range []uint32{
		1000,
		7000,
		13000,
		3_000_000_000,
		3_000_006_000,
		3_000_012_000,
	} {
		worker.updateH264TrackTimingLocked("chunk-001", "thermal", timestamp, time.Time{})
	}

	durationSeconds := worker.observedH264DurationSecondsLocked("chunk-001", "thermal")
	if durationSeconds < 0.26 || durationSeconds > 0.27 {
		t.Fatalf("observed duration = %f, want about 0.267s", durationSeconds)
	}
}

func TestH264TrackTimingPrefersWallClockDuration(t *testing.T) {
	worker := NewWorker(config.RecorderWorkerConfig{})
	base := time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC)
	for index := 0; index < 16; index++ {
		worker.updateH264TrackTimingLocked("chunk-001", "rgb", uint32(index*3000), base.Add(time.Duration(index)*time.Second/15))
	}

	durationSeconds := worker.observedH264DurationSecondsLocked("chunk-001", "rgb")
	if durationSeconds < 0.99 || durationSeconds > 1.01 {
		t.Fatalf("observed duration = %f, want about 1s", durationSeconds)
	}
	if fps := worker.observedH264FPSLocked("chunk-001", "rgb"); fps < 14.99 || fps > 15.01 {
		t.Fatalf("observed FPS = %f, want 15", fps)
	}
}

func TestH264PayloadContainsIDR(t *testing.T) {
	nonIDRPayload := []byte{
		0x00, 0x00, 0x00, 0x01, 0x41, 0x01,
	}
	if h264PayloadContainsIDR(nonIDRPayload) {
		t.Fatal("expected non-IDR payload to return false")
	}

	idrPayload := []byte{
		0x00, 0x00, 0x00, 0x01, 0x09, 0xf0,
		0x00, 0x00, 0x00, 0x01, 0x65, 0x88,
	}
	if !h264PayloadContainsIDR(idrPayload) {
		t.Fatal("expected IDR payload to return true")
	}
}

func TestH264PayloadContainsVCL(t *testing.T) {
	parameterSetPayload := []byte{
		0x00, 0x00, 0x00, 0x01, 0x67, 0x11,
		0x00, 0x00, 0x00, 0x01, 0x68, 0x22,
	}
	if h264PayloadContainsVCL(parameterSetPayload) {
		t.Fatal("expected SPS/PPS-only payload to have no VCL")
	}

	videoPayload := []byte{
		0x00, 0x00, 0x00, 0x01, 0x41, 0x33,
	}
	if !h264PayloadContainsVCL(videoPayload) {
		t.Fatal("expected slice payload to have VCL")
	}
}

func TestRemoveH264AccessUnitDelimitersDropsAUDOnly(t *testing.T) {
	payload := []byte{
		0x00, 0x00, 0x00, 0x01, 0x09, 0xf0,
	}
	if got := removeH264AccessUnitDelimiters(payload); len(got) != 0 {
		t.Fatalf("AUD-only payload length = %d, want 0", len(got))
	}

	payload = []byte{
		0x00, 0x00, 0x00, 0x01, 0x09, 0xf0,
		0x00, 0x00, 0x00, 0x01, 0x65, 0x88,
	}
	trimmed := removeH264AccessUnitDelimiters(payload)
	nalus := splitAnnexBNALUs(trimmed)
	if len(nalus) != 1 {
		t.Fatalf("filtered NALUs = %d, want 1", len(nalus))
	}
	naluType, ok := h264NALUType(nalus[0])
	if !ok || naluType != 5 {
		t.Fatalf("remaining NALU type = %d/%t, want IDR", naluType, ok)
	}
}

func TestTrimH264PayloadToRandomAccessPointKeepsParameterSets(t *testing.T) {
	payload := []byte{
		0x00, 0x00, 0x00, 0x01, 0x41, 0x11,
		0x00, 0x00, 0x00, 0x01, 0x67, 0x22,
		0x00, 0x00, 0x00, 0x01, 0x68, 0x33,
		0x00, 0x00, 0x00, 0x01, 0x65, 0x44,
		0x00, 0x00, 0x00, 0x01, 0x41, 0x55,
	}

	trimmed := trimH264PayloadToRandomAccessPoint(payload)
	nalus := splitAnnexBNALUs(trimmed)
	if len(nalus) != 4 {
		t.Fatalf("trimmed NALUs = %d, want 4", len(nalus))
	}
	expectedTypes := []byte{7, 8, 5, 1}
	for index, expectedType := range expectedTypes {
		actualType, ok := h264NALUType(nalus[index])
		if !ok || actualType != expectedType {
			t.Fatalf("NALU type at %d = %d/%t, want %d", index, actualType, ok, expectedType)
		}
	}
}

func TestFormatH264InputFPSFallsBackForInvalidValues(t *testing.T) {
	if got := formatH264InputFPS(15); got != "15.000" {
		t.Fatalf("formatH264InputFPS(15) = %q, want 15.000", got)
	}
	if got := formatH264InputFPS(0); got != "30.000" {
		t.Fatalf("formatH264InputFPS(0) = %q, want 30.000", got)
	}
}

func TestRecordingStorageMediaLabelMapsCanonicalVideoSlots(t *testing.T) {
	cases := map[string]string{
		"track.video_1": "rgb",
		"track.video_2": "thermal",
		"rgb":           "",
		"thermal":       "",
		"video":         "",
	}
	for input, want := range cases {
		if got := recordingStorageMediaLabel(input); got != want {
			t.Fatalf("recordingStorageMediaLabel(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestRecordingStorageAudioLabelMapsPrimaryCanonicalAudio(t *testing.T) {
	cases := map[string]string{
		"track.audio_1": "audio",
		"audio":         "",
		"track.audio_2": "",
	}
	for input, want := range cases {
		if got := recordingStorageAudioLabel(input); got != want {
			t.Fatalf("recordingStorageAudioLabel(%q) = %q, want %q", input, got, want)
		}
	}
}
