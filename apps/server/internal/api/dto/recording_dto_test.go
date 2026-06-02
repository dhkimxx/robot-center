package dto

import (
	"encoding/json"
	"testing"
	"time"

	"robot-center/apps/server/internal/domain"
)

func TestRecordingEnvelopeShape(t *testing.T) {
	now := time.Date(2026, 6, 2, 8, 0, 0, 0, time.UTC)
	chunk := RecordingChunkResponse{
		ID:          "chunk-001",
		MissionCode: "mission-001",
		RobotCode:   "robot-001",
		Status:      "recording",
		StartedAt:   now,
		EndedAt:     now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	assertRecordingJSONHasField(t, RecordingTargetsPayload([]domain.Mission{{MissionCode: "mission-001"}}), "targets")
	assertRecordingJSONHasField(t, RecordingsPayload([]RecordingChunkResponse{chunk}), "recordings")
	assertRecordingJSONHasField(t, RecordingChunkPayload(chunk), "chunk")
	assertRecordingJSONHasField(t, RecorderFinalizationJobsPayload([]domain.RecordingFinalizationJob{{ID: "job-001", Chunk: domain.RecordingChunk{ID: "chunk-001"}}}), "jobs")
	assertRecordingJSONHasField(t, OKPayload(), "ok")
}

func TestRecorderFinalizationJobPayloadShape(t *testing.T) {
	now := time.Date(2026, 6, 2, 8, 0, 0, 0, time.UTC)
	lockedUntil := now.Add(time.Minute)
	completedAt := now.Add(2 * time.Minute)

	payload := RecorderFinalizationJobsPayload([]domain.RecordingFinalizationJob{
		{
			ID:                 "job-001",
			RecordingChunkID:   "chunk-001",
			RecordingSessionID: "session-001",
			MissionID:          "mission-id",
			RobotID:            "robot-id",
			Status:             "claimed",
			Reason:             "retry",
			Attempts:           2,
			LockedBy:           "worker-001",
			LockedUntil:        &lockedUntil,
			LastError:          "upload failed",
			CreatedAt:          now,
			UpdatedAt:          now,
			CompletedAt:        &completedAt,
			Chunk: domain.RecordingChunk{
				ID:                 "chunk-001",
				RecordingSessionID: "session-001",
				MissionID:          "mission-id",
				MissionCode:        "mission-001",
				RobotCode:          "robot-001",
				Status:             "recording",
				CreatedAt:          now,
				UpdatedAt:          now,
			},
		},
	})

	fields := recordingJSONFields(t, payload)
	jobs, ok := fields["jobs"].([]any)
	if !ok || len(jobs) != 1 {
		t.Fatalf("expected one job, got %#v", fields)
	}
	job := jobs[0].(map[string]any)
	for _, field := range []string{
		"id",
		"recordingChunkId",
		"recordingSessionId",
		"missionId",
		"robotId",
		"status",
		"reason",
		"attempts",
		"lockedBy",
		"lockedUntil",
		"lastError",
		"createdAt",
		"updatedAt",
		"completedAt",
		"chunk",
	} {
		if _, ok := job[field]; !ok {
			t.Fatalf("expected finalization job field %q in %#v", field, job)
		}
	}
	chunk := job["chunk"].(map[string]any)
	if chunk["missionCode"] != "mission-001" || chunk["robotCode"] != "robot-001" {
		t.Fatalf("expected nested recording chunk DTO, got %#v", chunk)
	}
}

func TestRecorderRequestShape(t *testing.T) {
	now := time.Date(2026, 6, 2, 8, 0, 0, 0, time.UTC)
	sizeBytes := int64(1024)

	for _, field := range []string{"missionCode", "robotCode", "chunkDurationSeconds", "tickAt"} {
		assertRecordingJSONHasField(t, RecorderTickRequest{
			MissionCode:          "mission-001",
			RobotCode:            "robot-001",
			ChunkDurationSeconds: 60,
			TickAt:               now,
		}, field)
	}
	for _, field := range []string{"sizeBytes", "checksum", "workerId", "attempt"} {
		assertRecordingJSONHasField(t, RecorderUploadRequest{
			SizeBytes: &sizeBytes,
			Checksum:  "checksum",
			WorkerID:  "worker-001",
			Attempt:   1,
		}, field)
	}
	for _, field := range []string{"workerId", "limit", "lockDurationSeconds"} {
		assertRecordingJSONHasField(t, RecorderFinalizationClaimRequest{
			WorkerID:            "worker-001",
			Limit:               10,
			LockDurationSeconds: 30,
		}, field)
	}
	for _, field := range []string{"workerId", "attempt", "reason"} {
		assertRecordingJSONHasField(t, RecorderFinalizationStatusRequest{
			WorkerID: "worker-001",
			Attempt:  1,
			Reason:   "partial",
		}, field)
	}
}

func assertRecordingJSONHasField(t *testing.T, value any, field string) {
	t.Helper()
	fields := recordingJSONFields(t, value)
	if _, ok := fields[field]; !ok {
		t.Fatalf("expected field %q in JSON, got %#v", field, fields)
	}
}

func recordingJSONFields(t *testing.T, value any) map[string]any {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal value: %v", err)
	}
	var fields map[string]any
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatalf("unmarshal value: %v", err)
	}
	return fields
}
