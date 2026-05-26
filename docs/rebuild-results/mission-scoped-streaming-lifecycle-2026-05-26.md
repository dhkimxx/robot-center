# Mission-Scoped Streaming Lifecycle - 2026-05-26

## Scope

- Changed UI streaming labels to use mission-scoped streaming status instead of global robot status.
- Blocked busy robots in mission creation UI and backend create/start paths.
- Closed SFU mission rooms when missions end.
- Marked streaming status as `stopped` when missions end.
- Updated Python mock robots to stop publishing when the active mission disappears.

## Runtime Rules

```text
Robot online != Mission streaming
```

- `robots.status` is treated as device connectivity state.
- `streaming_statuses.status` is treated as mission room media state.
- A robot is shown as `송출 중` only when:
  - mission status is `active`
  - streaming status mission id matches the mission
  - streaming status robot code matches the robot
  - streaming status is `streaming` or `publishing`
  - `sentAt` is fresh within 30 seconds

Ended missions now show `임무 종료` even if the same robot is streaming in another active mission.

## Verification

```bash
cd apps/server && go test ./...
cd apps/server && go vet ./...
cd apps/web && npm test -- --run
cd apps/web && npm run build
python3 -m py_compile apps/mock-robot-python/mock_robot.py
```

Runtime:

- Restarted app-server, recorder-worker, TURN, PostgreSQL, MinIO.
- Started Python mock robots for `robot-001` and `robot-002`.
- Verified active `mission-004` showed both robots as `송출 중`.
- Verified creating another mission with `robot-001` returned `409 Conflict`.
- Ended `mission-004`.
- Verified both robots changed to `stopped`.
- Verified mock logs contain `active mission disappeared; stopping publish` and `publish stopped`.
- Started `mission-005` with both Python mock robots.
- Verified UI:
  - `mission-005` shows `robot-001 송출 중 / robot-002 송출 중`.
  - ended missions show `임무 종료`.
  - no ended mission row shows `송출 중`.
  - `mission-005/control` renders RGB, Thermal, GPS/sensor data.

## Current Demo State

```text
UI: http://192.168.20.32:18080
Active mission: mission-005
Robots: robot-001, robot-002
Python mock robot sessions: running
Recorder: connected to mission-005
```
