package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store/model"
	repo "robot-center/apps/server/internal/store/port"
)

const (
	defaultMissionEventLimit = 100
	maxMissionEventLimit     = 500
)

func (s *Store) SaveMissionEventEnvelope(ctx context.Context, envelope domain.MissionEventEnvelope) ([]domain.MissionEvent, error) {
	envelope.RobotCode = strings.TrimSpace(envelope.RobotCode)
	envelope.MissionID = strings.TrimSpace(envelope.MissionID)
	if envelope.RobotCode == "" || envelope.MissionID == "" || len(envelope.Events) == 0 {
		return nil, repo.ErrInvalidState
	}
	missionID, err := s.resolveMissionID(ctx, envelope.MissionID)
	if err != nil {
		return nil, err
	}
	robotID, err := s.findRobotID(ctx, envelope.RobotCode)
	if err != nil {
		return nil, err
	}
	receivedAt := envelope.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = time.Now().UTC()
	}
	if len(envelope.RawMessage) == 0 {
		envelope.RawMessage = []byte("{}")
	}

	records := make([]model.EventModel, 0, len(envelope.Events))
	for _, event := range envelope.Events {
		event.EventType = strings.TrimSpace(event.EventType)
		if event.EventType == "" {
			continue
		}
		timestamp := event.Timestamp
		if timestamp.IsZero() {
			timestamp = receivedAt
		}
		records = append(records, model.EventModel{
			MissionID:      missionID,
			RobotID:        &robotID,
			EventID:        optionalString(event.EventID),
			EventType:      event.EventType,
			EventCategory:  normalizeEventCategory(event.EventType, event.EventCategory),
			TrackID:        optionalString(event.TrackID),
			Severity:       normalizeEventSeverity(event.Severity),
			Title:          normalizeEventTitle(event.Title, event.EventType),
			Description:    optionalString(event.Description),
			OccurredAt:     timestamp.UTC(),
			ReceivedAt:     receivedAt,
			DetectionCount: event.DetectionCount,
			Values:         jsonWithDefault(event.Values),
			RawMessage:     jsonWithDefault(envelope.RawMessage),
			RawPayload:     jsonWithDefault(envelope.RawMessage),
		})
	}
	if len(records) == 0 {
		return []domain.MissionEvent{}, nil
	}
	if err := s.db.WithContext(ctx).Create(&records).Error; err != nil {
		return nil, err
	}
	return eventRecordsToDomain(records, envelope.RobotCode), nil
}

func (s *Store) ListMissionEvents(ctx context.Context, query repo.EventQuery) ([]domain.MissionEvent, error) {
	missionID, err := s.resolveMissionID(ctx, query.MissionID)
	if err != nil {
		return nil, err
	}
	robotID := ""
	if strings.TrimSpace(query.RobotCode) != "" {
		robotID, err = s.findRobotIDIncludingArchived(ctx, query.RobotCode)
		if errors.Is(err, repo.ErrNotFound) {
			return []domain.MissionEvent{}, nil
		}
		if err != nil {
			return nil, err
		}
	}
	limit := query.Limit
	if limit <= 0 || limit > maxMissionEventLimit {
		limit = defaultMissionEventLimit
	}

	var rows []eventQueryRow
	db := s.db.WithContext(ctx).
		Table("events e").
		Select(`
			e.id,
			e.mission_id,
			r.robot_code,
			e.event_id,
			e.event_type,
			e.event_category,
			e.track_id,
			e.severity,
			e.title,
			e.description,
			e.occurred_at,
			e.received_at,
			e.detection_count,
			e."values",
			e.raw_message,
			e.created_at,
			e.updated_at
		`).
		Joins("LEFT JOIN robots r ON r.id = e.robot_id").
		Where("e.mission_id = ?", strings.TrimSpace(missionID)).
		Order("e.occurred_at DESC, e.created_at DESC").
		Limit(limit)
	if robotID != "" {
		db = db.Where("e.robot_id = ?", robotID)
	}
	eventType := strings.TrimSpace(query.EventType)
	eventCategory := strings.TrimSpace(query.EventCategory)
	trackID := strings.TrimSpace(query.TrackID)
	if eventType != "" {
		db = db.Where("e.event_type = ?", eventType)
	}
	if eventCategory != "" {
		db = db.Where("e.event_category = ?", eventCategory)
	}
	if trackID != "" {
		db = db.Where("e.track_id = ?", trackID)
	}
	if !query.IncludeDetections && eventType == "" && eventCategory == "" && trackID == "" {
		db = db.Where("e.event_type <> ?", domain.EventTypeDetectionObject)
	}
	if err := db.Scan(&rows).Error; err != nil {
		return nil, err
	}
	events := make([]domain.MissionEvent, 0, len(rows))
	for _, row := range rows {
		events = append(events, row.toDomain())
	}
	return events, nil
}

