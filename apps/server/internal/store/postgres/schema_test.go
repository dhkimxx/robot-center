package postgres

import "testing"

func TestAutoMigrateAppliesBaseModelSchema(t *testing.T) {
	store := newPostgresTestStore(t)

	for _, tableName := range baseModelTableNames() {
		assertBaseModelColumn(t, store, tableName, "id")
		assertBaseModelColumn(t, store, tableName, "created_at")
		assertBaseModelColumn(t, store, tableName, "updated_at")
		assertUpdatedAtTrigger(t, store, tableName)
	}
}

func TestAutoMigrateRemovesLegacySensorSampleValueColumns(t *testing.T) {
	store := newPostgresTestStore(t)

	assertColumnPresent(t, store, "sensor_samples", "values")
	for _, columnName := range []string{
		"numeric_value",
		"text_value",
		"bool_value",
		"vector_value",
		"object_value",
	} {
		assertColumnMissing(t, store, "sensor_samples", columnName)
	}
}

func assertColumnPresent(t *testing.T, store *Store, tableName string, columnName string) {
	t.Helper()

	var count int
	if err := store.sqlDB.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = 'public'
			AND table_name = $1
			AND column_name = $2
	`, tableName, columnName).Scan(&count); err != nil {
		t.Fatalf("query %s.%s schema: %v", tableName, columnName, err)
	}
	if count != 1 {
		t.Fatalf("expected %s.%s to exist once, got %d", tableName, columnName, count)
	}
}

func assertBaseModelColumn(t *testing.T, store *Store, tableName string, columnName string) {
	t.Helper()

	var isNullable string
	var columnDefault *string
	if err := store.sqlDB.QueryRow(`
		SELECT is_nullable, column_default
		FROM information_schema.columns
		WHERE table_schema = 'public'
			AND table_name = $1
			AND column_name = $2
	`, tableName, columnName).Scan(&isNullable, &columnDefault); err != nil {
		t.Fatalf("query %s.%s schema: %v", tableName, columnName, err)
	}
	if isNullable != "NO" {
		t.Fatalf("expected %s.%s to be NOT NULL, got nullable=%s", tableName, columnName, isNullable)
	}
	if columnDefault == nil || *columnDefault == "" {
		t.Fatalf("expected %s.%s to have a default", tableName, columnName)
	}
}

func assertColumnMissing(t *testing.T, store *Store, tableName string, columnName string) {
	t.Helper()

	var count int
	if err := store.sqlDB.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = 'public'
			AND table_name = $1
			AND column_name = $2
	`, tableName, columnName).Scan(&count); err != nil {
		t.Fatalf("query %s.%s schema: %v", tableName, columnName, err)
	}
	if count != 0 {
		t.Fatalf("expected %s.%s to be absent, got %d", tableName, columnName, count)
	}
}

func assertUpdatedAtTrigger(t *testing.T, store *Store, tableName string) {
	t.Helper()

	var count int
	if err := store.sqlDB.QueryRow(`
		SELECT COUNT(*)
		FROM pg_trigger
		WHERE tgrelid = $1::regclass
			AND tgname = $2
			AND NOT tgisinternal
	`, "public."+tableName, tableName+"_set_updated_at").Scan(&count); err != nil {
		t.Fatalf("query %s updated_at trigger: %v", tableName, err)
	}
	if count != 1 {
		t.Fatalf("expected %s updated_at trigger to exist once, got %d", tableName, count)
	}
}
