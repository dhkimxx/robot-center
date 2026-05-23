package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"gorm.io/gorm"

	"robot-center/apps/server/internal/domain"
)

func (s *PostgresStore) CreateRobot(ctx context.Context, input CreateRobotInput) (domain.Robot, domain.RobotConnectionInfo, error) {
	if input.DisplayName == "" {
		return domain.Robot{}, domain.RobotConnectionInfo{}, errors.New("displayName is required")
	}

	token := "rb_p0_" + randomHex(18)
	var robot domain.Robot

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		robotCode, err := s.nextCodeWithGorm(tx, "robot", robotRecord{}.TableName())
		if err != nil {
			return err
		}

		record := robotRecord{
			RobotCode:   robotCode,
			DisplayName: input.DisplayName,
			ModelName:   stringPointer(input.ModelName),
			Status:      "offline",
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		tokenPlaintext := token
		tokenRecord := robotConnectionTokenRecord{
			RobotID:        record.ID,
			TokenHash:      hashToken(token),
			TokenPlaintext: &tokenPlaintext,
			Name:           "default",
			IsActive:       true,
		}
		if err := tx.Create(&tokenRecord).Error; err != nil {
			return err
		}

		robot = record.toDomainRobot()
		return nil
	})
	if err != nil {
		return domain.Robot{}, domain.RobotConnectionInfo{}, err
	}

	return robot, domain.RobotConnectionInfo{
		ServerURL:  s.serverURL,
		RobotCode:  robot.RobotCode,
		RobotToken: token,
	}, nil
}

func (s *PostgresStore) ListRobots(ctx context.Context) ([]domain.Robot, error) {
	var records []robotRecord
	if err := s.db.WithContext(ctx).Where("archived_at IS NULL").Order("robot_code").Find(&records).Error; err != nil {
		return nil, err
	}

	robots := make([]domain.Robot, 0, len(records))
	for _, record := range records {
		robots = append(robots, record.toDomainRobot())
	}
	return robots, nil
}

func (s *PostgresStore) UpdateRobot(ctx context.Context, robotCode string, input UpdateRobotInput) (domain.Robot, error) {
	if input.DisplayName == "" {
		return domain.Robot{}, errors.New("displayName is required")
	}

	var robot domain.Robot
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		record, err := s.findRobotRecordByCodeWithGorm(tx, robotCode)
		if err != nil {
			return err
		}
		if err := tx.Model(&robotRecord{}).
			Where("id = ?", record.ID).
			Updates(map[string]any{
				"display_name": input.DisplayName,
				"model_name":   stringOrNil(input.ModelName),
				"updated_at":   gorm.Expr("now()"),
			}).Error; err != nil {
			return err
		}
		updatedRecord, err := s.findRobotRecordByIDWithGorm(tx, record.ID)
		if err != nil {
			return err
		}
		robot = updatedRecord.toDomainRobot()
		return nil
	})
	if err != nil {
		return domain.Robot{}, err
	}
	return robot, nil
}

func (s *PostgresStore) ArchiveRobot(ctx context.Context, robotCode string) (domain.Robot, error) {
	var robot domain.Robot
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		record, err := s.findRobotRecordByCodeWithGorm(tx, robotCode)
		if err != nil {
			return err
		}

		var openMissionCount int64
		if err := tx.Table("mission_robots AS mr").
			Joins("JOIN missions m ON m.id = mr.mission_id").
			Where("mr.robot_id = ? AND mr.status != ? AND m.status IN ?", record.ID, "removed", []string{"ready", "active"}).
			Count(&openMissionCount).Error; err != nil {
			return err
		}
		if openMissionCount > 0 {
			return ErrInvalidState
		}

		if err := tx.Model(&robotRecord{}).
			Where("id = ?", record.ID).
			Updates(map[string]any{
				"status":      "offline",
				"archived_at": gorm.Expr("now()"),
				"updated_at":  gorm.Expr("now()"),
			}).Error; err != nil {
			return err
		}
		if err := tx.Model(&robotConnectionTokenRecord{}).
			Where("robot_id = ?", record.ID).
			Update("is_active", false).Error; err != nil {
			return err
		}
		updatedRecord, err := s.findRobotRecordByIDIncludingArchivedWithGorm(tx, record.ID)
		if err != nil {
			return err
		}
		robot = updatedRecord.toDomainRobot()
		return nil
	})
	if err != nil {
		return domain.Robot{}, err
	}
	return robot, nil
}

