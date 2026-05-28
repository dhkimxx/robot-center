package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
	"robot-center/apps/server/internal/testsupport/postgrestest"
)

func TestRecordingServiceIntegrationUploadsFilesAndManifestTransactionally(t *testing.T) {
	services, dsn := newRecordingServiceIntegrationFixture(t)
	fixture := createRecordingServiceMissionFixture(t, services)
	ctx := context.Background()
	tickAt := time.Date(2026, 5, 28, 4, 0, 0, 0, time.UTC)

	result, err := services.Recording.ApplyRecordingTick(ctx, store.RecordingTickInput{
		MissionCode:          fixture.Mission.MissionCode,
		RobotCode:            fixture.Robot.RobotCode,
		ChunkDurationSeconds: 600,
		TickAt:               tickAt,
	})
	if err != nil {
		t.Fatalf("apply recording tick: %v", err)
	}

	rgbSize := int64(1024)
	fileChunk, err := services.Recording.MarkRecordingFileUploaded(ctx, result.Chunk.ID, "rgb_audio_mp4", store.RecordingUploadMetadata{
		SizeBytes: &rgbSize,
		Checksum:  "sha256:rgb",
	})
	if err != nil {
		t.Fatalf("mark recording file uploaded: %v", err)
	}
	if fileChunk.Status != "recording" || !fileChunk.AvailableFileTypes["rgb_audio_mp4"] {
		t.Fatalf("expected available rgb file without completed chunk, got %#v", fileChunk)
	}

	manifestSize := int64(2048)
	uploadedChunk, err := services.Recording.MarkRecordingChunkUploaded(ctx, result.Chunk.ID, store.RecordingUploadMetadata{
		SizeBytes: &manifestSize,
		Checksum:  "sha256:manifest",
	})
	if err != nil {
		t.Fatalf("mark recording chunk uploaded: %v", err)
	}
	if uploadedChunk.Status != "uploaded" || !uploadedChunk.AvailableFileTypes["manifest"] || !uploadedChunk.AvailableFileTypes["rgb_audio_mp4"] {
		t.Fatalf("expected uploaded chunk with preserved file availability, got %#v", uploadedChunk)
	}

	if storageObjectCount := countRecordingServiceStorageObjects(t, dsn, uploadedChunk.ID); storageObjectCount != 2 {
		t.Fatalf("expected file and manifest storage objects, got %d", storageObjectCount)
	}
}

func TestRecordingServiceIntegrationRejectsUploadFromWrongFinalizationClaim(t *testing.T) {
	services, dsn := newRecordingServiceIntegrationFixture(t)
	fixture := createRecordingServiceMissionFixture(t, services)
	ctx := context.Background()
	tickAt := time.Date(2026, 5, 28, 5, 0, 0, 0, time.UTC)

	result, err := services.Recording.ApplyRecordingTick(ctx, store.RecordingTickInput{
		MissionCode:          fixture.Mission.MissionCode,
		RobotCode:            fixture.Robot.RobotCode,
		ChunkDurationSeconds: 600,
		TickAt:               tickAt,
	})
	if err != nil {
		t.Fatalf("apply recording tick: %v", err)
	}
	if _, err := services.Missions.EndMission(ctx, fixture.Mission.MissionCode); err != nil {
		t.Fatalf("end mission: %v", err)
	}

	jobs, err := services.Recording.ClaimFinalizationJobs(ctx, "worker-a", 10, time.Minute)
	if err != nil {
		t.Fatalf("claim finalization jobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected one finalization job, got %#v", jobs)
	}

	sizeBytes := int64(1024)
	_, err = services.Recording.MarkRecordingFileUploaded(ctx, result.Chunk.ID, "rgb_audio_mp4", store.RecordingUploadMetadata{
		SizeBytes: &sizeBytes,
		WorkerID:  "worker-b",
		Attempt:   jobs[0].Attempts,
	})
	if !errors.Is(err, store.ErrInvalidState) {
		t.Fatalf("expected invalid state for wrong worker upload claim, got %v", err)
	}
	if storageObjectCount := countRecordingServiceStorageObjects(t, dsn, result.Chunk.ID); storageObjectCount != 0 {
		t.Fatalf("wrong worker upload should not create storage objects, got %d", storageObjectCount)
	}

	uploadedChunk, err := services.Recording.MarkRecordingFileUploaded(ctx, result.Chunk.ID, "rgb_audio_mp4", store.RecordingUploadMetadata{
		SizeBytes: &sizeBytes,
		WorkerID:  "worker-a",
		Attempt:   jobs[0].Attempts,
	})
	if err != nil {
		t.Fatalf("mark recording file uploaded with valid claim: %v", err)
	}
	if !uploadedChunk.AvailableFileTypes["rgb_audio_mp4"] {
		t.Fatalf("expected rgb file availability after valid claim, got %#v", uploadedChunk)
	}
}

type recordingServiceIntegrationFixture struct {
	Robot   domain.Robot
	Mission domain.Mission
}

func newRecordingServiceIntegrationFixture(t *testing.T) (*Services, string) {
	t.Helper()
	postgresContainer := postgrestest.Start(t)
	repository, err := store.NewPostgresStore(context.Background(), store.PostgresConfig{
		DSN:       postgresContainer.DSN,
		ServerURL: "http://test-server",
	})
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	t.Cleanup(func() {
		if err := repository.Close(); err != nil {
			t.Fatalf("close postgres store: %v", err)
		}
	})
	return NewServices(repository), postgresContainer.DSN
}

func createRecordingServiceMissionFixture(t *testing.T, services *Services) recordingServiceIntegrationFixture {
	t.Helper()
	ctx := context.Background()
	robot, _, err := services.Robots.CreateRobot(ctx, store.CreateRobotInput{
		DisplayName: "Recording Service Robot",
		ModelName:   "Test Model",
	})
	if err != nil {
		t.Fatalf("create robot: %v", err)
	}
	mission, err := services.Missions.CreateMission(ctx, store.CreateMissionInput{
		Name:        "Recording Service Mission",
		MissionType: "mountain_rescue",
		RobotCodes:  []string{robot.RobotCode},
	})
	if err != nil {
		t.Fatalf("create mission: %v", err)
	}
	startedMission, err := services.Missions.StartMission(ctx, mission.MissionCode)
	if err != nil {
		t.Fatalf("start mission: %v", err)
	}
	return recordingServiceIntegrationFixture{
		Robot:   robot,
		Mission: startedMission,
	}
}

func countRecordingServiceStorageObjects(t *testing.T, dsn string, chunkID string) int {
	t.Helper()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres verification db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("resolve postgres verification db: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Fatalf("close postgres verification db: %v", err)
		}
	})

	var count int
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM storage_objects
		WHERE recording_chunk_id = ?::uuid
	`, chunkID).Scan(&count).Error; err != nil {
		t.Fatalf("count storage objects: %v", err)
	}
	return count
}
