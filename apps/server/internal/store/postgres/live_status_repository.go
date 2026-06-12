package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"robot-center/apps/server/internal/domain"
	repo "robot-center/apps/server/internal/store/port"
)

func (s *Store) GetMissionLiveStatusSnapshot(ctx context.Context, missionCode string) (repo.MissionLiveStatusSnapshot, error) {
	missionCode = strings.TrimSpace(missionCode)
	runner := s.sqlRunner()
	mission, err := s.findMissionByCode(ctx, runner, missionCode)
	if err != nil {
		return repo.MissionLiveStatusSnapshot{}, err
	}
	robots, err := s.listMissionAssignedRobots(ctx, runner, missionCode)
	if err != nil {
		return repo.MissionLiveStatusSnapshot{}, err
	}
	recordingChunks, err := s.listMissionLiveRecordingChunks(ctx, runner, missionCode)
	if err != nil {
		return repo.MissionLiveStatusSnapshot{}, err
	}
	streamSessions, err := s.listRobotStreamSessionsForMission(ctx, runner, missionCode)
	if err != nil {
		return repo.MissionLiveStatusSnapshot{}, err
	}
	return repo.MissionLiveStatusSnapshot{
		Mission:         mission,
		Robots:          robots,
		RecordingChunks: recordingChunks,
		StreamSessions:  streamSessions,
	}, nil
}

func (s *Store) listMissionAssignedRobots(ctx context.Context, runner sqlContextRunner, missionCode string) ([]domain.Robot, error) {
	rows, err := runner.QueryContext(ctx, `
		SELECT
			r.id::text,
			r.robot_code,
			r.display_name,
			COALESCE(r.model_name, ''),
			r.device_state,
			r.last_seen_at,
			r.created_at,
			r.updated_at
		FROM missions m
		JOIN mission_robots mr ON mr.mission_id = m.id AND mr.status != 'removed'
		JOIN robots r ON r.id = mr.robot_id
		WHERE m.mission_code = $1
			AND r.archived_at IS NULL
		ORDER BY mr.created_at, r.robot_code
	`, strings.TrimSpace(missionCode))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	robots := make([]domain.Robot, 0)
	for rows.Next() {
		robot, err := scanRobot(rows)
		if err != nil {
			return nil, err
		}
		robots = append(robots, robot)
	}
	return robots, rows.Err()
}

func (s *Store) listMissionLiveRecordingChunks(ctx context.Context, runner sqlContextRunner, missionCode string) ([]domain.RecordingChunk, error) {
	rows, err := runner.QueryContext(ctx, `
		SELECT
			id,
			recording_session_id,
			mission_id,
			mission_code,
			robot_code,
			chunk_index,
			status,
			started_at,
			ended_at,
			duration_seconds,
			metadata,
			created_at,
			updated_at
		FROM (
			SELECT
				rc.id::text AS id,
				rc.recording_session_id::text AS recording_session_id,
				m.id::text AS mission_id,
				m.mission_code AS mission_code,
				r.robot_code AS robot_code,
				rc.chunk_index AS chunk_index,
				rc.status AS status,
				rc.started_at AS started_at,
				rc.ended_at AS ended_at,
				COALESCE(rc.duration_seconds::int, 0) AS duration_seconds,
				rc.metadata AS metadata,
				rc.created_at AS created_at,
				rc.updated_at AS updated_at,
				row_number() OVER (
					PARTITION BY r.robot_code
					ORDER BY
						CASE WHEN rc.status = 'recording' THEN 0 ELSE 1 END,
						rc.chunk_index DESC,
						rc.started_at DESC,
						rc.updated_at DESC,
						rc.id DESC
				) AS live_rank
			FROM recording_chunks rc
			JOIN recording_sessions rs ON rs.id = rc.recording_session_id
			JOIN missions m ON m.id = rc.mission_id
			JOIN robots r ON r.id = rc.robot_id
			WHERE m.mission_code = $1
				AND NOT (
					rs.ended_at IS NOT NULL
					AND rc.status IN ('recording', 'pending')
				)
		) ranked_chunks
		WHERE live_rank = 1
		ORDER BY started_at DESC, id DESC
	`, strings.TrimSpace(missionCode))
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

func (s *Store) findMissionByCode(ctx context.Context, runner sqlContextRunner, missionCode string) (domain.Mission, error) {
	row := runner.QueryRowContext(ctx, `
		SELECT
			m.id::text,
			m.mission_code,
			m.name,
			m.mission_type,
			m.status,
			COALESCE(m.site_note, ''),
			COALESCE(string_agg(r.robot_code, ',' ORDER BY mr.created_at, r.robot_code) FILTER (WHERE r.robot_code IS NOT NULL), ''),
			m.started_at,
			m.ended_at,
			m.created_at,
			m.updated_at
		FROM missions m
		LEFT JOIN mission_robots mr ON mr.mission_id = m.id AND mr.status != 'removed'
		LEFT JOIN robots r ON r.id = mr.robot_id
		WHERE m.mission_code = $1
		GROUP BY m.id, m.mission_code, m.name, m.mission_type, m.status, m.site_note, m.started_at, m.ended_at, m.created_at, m.updated_at
	`, strings.TrimSpace(missionCode))
	mission, err := scanMissionWithRobotCodes(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Mission{}, repo.ErrNotFound
	}
	return mission, err
}
