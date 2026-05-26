---
title: "live-status-ssot-phase1"
created: 2026-05-26
updated: 2026-05-26
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "log"
tags: ["live-status", "ssot", "webrtc", "recording", "sfu"]
history:
  - "2026-05-26 danya.kim <danya.kim@thundersoft.com>: initial entry"
---

# Live Status SSOT Phase 1

## 목적

관제 Live 화면의 `송출`, `녹화`, `연결` 상태를 각각 다른 API/DB row로 추측하지 않고, app-server가 합성한 단일 상태 API로 통일했다.

기준 API:

```http
GET /api/missions/{missionCode}/live-status
```

## 구현 요약

- app-server에 mission live-status API를 추가했다.
- stream 상태는 app-server 내부 SFU observed publisher 기준으로 계산한다.
- recording 상태는 recorder-worker `/healthz`의 room/robot runtime 기준으로 계산한다.
- `recording_chunks`는 replay/history 참고 metadata로만 사용한다.
- `streaming_statuses`는 codec, 해상도, FPS 같은 optional metadata로만 유지한다.
- Live 관제 UI의 robot dropdown/chip 상태는 `live-status`를 우선 사용한다.

## 상태 판단 기준

| 상태 | 기준 |
| --- | --- |
| Connection | robot heartbeat 또는 `robots.status` |
| Stream | SFU observed publisher의 `lastTrackAt` 또는 `lastDataAt` freshness |
| Recording | recorder-worker `/healthz` robot runtime의 fresh track/data 수신 |
| Latest Chunk | `recording_chunks` 최신 row. replay/history 참고용 |

중요한 예외:

- `recording_chunks.status = recording`이어도 recorder runtime이 없으면 live 상태는 `recording`이 아니다.
- recorder-worker health 조회가 실패하거나 robot별 runtime이 없으면 `녹화 중`으로 표시하지 않는다.
- 종료된 mission이나 publisher가 없는 mission에서는 stale chunk row 때문에 `송출 중/녹화 중`이 뜨지 않아야 한다.

## 검증 결과

명령 검증:

```bash
cd apps/server && go test ./...
cd apps/server && go vet ./...
cd apps/web && npm test -- --run
cd apps/web && npm run build
```

런타임 검증:

```bash
HOST_IP=172.30.1.98 ./scripts/start.sh
APP_SERVER_URL=http://172.30.1.98:18080 MOCK_ROBOT_COUNT=2 MOCK_SHARED_MISSION=1 ./scripts/mock-robots-python.sh
./scripts/status.sh
```

확인한 상태:

- `mission-007`: `robot-001`, `robot-002` 모두 `stream.streaming`, `recording.recording`
- `mission-008`: `robot-003`, `robot-004` 모두 `stream.waiting`, `recording.idle`
- `mission-008`은 최신 chunk가 `uploaded`여도 Live UI에서 `녹화 중`으로 표시되지 않음
- Browser 관제 화면에서 `mission-007`은 `송출 중/녹화 중`, `mission-008`은 `송출 대기/녹화 대기`로 표시됨

## 남은 한계

- recorder runtime은 1차 구현에서 app-server가 recorder-worker `/healthz`를 짧은 timeout으로 조회한다.
- recorder runtime을 DB나 app-server memory report로 흡수하는 구조는 다음 단계에서 검토한다.
- `streaming_statuses` API는 제거하지 않고 optional media metadata 용도로 유지한다.
