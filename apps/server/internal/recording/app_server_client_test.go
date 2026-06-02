package recording

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/domain"
)

func TestHTTPAppServerClientUsesRecorderDTOs(t *testing.T) {
	now := time.Date(2026, 6, 2, 9, 0, 0, 0, time.UTC)
	lockedUntil := now.Add(2 * time.Minute)
	chunk := domain.RecordingChunk{
		ID:                 "chunk-001",
		RecordingSessionID: "session-001",
		MissionID:          "mission-id",
		MissionCode:        "mission-001",
		RobotCode:          "robot-001",
		ChunkIndex:         3,
		Status:             "recording",
		StartedAt:          now,
		EndedAt:            now.Add(time.Minute),
		DurationSeconds:    60,
		ManifestObjectKey:  "missions/mission-001/manifest.json",
		MediaObjectKeys:    map[string]string{"rgbMp4": "missions/mission-001/rgb.mp4"},
		AvailableFileTypes: map[string]bool{"manifest": true},
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	job := domain.RecordingFinalizationJob{
		ID:                 "job-001",
		RecordingChunkID:   chunk.ID,
		RecordingSessionID: chunk.RecordingSessionID,
		MissionID:          chunk.MissionID,
		RobotID:            "robot-id",
		Status:             "claimed",
		Attempts:           3,
		LockedBy:           "worker-001",
		LockedUntil:        &lockedUntil,
		CreatedAt:          now,
		UpdatedAt:          now,
		Chunk:              chunk,
	}
	mission := domain.Mission{
		ID:          "mission-id",
		MissionCode: "mission-001",
		Name:        "Mission 1",
		MissionType: "mountain_rescue",
		Status:      "active",
		RobotCode:   "robot-001",
		RobotCodes:  []string{"robot-001"},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	observedStatusRequest := false
	observedUploadRequest := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "GET /api/v1/recorder/recording-targets":
			writeClientTestJSON(t, w, dto.RecordingTargetsPayload([]domain.Mission{mission}))
		case "POST /api/v1/recorder/tick":
			request := decodeClientTestJSON[dto.RecorderTickRequest](t, r)
			if request.MissionCode != mission.MissionCode || request.RobotCode != mission.RobotCode || request.ChunkDurationSeconds != 60 {
				t.Fatalf("unexpected tick request: %#v", request)
			}
			writeClientTestJSON(t, w, dto.RecordingTick(domain.RecordingTickResult{
				Chunk:    chunk,
				Manifest: map[string]any{"chunkId": chunk.ID},
			}))
		case "POST /api/v1/recorder/finalization-jobs/claim":
			request := decodeClientTestJSON[dto.RecorderFinalizationClaimRequest](t, r)
			if request.WorkerID != "worker-001" || request.Limit != 8 || request.LockDurationSeconds != 120 {
				t.Fatalf("unexpected finalization claim request: %#v", request)
			}
			writeClientTestJSON(t, w, dto.RecorderFinalizationJobsPayload([]domain.RecordingFinalizationJob{job}))
		case "POST /api/v1/recorder/finalization-jobs/job-001/partial":
			request := decodeClientTestJSON[dto.RecorderFinalizationStatusRequest](t, r)
			if request.WorkerID != "worker-001" || request.Attempt != 3 || request.Reason != "missing media" {
				t.Fatalf("unexpected finalization status request: %#v", request)
			}
			observedStatusRequest = true
			writeClientTestJSON(t, w, dto.OKPayload())
		case "POST /api/v1/recorder/chunks/chunk-001/uploaded":
			request := decodeClientTestJSON[dto.RecorderUploadRequest](t, r)
			if request.SizeBytes == nil || *request.SizeBytes != 2048 || request.WorkerID != "worker-001" || request.Attempt != 3 {
				t.Fatalf("unexpected upload request: %#v", request)
			}
			observedUploadRequest = true
			writeClientTestJSON(t, w, dto.OKPayload())
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewHTTPAppServerClient(server.URL, server.Client())
	targets, err := client.FetchRecordingTargets(context.Background())
	if err != nil {
		t.Fatalf("fetch recording targets: %v", err)
	}
	if len(targets) != 1 || targets[0].MissionCode != mission.MissionCode || targets[0].RobotCodes[0] != mission.RobotCode {
		t.Fatalf("unexpected targets: %#v", targets)
	}

	tick, err := client.CreateRecordingTick(context.Background(), mission, time.Minute, now)
	if err != nil {
		t.Fatalf("create recording tick: %v", err)
	}
	if tick.Chunk.ID != chunk.ID || tick.Chunk.MediaObjectKeys["rgbMp4"] != "missions/mission-001/rgb.mp4" || tick.Manifest["chunkId"] != chunk.ID {
		t.Fatalf("unexpected tick result: %#v", tick)
	}

	jobs, err := client.ClaimRecordingFinalizationJobs(context.Background(), "worker-001", 8, 2*time.Minute)
	if err != nil {
		t.Fatalf("claim finalization jobs: %v", err)
	}
	if len(jobs) != 1 || jobs[0].ID != job.ID || jobs[0].Chunk.ID != chunk.ID || jobs[0].Chunk.AvailableFileTypes["manifest"] != true {
		t.Fatalf("unexpected finalization jobs: %#v", jobs)
	}

	uploadContext := RecordingUploadContext{WorkerID: "worker-001", Attempt: 3}
	if err := client.MarkRecordingFinalizationJobPartial(context.Background(), job.ID, uploadContext, "missing media"); err != nil {
		t.Fatalf("mark finalization partial: %v", err)
	}
	if err := client.MarkRecordingChunkUploaded(context.Background(), chunk.ID, 2048, uploadContext); err != nil {
		t.Fatalf("mark chunk uploaded: %v", err)
	}
	if !observedStatusRequest || !observedUploadRequest {
		t.Fatalf("expected status and upload DTO requests to be observed")
	}
}

func decodeClientTestJSON[T any](t *testing.T, r *http.Request) T {
	t.Helper()
	var payload T
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	return payload
}

func writeClientTestJSON(t *testing.T, w http.ResponseWriter, payload any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Fatalf("write response body: %v", err)
	}
}
