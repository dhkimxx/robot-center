package postgres

import (
	"context"
	"fmt"
	"strings"

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

func (s *Store) ListPrunableObjectStorageMetadata(ctx context.Context) ([]repo.StorageObjectPruneCandidate, error) {
	rows, err := s.sqlRunner().QueryContext(ctx, `
		SELECT so.object_key, COALESCE(so.size_bytes, 0)
		FROM storage_objects so
		LEFT JOIN recording_chunks rc ON rc.id = so.recording_chunk_id
		WHERE so.object_key <> ''
			AND (
				so.recording_chunk_id IS NULL
				OR (
					rc.id IS NOT NULL
					AND rc.status IN ('uploaded', 'partial', 'failed')
					AND NOT EXISTS (
						SELECT 1
						FROM recording_finalization_jobs rfj
						WHERE rfj.recording_chunk_id = rc.id
							AND rfj.status IN ('queued', 'processing')
					)
				)
			)
		ORDER BY so.created_at ASC, so.object_key ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := []repo.StorageObjectPruneCandidate{}
	for rows.Next() {
		var candidate repo.StorageObjectPruneCandidate
		if err := rows.Scan(&candidate.ObjectKey, &candidate.SizeBytes); err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return candidates, nil
}

func (s *Store) ResetPrunedObjectStorageMetadata(ctx context.Context, objectKeys []string) (repo.StorageMetadataResetResult, error) {
	objectKeys = compactObjectKeys(objectKeys)
	if len(objectKeys) == 0 {
		return repo.StorageMetadataResetResult{}, nil
	}
	if s.sqlTx == nil {
		tx, err := s.sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return repo.StorageMetadataResetResult{}, err
		}
		transactionalStore := *s
		transactionalStore.sqlTx = tx
		result, err := transactionalStore.ResetPrunedObjectStorageMetadata(ctx, objectKeys)
		if err != nil {
			_ = tx.Rollback()
			return repo.StorageMetadataResetResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return repo.StorageMetadataResetResult{}, err
		}
		return result, nil
	}

	placeholders, args := makeObjectKeyPlaceholders(objectKeys)
	runner := s.sqlRunner()
	recordingChunkResult, err := runner.ExecContext(ctx, fmt.Sprintf(`
		UPDATE recording_chunks
		SET
			manifest_object_id = NULL,
			metadata = jsonb_set(metadata - 'availableFileTypes', '{availableFileTypes}', '{}'::jsonb, true),
			status = CASE WHEN status = 'uploaded' THEN 'partial' ELSE status END,
			updated_at = now()
		WHERE id IN (
			SELECT DISTINCT recording_chunk_id
			FROM storage_objects
			WHERE object_key IN (%s)
				AND recording_chunk_id IS NOT NULL
		)
	`, placeholders), args...)
	if err != nil {
		return repo.StorageMetadataResetResult{}, err
	}

	storageObjectResult, err := runner.ExecContext(ctx, fmt.Sprintf(`
		DELETE FROM storage_objects
		WHERE object_key IN (%s)
	`, placeholders), args...)
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

func compactObjectKeys(objectKeys []string) []string {
	seen := map[string]struct{}{}
	compacted := make([]string, 0, len(objectKeys))
	for _, objectKey := range objectKeys {
		objectKey = strings.TrimSpace(objectKey)
		if objectKey == "" {
			continue
		}
		if _, ok := seen[objectKey]; ok {
			continue
		}
		seen[objectKey] = struct{}{}
		compacted = append(compacted, objectKey)
	}
	return compacted
}

func makeObjectKeyPlaceholders(objectKeys []string) (string, []any) {
	placeholders := make([]string, 0, len(objectKeys))
	args := make([]any, 0, len(objectKeys))
	for index, objectKey := range objectKeys {
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+1))
		args = append(args, objectKey)
	}
	return strings.Join(placeholders, ", "), args
}