func (s *PostgresStore) GetRobotConnectionInfo(ctx context.Context, robotCode string) (domain.RobotConnectionInfo, error) {
	var token string
	err := s.sqlDB.QueryRowContext(ctx, `
		SELECT COALESCE(rt.token_plaintext, '')
		FROM robots r
		JOIN robot_tokens rt ON rt.robot_id = r.id
		WHERE r.robot_code = $1 AND r.archived_at IS NULL AND rt.is_active = true
		ORDER BY rt.created_at DESC
		LIMIT 1
	`, strings.TrimSpace(robotCode)).Scan(&token)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.RobotConnectionInfo{}, ErrNotFound
	}
	if err != nil {
		return domain.RobotConnectionInfo{}, err
	}
	if token == "" {
		return domain.RobotConnectionInfo{}, errors.New("robot token plaintext is unavailable")
	}
	return domain.RobotConnectionInfo{
		ServerURL:  s.serverURL,
		RobotCode:  strings.TrimSpace(robotCode),
		RobotToken: token,
	}, nil
}

func (s *PostgresStore) RotateRobotConnectionToken(ctx context.Context, robotCode string) (domain.RobotConnectionInfo, error) {
	token := "rb_p0_" + randomHex(18)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		robot, err := s.findRobotRecordByCodeWithGorm(tx, robotCode)
		if err != nil {
			return err
		}
		if err := tx.Model(&robotConnectionTokenRecord{}).
			Where("robot_id = ?", robot.ID).
			Update("is_active", false).Error; err != nil {
			return err
		}
		tokenPlaintext := token
		tokenRecord := robotConnectionTokenRecord{
			RobotID:        robot.ID,
			TokenHash:      hashToken(token),
			TokenPlaintext: &tokenPlaintext,
			Name:           "rotated",
			IsActive:       true,
		}
		return tx.Create(&tokenRecord).Error
	})
	if err != nil {
		return domain.RobotConnectionInfo{}, err
	}
	return domain.RobotConnectionInfo{
		ServerURL:  s.serverURL,
		RobotCode:  strings.TrimSpace(robotCode),
		RobotToken: token,
	}, nil
}

func (s *PostgresStore) ApplyHeartbeat(ctx context.Context, input HeartbeatInput, bearerToken string) (domain.Robot, error) {
	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return domain.Robot{}, err
	}
	defer rollbackUnlessCommitted(tx)

	robot, err := s.authorizeRobot(ctx, tx, input.RobotCode, bearerToken)
	if err != nil {
		return domain.Robot{}, err
	}
	status := input.State
	if status == "" {
		status = "online"
	}
	err = tx.QueryRowContext(ctx, `
		UPDATE robots
		SET status = $2, last_seen_at = now(), updated_at = now()
		WHERE id = $1::uuid
		RETURNING id::text, robot_code, display_name, COALESCE(model_name, ''), status, last_seen_at, last_streaming_at, created_at, updated_at
	`, robot.ID, status).Scan(
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
	if err != nil {
		return domain.Robot{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Robot{}, err
	}
	return robot, nil
}

func (s *PostgresStore) findRobotRecordByCodeWithGorm(tx *gorm.DB, robotCode string) (robotRecord, error) {
	var record robotRecord
	err := tx.Where("robot_code = ? AND archived_at IS NULL", strings.TrimSpace(robotCode)).Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return robotRecord{}, ErrNotFound
	}
	return record, err
}

