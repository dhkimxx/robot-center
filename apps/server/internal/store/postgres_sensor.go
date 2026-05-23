package store

import (
	"context"
	"strings"
	"time"

	"robot-center/apps/server/internal/domain"
)

func (s *PostgresStore) SaveSensorReading(ctx context.Context, reading domain.SensorReading) (domain.SensorReading, error) {
	robotID, err := s.findRobotID(ctx, reading.RobotCode)
	if err != nil {
		return domain.SensorReading{}, err
	}
	if reading.ReceivedAt.IsZero() {
		reading.ReceivedAt = time.Now().UTC()
	}
	if len(reading.RawPayload) == 0 {
		reading.RawPayload = []byte("{}")
	}

	err = s.sqlDB.QueryRowContext(ctx, `
		INSERT INTO sensor_readings (
			mission_id, robot_id, sequence, sent_at, received_at, temperature_celsius,
			humidity_percent, oxygen_percent, co_ppm, ch4_ppm, raw_payload
		)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb)
		RETURNING id::text, received_at
	`, reading.MissionID, robotID, reading.Sequence, reading.SentAt, reading.ReceivedAt,
		nullableFloat(reading.TemperatureCelsius), nullableFloat(reading.HumidityPercent),
		nullableFloat(reading.OxygenPercent), nullableFloat(reading.COPpm), nullableFloat(reading.CH4Ppm),
		string(reading.RawPayload),
	).Scan(&reading.ID, &reading.ReceivedAt)
	if err != nil {
		return domain.SensorReading{}, err
	}
	return reading, nil
}

func (s *PostgresStore) ListSensorReadings(ctx context.Context, missionID string) ([]domain.SensorReading, error) {
	rows, err := s.sqlDB.QueryContext(ctx, `
		SELECT
			sr.id::text,
			r.robot_code,
			sr.mission_id::text,
			COALESCE(sr.sequence, 0),
			sr.sent_at,
			sr.received_at,
			sr.temperature_celsius::float8,
			sr.humidity_percent::float8,
			sr.oxygen_percent::float8,
			sr.co_ppm::float8,
			sr.ch4_ppm::float8,
			sr.raw_payload
		FROM sensor_readings sr
		JOIN robots r ON r.id = sr.robot_id
		WHERE sr.mission_id = $1::uuid
		ORDER BY sr.received_at DESC
		LIMIT 300
	`, strings.TrimSpace(missionID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.SensorReading, 0)
	for rows.Next() {
		item, err := scanSensorReading(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
