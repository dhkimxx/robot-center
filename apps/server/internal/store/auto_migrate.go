package store

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *PostgresStore) runAutoMigrations(ctx context.Context) error {
	db := s.db.WithContext(ctx)
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto`).Error; err != nil {
		return err
	}
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS postgis`).Error; err != nil {
		return err
	}
	if err := db.AutoMigrate(
		&userRecord{},
		&robotRecord{},
		&robotConnectionTokenRecord{},
		&missionRecord{},
		&missionRobotRecord{},
		&robotSessionRecord{},
		&browserSessionRecord{},
		&recorderSessionRecord{},
		&streamingStatusRecord{},
		&sensorDescriptorRecord{},
		&sensorSampleRecord{},
		&recordingSessionRecord{},
		&recordingChunkRecord{},
		&storageObjectRecord{},
		&eventRecord{},
		&controlCommandRecord{},
		&controlAckRecord{},
	); err != nil {
		return err
	}
	if err := s.applyPostAutoMigrateDDL(db); err != nil {
		return err
	}
	return s.seedBootstrapUser(db)
}

func (s *PostgresStore) applyPostAutoMigrateDDL(db *gorm.DB) error {
	statements := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS mission_robots_active_unique
			ON mission_robots(mission_id, robot_id)
			WHERE status != 'removed'`,
		`CREATE INDEX IF NOT EXISTS events_geom_idx
			ON events USING gist(geom)`,
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'recording_chunks_manifest_object_fk'
			) THEN
				ALTER TABLE recording_chunks
					ADD CONSTRAINT recording_chunks_manifest_object_fk
					FOREIGN KEY (manifest_object_id) REFERENCES storage_objects(id) ON DELETE SET NULL;
			END IF;
		END $$`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) seedBootstrapUser(db *gorm.DB) error {
	user := userRecord{
		LoginID:      "operator",
		PasswordHash: "seed-password-placeholder",
		DisplayName:  "PoC Operator",
		Role:         "operator",
		IsActive:     true,
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "login_id"}},
		DoNothing: true,
	}).Create(&user).Error
}
