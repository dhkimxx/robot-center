package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"robot-center/apps/server/internal/domain"
	repo "robot-center/apps/server/internal/store/port"
)

func (s *Store) QueueRecordingFinalizationJobsForInactiveMissions(ctx context.Context) (int64, error) {
	runner := s.sqlRunner()
	var queued int64
	err := runner.QueryRowContext(ctx, `
		WITH candidates AS (
			SELECT
				rc.id,
				rc.recording_session_id,
				rc.mission_id,
				rc.robot_id
			FROM recording_chunks rc
			JOIN missions m ON m.id = rc.mission_id
			WHERE m.status != 'active'
				AND rc.status IN ('recording', 'pending', 'stopped')
		),
		inserted AS (
			INSERT INTO recording_finalization_jobs (
				recording_chunk_id,
				recording_session_id,
				mission_id,
				robot_id,
				status,
				reason,
				metadata,
				created_at,
				updated_at
			)
			SELECT
				id,
				recording_session_id,
				mission_id,
				robot_id,
				'queued',
				'mission_inactive',
				'{}'::jsonb,
				now(),
				now()
			FROM candidates
			ON CONFLICT (recording_chunk_id) DO NOTHING
			RETURNING recording_chunk_id
		),
		updated_chunks AS (
			UPDATE recording_chunks rc
			SET status = 'finalizing', updated_at = now()
			WHERE rc.id IN (SELECT id FROM candidates)
				AND rc.status IN ('recording', 'pending', 'stopped')
			RETURNING rc.recording_session_id
		),
		updated_sessions AS (
			UPDATE recording_sessions rs
			SET status = 'finalizing', ended_at = COALESCE(rs.ended_at, now())
			WHERE rs.id IN (SELECT recording_session_id FROM updated_chunks)
				AND rs.status IN ('recording', 'pending')
			RETURNING 1
		)
		SELECT COUNT(*) FROM inserted
	`).Scan(&queued)
	if err != nil {
		return 0, err
	}
	return queued, nil
}

func (s *Store) ClaimRecordingFinalizationJobs(ctx context.Context, workerID string, limit int, lockDuration time.Duration) ([]domain.RecordingFinalizationJob, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	if lockDuration <= 0 {
		lockDuration = 2 * time.Minute
	}
	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		workerID = "recorder-worker"
	}
	lockedUntil := time.Now().UTC().Add(lockDuration)
	rows, err := s.sqlRunner().QueryContext(ctx, `
		WITH claimable AS (
			SELECT rfj.id
			FROM recording_finalization_jobs rfj
			JOIN recording_chunks rc ON rc.id = rfj.recording_chunk_id
			WHERE rc.status IN ('recording', 'pending', 'finalizing')
				AND (
					rfj.status = 'queued'
					OR (rfj.status = 'processing' AND rfj.locked_until < now())
				)
			ORDER BY rfj.created_at
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE recording_finalization_jobs rfj
		SET
			status = 'processing',
			attempts = attempts + 1,
			locked_by = $2,
			locked_until = $3,
			updated_at = now()
		FROM claimable
		WHERE rfj.id = claimable.id
		RETURNING
			rfj.id::text,
			rfj.recording_chunk_id::text,
			rfj.recording_session_id::text,
			rfj.mission_id::text,
			rfj.robot_id::text,
			rfj.status,
			COALESCE(rfj.reason, ''),
			rfj.attempts,
			COALESCE(rfj.locked_by, ''),
			rfj.locked_until,
			COALESCE(rfj.last_error, ''),
			rfj.created_at,
			rfj.updated_at,
			rfj.completed_at
	`, limit, workerID, lockedUntil)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]domain.RecordingFinalizationJob, 0)
	for rows.Next() {
		job, err := scanRecordingFinalizationJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for index := range jobs {
		chunk, err := s.getRecordingChunkByID(ctx, s.sqlRunner(), jobs[index].RecordingChunkID)
		if err != nil {
			return nil, err
		}
		jobs[index].Chunk = chunk
	}
	return jobs, nil
}

func (s *Store) MarkRecordingFinalizationJobCompleted(ctx context.Context, jobID string, workerID string, attempt int) error {
	return s.markRecordingFinalizationJob(ctx, jobID, "completed", workerID, attempt, "")
}

func (s *Store) MarkRecordingFinalizationJobPartial(ctx context.Context, jobID string, workerID string, attempt int, reason string) error {
	return s.markRecordingFinalizationJob(ctx, jobID, "partial", workerID, attempt, reason)
}

func (s *Store) MarkRecordingFinalizationJobFailed(ctx context.Context, jobID string, workerID string, attempt int, reason string) error {
	return s.markRecordingFinalizationJob(ctx, jobID, "failed", workerID, attempt, reason)
}

