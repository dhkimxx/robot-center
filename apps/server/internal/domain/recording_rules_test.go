package domain

import (
	"testing"
	"time"
)

func TestNewRecordingChunkWindow(t *testing.T) {
	base := time.Date(2026, 5, 23, 1, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		tickAt       time.Time
		wantIndex    int
		wantStarted  time.Time
		wantEnded    time.Time
		wantDuration int
	}{
		{
			name:         "first chunk",
			tickAt:       base.Add(3 * time.Minute),
			wantIndex:    0,
			wantStarted:  base,
			wantEnded:    base.Add(10 * time.Minute),
			wantDuration: 600,
		},
		{
			name:         "later chunk",
			tickAt:       base.Add(22 * time.Minute),
			wantIndex:    2,
			wantStarted:  base.Add(20 * time.Minute),
			wantEnded:    base.Add(30 * time.Minute),
			wantDuration: 600,
		},
		{
			name:         "negative elapsed clamps to first chunk",
			tickAt:       base.Add(-1 * time.Minute),
			wantIndex:    0,
			wantStarted:  base,
			wantEnded:    base.Add(10 * time.Minute),
			wantDuration: 600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			window := NewRecordingChunkWindow(base, tt.tickAt, 600)
			if window.Index != tt.wantIndex {
				t.Fatalf("Index = %d, want %d", window.Index, tt.wantIndex)
			}
			if !window.StartedAt.Equal(tt.wantStarted) {
				t.Fatalf("StartedAt = %s, want %s", window.StartedAt, tt.wantStarted)
			}
			if !window.EndedAt.Equal(tt.wantEnded) {
				t.Fatalf("EndedAt = %s, want %s", window.EndedAt, tt.wantEnded)
			}
			if window.DurationSeconds != tt.wantDuration {
				t.Fatalf("DurationSeconds = %d, want %d", window.DurationSeconds, tt.wantDuration)
			}
		})
	}
}

func TestNewRecordingObjectKeys(t *testing.T) {
	startedAt := time.Date(2026, 5, 23, 1, 0, 0, 0, time.UTC)
	endedAt := startedAt.Add(10 * time.Minute)

	keys := NewRecordingObjectKeys("mission-001", "robot-001", startedAt, endedAt)

	wantBase := "missions/mission-001/robots/robot-001/recordings/2026/05/23/20260523T010000Z_20260523T011000Z"
	expected := map[string]string{
		"manifest":  wantBase + "_manifest.json",
		"rgbMp4":    wantBase + "_rgb_h264_opus.mp4",
		"thermal":   wantBase + "_thermal_h264.mp4",
		"sensor":    wantBase + "_sensor.jsonl",
		"telemetry": wantBase + "_telemetry.jsonl",
	}
	for key, want := range expected {
		if keys[key] != want {
			t.Fatalf("keys[%q] = %q, want %q", key, keys[key], want)
		}
	}
}

func TestNewRecordingManifest(t *testing.T) {
	startedAt := time.Date(2026, 5, 23, 1, 0, 0, 0, time.UTC)
	endedAt := startedAt.Add(10 * time.Minute)
	chunk := RecordingChunk{
		ID:                 "rec-001",
		RecordingSessionID: "session-mission-001-robot-001",
		MissionID:          "mission-id-001",
		MissionCode:        "mission-001",
		RobotCode:          "robot-001",
		ChunkIndex:         3,
		Status:             "uploaded",
		StartedAt:          startedAt,
		EndedAt:            endedAt,
		MediaObjectKeys:    NewRecordingObjectKeys("mission-001", "robot-001", startedAt, endedAt),
		AvailableFileTypes: map[string]bool{"manifest": true, "rgbMp4": true},
	}

	manifest := NewRecordingManifest(chunk)

	if manifest["schemaVersion"] != "1.0" {
		t.Fatalf("schemaVersion = %v, want 1.0", manifest["schemaVersion"])
	}
	if manifest["chunkId"] != chunk.ID {
		t.Fatalf("chunkId = %v, want %s", manifest["chunkId"], chunk.ID)
	}
	if manifest["chunkIndex"] != chunk.ChunkIndex {
		t.Fatalf("chunkIndex = %v, want %d", manifest["chunkIndex"], chunk.ChunkIndex)
	}
	if manifest["startedAt"] != startedAt.Format(time.RFC3339Nano) {
		t.Fatalf("startedAt = %v", manifest["startedAt"])
	}
	codecPolicy, ok := manifest["codecPolicy"].(map[string]string)
	if !ok {
		t.Fatalf("codecPolicy has type %T", manifest["codecPolicy"])
	}
	if codecPolicy["video"] != "h264" || codecPolicy["audio"] != "opus" {
		t.Fatalf("codecPolicy = %#v", codecPolicy)
	}
}
