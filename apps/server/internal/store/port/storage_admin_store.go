package port

import "context"

type StorageMetadataResetResult struct {
	StorageObjectRowsDeleted int64
	RecordingChunksReset     int64
}

type StorageObjectPruneCandidate struct {
	ObjectKey string
	SizeBytes int64
}

type StorageAdminStore interface {
	ResetObjectStorageMetadata(ctx context.Context) (StorageMetadataResetResult, error)
	ListPrunableObjectStorageMetadata(ctx context.Context) ([]StorageObjectPruneCandidate, error)
	ResetPrunedObjectStorageMetadata(ctx context.Context, objectKeys []string) (StorageMetadataResetResult, error)
}
