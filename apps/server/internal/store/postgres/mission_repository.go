package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store/model"
	repo "robot-center/apps/server/internal/store/port"
	"robot-center/apps/server/internal/utils"
)

func (s *Store) CreateMission(ctx context.Context, input repo.CreateMissionInput) (domain.Mission, error) {
	if input.Name == "" {
		return domain.Mission{}, errors.New("name is required")
	}
	if input.MissionType == "" {
		return domain.Mission{}, errors.New("missionType is required")
	}
	robotCodes := repo.NormalizeMissionRobotCodes(input)

	var mission domain.Mission
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		robotIDs := make([]string, 0, len(robotCodes))
		for _, robotCode := range robotCodes {
			robot, err := s.findRobotRecordByCodeWithGorm(tx, robotCode)
			if errors.Is(err, repo.ErrNotFound) {
				return fmt.Errorf("robotCode %s not found", robotCode)
			}
			if err != nil {
				return err
			}
			robotIDs = append(robotIDs, robot.ID)
		}
		conflicts, err := s.findBusyMissionCreateConflicts(tx, robotIDs)
		if err != nil {
			return err
		}
		if len(conflicts) > 0 {
			return &repo.MissionStartConflictError{Conflicts: conflicts}
		}

		missionCode, err := s.nextCodeWithGorm(tx, "mission", model.MissionModel{}.TableName())
		if err != nil {
			return err
		}

		record := model.MissionModel{
			MissionCode: missionCode,
			Name:        input.Name,
			MissionType: input.MissionType,
			Status:      "ready",
			SiteNote:    stringPointer(input.SiteNote),
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		for _, robotID := range robotIDs {
			assignment := model.MissionRobotModel{
				MissionID: record.ID,
				RobotID:   robotID,
				Role:      "primary",
				Status:    "assigned",
			}
			if err := tx.Create(&assignment).Error; err != nil {
				return err
			}
		}

		mission = record.ToDomainMission(robotCodes)
		return nil
	})
	if err != nil {
		return domain.Mission{}, err
	}
	return mission, nil
}

func (s *Store) ListMissions(ctx context.Context) ([]domain.Mission, error) {
	return s.listMissionsWithGorm(s.db.WithContext(ctx))
}

func (s *Store) StartMission(ctx context.Context, missionCode string) (domain.Mission, error) {
	return s.transitionMission(ctx, strings.TrimSpace(missionCode), "ready", "active")
}

func (s *Store) EndMission(ctx context.Context, missionCode string) (domain.Mission, error) {
	return s.transitionMission(ctx, strings.TrimSpace(missionCode), "active", "ended")
}

func (s *Store) FindActiveMissionForRobot(ctx context.Context, robotCode string, bearerToken string) (domain.Mission, bool, error) {
	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return domain.Mission{}, false, err
	}
	defer rollbackUnlessCommitted(tx)

	robot, err := s.authorizeRobot(ctx, tx, robotCode, bearerToken)
	if err != nil {
		return domain.Mission{}, false, err
	}

	row := tx.QueryRowContext(ctx, `
		SELECT
			m.id::text,
			m.mission_code,
			m.name,
			m.mission_type,
			m.status,
			COALESCE(m.site_note, ''),
			COALESCE(string_agg(r_all.robot_code, ',' ORDER BY mr_all.created_at, r_all.robot_code) FILTER (WHERE r_all.robot_code IS NOT NULL), ''),
			m.started_at,
			m.ended_at,
			m.created_at,
			m.updated_at
		FROM missions m
		JOIN mission_robots mr_match ON mr_match.mission_id = m.id AND mr_match.status != 'removed'
		JOIN robots r_match ON r_match.id = mr_match.robot_id
		LEFT JOIN mission_robots mr_all ON mr_all.mission_id = m.id AND mr_all.status != 'removed'
		LEFT JOIN robots r_all ON r_all.id = mr_all.robot_id
		WHERE r_match.robot_code = $1 AND m.status = 'active'
		GROUP BY m.id, m.mission_code, m.name, m.mission_type, m.status, m.site_note, m.started_at, m.ended_at, m.created_at, m.updated_at
		ORDER BY m.started_at DESC NULLS LAST
		LIMIT 1
	`, robot.RobotCode)
	mission, err := scanMissionWithRobotCodes(row)
	if errors.Is(err, sql.ErrNoRows) {
		if err := tx.Commit(); err != nil {
			return domain.Mission{}, false, err
		}
		return domain.Mission{}, false, nil
	}
	if err != nil {
		return domain.Mission{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Mission{}, false, err
	}
	mission.RobotCode = robot.RobotCode
	return mission, true, nil
}

