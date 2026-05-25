# Mission Room Selective Subscribe PoC 문서화 결과

## 구현 범위

- app-server 내부 SFU, mission room, multi robot, multi operator 구조를 harness 문서에 반영했다.
- Robot은 mission room에 한 번만 publish 하고 operator subscriber가 선택한 robotCode만 subscribe 하는 내부 PoC 동작을 명시했다.
- Recorder subscriber는 같은 mission room의 모든 robotCode를 subscribe 하는 기준을 명시했다.
- Selective subscribe signaling은 내부 PoC 임시 동작이며 장기 API/schema 계약이 아님을 명시했다.
- 런타임 QA 체크리스트에 operator 선택, recorder all-subscribe, publisher 중복 방지, subscriber 독립성 검증 항목을 추가했다.
- SFU operator subscriber session에 선택 robotCode 상태를 추가했다.
- operator subscriber는 선택 robotCode의 RGB/Thermal/Audio track과 sensor/telemetry DataChannel message만 수신한다.
- recorder subscriber는 기존처럼 mission room 전체 robotCode를 수신한다.
- mission start 시 배정 robot이 다른 active mission에 있으면 409 Conflict와 conflict list를 반환한다.
- Python mock shared mission 실행 시 기존 단일 active mission과 충돌하면 해당 mission을 종료하고 shared mission을 시작하도록 보완했다.
- 관제 UI는 `관제 로봇` 선택값을 signaling join/query와 `select-robot` 내부 메시지로 전달한다.
- 관제 UI 수신 상태는 선택 로봇 기준으로만 `연결됨`을 표시한다.

## 변경 문서

- `docs/harness/20260522-webrtc-sfu-topology.md`
- `docs/harness/20260523-multi-robot-sfu-checklist.md`
- `docs/harness/20260522-harness-index.md`
- `docs/rebuild-results/mission-room-selective-subscribe-poc-2026-05-23.md`

## 변경 코드

- `apps/server/internal/sfu/*`
- `apps/server/internal/store/*`
- `apps/server/internal/api/server.go`
- `apps/server/internal/api/server_test.go`
- `apps/web/src/hooks/useControlCenterController.js`
- `apps/web/src/domains/missions/MissionsScreen.jsx`
- `scripts/python-mock-robots-up.sh`

## 검증 방법

정적 검증:

```bash
cd /Users/dhkim/workspace/sst/robot-center/apps/server
go test -count=1 ./...
go vet ./...

cd /Users/dhkim/workspace/sst/robot-center/apps/web
npm run build
```

런타임 검증:

```bash
cd /Users/dhkim/workspace/sst/robot-center
./scripts/dev-up.sh
MOCK_ROBOT_COUNT=2 MOCK_SHARED_MISSION=1 APP_SERVER_URL=http://127.0.0.1:18080 ./scripts/python-mock-robots-up.sh
./scripts/dev-status.sh
```

## 실제 확인 결과

- app-server, recorder-worker health OK.
- PostgreSQL, MinIO healthy.
- TURN relay-only config 확인: `turn:172.30.1.41:3478?transport=udp`.
- Python mock robot 2대 실행:
  - `robot-001 -> mission-009`
  - `robot-002 -> mission-009`
- `mission-009` SFU room:
  - robot publisher 2대
  - recorder subscriber 1개
  - published track 6개: `robot-001/robot-002` 각각 `audio/rgb/thermal`
- recorder-worker:
  - `mission-009` signaling stable
  - ICE connected
  - trackCount 6
  - DataChannel 2개
  - `robotCodes`: `robot-001`, `robot-002`
  - sensor/telemetry persisted count 증가 확인
- browser A/B headless runtime:
  - browser A selected `mission-009:robot-001`
  - browser B selected `mission-009:robot-002`
  - 두 browser 모두 RGB `1280x720`, Thermal `640x480` video readyState 4 확인
  - browser A UI 상태: `선택 robot-001 연결됨 / 연결 1/2대`
  - browser B UI 상태: `선택 robot-002 연결됨 / 연결 1/2대`
  - SFU room에서 operator subscriber별 selected robotCode가 각각 `robot-001`, `robot-002`로 표시됨

## 사용자 확인 URL

- UI: `http://172.30.1.41:18080`
- Local UI: `http://127.0.0.1:18080`
- MinIO console: `http://127.0.0.1:9001`

## 남은 한계

- Selective subscribe signaling의 JSON schema는 이 문서에서 정의하지 않았다.
- 장기 subscription API 계약, 재협상 정책, track metadata format은 별도 설계가 필요하다.
- Android Mock Robot은 이번 검증에서 사용하지 않았다. 현재 adb device unavailable 상태라 Python mock robot 2대로 대체 검증했다.
