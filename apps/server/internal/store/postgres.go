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
	db, sqlDB, err := openPostgresWithRetry(ctx, cfg.DSN)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	store := &PostgresStore{
		db:          db,
		sqlDB:       sqlDB,
		serverURL:   cfg.ServerURL,
		minioBucket: stringWithDefault(strings.TrimSpace(cfg.MinIOBucket), "robot-center"),
	}
	if err := store.runAutoMigrations(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return store, nil
}

func openPostgresWithRetry(ctx context.Context, dsn string) (*gorm.DB, *sql.DB, error) {
	var lastErr error
	deadline := time.Now().Add(30 * time.Second)
	for {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			sqlDB, dbErr := db.DB()
			if dbErr == nil {
				if pingErr := sqlDB.PingContext(ctx); pingErr == nil {
					return db, sqlDB, nil
				} else {
					lastErr = pingErr
				}
				_ = sqlDB.Close()
			} else {
				lastErr = dbErr
			}
		} else {
			lastErr = err
		}

		if time.Now().After(deadline) {
			return nil, nil, lastErr
		}
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
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
