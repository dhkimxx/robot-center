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

func TestRecordingStorageMediaLabelMapsCanonicalVideoSlots(t *testing.T) {
	cases := map[string]string{
		"track.video_1": "rgb",
		"track.video_2": "thermal",
		"rgb":           "rgb",
		"thermal":       "thermal",
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
		"audio":         "audio",
		"track.audio_2": "",
	}
	for input, want := range cases {
		if got := recordingStorageAudioLabel(input); got != want {
			t.Fatalf("recordingStorageAudioLabel(%q) = %q, want %q", input, got, want)
		}
	}
}
