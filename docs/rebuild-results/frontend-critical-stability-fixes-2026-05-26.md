# Frontend Critical Stability Fixes - 2026-05-26

## Scope

- Hardened WebRTC signaling message handling in the operator live connection client.
- Added timeout and malformed-response handling to the shared JSON API client.
- Replaced overlapping interval polling with sequential polling and stale-response guards.

## Changed Files

- `apps/web/src/api/controlCenterApi.js`
- `apps/web/src/api/controlCenterApi.test.js`
- `apps/web/src/domains/live/liveConnectionClient.js`
- `apps/web/src/domains/live/liveConnectionClient.test.js`
- `apps/web/src/domains/live/liveConnectionHandlers.js`
- `apps/web/src/hooks/useControlCenterData.js`

## Verification

- `cd apps/web && npm test -- --run`
  - Result: 8 test files passed, 21 tests passed.
- `cd apps/web && npm run build`
  - Result: build passed.
- `./scripts/dev-status.sh`
  - Result: app-server OK, recorder-worker OK, system OK.
- Browser verification at `http://127.0.0.1:18080`
  - `/missions`: rendered without error boundary.
  - `/missions/mission-006/control`: rendered without error boundary.
  - Live RGB video: 1280x720, readyState 4.
  - Live Thermal video: 640x480, readyState 4.
  - GPS and sensor values visible.
  - `/robots`, `/recordings`, `/system`: rendered without error boundary.

## Notes

- `react-icons` bundle size cleanup was intentionally left out because this pass only covered required stability fixes.
