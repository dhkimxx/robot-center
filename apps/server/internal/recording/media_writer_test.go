package recording

import (
	"testing"

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
		worker.updateH264TrackTimingLocked("chunk-001", "rgb", uint32(index*6000))
	}

	got := worker.observedH264FPSLocked("chunk-001", "rgb")
	if got < 14.99 || got > 15.01 {
		t.Fatalf("observed FPS = %f, want 15", got)
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
