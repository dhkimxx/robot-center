package postgres

import (
	"context"
	"fmt"

	"robot-center/apps/server/internal/store/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *Store) runAutoMigrations(ctx context.Context) error {
	db := s.db.WithContext(ctx)
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto`).Error; err != nil {
		return err
	}
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS postgis`).Error; err != nil {
		return err
	}
	if err := db.AutoMigrate(
		&model.UserModel{},
		&model.RobotModel{},
		&model.RobotConnectionTokenModel{},
		&model.MissionModel{},
		&model.MissionRobotModel{},
		&model.RobotSessionModel{},
		&model.BrowserSessionModel{},
		&model.RecorderSessionModel{},
		&model.SensorDescriptorModel{},
		&model.SensorSampleModel{},
		&model.RecordingSessionModel{},
		&model.RecordingChunkModel{},
		&model.RecordingFinalizationJobModel{},
		&model.StorageObjectModel{},
		&model.EventModel{},
		&model.ControlCommandModel{},
		&model.ControlAckModel{},
	); err != nil {
		return err
	}
	if err := s.applyPostAutoMigrateDDL(db); err != nil {
		return err
	}
	return s.seedBootstrapUser(db)
}

func (s *Store) applyPostAutoMigrateDDL(db *gorm.DB) error {
	statements := []string{
		robotDeviceStateMigrationStatement(),
		`CREATE UNIQUE INDEX IF NOT EXISTS mission_robots_active_unique
			ON mission_robots(mission_id, robot_id)
			WHERE status != 'removed'`,
		`CREATE UNIQUE INDEX IF NOT EXISTS robot_tokens_active_unique
			ON robot_tokens(robot_id)
			WHERE is_active = true`,
		`CREATE UNIQUE INDEX IF NOT EXISTS recording_sessions_open_unique
			ON recording_sessions(mission_id, robot_id)
			WHERE ended_at IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS sensor_descriptors_id_mission_robot_sensor_unique
			ON sensor_descriptors(id, mission_id, robot_id, sensor_id)`,
		`CREATE INDEX IF NOT EXISTS sensor_samples_descriptor_received_idx
			ON sensor_samples(descriptor_id, received_at DESC)`,
		`UPDATE sensor_samples ss
			SET descriptor_id = sd.id
			FROM sensor_descriptors sd
			WHERE ss.descriptor_id IS NULL
				AND sd.mission_id = ss.mission_id
				AND sd.robot_id = ss.robot_id
				AND sd.sensor_id = ss.sensor_id`,
		`ALTER TABLE sensor_samples
			ALTER COLUMN descriptor_id SET NOT NULL`,
		sensorSampleValueColumnMigrationStatement(),
		`CREATE INDEX IF NOT EXISTS events_geom_idx
			ON events USING gist(geom)`,
		dropEmptyTableStatement("sensor_readings"),
		dropEmptyTableStatement("telemetry_snapshots"),
	}
	statements = append(statements, baseModelColumnStatements()...)
	statements = append(statements, updatedAtTriggerStatements()...)
	statements = append(statements, foreignKeyConstraintStatements()...)
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func sensorSampleValueColumnMigrationStatement() string {
	return `DO $$
		BEGIN
			IF to_regclass('public.sensor_samples') IS NOT NULL THEN
				IF NOT EXISTS (
					SELECT 1
					FROM information_schema.columns
					WHERE table_schema = 'public'
						AND table_name = 'sensor_samples'
						AND column_name = 'values'
				) THEN
					ALTER TABLE sensor_samples ADD COLUMN "values" jsonb;
				END IF;

				IF EXISTS (
					SELECT 1
					FROM information_schema.columns
					WHERE table_schema = 'public'
						AND table_name = 'sensor_samples'
						AND column_name = 'object_value'
				) THEN
					UPDATE sensor_samples
					SET "values" = object_value
					WHERE "values" IS NULL
						AND object_value IS NOT NULL;
				END IF;

				ALTER TABLE sensor_samples DROP COLUMN IF EXISTS numeric_value;
				ALTER TABLE sensor_samples DROP COLUMN IF EXISTS text_value;
				ALTER TABLE sensor_samples DROP COLUMN IF EXISTS bool_value;
				ALTER TABLE sensor_samples DROP COLUMN IF EXISTS vector_value;
				ALTER TABLE sensor_samples DROP COLUMN IF EXISTS object_value;
			END IF;
		END $$`
}

func baseModelTableNames() []string {
	return []string{
		model.UserModel{}.TableName(),
		model.RobotModel{}.TableName(),
		model.RobotConnectionTokenModel{}.TableName(),
		model.MissionModel{}.TableName(),
		model.MissionRobotModel{}.TableName(),
		model.RobotSessionModel{}.TableName(),
		model.BrowserSessionModel{}.TableName(),
		model.RecorderSessionModel{}.TableName(),
		model.SensorDescriptorModel{}.TableName(),
		model.SensorSampleModel{}.TableName(),
		model.RecordingSessionModel{}.TableName(),
		model.RecordingChunkModel{}.TableName(),
		model.RecordingFinalizationJobModel{}.TableName(),
		model.StorageObjectModel{}.TableName(),
		model.EventModel{}.TableName(),
		model.ControlCommandModel{}.TableName(),
		model.ControlAckModel{}.TableName(),
	}
}

func baseModelColumnStatements() []string {
	statements := make([]string, 0, len(baseModelTableNames())*2)
	for _, tableName := range baseModelTableNames() {
		statements = append(statements, fmt.Sprintf(`DO $$
			BEGIN
				IF to_regclass('public.%s') IS NOT NULL THEN
					IF EXISTS (
						SELECT 1
						FROM information_schema.columns
						WHERE table_schema = 'public'
							AND table_name = '%s'
							AND column_name = 'created_at'
					) THEN
						EXECUTE 'UPDATE public.%s SET created_at = now() WHERE created_at IS NULL';
						ALTER TABLE %s ALTER COLUMN created_at SET DEFAULT now();
						ALTER TABLE %s ALTER COLUMN created_at SET NOT NULL;
					END IF;

					IF EXISTS (
						SELECT 1
						FROM information_schema.columns
						WHERE table_schema = 'public'
							AND table_name = '%s'
							AND column_name = 'updated_at'
					) THEN
						EXECUTE 'UPDATE public.%s SET updated_at = COALESCE(created_at, now()) WHERE updated_at IS NULL';
						ALTER TABLE %s ALTER COLUMN updated_at SET DEFAULT now();
						ALTER TABLE %s ALTER COLUMN updated_at SET NOT NULL;
					END IF;
				END IF;
			END $$`, tableName, tableName, tableName, tableName, tableName, tableName, tableName, tableName, tableName))
	}
	return statements
}

func updatedAtTriggerStatements() []string {
	statements := []string{`CREATE OR REPLACE FUNCTION set_updated_at_timestamp()
		RETURNS trigger AS $$
		BEGIN
			NEW.updated_at = now();
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql`}
	for _, tableName := range baseModelTableNames() {
		triggerName := tableName + "_set_updated_at"
		statements = append(statements, fmt.Sprintf(`DO $$
			BEGIN
				IF to_regclass('public.%s') IS NOT NULL
					AND EXISTS (
						SELECT 1
						FROM information_schema.columns
						WHERE table_schema = 'public'
							AND table_name = '%s'
							AND column_name = 'updated_at'
					)
					AND NOT EXISTS (
						SELECT 1
						FROM pg_trigger
						WHERE tgname = '%s'
					)
				THEN
					CREATE TRIGGER %s
						BEFORE UPDATE ON %s
						FOR EACH ROW
						EXECUTE FUNCTION set_updated_at_timestamp();
				END IF;
			END $$`, tableName, tableName, triggerName, triggerName, tableName))
	}
	return statements
}

func robotDeviceStateMigrationStatement() string {
	return `DO $$
		BEGIN
			IF to_regclass('public.robots') IS NOT NULL THEN
				IF NOT EXISTS (
					SELECT 1
					FROM information_schema.columns
					WHERE table_schema = 'public'
						AND table_name = 'robots'
						AND column_name = 'device_state'
				) THEN
					ALTER TABLE robots ADD COLUMN device_state text;
				END IF;

				IF EXISTS (
					SELECT 1
					FROM information_schema.columns
					WHERE table_schema = 'public'
						AND table_name = 'robots'
						AND column_name = 'status'
				) THEN
					UPDATE robots
					SET device_state = CASE
						WHEN status = 'fault' THEN 'fault'
						WHEN status = 'offline' THEN 'offline'
						ELSE 'online'
					END;
					ALTER TABLE robots DROP COLUMN status;
				END IF;

				UPDATE robots
				SET device_state = CASE
					WHEN device_state = 'fault' THEN 'fault'
					WHEN device_state = 'offline' THEN 'offline'
					ELSE 'online'
				END
				WHERE device_state IS NULL
					OR device_state NOT IN ('online', 'offline', 'fault');

				ALTER TABLE robots ALTER COLUMN device_state SET DEFAULT 'offline';
				ALTER TABLE robots ALTER COLUMN device_state SET NOT NULL;
			END IF;
		END $$`
}

type foreignKeyConstraint struct {
	Name            string
	Table           string
	Columns         string
	ReferenceTable  string
	ReferenceColumn string
	OnDelete        string
}

func foreignKeyConstraintStatements() []string {
	constraints := []foreignKeyConstraint{
		{Name: "missions_created_by_fk", Table: "missions", Columns: "created_by", ReferenceTable: "users", ReferenceColumn: "id", OnDelete: "SET NULL"},
		{Name: "robot_tokens_robot_fk", Table: "robot_tokens", Columns: "robot_id", ReferenceTable: "robots", ReferenceColumn: "id", OnDelete: "CASCADE"},
		{Name: "mission_robots_mission_fk", Table: "mission_robots", Columns: "mission_id", ReferenceTable: "missions", ReferenceColumn: "id", OnDelete: "CASCADE"},
		{Name: "mission_robots_robot_fk", Table: "mission_robots", Columns: "robot_id", ReferenceTable: "robots", ReferenceColumn: "id"},
		{Name: "robot_sessions_robot_fk", Table: "robot_sessions", Columns: "robot_id", ReferenceTable: "robots", ReferenceColumn: "id", OnDelete: "CASCADE"},
		{Name: "robot_sessions_mission_fk", Table: "robot_sessions", Columns: "mission_id", ReferenceTable: "missions", ReferenceColumn: "id", OnDelete: "SET NULL"},
		{Name: "browser_sessions_mission_fk", Table: "browser_sessions", Columns: "mission_id", ReferenceTable: "missions", ReferenceColumn: "id", OnDelete: "CASCADE"},
		{Name: "browser_sessions_user_fk", Table: "browser_sessions", Columns: "user_id", ReferenceTable: "users", ReferenceColumn: "id", OnDelete: "SET NULL"},
		{Name: "recorder_sessions_mission_fk", Table: "recorder_sessions", Columns: "mission_id", ReferenceTable: "missions", ReferenceColumn: "id", OnDelete: "CASCADE"},
		{Name: "sensor_descriptors_mission_fk", Table: "sensor_descriptors", Columns: "mission_id", ReferenceTable: "missions", ReferenceColumn: "id", OnDelete: "CASCADE"},
		{Name: "sensor_descriptors_robot_fk", Table: "sensor_descriptors", Columns: "robot_id", ReferenceTable: "robots", ReferenceColumn: "id"},
		{Name: "sensor_samples_descriptor_identity_fk", Table: "sensor_samples", Columns: "descriptor_id, mission_id, robot_id, sensor_id", ReferenceTable: "sensor_descriptors", ReferenceColumn: "id, mission_id, robot_id, sensor_id", OnDelete: "CASCADE"},
		{Name: "recording_sessions_mission_fk", Table: "recording_sessions", Columns: "mission_id", ReferenceTable: "missions", ReferenceColumn: "id", OnDelete: "CASCADE"},
		{Name: "recording_sessions_robot_fk", Table: "recording_sessions", Columns: "robot_id", ReferenceTable: "robots", ReferenceColumn: "id"},
		{Name: "recording_sessions_recorder_session_fk", Table: "recording_sessions", Columns: "recorder_session_id", ReferenceTable: "recorder_sessions", ReferenceColumn: "id", OnDelete: "SET NULL"},
		{Name: "recording_chunks_recording_session_fk", Table: "recording_chunks", Columns: "recording_session_id", ReferenceTable: "recording_sessions", ReferenceColumn: "id", OnDelete: "CASCADE"},
		{Name: "recording_chunks_mission_fk", Table: "recording_chunks", Columns: "mission_id", ReferenceTable: "missions", ReferenceColumn: "id", OnDelete: "CASCADE"},
		{Name: "recording_chunks_robot_fk", Table: "recording_chunks", Columns: "robot_id", ReferenceTable: "robots", ReferenceColumn: "id"},
		{Name: "recording_chunks_manifest_object_fk", Table: "recording_chunks", Columns: "manifest_object_id", ReferenceTable: "storage_objects", ReferenceColumn: "id", OnDelete: "SET NULL"},
		{Name: "recording_finalization_jobs_chunk_fk", Table: "recording_finalization_jobs", Columns: "recording_chunk_id", ReferenceTable: "recording_chunks", ReferenceColumn: "id", OnDelete: "CASCADE"},
		{Name: "recording_finalization_jobs_session_fk", Table: "recording_finalization_jobs", Columns: "recording_session_id", ReferenceTable: "recording_sessions", ReferenceColumn: "id", OnDelete: "CASCADE"},
		{Name: "recording_finalization_jobs_mission_fk", Table: "recording_finalization_jobs", Columns: "mission_id", ReferenceTable: "missions", ReferenceColumn: "id", OnDelete: "CASCADE"},
		{Name: "recording_finalization_jobs_robot_fk", Table: "recording_finalization_jobs", Columns: "robot_id", ReferenceTable: "robots", ReferenceColumn: "id"},
		{Name: "storage_objects_mission_fk", Table: "storage_objects", Columns: "mission_id", ReferenceTable: "missions", ReferenceColumn: "id", OnDelete: "CASCADE"},
		{Name: "storage_objects_robot_fk", Table: "storage_objects", Columns: "robot_id", ReferenceTable: "robots", ReferenceColumn: "id", OnDelete: "SET NULL"},
		{Name: "storage_objects_recording_chunk_fk", Table: "storage_objects", Columns: "recording_chunk_id", ReferenceTable: "recording_chunks", ReferenceColumn: "id", OnDelete: "SET NULL"},
		{Name: "events_mission_fk", Table: "events", Columns: "mission_id", ReferenceTable: "missions", ReferenceColumn: "id", OnDelete: "CASCADE"},
		{Name: "events_robot_fk", Table: "events", Columns: "robot_id", ReferenceTable: "robots", ReferenceColumn: "id", OnDelete: "SET NULL"},
		{Name: "events_related_storage_object_fk", Table: "events", Columns: "related_storage_object_id", ReferenceTable: "storage_objects", ReferenceColumn: "id", OnDelete: "SET NULL"},
		{Name: "control_commands_mission_fk", Table: "control_commands", Columns: "mission_id", ReferenceTable: "missions", ReferenceColumn: "id", OnDelete: "CASCADE"},
		{Name: "control_commands_robot_fk", Table: "control_commands", Columns: "robot_id", ReferenceTable: "robots", ReferenceColumn: "id"},
		{Name: "control_commands_requested_by_fk", Table: "control_commands", Columns: "requested_by", ReferenceTable: "users", ReferenceColumn: "id", OnDelete: "SET NULL"},
		{Name: "control_acks_command_fk", Table: "control_acks", Columns: "control_command_id", ReferenceTable: "control_commands", ReferenceColumn: "id", OnDelete: "CASCADE"},
		{Name: "control_acks_robot_fk", Table: "control_acks", Columns: "robot_id", ReferenceTable: "robots", ReferenceColumn: "id"},
	}
	statements := make([]string, 0, len(constraints))
	for _, constraint := range constraints {
		statements = append(statements, constraint.statement())
	}
	return statements
}

func (constraint foreignKeyConstraint) statement() string {
	onDelete := ""
	if constraint.OnDelete != "" {
		onDelete = " ON DELETE " + constraint.OnDelete
	}
	return fmt.Sprintf(`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = '%s'
			) THEN
				ALTER TABLE %s
					ADD CONSTRAINT %s
					FOREIGN KEY (%s) REFERENCES %s(%s)%s;
			END IF;
		END $$`,
		constraint.Name,
		constraint.Table,
		constraint.Name,
		constraint.Columns,
		constraint.ReferenceTable,
		constraint.ReferenceColumn,
		onDelete,
	)
}

func dropEmptyTableStatement(tableName string) string {
	return fmt.Sprintf(`DO $$
		DECLARE
			row_count bigint;
		BEGIN
			IF to_regclass('public.%s') IS NOT NULL THEN
				EXECUTE 'SELECT count(*) FROM public.%s' INTO row_count;
				IF row_count = 0 THEN
					EXECUTE 'DROP TABLE public.%s';
				END IF;
			END IF;
		END $$`, tableName, tableName, tableName)
}

func (s *Store) seedBootstrapUser(db *gorm.DB) error {
	user := model.UserModel{
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
