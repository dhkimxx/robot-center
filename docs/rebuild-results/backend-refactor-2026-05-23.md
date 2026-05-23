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

After restarting the dev stack with `SKIP_ANDROID=1 ./scripts/dev-up.sh`, the runtime URLs are host-reachable:

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