func (s *Store) transitionMission(ctx context.Context, missionCode string, expectedStatus string, nextStatus string) (domain.Mission, error) {
	var mission domain.Mission
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row missionTransitionQueryRow
		err := tx.Table("missions AS m").
			Select(`
				m.id::text AS id,
				m.status AS status
			`).
			Where("m.mission_code = ?", missionCode).
			Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "m"}}).
			Take(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return repo.ErrNotFound
		}
		if err != nil {
			return err
		}
		if row.Status != expectedStatus {
			return repo.ErrInvalidState
		}

		switch nextStatus {
		case "active":
			if err := s.lockMissionRobotsForStart(tx, row.ID); err != nil {
				return err
			}
			conflicts, err := s.findActiveMissionConflictsForStart(tx, row.ID)
			if err != nil {
				return err
			}
			if len(conflicts) > 0 {
				return &repo.MissionStartConflictError{Conflicts: conflicts}
			}
			if err := tx.Model(&model.MissionModel{}).
				Where("id = ?", row.ID).
				Updates(map[string]any{
					"status":     "active",
					"started_at": gorm.Expr("now()"),
					"updated_at": gorm.Expr("now()"),
				}).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.MissionRobotModel{}).
				Where("mission_id = ? AND status != ?", row.ID, "removed").
				Updates(map[string]any{
					"status":     "active",
					"joined_at":  gorm.Expr("COALESCE(joined_at, now())"),
					"updated_at": gorm.Expr("now()"),
				}).Error; err != nil {
				return err
			}
		case "ended":
			if err := tx.Model(&model.MissionModel{}).
				Where("id = ?", row.ID).
				Updates(map[string]any{
					"status":     "ended",
					"ended_at":   gorm.Expr("now()"),
					"updated_at": gorm.Expr("now()"),
				}).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.MissionRobotModel{}).
				Where("mission_id = ? AND status != ?", row.ID, "removed").
				Updates(map[string]any{
					"status":     "completed",
					"left_at":    gorm.Expr("now()"),
					"updated_at": gorm.Expr("now()"),
				}).Error; err != nil {
				return err
			}
		default:
			return repo.ErrInvalidState
		}

		updatedMission, err := s.findMissionByCodeWithGorm(tx, missionCode)
		if err != nil {
			return err
		}
		mission = updatedMission
		return nil
	})
	if err != nil {
		return domain.Mission{}, err
	}
	return mission, nil
}

func (s *Store) ValidateActiveMissionRobot(ctx context.Context, missionCode string, robotCode string) error {
	var count int64
	err := s.db.WithContext(ctx).
		Table("missions AS m").
		Joins("JOIN mission_robots mr ON mr.mission_id = m.id").
		Joins("JOIN robots r ON r.id = mr.robot_id").
		Where("m.mission_code = ? AND m.status = ? AND r.robot_code = ? AND mr.status = ?", strings.TrimSpace(missionCode), "active", strings.TrimSpace(robotCode), "active").
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return repo.ErrInvalidState
	}
	return nil
}

func (s *Store) lockMissionRobotsForStart(tx *gorm.DB, missionID string) error {
	var robotIDs []string
	return tx.Raw(`
		SELECT r.id::text
		FROM mission_robots mr
		JOIN robots r ON r.id = mr.robot_id
		WHERE mr.mission_id = ? AND mr.status != 'removed'
		ORDER BY r.robot_code
		FOR UPDATE OF r
	`, missionID).Scan(&robotIDs).Error
}

func (s *Store) findActiveMissionConflictsForStart(tx *gorm.DB, missionID string) ([]repo.MissionStartConflict, error) {
	var conflicts []repo.MissionStartConflict
	err := tx.Raw(`
		SELECT
			r.robot_code AS robot_code,
			active_m.mission_code AS active_mission_code
		FROM mission_robots target_mr
		JOIN robots r ON r.id = target_mr.robot_id
		JOIN mission_robots active_mr ON active_mr.robot_id = target_mr.robot_id
		JOIN missions active_m ON active_m.id = active_mr.mission_id
		WHERE target_mr.mission_id = ?
			AND target_mr.status != 'removed'
			AND active_m.id != ?
			AND active_m.status = 'active'
			AND active_mr.status != 'removed'
		ORDER BY r.robot_code, active_m.started_at DESC, active_m.mission_code
	`, missionID, missionID).Scan(&conflicts).Error
	if err != nil {
		return nil, err
	}
	return conflicts, nil
}

func (s *Store) findBusyMissionCreateConflicts(tx *gorm.DB, robotIDs []string) ([]repo.MissionStartConflict, error) {
	if len(robotIDs) == 0 {
		return nil, nil
	}
	var activeConflicts []repo.MissionStartConflict
	if err := tx.Raw(`
		SELECT
			r.robot_code AS robot_code,
			active_m.mission_code AS active_mission_code
		FROM robots r
		JOIN mission_robots active_mr ON active_mr.robot_id = r.id
		JOIN missions active_m ON active_m.id = active_mr.mission_id
		WHERE r.id IN ?
			AND active_m.status = 'active'
			AND active_mr.status != 'removed'
		ORDER BY r.robot_code, active_m.started_at DESC, active_m.mission_code
	`, robotIDs).Scan(&activeConflicts).Error; err != nil {
		return nil, err
	}

	return activeConflicts, nil
}

