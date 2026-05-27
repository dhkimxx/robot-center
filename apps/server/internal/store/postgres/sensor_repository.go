package postgres

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store/model"
	repo "robot-center/apps/server/internal/store/port"
	"robot-center/apps/server/internal/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *Store) SaveSensorEnvelope(ctx context.Context, envelope domain.SensorEnvelope) ([]domain.SensorSample, error) {
	envelope.RobotCode = strings.TrimSpace(envelope.RobotCode)
	envelope.MissionID = strings.TrimSpace(envelope.MissionID)
	envelope.ChannelRole = strings.TrimSpace(envelope.ChannelRole)
	if envelope.RobotCode == "" || envelope.MissionID == "" {
		return nil, repo.ErrInvalidState
	}
	missionID, err := s.resolveMissionID(ctx, envelope.MissionID)
	if err != nil {
		return nil, err
	}
	envelope.MissionID = missionID
	if envelope.ReceivedAt.IsZero() {
		envelope.ReceivedAt = time.Now().UTC()
	}
	if len(envelope.RawPayload) == 0 {
		envelope.RawPayload = []byte("{}")
	}
	robotID, err := s.findRobotID(ctx, envelope.RobotCode)
	if err != nil {
		return nil, err
	}

	db := s.db.WithContext(ctx)
	descriptorIDs := map[string]string{}
	if len(envelope.Descriptors) > 0 {
		records := make([]model.SensorDescriptorModel, 0, len(envelope.Descriptors))
		for _, descriptor := range envelope.Descriptors {
			descriptor.SensorID = strings.TrimSpace(descriptor.SensorID)
			if descriptor.SensorID == "" {
				continue
			}
			channelRole := utils.FirstNonEmptyString(descriptor.ChannelRole, envelope.ChannelRole, "channel.telemetry")
			metadata := descriptor.Metadata
			if len(metadata) == 0 {
				metadata = []byte("{}")
			}
			displayName := utils.FirstNonEmptyString(descriptor.DisplayName, descriptor.SensorID)
			sensorType := utils.FirstNonEmptyString(descriptor.SensorType, "unknown")
			valueType := utils.FirstNonEmptyString(descriptor.ValueType, "object")
			records = append(records, model.SensorDescriptorModel{
				MissionID:    envelope.MissionID,
				RobotID:      robotID,
				SensorID:     descriptor.SensorID,
				ChannelRole:  channelRole,
				DisplayName:  displayName,
				SensorType:   sensorType,
				ValueType:    valueType,
				Unit:         optionalString(descriptor.Unit),
				SampleRateHz: descriptor.SampleRateHz,
				Enabled:      descriptor.Enabled,
				Metadata:     metadata,
				FirstSeenAt:  envelope.ReceivedAt,
				LastSeenAt:   envelope.ReceivedAt,
			})
		}
		if len(records) > 0 {
			if err := db.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "mission_id"},
					{Name: "robot_id"},
					{Name: "sensor_id"},
				},
				DoUpdates: clause.AssignmentColumns([]string{
					"channel_role",
					"display_name",
					"sensor_type",
					"value_type",
					"unit",
					"sample_rate_hz",
					"enabled",
					"metadata",
					"last_seen_at",
				}),
			}).Create(&records).Error; err != nil {
				return nil, err
			}
			sensorIDs := make([]string, 0, len(records))
			for _, record := range records {
				sensorIDs = append(sensorIDs, record.SensorID)
			}
			var persisted []model.SensorDescriptorModel
			if err := db.
				Where("mission_id = ? AND robot_id = ? AND sensor_id IN ?", envelope.MissionID, robotID, sensorIDs).
				Find(&persisted).Error; err != nil {
				return nil, err
			}
			for _, descriptor := range persisted {
				descriptorIDs[descriptor.SensorID] = descriptor.ID
			}
		}
	}
	sampleDescriptorIDs, err := s.ensureSensorDescriptorIDsForSamples(ctx, envelope, robotID)
	if err != nil {
		return nil, err
	}
	for sensorID, descriptorID := range sampleDescriptorIDs {
		descriptorIDs[sensorID] = descriptorID
	}

	records := make([]model.SensorSampleModel, 0, len(envelope.Samples))
	for _, sample := range envelope.Samples {
		sample.SensorID = strings.TrimSpace(sample.SensorID)
		if sample.SensorID == "" {
			continue
		}
		descriptorID := descriptorIDs[sample.SensorID]
		if descriptorID == "" {
			return nil, repo.ErrInvalidState
		}
		receivedAt := sample.ReceivedAt
		if receivedAt.IsZero() {
			receivedAt = envelope.ReceivedAt
		}
		rawPayload := sample.RawPayload
		if len(rawPayload) == 0 {
			rawPayload = envelope.RawPayload
		}
		records = append(records, model.SensorSampleModel{
			DescriptorID: descriptorID,
			MissionID:    envelope.MissionID,
			RobotID:      robotID,
			SensorID:     sample.SensorID,
			ChannelRole:  utils.FirstNonEmptyString(sample.ChannelRole, envelope.ChannelRole, "channel.telemetry"),
			MessageID:    optionalString(utils.FirstNonEmptyString(sample.MessageID, envelope.MessageID)),
			Sequence:     optionalInt64(utils.FirstNonZeroInt64(sample.Sequence, envelope.Sequence)),
			SentAt:       utils.FirstTimePointer(sample.SentAt, envelope.SentAt),
			ReceivedAt:   receivedAt,
			NumericValue: sample.NumericValue,
			TextValue:    optionalString(sample.TextValue),
			BoolValue:    sample.BoolValue,
			VectorValue:  emptyJSONToNil(sample.VectorValue),
			ObjectValue:  emptyJSONToNil(sample.ObjectValue),
			ObjectKey:    optionalString(sample.ObjectKey),
			RawPayload:   rawPayload,
		})
	}
	if len(records) == 0 {
		return []domain.SensorSample{}, nil
	}
	if err := db.Create(&records).Error; err != nil {
		return nil, err
	}
	return s.ListSensorSamples(ctx, envelope.MissionID, envelope.RobotCode, "", len(records))
}