func (s *PostgresStore) findRobotRecordByIDWithGorm(tx *gorm.DB, robotID string) (robotRecord, error) {
	var record robotRecord
	err := tx.Where("id = ? AND archived_at IS NULL", strings.TrimSpace(robotID)).Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return robotRecord{}, ErrNotFound
	}
	return record, err
}

func (s *PostgresStore) findRobotRecordByIDIncludingArchivedWithGorm(tx *gorm.DB, robotID string) (robotRecord, error) {
	var record robotRecord
	err := tx.Where("id = ?", strings.TrimSpace(robotID)).Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return robotRecord{}, ErrNotFound
	}
	return record, err
}

func (s *PostgresStore) authorizeRobotWithGorm(tx *gorm.DB, robotCode string, bearerToken string) (domain.Robot, error) {
	trimmedRobotCode := strings.TrimSpace(robotCode)
	trimmedToken := strings.TrimSpace(bearerToken)
	if trimmedRobotCode == "" || trimmedToken == "" {
		return domain.Robot{}, ErrUnauthorized
	}

	tokenHash := hashToken(trimmedToken)
	var record robotRecord
	err := tx.Table("robots AS r").
		Select(`
			r.id::text AS id,
			r.robot_code AS robot_code,
			r.display_name AS display_name,
			r.model_name AS model_name,
			r.status AS status,
			r.last_seen_at AS last_seen_at,
			r.last_streaming_at AS last_streaming_at,
			r.created_at AS created_at,
			r.updated_at AS updated_at
		`).
		Joins("JOIN robot_tokens rt ON rt.robot_id = r.id").
		Where("r.robot_code = ? AND r.archived_at IS NULL AND rt.token_hash = ? AND rt.is_active = ?", trimmedRobotCode, tokenHash, true).
		Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Robot{}, ErrUnauthorized
	}
	if err != nil {
		return domain.Robot{}, err
	}

	if err := tx.Model(&robotConnectionTokenRecord{}).
		Where("robot_id = ? AND token_hash = ?", record.ID, tokenHash).
		Update("last_used_at", gorm.Expr("now()")).Error; err != nil {
		return domain.Robot{}, err
	}
	return record.toDomainRobot(), nil
}

func (s *PostgresStore) authorizeRobot(ctx context.Context, tx *sql.Tx, robotCode string, bearerToken string) (domain.Robot, error) {
	if strings.TrimSpace(robotCode) == "" || strings.TrimSpace(bearerToken) == "" {
		return domain.Robot{}, ErrUnauthorized
	}
	row := tx.QueryRowContext(ctx, `
		SELECT r.id::text, r.robot_code, r.display_name, COALESCE(r.model_name, ''), r.status,
		       r.last_seen_at, r.last_streaming_at, r.created_at, r.updated_at
			FROM robots r
			JOIN robot_tokens rt ON rt.robot_id = r.id
			WHERE r.robot_code = $1 AND r.archived_at IS NULL AND rt.token_hash = $2 AND rt.is_active = true
			LIMIT 1
	`, strings.TrimSpace(robotCode), hashToken(strings.TrimSpace(bearerToken)))
	robot, err := scanRobot(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Robot{}, ErrUnauthorized
	}
	if err != nil {
		return domain.Robot{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE robot_tokens SET last_used_at = now()
		WHERE robot_id = $1::uuid AND token_hash = $2
	`, robot.ID, hashToken(strings.TrimSpace(bearerToken))); err != nil {
		return domain.Robot{}, err
	}
	return robot, nil
}

func (s *PostgresStore) findRobotID(ctx context.Context, robotCode string) (string, error) {
	var robotID string
	err := s.sqlDB.QueryRowContext(ctx, `SELECT id::text FROM robots WHERE robot_code = $1 AND archived_at IS NULL`, strings.TrimSpace(robotCode)).Scan(&robotID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return robotID, err
}