func mergeMissionStartConflicts(conflictGroups ...[]repo.MissionStartConflict) []repo.MissionStartConflict {
	seen := map[string]struct{}{}
	merged := make([]repo.MissionStartConflict, 0)
	for _, conflicts := range conflictGroups {
		for _, conflict := range conflicts {
			key := conflict.RobotCode + "\x00" + conflict.ActiveMissionCode
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			merged = append(merged, conflict)
		}
	}
	return merged
}

func (s *Store) listMissionsWithGorm(tx *gorm.DB) ([]domain.Mission, error) {
	var rows []missionQueryRow
	err := tx.Table("missions AS m").
		Select(`
			m.id::text AS id,
			m.mission_code AS mission_code,
			m.name AS name,
			m.mission_type AS mission_type,
			m.status AS status,
			m.site_note AS site_note,
			COALESCE(string_agg(r.robot_code, ',' ORDER BY mr.created_at, r.robot_code) FILTER (WHERE r.robot_code IS NOT NULL), '') AS robot_codes,
			m.started_at AS started_at,
			m.ended_at AS ended_at,
			m.created_at AS created_at,
			m.updated_at AS updated_at
		`).
		Joins("LEFT JOIN mission_robots mr ON mr.mission_id = m.id AND mr.status != 'removed'").
		Joins("LEFT JOIN robots r ON r.id = mr.robot_id").
		Group("m.id, m.mission_code, m.name, m.mission_type, m.status, m.site_note, m.started_at, m.ended_at, m.created_at, m.updated_at").
		Order("m.created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	missions := make([]domain.Mission, 0, len(rows))
	for _, row := range rows {
		missions = append(missions, row.ToDomainMission())
	}
	return missions, nil
}

func (s *Store) findMissionByCodeWithGorm(tx *gorm.DB, missionCode string) (domain.Mission, error) {
	var row missionQueryRow
	err := tx.Table("missions AS m").
		Select(`
			m.id::text AS id,
			m.mission_code AS mission_code,
			m.name AS name,
			m.mission_type AS mission_type,
			m.status AS status,
			m.site_note AS site_note,
			COALESCE(string_agg(r.robot_code, ',' ORDER BY mr.created_at, r.robot_code) FILTER (WHERE r.robot_code IS NOT NULL), '') AS robot_codes,
			m.started_at AS started_at,
			m.ended_at AS ended_at,
			m.created_at AS created_at,
			m.updated_at AS updated_at
		`).
		Joins("LEFT JOIN mission_robots mr ON mr.mission_id = m.id AND mr.status != 'removed'").
		Joins("LEFT JOIN robots r ON r.id = mr.robot_id").
		Where("m.mission_code = ?", strings.TrimSpace(missionCode)).
		Group("m.id, m.mission_code, m.name, m.mission_type, m.status, m.site_note, m.started_at, m.ended_at, m.created_at, m.updated_at").
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Mission{}, repo.ErrNotFound
	}
	if err != nil {
		return domain.Mission{}, err
	}
	return row.ToDomainMission(), nil
}

type missionQueryRow struct {
	ID          string
	MissionCode string
	Name        string
	MissionType string
	Status      string
	SiteNote    sql.NullString
	RobotCodes  sql.NullString
	StartedAt   sql.NullTime
	EndedAt     sql.NullTime
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (row missionQueryRow) ToDomainMission() domain.Mission {
	robotCodes := robotCodesFromString(stringFromNull(row.RobotCodes))
	return domain.Mission{
		ID:          row.ID,
		MissionCode: row.MissionCode,
		Name:        row.Name,
		MissionType: row.MissionType,
		Status:      row.Status,
		SiteNote:    stringFromNull(row.SiteNote),
		RobotCode:   utils.FirstString(robotCodes),
		RobotCodes:  robotCodes,
		StartedAt:   timePointer(row.StartedAt),
		EndedAt:     timePointer(row.EndedAt),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

type missionTransitionQueryRow struct {
	ID     string
	Status string
}

func (s *Store) findMissionByCode(ctx context.Context, tx *sql.Tx, missionCode string) (domain.Mission, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT
			m.id::text,
			m.mission_code,
			m.name,
			m.mission_type,
			m.status,
			COALESCE(m.site_note, ''),
			COALESCE(string_agg(r.robot_code, ',' ORDER BY mr.created_at, r.robot_code) FILTER (WHERE r.robot_code IS NOT NULL), ''),
			m.started_at,
			m.ended_at,
			m.created_at,
			m.updated_at
		FROM missions m
		LEFT JOIN mission_robots mr ON mr.mission_id = m.id AND mr.status != 'removed'
		LEFT JOIN robots r ON r.id = mr.robot_id
		WHERE m.mission_code = $1
		GROUP BY m.id, m.mission_code, m.name, m.mission_type, m.status, m.site_note, m.started_at, m.ended_at, m.created_at, m.updated_at
	`, strings.TrimSpace(missionCode))
	mission, err := scanMissionWithRobotCodes(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Mission{}, repo.ErrNotFound
	}
	return mission, err
}
