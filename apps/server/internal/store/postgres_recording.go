package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"robot-center/apps/server/internal/domain"
)

type recordingChunkMetadata struct {
	ManifestObjectKey  string            `json:"manifestObjectKey"`
	MediaObjectKeys    map[string]string `json:"mediaObjectKeys"`
	AvailableFileTypes map[string]bool   `json:"availableFileTypes,omitempty"`
}

type sqlContextRunner interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (s *PostgresStore) sqlRunner() sqlContextRunner {
	if s.sqlTx != nil {
		return s.sqlTx
	}
	return s.sqlDB
}

func (s *PostgresStore) FindRecordingTarget(ctx context.Context, missionCode string, robotCode string) (RecordingTarget, error) {
	mission, robotID, err := s.findMissionRobotForRecording(ctx, s.sqlRunner(), missionCode, robotCode)
	if err != nil {
		return RecordingTarget{}, err
	}
	return RecordingTarget{
		Mission:   mission,
		RobotID:   robotID,
		RobotCode: mission.RobotCode,
	}, nil
}

func (s *PostgresStore) FindOrCreateRecordingSession(ctx context.Context, missionID string, robotID string, chunkDurationSeconds int, startedAt time.Time) (RecordingSession, error) {
	return s.findOrCreateRecordingSession(ctx, s.sqlRunner(), missionID, robotID, chunkDurationSeconds, startedAt)
}

func (s *PostgresStore) FindRecordingChunk(ctx context.Context, recordingSessionID string, chunkIndex int) (domain.RecordingChunk, bool, error) {
	return s.findRecordingChunk(ctx, s.sqlRunner(), recordingSessionID, chunkIndex)
}

func (s *PostgresStore) CreateRecordingChunk(ctx context.Context, input CreateRecordingChunkInput) (domain.RecordingChunk, error) {
	metadata := recordingChunkMetadata{
		ManifestObjectKey: input.MediaObjectKeys["manifest"],
		MediaObjectKeys:   input.MediaObjectKeys,
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return domain.RecordingChunk{}, err
	}

	var chunk domain.RecordingChunk
	err = s.sqlRunner().QueryRowContext(ctx, `
		INSERT INTO recording_chunks (
			recording_session_id, mission_id, robot_id, chunk_index, status,
			started_at, ended_at, duration_seconds, metadata, created_at, updated_at
		)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, 'recording', $5, $6, $7, $8::jsonb, $9, $10)
		RETURNING id::text, status, created_at, updated_at
	`, input.RecordingSessionID, input.MissionID, input.RobotID, input.Window.Index, input.Window.StartedAt, input.Window.EndedAt, input.Window.DurationSeconds, string(metadataJSON), input.CreatedAt, input.UpdatedAt).Scan(
		&chunk.ID,
		&chunk.Status,
		&chunk.CreatedAt,
		&chunk.UpdatedAt,
	)
	if err != nil {
		return domain.RecordingChunk{}, err
	}
	chunk.MissionID = input.MissionID
	chunk.RecordingSessionID = input.RecordingSessionID
	chunk.MissionCode = input.MissionCode
	chunk.RobotCode = input.RobotCode
	chunk.ChunkIndex = input.Window.Index
	chunk.StartedAt = input.Window.StartedAt
	chunk.EndedAt = input.Window.EndedAt
	chunk.DurationSeconds = input.Window.DurationSeconds
	chunk.ManifestObjectKey = metadata.ManifestObjectKey
	chunk.MediaObjectKeys = metadata.MediaObjectKeys
	chunk.AvailableFileTypes = map[string]bool{}
	return chunk, nil
}

