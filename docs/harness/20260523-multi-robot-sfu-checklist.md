---
title: "multi-robot-sfu-checklist"
created: 2026-05-23
updated: 2026-05-23
author: "dhkimxx <dhkimxx@naver.com>"
editors: ["dhkimxx <dhkimxx@naver.com>"]
type: "checklist"
tags: ["harness", "webrtc", "sfu", "multi-robot", "regression"]
history:
- "2026-05-23 dhkimxx <dhkimxx@naver.com>: initial mission-level multi-robot SFU validation checklist"
---
# Multi-Robot SFU Checklist

이 문서는 mission 단위 SFU room 회귀 검증 기준만 정리한다.

상세 REST JSON, signaling JSON, DataChannel payload schema, track naming, codec 정책은 이 문서에서 확정하지 않는다. 실제 수신값은 status/metadata로 확인한다.

## 목표 구조

```text
mission-001 room
  publishers:
    robot-001
    robot-002
    robot-003

  subscribers:
    browser A
    browser B
    recorder-worker
```

핵심 불변 조건:

- room id는 mission 단위 `missionCode`다.
- 같은 mission room에 2대 이상의 robot publisher가 동시에 들어올 수 있어야 한다.
- robot별 구분은 room id가 아니라 `robotCode`로 유지한다.
- sensor/telemetry payload에는 `robotCode`가 유지되어야 한다.
- recorder-worker는 같은 mission room을 subscribe하되 chunk/file metadata는 robotCode별로 나뉘어야 한다.
- browser A/B와 recorder-worker는 서로 독립적인 subscriber다.

## 필수 수동/통합 시나리오

1. `robot-001`, `robot-002`를 등록하고 각각 별도 token을 확인한다.
2. `mission-001`에 `robot-001`, `robot-002`를 배정한다.
3. `mission-001`을 start 한다.
4. Python 또는 Android Mock Robot 2대를 각각 다른 `robotCode`와 token으로 실행한다.
5. 두 mock이 같은 `roomId=mission-001`로 publish 하는지 확인한다.
6. browser A와 browser B에서 같은 mission live 화면에 subscribe 한다.
7. recorder-worker가 같은 mission room에 recorder subscriber로 붙는지 확인한다.
8. 각 browser에서 robot별 RGB/Thermal/Audio 수신 상태를 확인한다.
9. 각 browser에서 sensor/telemetry가 robotCode별로 구분되는지 확인한다.
10. recorder chunk가 `robot-001`, `robot-002` 각각에 대해 생성되는지 확인한다.
11. MinIO object와 PostgreSQL metadata가 robotCode별 recording chunk/file을 가리키는지 확인한다.

현재 mock runner는 `MOCK_SHARED_MISSION=1` 기본값으로 같은 mission room publish smoke를 지원한다. Backend multi-robot assignment API가 준비되기 전에는 이 mode가 room-level WebRTC 검증을 보조한다. 최종 통합 pass 기준은 robot gateway mission 조회가 두 robotCode 모두에게 같은 active mission을 반환하는 상태다.

## 자동 검증 명령

서버 단위:

```bash
cd /Users/dhkim/workspace/sst/robot-center/apps/server
go test ./...
go vet ./...
```

Web build:

```bash
cd /Users/dhkim/workspace/sst/robot-center/apps/web
npm run build -- --outDir /tmp/robot-center-web-dist --emptyOutDir
```

개발 스택 기동:

```bash
cd /Users/dhkim/workspace/sst/robot-center
SKIP_ANDROID=1 ./scripts/dev-up.sh
./scripts/dev-status.sh
```

Python mock 2대 실행:

```bash
cd /Users/dhkim/workspace/sst/robot-center
MOCK_ROBOT_COUNT=2 MOCK_SHARED_MISSION=1 ./scripts/python-mock-robots-up.sh
```

특정 mission으로 고정해서 실행:

```bash
cd /Users/dhkim/workspace/sst/robot-center
MOCK_ROBOT_COUNT=2 MOCK_SHARED_MISSION=1 MOCK_MISSION_CODE=mission-001 ./scripts/python-mock-robots-up.sh
```

API 상태 확인:

```bash
curl -fsS http://127.0.0.1:18080/api/system/status | /usr/bin/python3 -m json.tool
curl -fsS http://127.0.0.1:18080/api/streaming-statuses | /usr/bin/python3 -m json.tool
curl -fsS http://127.0.0.1:18082/healthz | /usr/bin/python3 -m json.tool
```

PostgreSQL recording chunk 확인:

```bash
cd /Users/dhkim/workspace/sst/robot-center
docker compose -f deploy/docker-compose.yml exec postgres \
  psql -U robot_center -d robot_center \
  -c "select r.robot_code, count(*) as chunks from recording_chunks rc join robots r on r.id = rc.robot_id group by r.robot_code order by r.robot_code;"
```

MinIO object 확인:

```bash
cd /Users/dhkim/workspace/sst/robot-center
docker compose -f deploy/docker-compose.yml exec minio \
  mc ls local/robot-center-poc/missions --recursive | tail -80
```

## 통과 기준

- `go test ./...`, `go vet ./...`, web build가 통과한다.
- app-server와 recorder-worker health가 OK다.
- `mission-001` room에 robot publisher 2대 이상, browser subscriber 2대, recorder subscriber 1대가 관찰된다.
- streaming status 또는 SFU room summary에서 robotCode별 publisher 상태를 구분할 수 있다.
- browser A/B에서 같은 mission의 robot 2대 이상을 볼 수 있다.
- recorder-worker가 robotCode별 chunk/file metadata를 생성한다.
- MinIO object path 또는 metadata에서 mission과 robotCode 구분을 확인할 수 있다.

## 주요 회귀 지점

- room id가 다시 `missionCode__robotCode`로 분기되어 browser/recorder가 robot별 별도 room만 보게 되는 문제
- robotCode가 DataChannel payload 또는 recording metadata에서 사라지는 문제
- browser A 연결 실패가 browser B 또는 recorder-worker 연결을 끊는 문제
- recorder-worker 재시작이 mission live subscriber 전체를 끊는 문제
- robot-001 chunk와 robot-002 chunk가 같은 recording row나 object key로 덮이는 문제
