package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"gorm.io/gorm"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store/model"
	repo "robot-center/apps/server/internal/store/port"
	"robot-center/apps/server/internal/utils"
)

func (s *Store) CreateRobot(ctx context.Context, input repo.CreateRobotInput) (domain.Robot, domain.RobotConnectionInfo, error) {
	if input.DisplayName == "" {
		return domain.Robot{}, domain.RobotConnectionInfo{}, errors.New("displayName is required")
	}

	token := "rb_p0_" + utils.RandomHex(18)
	var robot domain.Robot

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		robotCode, err := s.nextCodeWithGorm(tx, "robot", model.RobotModel{}.TableName())
		if err != nil {
			return err
		}

		record := model.RobotModel{
			RobotCode:   robotCode,
			DisplayName: input.DisplayName,
			ModelName:   stringPointer(input.ModelName),
			DeviceState: string(domain.RobotDeviceStateOffline),
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		tokenPlaintext := token
		tokenRecord := model.RobotConnectionTokenModel{
			RobotID:        record.ID,
			TokenHash:      utils.HashToken(token),
			TokenPlaintext: &tokenPlaintext,
			Name:           "default",
			IsActive:       true,
		}
		if err := tx.Create(&tokenRecord).Error; err != nil {
			return err
		}

		robot = record.ToDomainRobot()
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

func (s *Store) ListRobots(ctx context.Context) ([]domain.Robot, error) {
	var records []model.RobotModel
	if err := s.db.WithContext(ctx).Where("archived_at IS NULL").Order("robot_code").Find(&records).Error; err != nil {
		return nil, err
	}

	robots := make([]domain.Robot, 0, len(records))
	for _, record := range records {
		robots = append(robots, record.ToDomainRobot())
	}
	return robots, nil
}

func (s *Store) UpdateRobot(ctx context.Context, robotCode string, input repo.UpdateRobotInput) (domain.Robot, error) {
	if input.DisplayName == "" {
		return domain.Robot{}, errors.New("displayName is required")
	}

	var robot domain.Robot
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		record, err := s.findRobotRecordByCodeWithGorm(tx, robotCode)
		if err != nil {
			return err
		}
		if err := tx.Model(&model.RobotModel{}).
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
		robot = updatedRecord.ToDomainRobot()
		return nil
	})
	if err != nil {
		return domain.Robot{}, err
	}
	return robot, nil
}

func (s *Store) ArchiveRobot(ctx context.Context, robotCode string) (domain.Robot, error) {
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
			return repo.ErrInvalidState
		}

		if err := tx.Model(&model.RobotModel{}).
			Where("id = ?", record.ID).
			Updates(map[string]any{
				"device_state": string(domain.RobotDeviceStateOffline),
				"archived_at":  gorm.Expr("now()"),
				"updated_at":   gorm.Expr("now()"),
			}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.RobotConnectionTokenModel{}).
			Where("robot_id = ?", record.ID).
			Update("is_active", false).Error; err != nil {
			return err
		}
		updatedRecord, err := s.findRobotRecordByIDIncludingArchivedWithGorm(tx, record.ID)
		if err != nil {
			return err
		}
		robot = updatedRecord.ToDomainRobot()
		return nil
	})
	if err != nil {
		return domain.Robot{}, err
	}
	return robot, nil
}