func (s *Store) markRecordingFinalizationJob(ctx context.Context, jobID string, status string, workerID string, attempt int, reason string) error {
	runner := s.sqlRunner()
	var chunkID string
	var completedAtExpression string
	if status == "completed" || status == "partial" {
		completedAtExpression = "now()"
	} else {
		completedAtExpression = "completed_at"
	}
	err := runner.QueryRowContext(ctx, `
		UPDATE recording_finalization_jobs
		SET
			status = $2,
			reason = NULLIF($3, ''),
			last_error = NULLIF($3, ''),
			locked_until = NULL,
			completed_at = `+completedAtExpression+`,
			updated_at = now()
		WHERE id = $1::uuid
			AND (
				(
					status = 'processing'
					AND locked_by = NULLIF($4, '')
					AND attempts = $5
				)
				OR (
					$2 = 'completed'
					AND status = 'completed'
					AND locked_by = NULLIF($4, '')
					AND attempts = $5
				)
			)
		RETURNING recording_chunk_id::text
	`, strings.TrimSpace(jobID), status, strings.TrimSpace(reason), strings.TrimSpace(workerID), attempt).Scan(&chunkID)
	if errors.Is(err, sql.ErrNoRows) {
		return repo.ErrInvalidState
	}
	if err != nil {
		return err
	}
	if status == "completed" {
		return s.refreshRecordingSessionStatusForChunk(ctx, runner, chunkID)
	}
	if _, err := runner.ExecContext(ctx, `
		UPDATE recording_chunks
		SET status = $2, updated_at = now()
		WHERE id = $1::uuid
	`, chunkID, status); err != nil {
		return err
	}
	return s.refreshRecordingSessionStatusForChunk(ctx, runner, chunkID)
}

func (s *Store) validateRecordingFinalizationUploadClaim(ctx context.Context, runner sqlContextRunner, chunkID string, uploadMetadata repo.RecordingUploadMetadata) error {
	var status string
	var lockedBy string
	var attempts int
	err := runner.QueryRowContext(ctx, `
		SELECT status, COALESCE(locked_by, ''), attempts
		FROM recording_finalization_jobs
		WHERE recording_chunk_id = $1::uuid
	`, strings.TrimSpace(chunkID)).Scan(&status, &lockedBy, &attempts)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	switch status {
	case "queued":
		return nil
	case "processing":
		if strings.TrimSpace(uploadMetadata.WorkerID) == "" || uploadMetadata.Attempt <= 0 {
			return repo.ErrInvalidState
		}
		if lockedBy != strings.TrimSpace(uploadMetadata.WorkerID) || attempts != uploadMetadata.Attempt {
			return repo.ErrInvalidState
		}
		return nil
	default:
		return repo.ErrInvalidState
	}
}

func (s *Store) refreshRecordingSessionStatusForChunk(ctx context.Context, runner sqlContextRunner, chunkID string) error {
	var sessionID string
	if err := runner.QueryRowContext(ctx, `
		SELECT recording_session_id::text
		FROM recording_chunks
		WHERE id = $1::uuid
	`, strings.TrimSpace(chunkID)).Scan(&sessionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo.ErrNotFound
		}
		return err
	}
	_, err := runner.ExecContext(ctx, `
		WITH session_state AS (
			SELECT
				rs.id,
				EXISTS (
					SELECT 1 FROM recording_chunks rc
					WHERE rc.recording_session_id = rs.id
						AND rc.status IN ('recording', 'pending')
				) AS has_open_chunk,
				EXISTS (
					SELECT 1 FROM recording_chunks rc
					WHERE rc.recording_session_id = rs.id
						AND rc.status IN ('recording', 'pending', 'finalizing')
				) AS has_unfinished_chunk,
				EXISTS (
					SELECT 1 FROM recording_chunks rc
					WHERE rc.recording_session_id = rs.id
						AND rc.status = 'failed'
				) AS has_failed_chunk,
				EXISTS (
					SELECT 1 FROM recording_chunks rc
					WHERE rc.recording_session_id = rs.id
						AND rc.status = 'partial'
				) AS has_partial_chunk
			FROM recording_sessions rs
			WHERE rs.id = $1::uuid
		)
		UPDATE recording_sessions rs
		SET
			status = CASE
				WHEN session_state.has_open_chunk AND rs.ended_at IS NULL THEN 'recording'
				WHEN session_state.has_unfinished_chunk THEN 'finalizing'
				WHEN session_state.has_failed_chunk THEN 'failed'
				WHEN session_state.has_partial_chunk THEN 'partial'
				ELSE 'uploaded'
			END,
			ended_at = CASE
				WHEN session_state.has_open_chunk AND rs.ended_at IS NULL THEN NULL
				ELSE COALESCE(rs.ended_at, now())
			END
		FROM session_state
		WHERE rs.id = $1::uuid
	`, sessionID)
	return err
}

func scanRecordingFinalizationJob(row scanner) (domain.RecordingFinalizationJob, error) {
	var job domain.RecordingFinalizationJob
	err := row.Scan(
		&job.ID,
		&job.RecordingChunkID,
		&job.RecordingSessionID,
		&job.MissionID,
		&job.RobotID,
		&job.Status,
		&job.Reason,
		&job.Attempts,
		&job.LockedBy,
		nullableTimeScanner(&job.LockedUntil),
		&job.LastError,
		&job.CreatedAt,
		&job.UpdatedAt,
		nullableTimeScanner(&job.CompletedAt),
	)
	if err != nil {
		return domain.RecordingFinalizationJob{}, err
	}
	return job, nil
}
