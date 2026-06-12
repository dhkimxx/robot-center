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

func (s *Store) StartRobotStreamSession(ctx context.Context, input repo.StartRobotStreamSessionInput) (domain.RobotStreamSession, error) {
	if s.sqlTx != nil {
		return s.startRobotStreamSessionWithRunner(ctx, s.sqlTx, input)
	}
	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return domain.RobotStreamSession{}, err
	}
	defer rollbackUnlessCommitted(tx)

	session, err := s.startRobotStreamSessionWithRunner(ctx, tx, input)
	if err != nil {
		return domain.RobotStreamSession{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.RobotStreamSession{}, err
	}
	return session, nil
}

func (s *Store) startRobotStreamSessionWithRunner(ctx context.Context, runner sqlContextRunner, input repo.StartRobotStreamSessionInput) (domain.RobotStreamSession, error) {
	startedAt := normalizedTimeOrNow(input.StartedAt)
	missionID, robotID, err := s.findActiveMissionRobotForStreamSession(ctx, runner, input.MissionCode, input.RobotCode)
	if err != nil {
		return domain.RobotStreamSession{}, err
	}

	publisherPeerID := strings.TrimSpace(input.PublisherPeerID)
	if publisherPeerID == "" {
		return domain.RobotStreamSession{}, errors.New("publisherPeerID is required")
	}

	if _, err := runner.ExecContext(ctx, `
		UPDATE robot_stream_sessions
		SET state = 'ended', ended_at = COALESCE(ended_at, $4), end_reason = COALESCE(end_reason, 'replaced')
		WHERE mission_id = $1::uuid
			AND robot_id = $2::uuid
			AND publisher_peer_id != $3
			AND ended_at IS NULL
	`, missionID, robotID, publisherPeerID, startedAt); err != nil {
		return domain.RobotStreamSession{}, err
	}

	row := runner.QueryRowContext(ctx, `
	WITH upserted AS (
		INSERT INTO robot_stream_sessions (
			mission_id, robot_id, publisher_peer_id, state,
			started_at, last_media_at, ended_at, end_reason, metadata
		)
		VALUES ($1::uuid, $2::uuid, $3, 'active', $4, $4, NULL, NULL, '{}'::jsonb)
		ON CONFLICT (publisher_peer_id)
		DO UPDATE SET
			mission_id = EXCLUDED.mission_id,
			robot_id = EXCLUDED.robot_id,
			state = 'active',
			started_at = LEAST(robot_stream_sessions.started_at, EXCLUDED.started_at),
			last_media_at = GREATEST(COALESCE(robot_stream_sessions.last_media_at, EXCLUDED.last_media_at), EXCLUDED.last_media_at),
			ended_at = NULL,
			end_reason = NULL
		RETURNING
			id,
			mission_id,
			robot_id,
			publisher_peer_id,
			state,
			started_at,
			last_media_at,
			ended_at,
			end_reason,
			created_at,
			updated_at
	)
	SELECT
		upserted.id::text,
		m.mission_code,
		upserted.mission_id::text,
		r.robot_code,
		upserted.robot_id::text,
		upserted.publisher_peer_id,
		upserted.state,
		upserted.started_at,
		upserted.last_media_at,
		upserted.ended_at,
		upserted.end_reason,
		upserted.created_at,
		upserted.updated_at
	FROM upserted
	JOIN missions m ON m.id = upserted.mission_id
	JOIN robots r ON r.id = upserted.robot_id
	`, missionID, robotID, publisherPeerID, startedAt)
	return scanRobotStreamSession(row)
}

func (s *Store) TouchRobotStreamSession(ctx context.Context, input repo.TouchRobotStreamSessionInput) error {
	publisherPeerID := strings.TrimSpace(input.PublisherPeerID)
	if publisherPeerID == "" {
		return nil
	}
	observedAt := normalizedTimeOrNow(input.ObservedAt)
	_, err := s.sqlRunner().ExecContext(ctx, `
		UPDATE robot_stream_sessions
		SET
			state = 'active',
			last_media_at = GREATEST(COALESCE(last_media_at, $2), $2)
		WHERE publisher_peer_id = $1
			AND ended_at IS NULL
	`, publisherPeerID, observedAt)
	return err
}

func (s *Store) EndRobotStreamSession(ctx context.Context, input repo.EndRobotStreamSessionInput) error {
	publisherPeerID := strings.TrimSpace(input.PublisherPeerID)
	if publisherPeerID == "" {
		return nil
	}
	endedAt := normalizedTimeOrNow(input.EndedAt)
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		reason = "closed"
	}
	_, err := s.sqlRunner().ExecContext(ctx, `
		UPDATE robot_stream_sessions
		SET
			state = 'ended',
			ended_at = COALESCE(ended_at, $2),
			end_reason = COALESCE(NULLIF(end_reason, ''), $3)
		WHERE publisher_peer_id = $1
			AND ended_at IS NULL
	`, publisherPeerID, endedAt, reason)
	return err
}

func (s *Store) ListRobotStreamSessionsForMission(ctx context.Context, missionCode string) ([]domain.RobotStreamSession, error) {
	return s.listRobotStreamSessionsForMission(ctx, s.sqlRunner(), missionCode)
}

func (s *Store) listRobotStreamSessionsForMission(ctx context.Context, runner sqlContextRunner, missionCode string) ([]domain.RobotStreamSession, error) {
	rows, err := runner.QueryContext(ctx, `
		SELECT
			rss.id::text,
			m.mission_code,
			rss.mission_id::text,
			r.robot_code,
			rss.robot_id::text,
			rss.publisher_peer_id,
			rss.state,
			rss.started_at,
			rss.last_media_at,
			rss.ended_at,
			rss.end_reason,
			rss.created_at,
			rss.updated_at
		FROM robot_stream_sessions rss
		JOIN missions m ON m.id = rss.mission_id
		JOIN robots r ON r.id = rss.robot_id
		WHERE m.mission_code = $1
		ORDER BY rss.started_at DESC, rss.created_at DESC
		LIMIT 300
	`, strings.TrimSpace(missionCode))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]domain.RobotStreamSession, 0)
	for rows.Next() {
		session, err := scanRobotStreamSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

func (s *Store) findActiveMissionRobotForStreamSession(ctx context.Context, runner sqlContextRunner, missionCode string, robotCode string) (string, string, error) {
	var missionID string
	var robotID string
	err := runner.QueryRowContext(ctx, `
		SELECT m.id::text, r.id::text
		FROM missions m
		JOIN mission_robots mr ON mr.mission_id = m.id AND mr.status != 'removed'
		JOIN robots r ON r.id = mr.robot_id
		WHERE m.mission_code = $1
			AND r.robot_code = $2
			AND m.status = 'active'
		LIMIT 1
	`, strings.TrimSpace(missionCode), strings.TrimSpace(robotCode)).Scan(&missionID, &robotID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", repo.ErrNotFound
	}
	return missionID, robotID, err
}

func scanRobotStreamSession(row scanner) (domain.RobotStreamSession, error) {
	var session domain.RobotStreamSession
	var endReason sql.NullString
	err := row.Scan(
		&session.ID,
		&session.MissionCode,
		&session.MissionID,
		&session.RobotCode,
		&session.RobotID,
		&session.PublisherPeerID,
		&session.State,
		&session.StartedAt,
		nullableTimeScanner(&session.LastMediaAt),
		nullableTimeScanner(&session.EndedAt),
		&endReason,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if endReason.Valid {
		session.EndReason = endReason.String
	}
	return session, err
}

func normalizedTimeOrNow(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value.UTC()
}
