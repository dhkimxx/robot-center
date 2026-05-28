package port

import "context"

type StorageMetadataResetResult struct {
	StorageObjectRowsDeleted int64
	RecordingChunksReset     int64
}

type StorageAdminStore interface {
	ResetObjectStorageMetadata(ctx context.Context) (StorageMetadataResetResult, error)
}