func (s *PostgresStore) MarkRecordingChunkUploaded(ctx context.Context, chunkID string, uploadMetadata RecordingUploadMetadata) (domain.RecordingChunk, error) {
	runner := s.sqlRunner()
	if err := s.validateRecordingFinalizationUploadClaim(ctx, runner, chunkID, uploadMetadata); err != nil {
		return domain.RecordingChunk{}, err
	}
	chunk, err := s.getRecordingChunkByID(ctx, runner, chunkID)
	if err != nil {
		return domain.RecordingChunk{}, err
	}
	metadata := recordingChunkMetadata{
		ManifestObjectKey:  chunk.ManifestObjectKey,
		MediaObjectKeys:    chunk.MediaObjectKeys,
		AvailableFileTypes: chunk.AvailableFileTypes,
	}
	if metadata.AvailableFileTypes == nil {
		metadata.AvailableFileTypes = map[string]bool{}
	}
	metadata.AvailableFileTypes["manifest"] = true
	manifestObjectID, err := s.upsertRecordingStorageObject(ctx, runner, chunk, "manifest", uploadMetadata)
	if err != nil {
		return domain.RecordingChunk{}, err
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return domain.RecordingChunk{}, err
	}
	if _, err := runner.ExecContext(ctx, `
		UPDATE recording_chunks
		SET status = 'uploaded', metadata = $2::jsonb, manifest_object_id = $3::uuid, updated_at = now()
		WHERE id = $1::uuid
	`, strings.TrimSpace(chunkID), string(metadataJSON), manifestObjectID); err != nil {
		return domain.RecordingChunk{}, err
	}
	if _, err := runner.ExecContext(ctx, `
		UPDATE recording_finalization_jobs
		SET status = 'completed', completed_at = COALESCE(completed_at, now()), locked_until = NULL, updated_at = now()
		WHERE recording_chunk_id = $1::uuid
			AND (
				status = 'queued'
				OR (
					status = 'processing'
					AND locked_by = NULLIF($2, '')
					AND attempts = $3
				)
			)
	`, strings.TrimSpace(chunkID), strings.TrimSpace(uploadMetadata.WorkerID), uploadMetadata.Attempt); err != nil {
		return domain.RecordingChunk{}, err
	}
	if err := s.refreshRecordingSessionStatusForChunk(ctx, runner, chunkID); err != nil {
		return domain.RecordingChunk{}, err
	}
	chunk, err = s.getRecordingChunkByID(ctx, runner, chunkID)
	if err != nil {
		return domain.RecordingChunk{}, err
	}
	return chunk, nil
}

func (s *PostgresStore) MarkRecordingFileUploaded(ctx context.Context, chunkID string, fileType string, uploadMetadata RecordingUploadMetadata) (domain.RecordingChunk, error) {
	runner := s.sqlRunner()
	if err := s.validateRecordingFinalizationUploadClaim(ctx, runner, chunkID, uploadMetadata); err != nil {
		return domain.RecordingChunk{}, err
	}
	chunk, err := s.getRecordingChunkByID(ctx, runner, chunkID)
	if err != nil {
		return domain.RecordingChunk{}, err
	}
	metadata := recordingChunkMetadata{
		ManifestObjectKey:  chunk.ManifestObjectKey,
		MediaObjectKeys:    chunk.MediaObjectKeys,
		AvailableFileTypes: chunk.AvailableFileTypes,
	}
	if metadata.AvailableFileTypes == nil {
		metadata.AvailableFileTypes = map[string]bool{}
	}
	normalizedFileType := strings.TrimSpace(fileType)
	metadata.AvailableFileTypes[normalizedFileType] = true
	if _, err := s.upsertRecordingStorageObject(ctx, runner, chunk, normalizedFileType, uploadMetadata); err != nil {
		return domain.RecordingChunk{}, err
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return domain.RecordingChunk{}, err
	}
	if _, err := runner.ExecContext(ctx, `
		UPDATE recording_chunks
		SET metadata = $2::jsonb, updated_at = now()
		WHERE id = $1::uuid
	`, strings.TrimSpace(chunkID), string(metadataJSON)); err != nil {
		return domain.RecordingChunk{}, err
	}
	chunk, err = s.getRecordingChunkByID(ctx, runner, chunkID)
	if err != nil {
		return domain.RecordingChunk{}, err
	}
	return chunk, nil
}

type recordingStorageObject struct {
	ObjectType  string
	ObjectKey   string
	ContentType string
}

