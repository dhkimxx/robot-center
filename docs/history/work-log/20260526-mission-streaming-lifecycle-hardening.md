---
title: "mission-streaming-lifecycle-hardening"
created: 2026-05-26
updated: '2026-05-26'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "log"
status: "verified"
tags: ["mission", "streaming", "webrtc", "sfu", "robot"]
history:
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: mission scoped streaming lifecycle hardening 기록"
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: moved into docs/history lifecycle structure"
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: renamed log file to YYYYMMDD-title format"
- '2026-05-26 danya.kim <danya.kim@thundersoft.com>: standardized work log path and filename'
- '2026-05-26 danya.kim <danya.kim@thundersoft.com>: script entrypoint names updated after scripts cleanup'
---

# Mission Streaming Lifecycle Hardening

## 구현 범위

- Robot streaming status 수신 시 `missionId`가 active mission assignment인지 확인한다.
- `streaming`/`publishing` 보고의 `roomId`는 해당 mission의 `missionCode`와 일치해야 한다.
- streaming freshness 판단은 robot `sentAt`이 아니라 server `updatedAt` 기준 30초 window를 사용한다.
- `/api/streaming-statuses` 응답에 `updatedAt`을 포함한다.
- UI의 송출 표시와 임무 생성 차단도 `updatedAt`과 mission room 일치를 기준으로 판단한다.
- 문서에 `missionId = DB UUID`, `roomId = missionCode`, 종료성 status stale 방어 규칙을 반영했다.

## 검증 명령

```bash
cd /Users/dhkim/workspace/sst/robot-center/apps/server
go test ./...

cd /Users/dhkim/workspace/sst/robot-center/apps/web
npm test

cd /Users/dhkim/workspace/sst/robot-center
./scripts/start.sh
./scripts/mock-robots-python.sh
curl -fsS http://127.0.0.1:18080/api/system/status
curl -fsS http://127.0.0.1:18080/api/streaming-statuses
```

## 실제 확인 결과

- Go server tests: `go test ./...` 통과
- Web tests: `npm test` 통과, 9 files / 29 tests
- Web build/start script: `./scripts/start.sh` 통과
- Python Mock Robot: robot-001, robot-002가 `mission-006` / `roomId=mission-006`에 publish
- SFU room: `mission-006`, robot publisher 2, recorder subscriber 1
- Recorder worker: `mission-006` subscriber `iceState=connected`, `trackCount=6`, `dataChannelCount=4`
- Streaming status: robot-001, robot-002 모두 `status=streaming`, `roomId=mission-006`, `updatedAt` 포함
- Room mismatch guard: `roomId=mission-005__robot-001` streaming report가 HTTP 409 `invalid state`로 거부됨
- Mission end check: `mission-005` 종료 시 SFU room 목록에서 제거되고 streaming status가 `stopped`로 전환됨
- Sensor latest: `mission-006`에서 telemetry/spatial latest sample 조회 확인

## 사용자 확인 URL

- UI: http://192.168.20.32:18080
- Local API: http://127.0.0.1:18080

## Python Mock Robot 상태

- robot-001, robot-002 실행 중.
- 현재 active mission은 `mission-006`.
- 두 robot 모두 mission 조회/override 응답의 `missionId`와 `roomId=mission-006`을 사용한다.

## 남은 한계

- 과거 작업 로그 삭제분 복구 여부는 별도 정책 결정이 필요하다.
- Browser MCP가 기존 browser lock으로 열리지 않아 브라우저 스크린샷 검증은 수행하지 못했다. API/Mock Robot/SFU 상태는 확인했다.
