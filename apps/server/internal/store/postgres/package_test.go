package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"robot-center/apps/server/internal/testsupport/postgrestest"
)

var sharedPostgresDSN string

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, terminate, err := postgrestest.StartContainer(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start PostgreSQL test container: %v\n", err)
		os.Exit(1)
	}
	sharedPostgresDSN = container.DSN

	code := m.Run()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := terminate(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "terminate PostgreSQL test container: %v\n", err)
		if code == 0 {
			code = 1
		}
	}
	shutdownCancel()
	os.Exit(code)
}
