# Control Center Live Refactor Result - 2026-05-25

## Scope

- Live connection reconnect race fixed with `LiveConnectionAttempt` / `attemptId` guards.
- WebSocket/RTCPeerConnection events now update session state only when the event attempt is current.
- Live connection status and close reason policy moved to `liveConnectionStates.js`.
- Live session state mutation moved to `useLiveSessionStore.js`.
- Live target selection and localStorage persistence moved to `useLiveTargetSelection.js`.
- Live connection registry moved to `useLiveConnectionRegistry.js`.
- Auto-connect policy moved to `useLiveAutoConnect.js`.
- `useLiveConnectionManager.js` reduced to a facade-level orchestration hook.
- App route/modal props grouped by domain instead of passing the full controller object.
- Pure function tests added for live payload mapping, live helpers, mission helpers, route helpers, and stale live attempt guards.
- Recording playback modal was moved into the recordings domain file boundary.

## Verification Commands

```bash
cd /Users/dhkim/workspace/sst/robot-center/apps/web
npm run build
npm test
```

```bash
cd /Users/dhkim/workspace/sst/robot-center
./scripts/dev-status.sh
```

## Verification Results

- `npm run build`: passed.
- `npm test -- --run`: passed, 5 test files / 13 tests.
- `app-server`: ok.
- `recorder-worker`: ok.
- `system`: ok.
- PostgreSQL container: healthy.
- MinIO container: healthy.
- Python mock robots: `robot-center-pyrobot-001`, `robot-center-pyrobot-002` running.
- Recorder subscriber: `mission-006`, ICE connected, 6 tracks, 4 data channels.

## Browser Checks

URL:

```text
http://127.0.0.1:18080/missions
http://127.0.0.1:18080/missions/mission-006/control
http://127.0.0.1:18080/recordings
```

Observed:

- `/missions` rendered without error boundary.
- `mission-006` control view rendered without error boundary.
- Initial selected robot `robot-001` connected.
- RGB video rendered at `1280x720`.
- Thermal video rendered at `640x480`.
- Position panel showed a current/last position state.
- Sensor panel rendered with live values.
- Robot selection changed to `robot-002`.
- After switching, `robot-002` connected and RGB/Thermal videos rendered.
- `/recordings` rendered robot-grouped recording list.
- Clicking `재생` opened an embedded playback modal with video ready state.
- Browser console warning/error logs: none.

## Notes

- The `재연결` button is intentionally hidden while the selected robot is connected.
- Verified control view re-entry: mission list -> same active mission control -> automatic connection resumed.
- The active development stack was left running for user verification.
