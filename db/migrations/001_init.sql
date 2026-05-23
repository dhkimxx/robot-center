CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  login_id text NOT NULL UNIQUE,
  password_hash text NOT NULL,
  display_name text NOT NULL,
  role text NOT NULL CHECK (role IN ('operator', 'commander', 'admin')),
  is_active boolean NOT NULL DEFAULT true,
  last_login_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS robots (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  robot_code text NOT NULL UNIQUE,
  display_name text NOT NULL,
  model_name text,
  status text NOT NULL DEFAULT 'offline',
  last_seen_at timestamptz,
  last_streaming_at timestamptz,
  archived_at timestamptz,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT robots_status_check CHECK (status IN ('offline', 'online', 'assigned', 'streaming', 'reconnecting', 'fault'))
);

CREATE TABLE IF NOT EXISTS robot_tokens (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  robot_id uuid NOT NULL REFERENCES robots(id) ON DELETE CASCADE,
  token_hash text NOT NULL,
  token_plaintext text,
  name text NOT NULL,
  is_active boolean NOT NULL DEFAULT true,
  last_used_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE robot_tokens
  ADD COLUMN IF NOT EXISTS token_plaintext text;

ALTER TABLE robots
  ADD COLUMN IF NOT EXISTS archived_at timestamptz;

CREATE TABLE IF NOT EXISTS missions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_code text NOT NULL UNIQUE,
  name text NOT NULL,
  mission_type text NOT NULL CHECK (mission_type IN ('mountain_rescue', 'collapse_site', 'underground_facility')),
  status text NOT NULL DEFAULT 'ready' CHECK (status IN ('ready', 'active', 'ended', 'cancelled')),
  created_by uuid REFERENCES users(id),
  site_note text,
  started_at timestamptz,
  ended_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS mission_robots (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id uuid NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  robot_id uuid NOT NULL REFERENCES robots(id) ON DELETE CASCADE,
  role text NOT NULL DEFAULT 'primary',
  status text NOT NULL DEFAULT 'assigned' CHECK (status IN ('assigned', 'active', 'completed', 'removed')),
  joined_at timestamptz,
  left_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS mission_robots_active_unique
  ON mission_robots(mission_id, robot_id)
  WHERE status != 'removed';

CREATE TABLE IF NOT EXISTS robot_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  robot_id uuid NOT NULL REFERENCES robots(id) ON DELETE CASCADE,
  mission_id uuid REFERENCES missions(id) ON DELETE SET NULL,
  state text NOT NULL,
  client_ip inet,
  user_agent text,
  connected_at timestamptz NOT NULL DEFAULT now(),
  last_heartbeat_at timestamptz NOT NULL DEFAULT now(),
  disconnected_at timestamptz,
  raw_payload jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS browser_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id uuid NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  state text NOT NULL,
  connected_at timestamptz NOT NULL DEFAULT now(),
  disconnected_at timestamptz,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS recorder_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id uuid NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  state text NOT NULL,
  started_at timestamptz NOT NULL DEFAULT now(),
  stopped_at timestamptz,
  last_error text,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS streaming_statuses (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  robot_id uuid NOT NULL REFERENCES robots(id) ON DELETE CASCADE,
  mission_id uuid REFERENCES missions(id) ON DELETE CASCADE,
  room_id text NOT NULL,
  status text NOT NULL,
  published_tracks jsonb NOT NULL DEFAULT '[]'::jsonb,
  published_data_channels jsonb NOT NULL DEFAULT '[]'::jsonb,
  sent_at timestamptz,
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(robot_id)
);

CREATE TABLE IF NOT EXISTS telemetry_snapshots (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id uuid NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  robot_id uuid NOT NULL REFERENCES robots(id) ON DELETE CASCADE,
  sequence bigint,
  sent_at timestamptz,
  received_at timestamptz NOT NULL DEFAULT now(),
  battery_percent numeric,
  network_quality text,
  position_type text,
  latitude double precision,
  longitude double precision,
  altitude_meter double precision,
  accuracy_meter double precision,
  heading_degree double precision,
  geom geometry(Point, 4326),
  raw_payload jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS telemetry_snapshots_mission_received_idx ON telemetry_snapshots(mission_id, received_at DESC);
CREATE INDEX IF NOT EXISTS telemetry_snapshots_robot_received_idx ON telemetry_snapshots(robot_id, received_at DESC);
CREATE INDEX IF NOT EXISTS telemetry_snapshots_geom_idx ON telemetry_snapshots USING gist(geom);

CREATE TABLE IF NOT EXISTS sensor_readings (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id uuid NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  robot_id uuid NOT NULL REFERENCES robots(id) ON DELETE CASCADE,
  sequence bigint,
  sent_at timestamptz,
  received_at timestamptz NOT NULL DEFAULT now(),
  temperature_celsius numeric,
  humidity_percent numeric,
  oxygen_percent numeric,
  co_ppm numeric,
  ch4_ppm numeric,
  raw_payload jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS sensor_readings_mission_received_idx ON sensor_readings(mission_id, received_at DESC);
CREATE INDEX IF NOT EXISTS sensor_readings_robot_received_idx ON sensor_readings(robot_id, received_at DESC);

CREATE TABLE IF NOT EXISTS recording_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id uuid NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  robot_id uuid NOT NULL REFERENCES robots(id) ON DELETE CASCADE,
  recorder_session_id uuid REFERENCES recorder_sessions(id) ON DELETE SET NULL,
  status text NOT NULL DEFAULT 'pending',
  chunk_duration_seconds integer NOT NULL DEFAULT 600,
  started_at timestamptz NOT NULL DEFAULT now(),
  ended_at timestamptz,
  last_error text,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  CONSTRAINT recording_sessions_status_check CHECK (status IN ('pending', 'recording', 'finalizing', 'uploaded', 'failed'))
);

CREATE TABLE IF NOT EXISTS recording_chunks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  recording_session_id uuid NOT NULL REFERENCES recording_sessions(id) ON DELETE CASCADE,
  mission_id uuid NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  robot_id uuid NOT NULL REFERENCES robots(id) ON DELETE CASCADE,
  chunk_index integer NOT NULL,
  status text NOT NULL DEFAULT 'pending',
  started_at timestamptz NOT NULL,
  ended_at timestamptz,
  duration_seconds numeric,
  manifest_object_id uuid,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT recording_chunks_status_check CHECK (status IN ('pending', 'recording', 'finalizing', 'uploaded', 'failed'))
);

CREATE UNIQUE INDEX IF NOT EXISTS recording_chunks_session_index_unique
  ON recording_chunks(recording_session_id, chunk_index);

CREATE TABLE IF NOT EXISTS storage_objects (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id uuid NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  robot_id uuid REFERENCES robots(id) ON DELETE SET NULL,
  recording_chunk_id uuid REFERENCES recording_chunks(id) ON DELETE SET NULL,
  object_type text NOT NULL,
  bucket text NOT NULL,
  object_key text NOT NULL UNIQUE,
  content_type text,
  size_bytes bigint,
  checksum text,
  started_at timestamptz,
  ended_at timestamptz,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'recording_chunks_manifest_object_fk'
  ) THEN
    ALTER TABLE recording_chunks
      ADD CONSTRAINT recording_chunks_manifest_object_fk
      FOREIGN KEY (manifest_object_id) REFERENCES storage_objects(id) ON DELETE SET NULL;
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id uuid NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  robot_id uuid REFERENCES robots(id) ON DELETE SET NULL,
  event_type text NOT NULL,
  severity text NOT NULL CHECK (severity IN ('info', 'notice', 'warning', 'critical')),
  title text NOT NULL,
  description text,
  occurred_at timestamptz NOT NULL,
  geom geometry(Point, 4326),
  related_storage_object_id uuid REFERENCES storage_objects(id) ON DELETE SET NULL,
  raw_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS events_mission_occurred_idx ON events(mission_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS events_geom_idx ON events USING gist(geom);

CREATE TABLE IF NOT EXISTS control_commands (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id uuid NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  robot_id uuid NOT NULL REFERENCES robots(id) ON DELETE CASCADE,
  requested_by uuid REFERENCES users(id) ON DELETE SET NULL,
  command_type text NOT NULL CHECK (command_type IN ('estop', 'return_to_home', 'ptz', 'waypoint')),
  status text NOT NULL DEFAULT 'requested',
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  requested_at timestamptz NOT NULL DEFAULT now(),
  sent_at timestamptz,
  completed_at timestamptz,
  failure_reason text,
  CONSTRAINT control_commands_status_check CHECK (status IN ('requested', 'sent', 'accepted', 'rejected', 'executing', 'succeeded', 'failed', 'timeout'))
);

CREATE TABLE IF NOT EXISTS control_acks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  control_command_id uuid NOT NULL REFERENCES control_commands(id) ON DELETE CASCADE,
  robot_id uuid NOT NULL REFERENCES robots(id) ON DELETE CASCADE,
  ack_status text NOT NULL CHECK (ack_status IN ('accepted', 'rejected', 'executing', 'succeeded', 'failed')),
  message text,
  received_at timestamptz NOT NULL DEFAULT now(),
  raw_payload jsonb NOT NULL DEFAULT '{}'::jsonb
);

INSERT INTO users (login_id, password_hash, display_name, role)
VALUES ('operator', 'seed-password-placeholder', 'PoC Operator', 'operator')
ON CONFLICT (login_id) DO NOTHING;
