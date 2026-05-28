package postgres

import (
	"context"

	repo "robot-center/apps/server/internal/store/port"
)

func (s *Store) ResetObjectStorageMetadata(ctx context.Context) (repo.StorageMetadataResetResult, error) {
	if s.sqlTx == nil {
		tx, err := s.sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return repo.StorageMetadataResetResult{}, err
		}
		transactionalStore := *s
		transactionalStore.sqlTx = tx
		result, err := transactionalStore.ResetObjectStorageMetadata(ctx)
		if err != nil {
			_ = tx.Rollback()
			return repo.StorageMetadataResetResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return repo.StorageMetadataResetResult{}, err
		}
		return result, nil
	}

	runner := s.sqlRunner()
	recordingChunkResult, err := runner.ExecContext(ctx, `
		UPDATE recording_chunks
		SET
			manifest_object_id = NULL,
			metadata = jsonb_set(metadata - 'availableFileTypes', '{availableFileTypes}', '{}'::jsonb, true),
			status = CASE WHEN status = 'uploaded' THEN 'partial' ELSE status END,
			updated_at = now()
		WHERE manifest_object_id IS NOT NULL
			OR metadata ? 'availableFileTypes'
			OR status = 'uploaded'
	`)
	if err != nil {
		return repo.StorageMetadataResetResult{}, err
	}
	storageObjectResult, err := runner.ExecContext(ctx, `DELETE FROM storage_objects`)
	if err != nil {
		return repo.StorageMetadataResetResult{}, err
	}

	recordingChunksReset, _ := recordingChunkResult.RowsAffected()
	storageObjectRowsDeleted, _ := storageObjectResult.RowsAffected()
	return repo.StorageMetadataResetResult{
		StorageObjectRowsDeleted: storageObjectRowsDeleted,
		RecordingChunksReset:     recordingChunksReset,
	}, nil
}
