package postgres

import (
	"context"
	"database/sql"
	"strings"

	"robot-center/apps/server/internal/domain"
	repo "robot-center/apps/server/internal/store/port"
)

func (s *Store) ListRecordingChunks(ctx context.Context) ([]domain.RecordingChunk, error) {
	rows, err := s.sqlDB.QueryContext(ctx, recordingChunkSelectSQL()+`
		JOIN recording_sessions rs ON rs.id = rc.recording_session_id
		WHERE NOT (
			rs.ended_at IS NOT NULL
			AND rc.status IN ('recording', 'pending')
		)
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

func (s *Store) SummarizeMissionRecordings(ctx context.Context, missionCode string) (repo.MissionRecordingSummary, error) {
	missionCode = strings.TrimSpace(missionCode)
	rows, err := s.sqlDB.QueryContext(ctx, `
		WITH chunk_files AS (
			SELECT
				r.robot_code,
				rc.id,
				rc.status,
				rc.started_at,
				rc.ended_at,
				rc.updated_at,
				COALESCE(BOOL_OR(so.object_type = 'rgb_audio_mp4'), false) AS has_rgb_audio_mp4,
				COALESCE(BOOL_OR(so.object_type = 'thermal_mp4'), false) AS has_thermal_mp4,
				COALESCE(BOOL_OR(so.object_type = 'sensor_jsonl'), false) AS has_sensor_jsonl,
				COALESCE(BOOL_OR(so.object_type = 'telemetry_jsonl'), false) AS has_telemetry_jsonl,
				COALESCE(BOOL_OR(so.object_type = 'manifest'), false) AS has_manifest
			FROM recording_chunks rc
			JOIN recording_sessions rs ON rs.id = rc.recording_session_id
			JOIN missions m ON m.id = rc.mission_id
			JOIN robots r ON r.id = rc.robot_id
			LEFT JOIN storage_objects so ON so.recording_chunk_id = rc.id
			WHERE m.mission_code = $1
				AND NOT (
					rs.ended_at IS NOT NULL
					AND rc.status IN ('recording', 'pending')
				)
			GROUP BY r.robot_code, rc.id, rc.status, rc.started_at, rc.ended_at, rc.updated_at
		)
		SELECT
			robot_code,
			COUNT(*)::int AS chunk_count,
			COUNT(*) FILTER (WHERE status = 'uploaded')::int AS uploaded_chunk_count,
			COUNT(*) FILTER (WHERE status IN ('recording', 'pending'))::int AS recording_chunk_count,
			COUNT(*) FILTER (WHERE status = 'finalizing')::int AS finalizing_chunk_count,
			COUNT(*) FILTER (WHERE status = 'uploaded' AND NOT (has_rgb_audio_mp4 OR has_thermal_mp4))::int AS partial_chunk_count,
			MIN(started_at),
			MAX(COALESCE(ended_at, updated_at, started_at)),
			COUNT(*) FILTER (WHERE has_rgb_audio_mp4)::int AS rgb_audio_mp4_count,
			COUNT(*) FILTER (WHERE has_thermal_mp4)::int AS thermal_mp4_count,
			COUNT(*) FILTER (WHERE has_sensor_jsonl)::int AS sensor_jsonl_count,
			COUNT(*) FILTER (WHERE has_telemetry_jsonl)::int AS telemetry_jsonl_count,
			COUNT(*) FILTER (WHERE has_manifest)::int AS manifest_count
		FROM chunk_files
		GROUP BY robot_code
		ORDER BY robot_code
	`, missionCode)
	if err != nil {
		return repo.MissionRecordingSummary{}, err
	}
	defer rows.Close()

	summary := repo.MissionRecordingSummary{
		MissionCode: missionCode,
		Robots:      []repo.MissionRecordingRobotSummary{},
	}
	for rows.Next() {
		var robotSummary repo.MissionRecordingRobotSummary
		var firstStartedAt sql.NullTime
		var lastEndedAt sql.NullTime
		var rgbAudioMP4Count int
		var thermalMP4Count int
		var sensorJSONLCount int
		var telemetryJSONLCount int
		var manifestCount int
		if err := rows.Scan(
			&robotSummary.RobotCode,
			&robotSummary.ChunkCount,
			&robotSummary.UploadedChunkCount,
			&robotSummary.RecordingChunkCount,
			&robotSummary.FinalizingChunkCount,
			&robotSummary.PartialChunkCount,
			&firstStartedAt,
			&lastEndedAt,
			&rgbAudioMP4Count,
			&thermalMP4Count,
			&sensorJSONLCount,
			&telemetryJSONLCount,
			&manifestCount,
		); err != nil {
			return repo.MissionRecordingSummary{}, err
		}
		robotSummary.FirstStartedAt = timePointer(firstStartedAt)
		robotSummary.LastEndedAt = timePointer(lastEndedAt)
		robotSummary.AvailableFileCounts = map[string]int{
			"rgb_audio_mp4":   rgbAudioMP4Count,
			"thermal_mp4":     thermalMP4Count,
			"sensor_jsonl":    sensorJSONLCount,
			"telemetry_jsonl": telemetryJSONLCount,
			"manifest":        manifestCount,
		}
		robotSummary.MissingFileCounts = makeMissingRecordingFileCounts(robotSummary.ChunkCount, robotSummary.AvailableFileCounts)
		summary.TotalChunks += robotSummary.ChunkCount
		summary.Robots = append(summary.Robots, robotSummary)
	}
	return summary, rows.Err()
}

func (s *Store) ListMissionRecordingChunks(ctx context.Context, query repo.MissionRecordingChunkQuery) (repo.MissionRecordingChunkPage, error) {
	query.MissionCode = strings.TrimSpace(query.MissionCode)
	query.RobotCode = strings.TrimSpace(query.RobotCode)
	if query.Limit <= 0 {
		query.Limit = 100
	}
	if query.Offset < 0 {
		query.Offset = 0
	}

	var total int
	if err := s.sqlDB.QueryRowContext(ctx, `
		SELECT COUNT(*)::int
		FROM recording_chunks rc
		JOIN recording_sessions rs ON rs.id = rc.recording_session_id
		JOIN missions m ON m.id = rc.mission_id
		JOIN robots r ON r.id = rc.robot_id
		WHERE m.mission_code = $1
			AND ($2 = '' OR r.robot_code = $2)
			AND NOT (
				rs.ended_at IS NOT NULL
				AND rc.status IN ('recording', 'pending')
			)
	`, query.MissionCode, query.RobotCode).Scan(&total); err != nil {
		return repo.MissionRecordingChunkPage{}, err
	}

	rows, err := s.sqlDB.QueryContext(ctx, recordingChunkSelectSQL()+`
		JOIN recording_sessions rs ON rs.id = rc.recording_session_id
		WHERE m.mission_code = $1
			AND ($2 = '' OR r.robot_code = $2)
			AND NOT (
				rs.ended_at IS NOT NULL
				AND rc.status IN ('recording', 'pending')
			)
		ORDER BY rc.started_at DESC, rc.id DESC
		LIMIT $3 OFFSET $4
	`, query.MissionCode, query.RobotCode, query.Limit, query.Offset)
	if err != nil {
		return repo.MissionRecordingChunkPage{}, err
	}
	defer rows.Close()

	chunks := make([]domain.RecordingChunk, 0)
	for rows.Next() {
		chunk, err := scanRecordingChunk(rows)
		if err != nil {
			return repo.MissionRecordingChunkPage{}, err
		}
		chunks = append(chunks, chunk)
	}
	if err := rows.Err(); err != nil {
		return repo.MissionRecordingChunkPage{}, err
	}
	return repo.MissionRecordingChunkPage{
		Chunks: chunks,
		Limit:  query.Limit,
		Offset: query.Offset,
		Total:  total,
	}, nil
}

func makeMissingRecordingFileCounts(chunkCount int, availableFileCounts map[string]int) map[string]int {
	missingFileCounts := make(map[string]int, len(availableFileCounts))
	for fileType, availableCount := range availableFileCounts {
		missingCount := chunkCount - availableCount
		if missingCount < 0 {
			missingCount = 0
		}
		missingFileCounts[fileType] = missingCount
	}
	return missingFileCounts
}
