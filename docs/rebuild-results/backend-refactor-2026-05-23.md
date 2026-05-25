# Backend Refactor Result - 2026-05-23

## Scope

Refactored backend code without changing the current PoC behavior.

Completed safe slices from the refactoring instruction:

- Store tidy first
- Recording domain rule extraction
- RecordingService normalization boundary
- PostgresStore file split by responsibility
- SFU Hub helper/registry split
- Recorder Worker app-server/object-storage/subscriber helper split

## Main Changes

### Store

- Moved shared store input/error types to `apps/server/internal/store/types.go`.
- Moved token helper logic to `apps/server/internal/store/token.go`.
- Split PostgreSQL persistence by responsibility:
  - `postgres_robot.go`
  - `postgres_mission.go`
  - `postgres_streaming.go`
  - `postgres_telemetry.go`
  - `postgres_sensor.go`
  - `postgres_recording.go`
  - `postgres_scan.go`

### Recording Domain

- Added shared recording rule helpers in `apps/server/internal/domain/recording_rules.go`.
- Added tests for:
  - chunk window calculation
  - MinIO object key generation
  - recording manifest generation
- Memory and PostgreSQL stores now share the same recording chunk/key/manifest rules.

### Service

- `RecordingService.ApplyRecordingTick` now owns request normalization:
  - trim mission/robot code
  - default chunk duration
  - default tick time
- Added service test coverage for this normalization behavior.

### SFU

- Split `hub.go` helper responsibilities into:
  - `types.go`
  - `signaling.go`
  - `media.go`
  - `data_channel.go`
  - `peer.go`
  - `room_registry.go`
  - `random.go`
- Kept Pion negotiation and multi-publisher behavior unchanged.

### Recorder Worker

- Split app-server HTTP calls into `app_server_client.go`.
- Split MinIO/object upload logic into `object_storage.go`.
- Split recorder subscriber helper/grouping logic into `subscriber_helpers.go`.
- Kept subscriber, media recording, muxing, and upload behavior unchanged.

## Verification

Executed:

```bash
cd /Users/dhkim/workspace/sst/robot-center/apps/server
go test ./...
go vet ./...
```

Result: passed.

## Second Pass Fixup

Additional risk fixes after review:

- Wrapped `RecordingService.MarkRecordingChunkUploaded` in the service transaction boundary.
- Wrapped `RecordingService.MarkRecordingFileUploaded` in the service transaction boundary.
- Added service tests that fail if upload status changes bypass the transaction repository.
- Fixed `002_recording_chunk_timestamps.sql` to add nullable columns, backfill from `started_at`, then set defaults and `NOT NULL`.
- Added `003_fix_recording_chunk_timestamps.sql` for already-applied PoC DBs.
- Applied the timestamp migration to the local PoC DB.
- Moved SFU subscriber track attach, DataChannel creation, and RTCP read loop into `subscriberSession`.

Verification:

```bash
cd /Users/dhkim/workspace/sst/robot-center/apps/server
go test ./...
go vet ./...
```

Result: passed.

Runtime verification:

- app-server: ok
- recorder-worker: ok
- robot publisher: Python mock `robot-001`
- mission room: `mission-008`
- SFU room peers: robot 1, operator 1, recorder 1
- published tracks: `robot-001:rgb`, `robot-001:thermal`, `robot-001:audio`
- recorder ICE: connected
- recorder DataChannels: 2
- browser/operator UI: connected
- browser video dimensions: RGB 1280x720, Thermal 640x480
- DB sample: `mission-008` uploaded chunk has `manifest_object_id` and 5 `storage_objects`
- DB backfill check: 404 existing recording chunks have `created_at = started_at`

## Operational Fixup

Additional operational fixes:

- Added `scripts/db-migrate.sh` so existing PostgreSQL volumes can receive `db/migrations/*.sql` without dropping data.
- Wired `scripts/dev-up.sh` to run `scripts/db-migrate.sh` after PostgreSQL starts and before app-server starts.
- Re-ran migrations against the local PoC DB.
- Added upload failure tests for `MarkRecordingChunkUploaded` and `MarkRecordingFileUploaded`.
  - transaction repository error is propagated
  - transaction is not committed
  - outside repository is not called
- Moved more SFU offer responsibility into session helpers:
  - publisher robot code resolution and local answer generation
  - subscriber offer deferral/begin logic

Verification:

```bash
cd /Users/dhkim/workspace/sst/robot-center/apps/server
go test -count=1 ./...
go vet ./...

cd /Users/dhkim/workspace/sst/robot-center
./scripts/db-migrate.sh
./scripts/dev-status.sh
```

Runtime result:

- app-server: ok
- recorder-worker: ok
- PostgreSQL: healthy
- MinIO: healthy
- Python mock robot: `robot-001` running
- SFU room: `mission-008`
- SFU peers: robot 1, operator 1, recorder 1
- published tracks: `robot-001:rgb`, `robot-001:thermal`, `robot-001:audio`
- recorder ICE: connected
- recorder DataChannels: 2
- browser/operator UI: connected
- browser video dimensions: RGB 1280x720, Thermal 640x480
- latest DB recording chunk: uploaded, has manifest object, 5 storage objects
- latest MinIO objects include manifest, RGB MP4, Thermal MP4, sensor JSONL, telemetry JSONL

Checked runtime status:

```bash
cd /Users/dhkim/workspace/sst/robot-center
./scripts/dev-status.sh
```

Observed:

- app-server: ok
- recorder-worker: ok
- system: ok
- PostgreSQL container: healthy
- MinIO container: healthy
- UI URL: `http://192.168.20.8:18080`

## Current Runtime Note

After restarting the dev stack with `./scripts/dev-up.sh`, the runtime URLs are host-reachable:

```text
UI:  http://192.168.20.8:18080
SFU: ws://192.168.20.8:18080/sfu/ws
TURN: turn:192.168.20.8:3478?transport=udp
```

Recorder subscriber rooms are connected, but there are no active robot tracks because Android adb is unavailable and Python mock robots are stopped.

## Remaining Work

- Move more recorder media-buffer responsibilities out of `media_writer.go` if the next change touches RTP/media internals.
- Add deeper WebRTC negotiation regression tests before further SFU session extraction.
- Start Android Mock Robot or Python mock robots before live media/telemetry demonstration.

## Second Refactor Pass

Additional scope completed from the second refactoring instruction:

- `RecordingService.ApplyRecordingTick` now owns the recording tick use case flow:
  - input normalization
  - target lookup
  - active mission / robot assignment validation
  - session lookup or creation
  - chunk window calculation
  - object key generation
  - existing chunk lookup
  - chunk creation
  - manifest generation
- `RecordingRepository` now exposes storage primitives instead of a single use-case method.
- `PostgresStore.ApplyRecordingTick` and `MemoryStore.ApplyRecordingTick` were removed.
- `recording_chunks.created_at` and `recording_chunks.updated_at` are real DB columns.
- Added migration `db/migrations/002_recording_chunk_timestamps.sql`.
- Recording upload status updates now set `updated_at = now()`.
- Recorder Worker now receives `AppServerClient`, `ObjectStorage`, and `MediaUploader` collaborators through construction.
- Added a Worker tick unit test that runs without HTTP or MinIO.
- SFU publisher/subscriber negotiation setup was moved into `publisherSession` and `subscriberSession` helpers while keeping Hub orchestration.

Verification:

```bash
cd /Users/dhkim/workspace/sst/robot-center/apps/server
go test ./...
go vet ./...
```

Result: passed.