type eventQueryRow struct {
	ID             string
	MissionID      string
	RobotCode      *string
	EventID        *string
	EventType      string
	EventCategory  string
	TrackID        *string
	Severity       string
	Title          string
	Description    *string
	OccurredAt     time.Time
	ReceivedAt     time.Time
	DetectionCount *int
	Values         json.RawMessage
	RawMessage     json.RawMessage
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (row eventQueryRow) toDomain() domain.MissionEvent {
	return domain.MissionEvent{
		ID:             row.ID,
		MissionID:      row.MissionID,
		RobotCode:      stringFromPointer(row.RobotCode),
		EventID:        stringFromPointer(row.EventID),
		EventType:      row.EventType,
		EventCategory:  row.EventCategory,
		TrackID:        stringFromPointer(row.TrackID),
		Severity:       row.Severity,
		Title:          row.Title,
		Description:    stringFromPointer(row.Description),
		Timestamp:      row.OccurredAt,
		ReceivedAt:     row.ReceivedAt,
		DetectionCount: row.DetectionCount,
		Values:         jsonWithDefault(row.Values),
		RawMessage:     jsonWithDefault(row.RawMessage),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

func eventRecordsToDomain(records []model.EventModel, robotCode string) []domain.MissionEvent {
	events := make([]domain.MissionEvent, 0, len(records))
	for _, record := range records {
		events = append(events, domain.MissionEvent{
			ID:             record.ID,
			MissionID:      record.MissionID,
			RobotCode:      robotCode,
			EventID:        stringFromPointer(record.EventID),
			EventType:      record.EventType,
			EventCategory:  record.EventCategory,
			TrackID:        stringFromPointer(record.TrackID),
			Severity:       record.Severity,
			Title:          record.Title,
			Description:    stringFromPointer(record.Description),
			Timestamp:      record.OccurredAt,
			ReceivedAt:     record.ReceivedAt,
			DetectionCount: record.DetectionCount,
			Values:         jsonWithDefault(record.Values),
			RawMessage:     jsonWithDefault(record.RawMessage),
			CreatedAt:      record.CreatedAt,
			UpdatedAt:      record.UpdatedAt,
		})
	}
	return events
}

func normalizeEventCategory(eventType string, eventCategory string) string {
	if strings.TrimSpace(eventCategory) != "" {
		return strings.TrimSpace(eventCategory)
	}
	switch {
	case eventType == domain.EventTypeDetectionObject:
		return domain.EventCategoryDetection
	case eventType == domain.EventTypeMissionEvent:
		return domain.EventCategoryMission
	case strings.HasPrefix(eventType, "alarm."):
		return domain.EventCategoryAlarm
	case strings.HasPrefix(eventType, "system."):
		return domain.EventCategorySystem
	default:
		return domain.EventCategoryMission
	}
}

func normalizeEventSeverity(severity string) string {
	switch strings.TrimSpace(severity) {
	case "notice", "warning", "critical":
		return strings.TrimSpace(severity)
	default:
		return "info"
	}
}

func normalizeEventTitle(title string, eventType string) string {
	if strings.TrimSpace(title) != "" {
		return strings.TrimSpace(title)
	}
	return strings.TrimSpace(eventType)
}

func jsonWithDefault(value json.RawMessage) json.RawMessage {
	if len(value) == 0 || !json.Valid(value) {
		return json.RawMessage(`{}`)
	}
	return append(json.RawMessage(nil), value...)
}