func (s *PostgresStore) upsertRecordingStorageObject(ctx context.Context, runner sqlContextRunner, chunk domain.RecordingChunk, fileType string, uploadMetadata RecordingUploadMetadata) (string, error) {
	storageObject, ok := recordingStorageObjectForFileType(chunk, fileType)
	if !ok {
		return "", fmt.Errorf("unsupported recording fileType %q", fileType)
	}
	if strings.TrimSpace(storageObject.ObjectKey) == "" {
		return "", fmt.Errorf("recording fileType %q has no object key", fileType)
	}

	var robotID string
	if err := runner.QueryRowContext(ctx, `
		SELECT id::text
		FROM robots
		WHERE robot_code = $1
	`, chunk.RobotCode).Scan(&robotID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}

	metadataJSON, err := json.Marshal(map[string]any{
		"status": "available",
		"source": "recorder-worker",
	})
	if err != nil {
		return "", err
	}
	sizeBytes := nullableInt64(uploadMetadata.SizeBytes)

	var storageObjectID string
	err = runner.QueryRowContext(ctx, `
		INSERT INTO storage_objects (
			mission_id, robot_id, recording_chunk_id, object_type, bucket,
			object_key, content_type, size_bytes, checksum, started_at, ended_at, metadata
		)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7, $8, $9, $10, $11, $12::jsonb)
		ON CONFLICT (object_key)
		DO UPDATE SET
			mission_id = EXCLUDED.mission_id,
			robot_id = EXCLUDED.robot_id,
			recording_chunk_id = EXCLUDED.recording_chunk_id,
			object_type = EXCLUDED.object_type,
			bucket = EXCLUDED.bucket,
			content_type = EXCLUDED.content_type,
			size_bytes = COALESCE(EXCLUDED.size_bytes, storage_objects.size_bytes),
			checksum = COALESCE(EXCLUDED.checksum, storage_objects.checksum),
			started_at = EXCLUDED.started_at,
			ended_at = EXCLUDED.ended_at,
			metadata = storage_objects.metadata || EXCLUDED.metadata
		RETURNING id::text
	`, chunk.MissionID, robotID, chunk.ID, storageObject.ObjectType, s.minioBucket, storageObject.ObjectKey, storageObject.ContentType, sizeBytes, nullString(uploadMetadata.Checksum), chunk.StartedAt, chunk.EndedAt, string(metadataJSON)).Scan(&storageObjectID)
	if err != nil {
		return "", err
	}
	return storageObjectID, nil
}

func recordingStorageObjectForFileType(chunk domain.RecordingChunk, fileType string) (recordingStorageObject, bool) {
	switch strings.TrimSpace(fileType) {
	case "manifest":
		return recordingStorageObject{
			ObjectType:  "manifest",
			ObjectKey:   chunk.ManifestObjectKey,
			ContentType: "application/json",
		}, true
	case "rgb_audio_mp4":
		return recordingStorageObject{
			ObjectType:  "rgb_audio_mp4",
			ObjectKey:   chunk.MediaObjectKeys["rgbMp4"],
			ContentType: "video/mp4",
		}, true
	case "thermal_mp4":
		return recordingStorageObject{
			ObjectType:  "thermal_mp4",
			ObjectKey:   chunk.MediaObjectKeys["thermal"],
			ContentType: "video/mp4",
		}, true
	case "sensor_jsonl":
		return recordingStorageObject{
			ObjectType:  "sensor_jsonl",
			ObjectKey:   chunk.MediaObjectKeys["sensor"],
			ContentType: "application/x-ndjson",
		}, true
	case "telemetry_jsonl":
		return recordingStorageObject{
			ObjectType:  "telemetry_jsonl",
			ObjectKey:   chunk.MediaObjectKeys["telemetry"],
			ContentType: "application/x-ndjson",
		}, true
	default:
		return recordingStorageObject{}, false
	}
}