func (s *Store) GetRobotConnectionInfo(ctx context.Context, robotCode string) (domain.RobotConnectionInfo, error) {
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
		return domain.RobotConnectionInfo{}, repo.ErrNotFound
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

func (s *Store) RotateRobotConnectionToken(ctx context.Context, robotCode string) (domain.RobotConnectionInfo, error) {
	token := "rb_p0_" + utils.RandomHex(18)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		robot, err := s.findRobotRecordByCodeWithGorm(tx, robotCode)
		if err != nil {
			return err
		}
		if err := tx.Model(&model.RobotConnectionTokenModel{}).
			Where("robot_id = ?", robot.ID).
			Update("is_active", false).Error; err != nil {
			return err
		}
		tokenPlaintext := token
		tokenRecord := model.RobotConnectionTokenModel{
			RobotID:        robot.ID,
			TokenHash:      utils.HashToken(token),
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

func (s *Store) ApplyHeartbeat(ctx context.Context, input repo.HeartbeatInput, bearerToken string) (domain.Robot, error) {
	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return domain.Robot{}, err
	}
	defer rollbackUnlessCommitted(tx)

	robot, err := s.authorizeRobot(ctx, tx, bearerToken)
	if err != nil {
		return domain.Robot{}, err
	}
	reportedState := strings.TrimSpace(input.State)
	if reportedState == "" {
		reportedState = string(domain.RobotDeviceStateOnline)
	}
	deviceState := domain.NormalizeRobotDeviceState(reportedState)
	var storedDeviceState string
	err = tx.QueryRowContext(ctx, `
		UPDATE robots
		SET device_state = $2, last_seen_at = now(), updated_at = now()
		WHERE id = $1::uuid
		RETURNING id::text, robot_code, display_name, COALESCE(model_name, ''), device_state, last_seen_at, created_at, updated_at
	`, robot.ID, string(deviceState)).Scan(
		&robot.ID,
		&robot.RobotCode,
		&robot.DisplayName,
		&robot.ModelName,
		&storedDeviceState,
		nullableTimeScanner(&robot.LastSeenAt),
		&robot.CreatedAt,
		&robot.UpdatedAt,
	)
	if err != nil {
		return domain.Robot{}, err
	}
	robot.DeviceState = domain.NormalizeRobotDeviceState(storedDeviceState)
	if err := tx.Commit(); err != nil {
		return domain.Robot{}, err
	}
	return robot, nil
}

func (s *Store) ResolveRobotByBearerToken(ctx context.Context, bearerToken string) (domain.Robot, error) {
	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return domain.Robot{}, err
	}
	defer rollbackUnlessCommitted(tx)

	robot, err := s.authorizeRobot(ctx, tx, bearerToken)
	if err != nil {
		return domain.Robot{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Robot{}, err
	}
	return robot, nil
}

func (s *Store) findRobotRecordByCodeWithGorm(tx *gorm.DB, robotCode string) (model.RobotModel, error) {
	var record model.RobotModel
	err := tx.Where("robot_code = ? AND archived_at IS NULL", strings.TrimSpace(robotCode)).Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.RobotModel{}, repo.ErrNotFound
	}
	return record, err
}

func (s *Store) findRobotRecordByIDWithGorm(tx *gorm.DB, robotID string) (model.RobotModel, error) {
	var record model.RobotModel
	err := tx.Where("id = ? AND archived_at IS NULL", strings.TrimSpace(robotID)).Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.RobotModel{}, repo.ErrNotFound
	}
	return record, err
}

func (s *Store) findRobotRecordByIDIncludingArchivedWithGorm(tx *gorm.DB, robotID string) (model.RobotModel, error) {
	var record model.RobotModel
	err := tx.Where("id = ?", strings.TrimSpace(robotID)).Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.RobotModel{}, repo.ErrNotFound
	}
	return record, err
}

func (s *Store) authorizeRobot(ctx context.Context, tx *sql.Tx, bearerToken string) (domain.Robot, error) {
	trimmedToken := strings.TrimSpace(bearerToken)
	if trimmedToken == "" {
		return domain.Robot{}, repo.ErrUnauthorized
	}
	row := tx.QueryRowContext(ctx, `
		SELECT r.id::text, r.robot_code, r.display_name, COALESCE(r.model_name, ''), r.device_state,
		       r.last_seen_at, r.created_at, r.updated_at
			FROM robots r
			JOIN robot_tokens rt ON rt.robot_id = r.id
			WHERE r.archived_at IS NULL AND rt.token_hash = $1 AND rt.is_active = true
			LIMIT 1
	`, utils.HashToken(trimmedToken))
	robot, err := scanRobot(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Robot{}, repo.ErrUnauthorized
	}
	if err != nil {
		return domain.Robot{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE robot_tokens SET last_used_at = now()
		WHERE robot_id = $1::uuid AND token_hash = $2
	`, robot.ID, utils.HashToken(trimmedToken)); err != nil {
		return domain.Robot{}, err
	}
	return robot, nil
}

func (s *Store) findRobotID(ctx context.Context, robotCode string) (string, error) {
	var robotID string
	err := s.sqlDB.QueryRowContext(ctx, `SELECT id::text FROM robots WHERE robot_code = $1 AND archived_at IS NULL`, strings.TrimSpace(robotCode)).Scan(&robotID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", repo.ErrNotFound
	}
	return robotID, err
}
