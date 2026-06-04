package api

import (
	"testing"

	"robot-center/apps/server/internal/config"
	"robot-center/apps/server/internal/domain"
)

func TestCreateStorageObjectURLUsesPublicMinIOURL(t *testing.T) {
	server := &Server{config: config.AppServerConfig{
		AppServerPublicURL: "http://192.168.20.12:18080",
		MinIOInternalURL:   "http://minio:9000",
		MinIOPublicURL:     "http://192.168.20.12:19000",
		MinIOBucket:        "robot-center",
	}}

	got := server.createStorageObjectURL("missions/mission-039/robots/robot-028/rgb video.mp4")
	want := "http://192.168.20.12:19000/robot-center/missions/mission-039/robots/robot-028/rgb%20video.mp4"
	if got != want {
		t.Fatalf("createStorageObjectURL() = %q, want %q", got, want)
	}
}

func TestCreateStorageObjectURLFallsBackToLegacyPublicHost(t *testing.T) {
	server := &Server{config: config.AppServerConfig{
		AppServerPublicURL: "http://center.local:18080",
		MinIOInternalURL:   "http://minio:9000",
		MinIOBucket:        "robot-center",
	}}

	got := server.createStorageObjectURL("missions/mission-001/rgb.mp4")
	want := "http://center.local:9000/robot-center/missions/mission-001/rgb.mp4"
	if got != want {
		t.Fatalf("createStorageObjectURL() = %q, want %q", got, want)
	}
}

func TestCreateRecordingResponseMarksVideoLessUploadedChunkPartial(t *testing.T) {
	server := &Server{}
	response := server.createOperatorRecordingResponse(domain.RecordingChunk{
		Status:            "uploaded",
		ManifestObjectKey: "missions/mission-039/manifest.json",
		MediaObjectKeys: map[string]string{
			"rgbMp4":    "missions/mission-039/rgb.mp4",
			"thermal":   "missions/mission-039/thermal.mp4",
			"telemetry": "missions/mission-039/telemetry.jsonl",
		},
		AvailableFileTypes: map[string]bool{
			"manifest":        true,
			"telemetry_jsonl": true,
		},
	})

	if response.Status != "partial" {
		t.Fatalf("expected video-less uploaded chunk to be partial, got %q", response.Status)
	}
}

func TestCreateRecordingResponseKeepsUploadedStatusWhenVideoExists(t *testing.T) {
	server := &Server{}
	response := server.createOperatorRecordingResponse(domain.RecordingChunk{
		Status: "uploaded",
		MediaObjectKeys: map[string]string{
			"rgbMp4":  "missions/mission-039/rgb.mp4",
			"thermal": "missions/mission-039/thermal.mp4",
		},
		AvailableFileTypes: map[string]bool{
			"rgb_audio_mp4": true,
		},
	})

	if response.Status != "uploaded" {
		t.Fatalf("expected uploaded chunk with video to stay uploaded, got %q", response.Status)
	}
}
