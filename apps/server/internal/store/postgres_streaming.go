package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"robot-center/apps/server/internal/domain"
)

func (s *PostgresStore) ApplyStreamingStatus(ctx context.Context, status domain.StreamingStatus, bearerToken string) (domain.Robot, error) {
	tracksJSON, err := json.Marshal(status.PublishedTracks)
	if err != nil {
		return domain.Robot{}, err
	}
	channelsJSON, err := json.Marshal(status.PublishedDataChannels)
	if err != nil {
		return domain.Robot{}, err
	}
	if status.SentAt.IsZero() {
		status.SentAt = time.Now().UTC()
	}
	status.RobotCode = strings.TrimSpace(status.RobotCode)
	status.MissionID = strings.TrimSpace(status.MissionID)
	status.RoomID = strings.TrimSpace(status.RoomID)
	status.Status = strings.TrimSpace(status.Status)

	var robot domain.Robot
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		authorizedRobot, err := s.authorizeRobotWithGorm(tx, status.RobotCode, bearerToken)
		if err != nil {
			return err
		}
		if isPublishingStatus(status.Status) {
			missionCode, err := s.findActiveStreamingMissionCodeForRobot(tx, authorizedRobot.ID, status.MissionID)
			if err != nil {
				return err
			}
			if status.RoomID != missionCode {
				return ErrInvalidState
			}
		}

		statusApplied := true
		if isTerminalStreamingStatus(status.Status) {
			applied, err := s.updateMatchingTerminalStreamingStatus(tx, authorizedRobot.ID, status, tracksJSON, channelsJSON)
			if err != nil {
				return err
			}
			statusApplied = applied
		} else {
			if err := s.upsertStreamingStatus(tx, authorizedRobot.ID, status, tracksJSON, channelsJSON); err != nil {
				return err
			}
		}

		if statusApplied {
			if err := tx.Model(&robotRecord{}).
				Where("id = ?", authorizedRobot.ID).
				Updates(map[string]any{
					"status":            normalizeRobotDeviceStatus(status.Status),
					"last_streaming_at": gorm.Expr("now()"),
					"updated_at":        gorm.Expr("now()"),
				}).Error; err != nil {
				return err
			}
		}

		updatedRobot, err := s.findRobotRecordByIDWithGorm(tx, authorizedRobot.ID)
		if err != nil {
			return err
		}
		robot = updatedRobot.toDomainRobot()
		return nil
	})
	if err != nil {
		return domain.Robot{}, err
	}
	return robot, nil
}

func (s *PostgresStore) upsertStreamingStatus(tx *gorm.DB, robotID string, status domain.StreamingStatus, tracksJSON []byte, channelsJSON []byte) error {
	updates := map[string]any{
		"mission_id":              gorm.Expr("EXCLUDED.mission_id"),
		"room_id":                 gorm.Expr("EXCLUDED.room_id"),
		"status":                  gorm.Expr("EXCLUDED.status"),
		"published_tracks":        gorm.Expr("EXCLUDED.published_tracks"),
		"published_data_channels": gorm.Expr("EXCLUDED.published_data_channels"),
		"sent_at":                 gorm.Expr("EXCLUDED.sent_at"),
		"updated_at":              gorm.Expr("now()"),
	}
	record := map[string]any{
		"robot_id":                robotID,
		"mission_id":              stringOrNil(status.MissionID),
		"room_id":                 status.RoomID,
		"status":                  status.Status,
		"published_tracks":        gorm.Expr("?::jsonb", string(tracksJSON)),
		"published_data_channels": gorm.Expr("?::jsonb", string(channelsJSON)),
		"sent_at":                 status.SentAt,
		"updated_at":              gorm.Expr("now()"),
	}
	return tx.Table(streamingStatusRecord{}.TableName()).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "robot_id"}},
			DoUpdates: clause.Assignments(updates),
		}).
		Create(record).Error
}

func (s *PostgresStore) updateMatchingTerminalStreamingStatus(tx *gorm.DB, robotID string, status domain.StreamingStatus, tracksJSON []byte, channelsJSON []byte) (bool, error) {
	query := tx.Table(streamingStatusRecord{}.TableName()).
		Where("robot_id = ? AND room_id = ?", robotID, status.RoomID)
	if status.MissionID == "" {
		query = query.Where("mission_id IS NULL")
	} else {
		query = query.Where("mission_id = ?", status.MissionID)
	}
	result := query.Updates(map[string]any{
		"status":                  status.Status,
		"published_tracks":        gorm.Expr("?::jsonb", string(tracksJSON)),
		"published_data_channels": gorm.Expr("?::jsonb", string(channelsJSON)),
		"sent_at":                 status.SentAt,
		"updated_at":              gorm.Expr("now()"),
	})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (s *PostgresStore) findActiveStreamingMissionCodeForRobot(tx *gorm.DB, robotID string, missionID string) (string, error) {
	var missionCode string
	err := tx.Table("missions AS m").
		Select("m.mission_code").
		Joins("JOIN mission_robots mr ON mr.mission_id = m.id").
		Where("m.id = ? AND m.status = ? AND mr.robot_id = ? AND mr.status = ?", missionID, "active", robotID, "active").
		Take(&missionCode).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrInvalidState
	}
	if err != nil {
		return "", err
	}
	return missionCode, nil
}

func (s *PostgresStore) ListStreamingStatuses(ctx context.Context) ([]domain.StreamingStatus, error) {
	var rows []streamingStatusQueryRow
	err := s.db.WithContext(ctx).
		Table("streaming_statuses AS ss").
		Select(`
			r.robot_code AS robot_code,
				ss.mission_id::text AS mission_id,
				ss.room_id AS room_id,
				ss.status AS status,
				ss.published_tracks AS published_tracks,
				ss.published_data_channels AS published_data_channels,
				ss.sent_at AS sent_at,
				ss.updated_at AS updated_at
			`).
		Joins("JOIN robots r ON r.id = ss.robot_id").
		Order("r.robot_code").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	statuses := make([]domain.StreamingStatus, 0, len(rows))
	for _, row := range rows {
		statuses = append(statuses, row.toDomainStreamingStatus())
	}
	return statuses, nil
}

func isPublishingStatus(status string) bool {
	return status == "streaming" || status == "publishing"
}

func isTerminalStreamingStatus(status string) bool {
	return status == "stopped" || status == "failed" || status == "stale"
}

type streamingStatusQueryRow struct {
	RobotCode             string
	MissionID             sql.NullString
	RoomID                string
	Status                string
	PublishedTracks       []byte
	PublishedDataChannels []byte
	SentAt                sql.NullTime
	UpdatedAt             time.Time
}

func (row streamingStatusQueryRow) toDomainStreamingStatus() domain.StreamingStatus {
	status := domain.StreamingStatus{
		RobotCode: row.RobotCode,
		MissionID: stringFromNull(row.MissionID),
		RoomID:    row.RoomID,
		Status:    row.Status,
		SentAt:    row.SentAt.Time,
		UpdatedAt: row.UpdatedAt,
	}
	_ = json.Unmarshal(row.PublishedTracks, &status.PublishedTracks)
	_ = json.Unmarshal(row.PublishedDataChannels, &status.PublishedDataChannels)
	return status
}
