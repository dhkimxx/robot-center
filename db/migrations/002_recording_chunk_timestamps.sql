ALTER TABLE recording_chunks
  ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now(),
  ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now();

UPDATE recording_chunks
SET
  created_at = COALESCE(created_at, started_at, now()),
  updated_at = COALESCE(updated_at, started_at, now());
