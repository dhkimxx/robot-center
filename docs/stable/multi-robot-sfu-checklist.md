---
title: "multi-robot-sfu-checklist"
created: 2026-05-23
updated: '2026-05-26'
author: "dhkimxx <dhkimxx@naver.com>"
editors: ["dhkimxx <dhkimxx@naver.com>", danya.kim <danya.kim@thundersoft.com>]
type: "checklist"
status: "stable"
tags: ["webrtc", "sfu", "multi-robot", "regression"]
history:
- "2026-05-23 dhkimxx <dhkimxx@naver.com>: initial mission-level multi-robot SFU validation checklist"
- "2026-05-23 dhkimxx <dhkimxx@naver.com>: added internal PoC selective subscribe QA checklist"
- "2026-05-25 Codex: aligned checklist with RobotStreamBundle slot roles"
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: moved into docs/stable lifecycle structure"
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: flattened from harness directory into stable docs"
- '2026-05-26 danya.kim <danya.kim@thundersoft.com>: flattened multi-robot SFU checklist from harness directory into stable docs'
---
# Multi-Robot SFU Selective Subscribe Checklist

이 문서는 mission 단위 SFU room과 내부 PoC selective subscribe 회귀 검증 기준만 정리한다.

상세 REST JSON, signaling JSON, DataChannel payload schema, track naming, codec 정책은 이 문서에서 확정하지 않는다. 실제 수신값은 status/metadata로 확인한다.

Selective subscribe signaling은 내부 PoC 임시 동작이다. 이 체크리스트는 런타임 검증 기준이며, 장기 API/schema 계약을 정의하지 않는다.

## 목표 구조

```text
mission-001 room
  publishers:
    robot-001
    robot-002
    robot-003

  subscribers:
    browser A selected robot-001
    browser B selected robot-002
    recorder-worker
```

핵심 불변 조건:

- room id는 mission 단위 `missionCode`다.
- 같은 mission room에 2대 이상의 robot publisher가 동시에 들어올 수 있어야 한다.
- robot별 구분은 room id가 아니라 `robotCode`로 유지한다.
- 각 robot은 subscriber 수와 무관하게 mission room에 한 번만 publish 한다.
- browser operator는 선택한 robotCode 하나만 subscribe 한다.
- 서로 다른 browser operator는 같은 mission room에서 서로 다른 robotCode를 선택할 수 있다.
- recorder-worker는 같은 mission room의 모든 robotCode를 subscribe 한다.
- sensor/telemetry payload에는 `robotCode`가 유지되어야 한다.
- recorder-worker는 같은 mission room을 subscribe하되 chunk/file metadata는 robotCode별로 나뉘어야 한다.
- browser A/B와 recorder-worker는 서로 독립적인 subscriber다.

## RobotStreamBundle 내부 슬롯

로봇 1대는 mission room 안에 1개의 `RobotStreamBundle`로 등록된다.

```text
RobotStreamBundle
  missionCode
  robotCode
  tracks:
    track.video_1
    track.video_2
    track.audio_1
    track.audio_2
  dataChannels:
    channel.telemetry
    channel.spatial
    channel.event
    channel.control
```

슬롯명은 의미를 고정하지 않는다. `track.video_1 = RGB`, `track.video_2 = Thermal` 같은 해석은 mock/display metadata에서만 다룬다.

현재 PoC 호환을 위해 `rgb`, `thermal`, `audio`, `sensor`, `telemetry` 같은 legacy label은 fallback alias로만 허용한다. 신규 mock/robot은 canonical slot label을 우선 사용한다.

Subscriber policy:

- operator는 selected robot의 bundle만 받는다.
- recorder는 mission room의 모든 robot bundle을 받는다.
- `select-robot`이 publish 중인 robot bundle을 찾지 못하면 성공 ACK가 아니라 error signal을 받는다.
- `channel.spatial`과 `channel.control`은 telemetry/event와 섞지 않는다.

## 필수 수동/통합 시나리오

