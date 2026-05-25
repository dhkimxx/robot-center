package store

import (
	"context"
	"strings"
	"time"

	"robot-center/apps/server/internal/domain"
)

func (s *PostgresStore) SaveTelemetry(ctx context.Context, snapshot domain.TelemetrySnapshot) (domain.TelemetrySnapshot, error) {
	robotID, err := s.findRobotID(ctx, snapshot.RobotCode)
	if err != nil {
		return domain.TelemetrySnapshot{}, err
	}
	if snapshot.ReceivedAt.IsZero() {
		snapshot.ReceivedAt = time.Now().UTC()
	}
	if len(snapshot.RawPayload) == 0 {
		snapshot.RawPayload = []byte("{}")
	}
	positionType := ""
	if snapshot.Latitude != nil && snapshot.Longitude != nil {
		positionType = "gps"
		snapshot.PositionAvailable = true
	}

	err = s.sqlDB.QueryRowContext(ctx, `
		INSERT INTO telemetry_snapshots (
			mission_id, robot_id, sequence, sent_at, received_at, battery_percent,
			network_quality, position_type, latitude, longitude, altitude_meter,
			accuracy_meter, heading_degree, geom, raw_payload
		)
		VALUES (
			$1::uuid, $2::uuid, $3, $4, $5, $6::double precision, $7, $8,
			$9::double precision, $10::double precision, $11::double precision,
			$12::double precision, $13::double precision,
			CASE
				WHEN $9::double precision IS NOT NULL AND $10::double precision IS NOT NULL
				THEN ST_SetSRID(ST_MakePoint($10::double precision, $9::double precision), 4326)
				ELSE NULL
			END,
			$14::jsonb
		)
		RETURNING id::text, received_at
	`, snapshot.MissionID, robotID, snapshot.Sequence, snapshot.SentAt, snapshot.ReceivedAt,
		nullableFloat(snapshot.BatteryPercent), snapshot.NetworkState, nullString(positionType),
		nullableFloat(snapshot.Latitude), nullableFloat(snapshot.Longitude), nullableFloat(snapshot.AltitudeMeter),
		nullableFloat(snapshot.AccuracyMeter), nullableFloat(snapshot.HeadingDegree), string(snapshot.RawPayload),
	).Scan(&snapshot.ID, &snapshot.ReceivedAt)
	if err != nil {
		return domain.TelemetrySnapshot{}, err
	}
	return snapshot, nil
}

func (s *PostgresStore) ListTelemetry(ctx context.Context, missionID string) ([]domain.TelemetrySnapshot, error) {
	rows, err := s.sqlDB.QueryContext(ctx, `
		SELECT
			ts.id::text,
			r.robot_code,
			ts.mission_id::text,
			COALESCE(ts.sequence, 0),
			ts.sent_at,
			ts.received_at,
			ts.battery_percent::float8,
			COALESCE(ts.network_quality, ''),
			ts.latitude,
			ts.longitude,
			ts.altitude_meter,
			ts.accuracy_meter,
			ts.heading_degree,
			ts.raw_payload
		FROM telemetry_snapshots ts
		JOIN robots r ON r.id = ts.robot_id
		WHERE ts.mission_id = $1::uuid
		ORDER BY ts.received_at DESC
		LIMIT 300
	`, strings.TrimSpace(missionID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.TelemetrySnapshot, 0)
	for rows.Next() {
		item, err := scanTelemetry(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
