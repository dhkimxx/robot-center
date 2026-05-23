package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresConfig struct {
	DSN         string
	ServerURL   string
	MinIOBucket string
}

type PostgresStore struct {
	db          *gorm.DB
	sqlDB       *sql.DB
	sqlTx       *sql.Tx
	serverURL   string
	minioBucket string
}

func stringWithDefault(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func NewPostgresStore(ctx context.Context, cfg PostgresConfig) (*PostgresStore, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	store := &PostgresStore{
		db:          db,
		sqlDB:       sqlDB,
		serverURL:   cfg.ServerURL,
		minioBucket: stringWithDefault(strings.TrimSpace(cfg.MinIOBucket), "robot-center"),
	}
	if err := store.ensureP0Schema(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return store, nil
}

func (s *PostgresStore) Close() error {
	return s.sqlDB.Close()
}

func (s *PostgresStore) WithTransaction(ctx context.Context, run func(ctx context.Context, repository Store) error) error {
	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	transactionalStore := *s
	transactionalStore.sqlTx = tx
	if err := run(ctx, &transactionalStore); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *PostgresStore) ensureP0Schema(ctx context.Context) error {
	return s.db.WithContext(ctx).Exec(`
		ALTER TABLE robots
			ADD COLUMN IF NOT EXISTS archived_at timestamptz;

		ALTER TABLE robot_tokens
			ADD COLUMN IF NOT EXISTS token_plaintext text;

		ALTER TABLE recording_chunks
			ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now(),
			ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now();

		CREATE TABLE IF NOT EXISTS streaming_statuses (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			robot_id uuid NOT NULL REFERENCES robots(id) ON DELETE CASCADE,
			mission_id uuid REFERENCES missions(id) ON DELETE CASCADE,
			room_id text NOT NULL,
			status text NOT NULL,
			published_tracks jsonb NOT NULL DEFAULT '[]'::jsonb,
			published_data_channels jsonb NOT NULL DEFAULT '[]'::jsonb,
			sent_at timestamptz,
			updated_at timestamptz NOT NULL DEFAULT now(),
			UNIQUE(robot_id)
		);
	`).Error
}

func (s *PostgresStore) nextCodeWithGorm(tx *gorm.DB, prefix string, tableName string) (string, error) {
	var count int64
	if err := tx.Table(tableName).Count(&count).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%03d", prefix, count+1), nil
}

func (s *PostgresStore) nextCode(ctx context.Context, tx *sql.Tx, prefix string, tableName string) (string, error) {
	var count int
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, tableName)
	if err := tx.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%03d", prefix, count+1), nil
}
