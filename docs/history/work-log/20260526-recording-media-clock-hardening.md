---
title: Recording Media Clock Hardening
status: history
date: 2026-05-26
owner: danya.kim
tags: ["recording", "recorder-worker", "webrtc", "rtp", "chunk"]
---

# Recording Media Clock Hardening

## Scope

- Recording chunk creation no longer happens only because a mission is active.
- Recorder-worker now caches active recording targets and opens a chunk when media packets arrive.
- Recording chunk windows are based on recording session start time, not mission start time.
- H264 MP4 muxing no longer uses a hard-coded 30 fps input rate. It uses observed RTP timestamp cadence when available.

## Implementation Notes

- `recording_sessions.started_at` is created from the first recorder tick, which is now triggered by media arrival.
- Existing open recording sessions keep their original `started_at`, and chunk indexes are computed from that session start.
- H264 packet timestamps are tracked per chunk and per track label. The observed input fps is passed to ffmpeg when muxing raw H264 snapshots.
- Data channel payloads are still attached only when a media-driven chunk already exists.

## Verification

```text
cd apps/server
go test ./...
```

Result: passed.

## Remaining Work

- This is not yet a full RTP-sample MP4 muxer. Raw H264 is still muxed through ffmpeg, now with observed RTP fps instead of fixed 30 fps.
- Chunk rollover is still time-window based. Keyframe-aware rollover with RTCP PLI/FIR is still required for production-grade recording.
- Track-level actual duration, frame count, packet count, and finalize reason are not yet persisted in DB.
