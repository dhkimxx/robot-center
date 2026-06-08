---
title: stale-recording-session-reuse
created: '2026-06-08'
updated: '2026-06-08'
author: danya.kim <danya.kim@thundersoft.com>
editors:
- danya.kim <danya.kim@thundersoft.com>
type: incident
tags:
- recording
- troubleshooting
- d-srv-07
history:
- '2026-06-08 danya.kim <danya.kim@thundersoft.com>: 최초 작성'
---
## Summary

Recorder playback validation exposed stale recording sessions being reused after an old active mission or recorder restart. New recording chunks could continue from a very large chunk index instead of starting a fresh session.

## Root Cause

The PostgreSQL recording session lookup reused any session with open recording chunks, even when the incoming media timestamp was outside the active chunk window.

## Fix

The repository now only recovers an existing session when the incoming tick timestamp is inside an open chunk window. Superseded open chunks are queued for finalization and a new recording session starts from chunk index 0.

## Verification

- `go test ./internal/recording ./internal/service ./internal/store/postgres ./internal/api`
- Python Mock Robot 2대 E2E with `mission-017`
- Mission end finalized both robot chunks to `uploaded`
- MinIO RGB MP4 returned `200 OK`
