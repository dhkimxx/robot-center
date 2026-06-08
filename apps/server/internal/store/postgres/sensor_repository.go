package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	var records []model.SensorSampleModel
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.upsertSensorDescriptors(ctx, tx, envelope, robotID); err != nil {
			return err
		}

		descriptorIDs, err := s.resolveSampleDescriptorIDs(ctx, tx, envelope, robotID)
		if err != nil {
			return err
		}

		records, err = makeSensorSampleRecords(envelope, robotID, descriptorIDs)
		if err != nil {
			return err
		}
		if len(records) == 0 {
			return nil
		}
		if err := tx.Create(&records).Error; err != nil {
			return err
		}
		return upsertLatestSensorSamples(tx, records)
	})
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return []domain.SensorSample{}, nil
	}
	return sensorSampleRecordsToDomain(records, envelope.RobotCode), nil
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
				sd.label,
				sd.sensor_type,
				sd.unit,
				sd.enabled,
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
	robotID := ""
	if strings.TrimSpace(robotCode) != "" {
		robotID, err = s.findRobotIDIncludingArchived(ctx, robotCode)
		if errors.Is(err, repo.ErrNotFound) {
			return []domain.SensorSample{}, nil
		}
		if err != nil {
			return nil, err
		}
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
				ss.sample_timestamp,
				ss.received_at,
				ss."values" AS "values",
			ss.object_key,
			ss.raw_payload
		`).
		Joins("JOIN robots r ON r.id = ss.robot_id").
		Where("ss.mission_id = ?", strings.TrimSpace(missionID)).
		Order("ss.received_at DESC").
		Limit(limit)
	if robotID != "" {
		query = query.Where("ss.robot_id = ?", robotID)
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
	robotID := ""
	if strings.TrimSpace(robotCode) != "" {
		robotID, err = s.findRobotIDIncludingArchived(ctx, robotCode)
		if errors.Is(err, repo.ErrNotFound) {
			return []domain.SensorLatest{}, nil
		}
		if err != nil {
			return nil, err
		}
	}
	var rows []sensorLatestQueryRow
	query := s.db.WithContext(ctx).
		Table("sensor_descriptors sd").
		Select(`
			sd.id AS descriptor_id,
			sd.mission_id,
			r.robot_code,
			sd.sensor_id,
				sd.channel_role AS descriptor_channel_role,
				sd.label,
				sd.sensor_type,
				sd.unit,
				sd.enabled,
			sd.first_seen_at,
			sd.last_seen_at,
			lss.sample_id,
				lss.channel_role AS sample_channel_role,
				lss.message_id,
				lss.sample_timestamp,
				lss.received_at,
				lss."values" AS "values",
			lss.object_key,
			lss.raw_payload
		`).
		Joins("JOIN robots r ON r.id = sd.robot_id").
		Joins("LEFT JOIN sensor_latest_samples lss ON lss.descriptor_id = sd.id").
		Where("sd.mission_id = ?", strings.TrimSpace(missionID)).
		Order("r.robot_code ASC, sd.sensor_id ASC")
	if robotID != "" {
		query = query.Where("sd.robot_id = ?", robotID)
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	latest := make([]domain.SensorLatest, 0, len(rows))
	for _, row := range rows {
		latest = append(latest, row.toDomain())
	}
	return latest, nil
}

func (s *Store) ClearSensorData(ctx context.Context) (repo.SensorDataClearResult, error) {
	if s.sqlTx == nil {
		tx, err := s.sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return repo.SensorDataClearResult{}, err
		}
		transactionalStore := *s
		transactionalStore.sqlTx = tx
		result, err := transactionalStore.ClearSensorData(ctx)
		if err != nil {
			_ = tx.Rollback()
			return repo.SensorDataClearResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return repo.SensorDataClearResult{}, err
		}
		return result, nil
	}

	runner := s.sqlRunner()
	if _, err := runner.ExecContext(ctx, `LOCK TABLE sensor_latest_samples, sensor_samples, sensor_descriptors IN ACCESS EXCLUSIVE MODE`); err != nil {
		return repo.SensorDataClearResult{}, err
	}
	sensorLatestSamplesDeleted, sensorSamplesDeleted, sensorDescriptorsDeleted, err := countSensorDataRows(ctx, runner)
	if err != nil {
		return repo.SensorDataClearResult{}, err
	}
	if _, err := runner.ExecContext(ctx, `TRUNCATE TABLE sensor_latest_samples, sensor_samples, sensor_descriptors`); err != nil {
		return repo.SensorDataClearResult{}, err
	}
	return repo.SensorDataClearResult{
		SensorLatestSamplesDeleted: sensorLatestSamplesDeleted,
		SensorSamplesDeleted:       sensorSamplesDeleted,
		SensorDescriptorsDeleted:   sensorDescriptorsDeleted,
	}, nil
}

func countSensorDataRows(ctx context.Context, runner sqlContextRunner) (int64, int64, int64, error) {
	var latestCount int64
	var sampleCount int64
	var descriptorCount int64
	err := runner.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM sensor_latest_samples),
			(SELECT COUNT(*) FROM sensor_samples),
			(SELECT COUNT(*) FROM sensor_descriptors)
	`).Scan(&latestCount, &sampleCount, &descriptorCount)
	if err != nil {
		return 0, 0, 0, err
	}
	return latestCount, sampleCount, descriptorCount, nil
}