1. `robot-001`, `robot-002`를 등록하고 각각 별도 token을 확인한다.
2. `mission-001`에 `robot-001`, `robot-002`를 배정한다.
3. `mission-001`을 start 한다.
4. Python 또는 Android Mock Robot 2대를 각각 다른 `robotCode`와 token으로 실행한다.
5. 두 mock이 같은 `roomId=mission-001`로 publish 하는지 확인한다.
6. browser A와 browser B에서 같은 mission live 화면에 operator subscriber로 join 한다.
7. browser A는 `robot-001`, browser B는 `robot-002`를 선택한다.
8. browser A가 `robot-001` RGB/Thermal/Audio와 sensor/telemetry만 수신하는지 확인한다.
9. browser B가 `robot-002` RGB/Thermal/Audio와 sensor/telemetry만 수신하는지 확인한다.
10. browser A 선택을 `robot-002`로 바꾸면 A 수신 대상만 바뀌고 B/recorder가 끊기지 않는지 확인한다.
11. recorder-worker가 같은 mission room에 recorder subscriber로 붙고 모든 robotCode를 수신하는지 확인한다.
12. recorder chunk가 `robot-001`, `robot-002` 각각에 대해 생성되는지 확인한다.
13. MinIO object와 PostgreSQL metadata가 robotCode별 recording chunk/file을 가리키는지 확인한다.
14. SFU 상태에서 robot publisher 수가 browser/recorder subscriber 수에 따라 증가하지 않는지 확인한다.

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
./scripts/dev-up.sh
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

Browser runtime 확인:

```text
1. browser A에서 mission live 화면을 열고 robot-001을 선택한다.
2. browser B에서 같은 mission live 화면을 열고 robot-002를 선택한다.
3. browser A video element와 telemetry panel이 robot-001 값만 표시하는지 확인한다.
4. browser B video element와 telemetry panel이 robot-002 값만 표시하는지 확인한다.
5. browser A 선택을 robot-002로 전환한 뒤 browser B와 recorder-worker가 계속 연결 상태인지 확인한다.
6. browser A를 닫아도 browser B와 recorder-worker가 계속 수신하는지 확인한다.
```

SFU 상태 확인:

```text
1. mission room id가 missionCode인지 확인한다.
2. publisher가 robotCode별 1개인지 확인한다.
3. browser subscriber별 selected robotCode가 기대값인지 확인한다.
4. recorder subscriber가 all robotCode subscription인지 확인한다.
5. publish track 수가 subscriber 수 증가로 중복 증가하지 않는지 확인한다.
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
- `mission-001` room에 robot publisher 2대 이상, browser operator subscriber 2대, recorder subscriber 1대가 관찰된다.
- streaming status 또는 SFU room summary에서 robotCode별 publisher 상태를 구분할 수 있다.
- browser A/B에서 같은 mission에 join 하되 각자 선택한 robotCode만 볼 수 있다.
- browser A/B의 robot 선택 변경이 다른 browser와 recorder subscriber를 끊지 않는다.
- recorder-worker는 같은 mission의 모든 robotCode를 수신한다.
- robot publisher는 subscriber 수와 무관하게 robotCode별 1개로 유지된다.
- recorder-worker가 robotCode별 chunk/file metadata를 생성한다.
- MinIO object path 또는 metadata에서 mission과 robotCode 구분을 확인할 수 있다.

## 주요 회귀 지점

- room id가 다시 `missionCode__robotCode`로 분기되어 browser/recorder가 robot별 별도 room만 보게 되는 문제
- browser operator가 선택하지 않은 robotCode media/data까지 수신하는 문제
- browser operator 선택 변경이 다른 subscriber 또는 publisher를 끊는 문제
- recorder-worker가 selected robotCode만 수신하고 전체 robotCode 저장을 놓치는 문제
- subscriber 수만큼 Robot publish peer connection이 중복 생성되는 문제
- robotCode가 DataChannel payload 또는 recording metadata에서 사라지는 문제
- browser A 연결 실패가 browser B 또는 recorder-worker 연결을 끊는 문제
- recorder-worker 재시작이 mission live subscriber 전체를 끊는 문제
- robot-001 chunk와 robot-002 chunk가 같은 recording row나 object key로 덮이는 문제