func (s *PostgresStore) ListRecordingChunks(ctx context.Context) ([]domain.RecordingChunk, error) {
	rows, err := s.sqlDB.QueryContext(ctx, recordingChunkSelectSQL()+`
		ORDER BY rc.started_at DESC
		LIMIT 300
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chunks := make([]domain.RecordingChunk, 0)
	for rows.Next() {
		chunk, err := scanRecordingChunk(rows)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, chunk)
	}
	return chunks, rows.Err()
}

func (s *PostgresStore) RecordingTargets(ctx context.Context) ([]domain.Mission, error) {
	rows, err := s.sqlDB.QueryContext(ctx, `
		SELECT
			m.id::text,
			m.mission_code,
			m.name,
			m.mission_type,
			m.status,
			COALESCE(m.site_note, ''),
			r.robot_code,
			m.started_at,
			m.ended_at,
			m.created_at,
			m.updated_at
		FROM missions m
		JOIN mission_robots mr ON mr.mission_id = m.id AND mr.status != 'removed'
		JOIN robots r ON r.id = mr.robot_id
		WHERE m.status = 'active'
		ORDER BY m.mission_code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	missions := make([]domain.Mission, 0)
	for rows.Next() {
		mission, err := scanMission(rows)
		if err != nil {
			return nil, err
		}
		missions = append(missions, mission)
	}
	return missions, rows.Err()
}

func (s *PostgresStore) findMissionRobotForRecording(ctx context.Context, runner sqlContextRunner, missionCode string, robotCode string) (domain.Mission, string, error) {
	row := runner.QueryRowContext(ctx, `
		SELECT
			m.id::text,
			m.mission_code,
			m.name,
			m.mission_type,
			m.status,
			COALESCE(m.site_note, ''),
			r.robot_code,
			m.started_at,
			m.ended_at,
			m.created_at,
			m.updated_at,
			r.id::text
		FROM missions m
		JOIN mission_robots mr ON mr.mission_id = m.id AND mr.status != 'removed'
		JOIN robots r ON r.id = mr.robot_id
		WHERE m.mission_code = $1 AND ($2 = '' OR r.robot_code = $2)
		ORDER BY mr.created_at, r.robot_code
		LIMIT 1
	`, strings.TrimSpace(missionCode), strings.TrimSpace(robotCode))
	mission, robotID, err := scanMissionWithRobotID(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Mission{}, "", ErrNotFound
	}
	return mission, robotID, err
}

func (s *PostgresStore) findOrCreateRecordingSession(ctx context.Context, runner sqlContextRunner, missionID string, robotID string, chunkDurationSeconds int, startedAt time.Time) (RecordingSession, error) {
	var recordingSession RecordingSession
	err := runner.QueryRowContext(ctx, `
		SELECT id::text, started_at
		FROM recording_sessions
		WHERE mission_id = $1::uuid AND robot_id = $2::uuid AND ended_at IS NULL
		ORDER BY started_at DESC
		LIMIT 1
	`, missionID, robotID).Scan(&recordingSession.ID, &recordingSession.StartedAt)
	if err == nil {
		_, updateErr := runner.ExecContext(ctx, `
			UPDATE recording_sessions
			SET status = 'recording', chunk_duration_seconds = $2
			WHERE id = $1::uuid
		`, recordingSession.ID, chunkDurationSeconds)
		return recordingSession, updateErr
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return RecordingSession{}, err
	}
	err = runner.QueryRowContext(ctx, `
		INSERT INTO recording_sessions (mission_id, robot_id, status, chunk_duration_seconds, started_at)
		VALUES ($1::uuid, $2::uuid, 'recording', $3, $4)
		RETURNING id::text, started_at
	`, missionID, robotID, chunkDurationSeconds, startedAt).Scan(&recordingSession.ID, &recordingSession.StartedAt)
	return recordingSession, err
}

func (s *PostgresStore) findRecordingChunk(ctx context.Context, runner sqlContextRunner, recordingSessionID string, chunkIndex int) (domain.RecordingChunk, bool, error) {
	row := runner.QueryRowContext(ctx, recordingChunkSelectSQL()+`
		WHERE rc.recording_session_id = $1::uuid AND rc.chunk_index = $2
	`, recordingSessionID, chunkIndex)
	chunk, err := scanRecordingChunk(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.RecordingChunk{}, false, nil
	}
	return chunk, err == nil, err
}

func (s *PostgresStore) getRecordingChunkByID(ctx context.Context, runner sqlContextRunner, chunkID string) (domain.RecordingChunk, error) {
	row := runner.QueryRowContext(ctx, recordingChunkSelectSQL()+`
		WHERE rc.id = $1::uuid
	`, strings.TrimSpace(chunkID))
	chunk, err := scanRecordingChunk(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.RecordingChunk{}, ErrNotFound
	}
	return chunk, err
}

func recordingChunkSelectSQL() string {
	return `
		SELECT
			rc.id::text,
			rc.recording_session_id::text,
			m.id::text,
			m.mission_code,
			r.robot_code,
			rc.chunk_index,
			rc.status,
			rc.started_at,
			rc.ended_at,
			COALESCE(rc.duration_seconds::int, 0),
			rc.metadata,
			rc.created_at,
			rc.updated_at
		FROM recording_chunks rc
		JOIN missions m ON m.id = rc.mission_id
		JOIN robots r ON r.id = rc.robot_id
	`
}