func (s *Store) upsertSensorDescriptors(ctx context.Context, db *gorm.DB, envelope domain.SensorEnvelope, robotID string) error {
	if len(envelope.Descriptors) == 0 {
		return nil
	}
	records := make([]model.SensorDescriptorModel, 0, len(envelope.Descriptors))
	for _, descriptor := range envelope.Descriptors {
		descriptor.SensorID = strings.TrimSpace(descriptor.SensorID)
		if descriptor.SensorID == "" {
			continue
		}
		channelRole := utils.FirstNonEmptyString(descriptor.ChannelRole, envelope.ChannelRole, "channel.telemetry")
		label := utils.FirstNonEmptyString(descriptor.Label, descriptor.SensorID)
		sensorType := domain.NormalizeSensorType(descriptor.SensorType, descriptor.SensorID)
		records = append(records, model.SensorDescriptorModel{
			MissionID:   envelope.MissionID,
			RobotID:     robotID,
			SensorID:    descriptor.SensorID,
			ChannelRole: channelRole,
			Label:       label,
			SensorType:  sensorType,
			Unit:        optionalString(descriptor.Unit),
			Enabled:     descriptor.Enabled,
			FirstSeenAt: envelope.ReceivedAt,
			LastSeenAt:  envelope.ReceivedAt,
		})
	}
	if len(records) == 0 {
		return nil
	}
	if err := db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "mission_id"},
			{Name: "robot_id"},
			{Name: "sensor_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"channel_role",
			"label",
			"sensor_type",
			"unit",
			"enabled",
			"last_seen_at",
		}),
	}).Create(&records).Error; err != nil {
		return err
	}
	return nil
}

func (s *Store) resolveSampleDescriptorIDs(ctx context.Context, db *gorm.DB, envelope domain.SensorEnvelope, robotID string) (map[string]string, error) {
	seenSensorIDs := map[string]struct{}{}
	sensorIDs := make([]string, 0, len(envelope.Samples))
	for _, sample := range envelope.Samples {
		sensorID := strings.TrimSpace(sample.SensorID)
		if sensorID == "" {
			continue
		}
		if _, exists := seenSensorIDs[sensorID]; exists {
			continue
		}
		seenSensorIDs[sensorID] = struct{}{}
		sensorIDs = append(sensorIDs, sensorID)
	}
	if len(sensorIDs) == 0 {
		return map[string]string{}, nil
	}

	descriptorIDs, err := findSensorDescriptorIDs(db.WithContext(ctx), envelope.MissionID, robotID, sensorIDs)
	if err != nil {
		return nil, err
	}

	for _, sensorID := range sensorIDs {
		if descriptorIDs[sensorID] != "" {
			continue
		}
		return nil, fmt.Errorf("%w: sensor descriptor is required before sample sensorId %q", repo.ErrInvalidState, sensorID)
	}
	return descriptorIDs, nil
}

func makeSensorSampleRecords(envelope domain.SensorEnvelope, robotID string, descriptorIDs map[string]string) ([]model.SensorSampleModel, error) {
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
			Timestamp:    sample.Timestamp,
			ReceivedAt:   receivedAt,
			Values:       emptyJSONToNil(sample.Values),
			ObjectKey:    optionalString(sample.ObjectKey),
			RawPayload:   rawPayload,
		})
	}
	return records, nil
}

