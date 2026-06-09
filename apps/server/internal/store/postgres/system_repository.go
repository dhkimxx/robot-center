package postgres

import (
	"context"
	"database/sql"

	repo "robot-center/apps/server/internal/store/port"
)

func (s *Store) GetDatabaseUsage(ctx context.Context) (repo.DatabaseUsageResult, error) {
	result := repo.DatabaseUsageResult{Status: "ok"}
	if err := s.sqlRunner().QueryRowContext(ctx, `
		SELECT current_database(), pg_database_size(current_database())
	`).Scan(&result.DatabaseName, &result.DatabaseSizeBytes); err != nil {
		return result, err
	}

	rows, err := s.sqlRunner().QueryContext(ctx, `
		SELECT
			c.relname,
			GREATEST(COALESCE(st.n_live_tup, c.reltuples)::bigint, 0),
			pg_total_relation_size(c.oid)
		FROM pg_class c
		LEFT JOIN pg_stat_user_tables st ON st.relid = c.oid
		WHERE c.relkind = 'r'
			AND c.relnamespace = 'public'::regnamespace
		ORDER BY pg_total_relation_size(c.oid) DESC, c.relname ASC
	`)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	result.Tables, err = scanDatabaseTableUsage(rows)
	if err != nil {
		return result, err
	}
	for _, table := range result.Tables {
		result.TrackedTableBytes += table.TotalBytes
	}
	return result, rows.Err()
}

func scanDatabaseTableUsage(rows *sql.Rows) ([]repo.DatabaseTableUsage, error) {
	tables := make([]repo.DatabaseTableUsage, 0)
	for rows.Next() {
		var table repo.DatabaseTableUsage
		if err := rows.Scan(&table.TableName, &table.RowCount, &table.TotalBytes); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}