func (s *Store) ListSensorDescriptors(ctx context.Context, missionID string, robotCode string) ([]domain.SensorDescriptor, error) {
	missionID, err := s.resolveMissionID(ctx, missionID)
	if err != nil {
		return nil, err
	}
	var rows []sensorDescriptorQueryRow
	query := s.db.WithContext(ctx).
		Table("sensor_descriptors sd").
		Select(`
			sd.id,
			sd.mission_id,
			r.robot_code,
			sd.sensor_id,
			sd.channel_role,
			sd.display_name,
			sd.sensor_type,
			sd.value_type,
			sd.unit,
			sd.sample_rate_hz,
			sd.enabled,
			sd.metadata,
			sd.first_seen_at,
			sd.last_seen_at
		`).
		Joins("JOIN robots r ON r.id = sd.robot_id").
		Where("sd.mission_id = ?", strings.TrimSpace(missionID)).
		Order("r.robot_code ASC, sd.sensor_id ASC")
	if strings.TrimSpace(robotCode) != "" {
		query = query.Where("r.robot_code = ?", strings.TrimSpace(robotCode))
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	descriptors := make([]domain.SensorDescriptor, 0, len(rows))
	for _, row := range rows {
		descriptors = append(descriptors, row.toDomain())
	}
	return descriptors, nil
}

func (s *Store) ListSensorSamples(ctx context.Context, missionID string, robotCode string, sensorID string, limit int) ([]domain.SensorSample, error) {
	missionID, err := s.resolveMissionID(ctx, missionID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var rows []sensorSampleQueryRow
	query := s.db.WithContext(ctx).
		Table("sensor_samples ss").
		Select(`
			ss.id,
			ss.descriptor_id,
			ss.mission_id,
			r.robot_code,
			ss.sensor_id,
			ss.channel_role,
			ss.message_id,
			ss.sequence,
			ss.sent_at,
			ss.received_at,
			ss.numeric_value,
			ss.text_value,
			ss.bool_value,
			ss.vector_value,
			ss.object_value,
			ss.object_key,
			ss.raw_payload
		`).
		Joins("JOIN robots r ON r.id = ss.robot_id").
		Where("ss.mission_id = ?", strings.TrimSpace(missionID)).
		Order("ss.received_at DESC").
		Limit(limit)
	if strings.TrimSpace(robotCode) != "" {
		query = query.Where("r.robot_code = ?", strings.TrimSpace(robotCode))
	}
	if strings.TrimSpace(sensorID) != "" {
		query = query.Where("ss.sensor_id = ?", strings.TrimSpace(sensorID))
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	samples := make([]domain.SensorSample, 0, len(rows))
	for _, row := range rows {
		samples = append(samples, row.toDomain())
	}
	return samples, nil
}

func (s *Store) ListLatestSensorSamples(ctx context.Context, missionID string, robotCode string) ([]domain.SensorLatest, error) {
	missionID, err := s.resolveMissionID(ctx, missionID)
	if err != nil {
		return nil, err
	}
	var rows []sensorLatestQueryRow
	query := s.db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (r.robot_code, sd.sensor_id)
			sd.id AS descriptor_id,
			sd.mission_id,
			r.robot_code,
			sd.sensor_id,
			sd.channel_role AS descriptor_channel_role,
			sd.display_name,
			sd.sensor_type,
			sd.value_type,
			sd.unit,
			sd.sample_rate_hz,
			sd.enabled,
			sd.metadata,
			sd.first_seen_at,
			sd.last_seen_at,
			ss.id AS sample_id,
			ss.channel_role AS sample_channel_role,
			ss.message_id,
			ss.sequence,
			ss.sent_at,
			ss.received_at,
			ss.numeric_value,
			ss.text_value,
			ss.bool_value,
			ss.vector_value,
			ss.object_value,
			ss.object_key,
			ss.raw_payload
		FROM sensor_descriptors sd
		JOIN robots r ON r.id = sd.robot_id
		LEFT JOIN sensor_samples ss ON ss.descriptor_id = sd.id
		WHERE sd.mission_id = ?
			AND (? = '' OR r.robot_code = ?)
		ORDER BY r.robot_code, sd.sensor_id, ss.received_at DESC NULLS LAST
	`, strings.TrimSpace(missionID), strings.TrimSpace(robotCode), strings.TrimSpace(robotCode))
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	latest := make([]domain.SensorLatest, 0, len(rows))
	for _, row := range rows {
		latest = append(latest, row.toDomain())
	}
	return latest, nil
}

func (s *Store) ensureSensorDescriptorIDsForSamples(ctx context.Context, envelope domain.SensorEnvelope, robotID string) (map[string]string, error) {
	samplesBySensorID := map[string]domain.SensorSample{}
	sensorIDs := make([]string, 0, len(envelope.Samples))
	for _, sample := range envelope.Samples {
		sensorID := strings.TrimSpace(sample.SensorID)
		if sensorID == "" {
			continue
		}
		if _, exists := samplesBySensorID[sensorID]; exists {
			continue
		}
		sample.SensorID = sensorID
		samplesBySensorID[sensorID] = sample
		sensorIDs = append(sensorIDs, sensorID)
	}
	if len(sensorIDs) == 0 {
		return map[string]string{}, nil
	}

	db := s.db.WithContext(ctx)
	descriptorIDs, err := findSensorDescriptorIDs(db, envelope.MissionID, robotID, sensorIDs)
	if err != nil {
		return nil, err
	}

	records := make([]model.SensorDescriptorModel, 0)
	for _, sensorID := range sensorIDs {
		if descriptorIDs[sensorID] != "" {
			continue
		}
		sample := samplesBySensorID[sensorID]
		records = append(records, model.SensorDescriptorModel{
			MissionID:   envelope.MissionID,
			RobotID:     robotID,
			SensorID:    sensorID,
			ChannelRole: utils.FirstNonEmptyString(sample.ChannelRole, envelope.ChannelRole, "channel.telemetry"),
			DisplayName: sensorID,
			SensorType:  domain.InferSensorTypeFromID(sensorID),
			ValueType:   domain.InferSensorValueType(sample),
			Enabled:     true,
			Metadata:    []byte(`{"source":"auto-sample"}`),
			FirstSeenAt: envelope.ReceivedAt,
			LastSeenAt:  envelope.ReceivedAt,
		})
	}
	if len(records) > 0 {
		if err := db.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "mission_id"},
				{Name: "robot_id"},
				{Name: "sensor_id"},
			},
			DoUpdates: clause.AssignmentColumns([]string{"last_seen_at"}),
		}).Create(&records).Error; err != nil {
			return nil, err
		}
		descriptorIDs, err = findSensorDescriptorIDs(db, envelope.MissionID, robotID, sensorIDs)
		if err != nil {
			return nil, err
		}
	}
	return descriptorIDs, nil
}

func findSensorDescriptorIDs(db *gorm.DB, missionID string, robotID string, sensorIDs []string) (map[string]string, error) {
	var persisted []model.SensorDescriptorModel
	if err := db.
		Where("mission_id = ? AND robot_id = ? AND sensor_id IN ?", missionID, robotID, sensorIDs).
		Find(&persisted).Error; err != nil {
		return nil, err
	}
	descriptorIDs := make(map[string]string, len(persisted))
	for _, descriptor := range persisted {
		descriptorIDs[descriptor.SensorID] = descriptor.ID
	}
	return descriptorIDs, nil
}

type sensorDescriptorQueryRow struct {
	ID           string
	MissionID    string
	RobotCode    string
	SensorID     string
	ChannelRole  string
	DisplayName  string
	SensorType   string
	ValueType    string
	Unit         *string
	SampleRateHz *float64
	Enabled      bool
	Metadata     json.RawMessage
	FirstSeenAt  time.Time
	LastSeenAt   time.Time
}

func (row sensorDescriptorQueryRow) toDomain() domain.SensorDescriptor {
	return domain.SensorDescriptor{
		ID:           row.ID,
		MissionID:    row.MissionID,
		RobotCode:    row.RobotCode,
		SensorID:     row.SensorID,
		ChannelRole:  row.ChannelRole,
		DisplayName:  row.DisplayName,
		SensorType:   row.SensorType,
		ValueType:    row.ValueType,
		Unit:         stringFromPointer(row.Unit),
		SampleRateHz: row.SampleRateHz,
		Enabled:      row.Enabled,
		Metadata:     row.Metadata,
		FirstSeenAt:  row.FirstSeenAt,
		LastSeenAt:   row.LastSeenAt,
	}
}

type sensorSampleQueryRow struct {
	ID           string
	DescriptorID *string
	MissionID    string
	RobotCode    string
	SensorID     string
	ChannelRole  string
	MessageID    *string
	Sequence     *int64
	SentAt       *time.Time
	ReceivedAt   time.Time
	NumericValue *float64
	TextValue    *string
	BoolValue    *bool
	VectorValue  json.RawMessage
	ObjectValue  json.RawMessage
	ObjectKey    *string
	RawPayload   json.RawMessage
}

func (row sensorSampleQueryRow) toDomain() domain.SensorSample {
	return domain.SensorSample{
		ID:           row.ID,
		DescriptorID: stringFromPointer(row.DescriptorID),
		MissionID:    row.MissionID,
		RobotCode:    row.RobotCode,
		SensorID:     row.SensorID,
		ChannelRole:  row.ChannelRole,
		MessageID:    stringFromPointer(row.MessageID),
		Sequence:     int64FromPointer(row.Sequence),
		SentAt:       row.SentAt,
		ReceivedAt:   row.ReceivedAt,
		NumericValue: row.NumericValue,
		TextValue:    stringFromPointer(row.TextValue),
		BoolValue:    row.BoolValue,
		VectorValue:  row.VectorValue,
		ObjectValue:  row.ObjectValue,
		ObjectKey:    stringFromPointer(row.ObjectKey),
		RawPayload:   row.RawPayload,
	}
}

type sensorLatestQueryRow struct {
	DescriptorID          string
	MissionID             string
	RobotCode             string
	SensorID              string
	DescriptorChannelRole string
	DisplayName           string
	SensorType            string
	ValueType             string
	Unit                  *string
	SampleRateHz          *float64
	Enabled               bool
	Metadata              json.RawMessage
	FirstSeenAt           time.Time
	LastSeenAt            time.Time
	SampleID              *string
	SampleChannelRole     *string
	MessageID             *string
	Sequence              *int64
	SentAt                *time.Time
	ReceivedAt            *time.Time
	NumericValue          *float64
	TextValue             *string
	BoolValue             *bool
	VectorValue           json.RawMessage
	ObjectValue           json.RawMessage
	ObjectKey             *string
	RawPayload            json.RawMessage
}

func (row sensorLatestQueryRow) toDomain() domain.SensorLatest {
	descriptor := domain.SensorDescriptor{
		ID:           row.DescriptorID,
		MissionID:    row.MissionID,
		RobotCode:    row.RobotCode,
		SensorID:     row.SensorID,
		ChannelRole:  row.DescriptorChannelRole,
		DisplayName:  row.DisplayName,
		SensorType:   row.SensorType,
		ValueType:    row.ValueType,
		Unit:         stringFromPointer(row.Unit),
		SampleRateHz: row.SampleRateHz,
		Enabled:      row.Enabled,
		Metadata:     row.Metadata,
		FirstSeenAt:  row.FirstSeenAt,
		LastSeenAt:   row.LastSeenAt,
	}
	if row.SampleID == nil {
		return domain.SensorLatest{Descriptor: descriptor}
	}
	receivedAt := time.Time{}
	if row.ReceivedAt != nil {
		receivedAt = *row.ReceivedAt
	}
	sample := domain.SensorSample{
		ID:           stringFromPointer(row.SampleID),
		DescriptorID: row.DescriptorID,
		MissionID:    row.MissionID,
		RobotCode:    row.RobotCode,
		SensorID:     row.SensorID,
		ChannelRole:  stringFromPointer(row.SampleChannelRole),
		MessageID:    stringFromPointer(row.MessageID),
		Sequence:     int64FromPointer(row.Sequence),
		SentAt:       row.SentAt,
		ReceivedAt:   receivedAt,
		NumericValue: row.NumericValue,
		TextValue:    stringFromPointer(row.TextValue),
		BoolValue:    row.BoolValue,
		VectorValue:  row.VectorValue,
		ObjectValue:  row.ObjectValue,
		ObjectKey:    stringFromPointer(row.ObjectKey),
		RawPayload:   row.RawPayload,
	}
	return domain.SensorLatest{Descriptor: descriptor, LatestSample: &sample}
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func optionalInt64(value int64) *int64 {
	if value == 0 {
		return nil
	}
	return &value
}

func emptyJSONToNil(value json.RawMessage) json.RawMessage {
	if len(value) == 0 {
		return nil
	}
	return value
}

func int64FromPointer(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func (s *Store) resolveMissionID(ctx context.Context, missionIDOrCode string) (string, error) {
	missionIDOrCode = strings.TrimSpace(missionIDOrCode)
	if missionIDOrCode == "" {
		return "", repo.ErrInvalidState
	}
	var missionID string
	err := s.sqlDB.QueryRowContext(ctx, `
		SELECT id::text
		FROM missions
		WHERE id::text = $1 OR mission_code = $1
		LIMIT 1
	`, missionIDOrCode).Scan(&missionID)
	if err != nil {
		return "", err
	}
	return missionID, nil
}
