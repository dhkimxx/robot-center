package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"robot-center/apps/server/internal/domain"
)

type scanner interface {
	Scan(dest ...any) error
}

func scanRobot(row scanner) (domain.Robot, error) {
	var robot domain.Robot
	err := row.Scan(
		&robot.ID,
		&robot.RobotCode,
		&robot.DisplayName,
		&robot.ModelName,
		&robot.Status,
		nullableTimeScanner(&robot.LastSeenAt),
		nullableTimeScanner(&robot.LastStreamingAt),
		&robot.CreatedAt,
		&robot.UpdatedAt,
	)
	return robot, err
}

func scanMission(row scanner) (domain.Mission, error) {
	var mission domain.Mission
	err := row.Scan(
		&mission.ID,
		&mission.MissionCode,
		&mission.Name,
		&mission.MissionType,
		&mission.Status,
		&mission.SiteNote,
		&mission.RobotCode,
		nullableTimeScanner(&mission.StartedAt),
		nullableTimeScanner(&mission.EndedAt),
		&mission.CreatedAt,
		&mission.UpdatedAt,
	)
	if mission.RobotCode != "" {
		mission.RobotCodes = []string{mission.RobotCode}
	}
	return mission, err
}

func scanMissionWithRobotCodes(row scanner) (domain.Mission, error) {
	var mission domain.Mission
	var robotCodesRaw string
	err := row.Scan(
		&mission.ID,
		&mission.MissionCode,
		&mission.Name,
		&mission.MissionType,
		&mission.Status,
		&mission.SiteNote,
		&robotCodesRaw,
		nullableTimeScanner(&mission.StartedAt),
		nullableTimeScanner(&mission.EndedAt),
		&mission.CreatedAt,
		&mission.UpdatedAt,
	)
	if err != nil {
		return domain.Mission{}, err
	}
	mission.RobotCodes = robotCodesFromString(robotCodesRaw)
	mission.RobotCode = firstString(mission.RobotCodes)
	return mission, nil
}

func scanMissionWithRobotID(row scanner) (domain.Mission, string, error) {
	var mission domain.Mission
	var robotID string
	err := row.Scan(
		&mission.ID,
		&mission.MissionCode,
		&mission.Name,
		&mission.MissionType,
		&mission.Status,
		&mission.SiteNote,
		&mission.RobotCode,
		nullableTimeScanner(&mission.StartedAt),
		nullableTimeScanner(&mission.EndedAt),
		&mission.CreatedAt,
		&mission.UpdatedAt,
		&robotID,
	)
	if mission.RobotCode != "" {
		mission.RobotCodes = []string{mission.RobotCode}
	}
	return mission, robotID, err
}

func scanTelemetry(row scanner) (domain.TelemetrySnapshot, error) {
	var item domain.TelemetrySnapshot
	var sentAt sql.NullTime
	var battery sql.NullFloat64
	var latitude sql.NullFloat64
	var longitude sql.NullFloat64
	var altitude sql.NullFloat64
	var accuracy sql.NullFloat64
	var heading sql.NullFloat64
	var raw []byte
	err := row.Scan(
		&item.ID,
		&item.RobotCode,
		&item.MissionID,
		&item.Sequence,
		&sentAt,
		&item.ReceivedAt,
		&battery,
		&item.NetworkState,
		&latitude,
		&longitude,
		&altitude,
		&accuracy,
		&heading,
		&raw,
	)
	if err != nil {
		return domain.TelemetrySnapshot{}, err
	}
	item.SentAt = timePointer(sentAt)
	item.BatteryPercent = floatPointerFromNull(battery)
	item.Latitude = floatPointerFromNull(latitude)
	item.Longitude = floatPointerFromNull(longitude)
	item.AltitudeMeter = floatPointerFromNull(altitude)
	item.AccuracyMeter = floatPointerFromNull(accuracy)
	item.HeadingDegree = floatPointerFromNull(heading)
	item.PositionAvailable = item.Latitude != nil && item.Longitude != nil
	item.RawPayload = append([]byte(nil), raw...)
	return item, nil
}

func scanSensorReading(row scanner) (domain.SensorReading, error) {
	var item domain.SensorReading
	var sentAt sql.NullTime
	var temperature sql.NullFloat64
	var humidity sql.NullFloat64
	var oxygen sql.NullFloat64
	var co sql.NullFloat64
	var ch4 sql.NullFloat64
	var raw []byte
	err := row.Scan(
		&item.ID,
		&item.RobotCode,
		&item.MissionID,
		&item.Sequence,
		&sentAt,
		&item.ReceivedAt,
		&temperature,
		&humidity,
		&oxygen,
		&co,
		&ch4,
		&raw,
	)
	if err != nil {
		return domain.SensorReading{}, err
	}
	item.SentAt = timePointer(sentAt)
	item.TemperatureCelsius = floatPointerFromNull(temperature)
	item.HumidityPercent = floatPointerFromNull(humidity)
	item.OxygenPercent = floatPointerFromNull(oxygen)
	item.COPpm = floatPointerFromNull(co)
	item.CH4Ppm = floatPointerFromNull(ch4)
	item.RawPayload = append([]byte(nil), raw...)
	return item, nil
}

func scanRecordingChunk(row scanner) (domain.RecordingChunk, error) {
	var chunk domain.RecordingChunk
	var metadataRaw []byte
	var metadata recordingChunkMetadata
	err := row.Scan(
		&chunk.ID,
		&chunk.RecordingSessionID,
		&chunk.MissionID,
		&chunk.MissionCode,
		&chunk.RobotCode,
		&chunk.ChunkIndex,
		&chunk.Status,
		&chunk.StartedAt,
		&chunk.EndedAt,
		&chunk.DurationSeconds,
		&metadataRaw,
		&chunk.CreatedAt,
		&chunk.UpdatedAt,
	)
	if err != nil {
		return domain.RecordingChunk{}, err
	}
	_ = json.Unmarshal(metadataRaw, &metadata)
	chunk.ManifestObjectKey = metadata.ManifestObjectKey
	chunk.MediaObjectKeys = metadata.MediaObjectKeys
	chunk.AvailableFileTypes = metadata.AvailableFileTypes
	if chunk.MediaObjectKeys == nil {
		chunk.MediaObjectKeys = map[string]string{}
	}
	if chunk.AvailableFileTypes == nil {
		chunk.AvailableFileTypes = map[string]bool{}
	}
	return chunk, nil
}

func rollbackUnlessCommitted(tx *sql.Tx) {
	_ = tx.Rollback()
}

func nullableTimeScanner(target **time.Time) any {
	return &nullableTime{target: target}
}

type nullableTime struct {
	target **time.Time
}

func (n *nullableTime) Scan(value any) error {
	if value == nil {
		*n.target = nil
		return nil
	}
	switch typedValue := value.(type) {
	case time.Time:
		copied := typedValue
		*n.target = &copied
		return nil
	default:
		return fmt.Errorf("unsupported nullable time type %T", value)
	}
}

func timePointer(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}

func floatPointerFromNull(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}
	return &value.Float64
}

func nullableFloat(value *float64) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}

func stringPointer(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func stringOrNil(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func stringFromNull(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func robotCodesFromString(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return normalizeRobotCodes(strings.Split(value, ","))
}

func sortRecordingChunks(chunks []domain.RecordingChunk) {
	sort.Slice(chunks, func(i, j int) bool {
		if chunks[i].StartedAt.Equal(chunks[j].StartedAt) {
			return chunks[i].MissionCode < chunks[j].MissionCode
		}
		return chunks[i].StartedAt.After(chunks[j].StartedAt)
	})
}
