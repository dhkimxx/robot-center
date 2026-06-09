package port

import "context"

type DatabaseUsageResult struct {
	Status            string
	DatabaseName      string
	DatabaseSizeBytes int64
	TrackedTableBytes int64
	Tables            []DatabaseTableUsage
}

type DatabaseTableUsage struct {
	TableName  string
	RowCount   int64
	TotalBytes int64
}

type SystemStore interface {
	GetDatabaseUsage(ctx context.Context) (DatabaseUsageResult, error)
}