func upsertLatestSensorSamples(db *gorm.DB, samples []model.SensorSampleModel) error {
	latestRecords := make([]model.SensorLatestSampleModel, 0, len(samples))
	for _, sample := range samples {
		latestRecords = append(latestRecords, model.SensorLatestSampleModel{
			SampleID:     sample.ID,
			DescriptorID: sample.DescriptorID,
			MissionID:    sample.MissionID,
			RobotID:      sample.RobotID,
			SensorID:     sample.SensorID,
			ChannelRole:  sample.ChannelRole,
			MessageID:    sample.MessageID,
			Timestamp:    sample.Timestamp,
			ReceivedAt:   sample.ReceivedAt,
			Values:       sample.Values,
			ObjectKey:    sample.ObjectKey,
			RawPayload:   sample.RawPayload,
		})
	}
	if len(latestRecords) == 0 {
		return nil
	}
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "mission_id"},
			{Name: "robot_id"},
			{Name: "sensor_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"sample_id",
			"descriptor_id",
			"channel_role",
			"message_id",
			"sample_timestamp",
			"received_at",
			"values",
			"object_key",
			"raw_payload",
			"updated_at",
		}),
		Where: clause.Where{Exprs: []clause.Expression{
			clause.Expr{SQL: "sensor_latest_samples.received_at <= EXCLUDED.received_at"},
		}},
	}).Create(&latestRecords).Error
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
	ID          string
	MissionID   string
	RobotCode   string
	SensorID    string
	ChannelRole string
	Label       string
	SensorType  string
	Unit        *string
	Enabled     bool
	FirstSeenAt time.Time
	LastSeenAt  time.Time
}

func (row sensorDescriptorQueryRow) toDomain() domain.SensorDescriptor {
	return domain.SensorDescriptor{
		ID:          row.ID,
		MissionID:   row.MissionID,
		RobotCode:   row.RobotCode,
		SensorID:    row.SensorID,
		ChannelRole: row.ChannelRole,
		Label:       row.Label,
		SensorType:  row.SensorType,
		Unit:        stringFromPointer(row.Unit),
		Enabled:     row.Enabled,
		FirstSeenAt: row.FirstSeenAt,
		LastSeenAt:  row.LastSeenAt,
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
	Timestamp    *time.Time `gorm:"column:sample_timestamp"`
	ReceivedAt   time.Time
	Values       json.RawMessage
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
		Timestamp:    row.Timestamp,
		ReceivedAt:   row.ReceivedAt,
		Values:       row.Values,
		ObjectKey:    stringFromPointer(row.ObjectKey),
		RawPayload:   row.RawPayload,
	}
}

func sensorSampleRecordsToDomain(records []model.SensorSampleModel, robotCode string) []domain.SensorSample {
	samples := make([]domain.SensorSample, 0, len(records))
	for _, record := range records {
		samples = append(samples, domain.SensorSample{
			ID:           record.ID,
			DescriptorID: record.DescriptorID,
			MissionID:    record.MissionID,
			RobotCode:    robotCode,
			SensorID:     record.SensorID,
			ChannelRole:  record.ChannelRole,
			MessageID:    stringFromPointer(record.MessageID),
			Timestamp:    record.Timestamp,
			ReceivedAt:   record.ReceivedAt,
			Values:       record.Values,
			ObjectKey:    stringFromPointer(record.ObjectKey),
			RawPayload:   record.RawPayload,
		})
	}
	return samples
}

type sensorLatestQueryRow struct {
	DescriptorID          string
	MissionID             string
	RobotCode             string
	SensorID              string
	DescriptorChannelRole string
	Label                 string
	SensorType            string
	Unit                  *string
	Enabled               bool
	FirstSeenAt           time.Time
	LastSeenAt            time.Time
	SampleID              *string
	SampleChannelRole     *string
	MessageID             *string
	Timestamp             *time.Time `gorm:"column:sample_timestamp"`
	ReceivedAt            *time.Time
	Values                json.RawMessage
	ObjectKey             *string
	RawPayload            json.RawMessage
}

func (row sensorLatestQueryRow) toDomain() domain.SensorLatest {
	descriptor := domain.SensorDescriptor{
		ID:          row.DescriptorID,
		MissionID:   row.MissionID,
		RobotCode:   row.RobotCode,
		SensorID:    row.SensorID,
		ChannelRole: row.DescriptorChannelRole,
		Label:       row.Label,
		SensorType:  row.SensorType,
		Unit:        stringFromPointer(row.Unit),
		Enabled:     row.Enabled,
		FirstSeenAt: row.FirstSeenAt,
		LastSeenAt:  row.LastSeenAt,
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
		Timestamp:    row.Timestamp,
		ReceivedAt:   receivedAt,
		Values:       row.Values,
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

func emptyJSONToNil(value json.RawMessage) json.RawMessage {
	if len(value) == 0 {
		return nil
	}
	return value
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
