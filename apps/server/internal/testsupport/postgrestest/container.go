package postgrestest

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
	postgresdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	databaseName = "robot_center_test"
	username     = "robot_center"
	password     = "robot_center"
)

type PostgresContainer struct {
	DSN string
}

var databaseSequence uint64

func Start(t *testing.T) PostgresContainer {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)

	container, terminate, err := StartContainer(ctx)
	if err != nil {
		t.Fatalf("start PostgreSQL test container: %v", err)
	}
	t.Cleanup(func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		if err := terminate(shutdownCtx); err != nil {
			t.Fatalf("terminate PostgreSQL test container: %v", err)
		}
	})
	return container
}

func StartContainer(ctx context.Context) (PostgresContainer, func(context.Context) error, error) {
	configureDockerHostForLocalRuntime()
	// The tests terminate each container explicitly. Disabling Ryuk avoids
	// reaper name conflicts when multiple packages start containers in parallel.
	if os.Getenv("TESTCONTAINERS_RYUK_DISABLED") == "" {
		_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}

	container, err := postgres.Run(ctx,
		"postgis/postgis:16-3.4",
		postgres.WithDatabase(databaseName),
		postgres.WithUsername(username),
		postgres.WithPassword(password),
	)
	if err != nil {
		return PostgresContainer{}, nil, err
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		return PostgresContainer{}, nil, err
	}
	return PostgresContainer{DSN: dsn}, func(ctx context.Context) error {
		return container.Terminate(ctx)
	}, nil
}

func CreateDatabase(t *testing.T, baseDSN string) string {
	t.Helper()

	databaseName := fmt.Sprintf("robot_center_test_%d", atomic.AddUint64(&databaseSequence, 1))
	adminDB := openAdminDatabase(t, baseDSN)
	if err := adminDB.Exec("CREATE DATABASE " + quoteIdentifier(databaseName)).Error; err != nil {
		t.Fatalf("create PostgreSQL test database %s: %v", databaseName, err)
	}
	closeGormDatabase(t, adminDB)

	t.Cleanup(func() {
		cleanupDB := openAdminDatabase(t, baseDSN)
		defer closeGormDatabase(t, cleanupDB)
		if err := cleanupDB.Exec(`
			SELECT pg_terminate_backend(pid)
			FROM pg_stat_activity
			WHERE datname = ?
				AND pid <> pg_backend_pid()
		`, databaseName).Error; err != nil {
			t.Fatalf("terminate PostgreSQL test database connections %s: %v", databaseName, err)
		}
		if err := cleanupDB.Exec("DROP DATABASE IF EXISTS " + quoteIdentifier(databaseName)).Error; err != nil {
			t.Fatalf("drop PostgreSQL test database %s: %v", databaseName, err)
		}
	})

	return dsnWithDatabase(t, baseDSN, databaseName)
}

func openAdminDatabase(t *testing.T, dsn string) *gorm.DB {
	t.Helper()

	var lastErr error
	deadline := time.Now().Add(30 * time.Second)
	for {
		db, err := gorm.Open(postgresdriver.Open(dsn), &gorm.Config{})
		if err == nil {
			sqlDB, dbErr := db.DB()
			if dbErr == nil {
				if pingErr := sqlDB.Ping(); pingErr == nil {
					return db
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
			t.Fatalf("open PostgreSQL admin database: %v", lastErr)
		}
		time.Sleep(time.Second)
	}
}

func closeGormDatabase(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("resolve PostgreSQL admin sql DB: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close PostgreSQL admin database: %v", err)
	}
}

func dsnWithDatabase(t *testing.T, dsn string, databaseName string) string {
	t.Helper()
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse PostgreSQL test DSN: %v", err)
	}
	parsed.Path = "/" + databaseName
	return parsed.String()
}

func quoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func configureDockerHostForLocalRuntime() {
	if os.Getenv("DOCKER_HOST") != "" {
		return
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	orbstackSocket := filepath.Join(homeDir, ".orbstack", "run", "docker.sock")
	if _, err := os.Stat(orbstackSocket); err == nil {
		_ = os.Setenv("DOCKER_HOST", "unix://"+orbstackSocket)
	}
}
