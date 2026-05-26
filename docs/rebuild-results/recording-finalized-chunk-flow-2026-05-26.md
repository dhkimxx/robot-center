# Recording Finalized Chunk Flow - 2026-05-26

## Scope

- Changed recorder-worker so the current in-progress chunk is not uploaded as a finalized MP4 on every poll tick.
- The worker now keeps writing media to the active chunk and finalizes only when:
  - the recorder tick moves to a new chunk, or
  - the recording target disappears, such as mission end.
- Failed finalization is kept in a pending queue and retried on later ticks.

## Implementation Notes

- `worker.tick` now updates active chunks first, then queues previous chunks for finalization.
- `media_writer` now owns active chunk rotation helpers and pending finalization tracking.
- Audio writers are closed by chunk ID before muxing so a new active chunk writer is not closed accidentally.
- Existing upload APIs are reused:
  - file upload status: `/api/recorder/chunks/{chunkID}/files/{fileType}/uploaded`
  - manifest/chunk upload status: `/api/recorder/chunks/{chunkID}/uploaded`

## Verification

Commands:

```bash
cd /Users/dhkim/workspace/sst/robot-center/apps/server
go test ./...
go vet ./...
```

Runtime:

```bash
cd /Users/dhkim/workspace/sst/robot-center
SKIP_WEB_BUILD=1 ./scripts/dev-up.sh
./scripts/python-mock-robots-up.sh
./scripts/dev-status.sh
```

Observed:

- app-server: OK
- recorder-worker: OK
- PostgreSQL: healthy
- MinIO: healthy
- Python mock robots: `robot-001`, `robot-002` running
- recorder subscriber room: `mission-006`
- recorder ICE state: `connected`
- recorder tracks: 6
- recorder data channels: 4
- latest current chunks:
  - `mission-006 / robot-001 / chunk_index=75 / status=recording`
  - `mission-006 / robot-002 / chunk_index=75 / status=recording`
- latest MinIO storage objects remained on previously finalized chunks; current chunk 75 was not uploaded as a finalized MP4 immediately.

User URL:

```text
http://192.168.20.32:18080
```

## Remaining Limits

- MP4 duration accuracy still depends on the current H264/ffmpeg muxing path.
- Existing chunks created by the previous snapshot uploader remain in DB/MinIO.
- If the recorder process is killed before it can finalize pending local chunks, recovery of stale local spool files is not implemented yet.
