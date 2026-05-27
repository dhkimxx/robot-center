---
title: "robot-team-webrtc-send-test-guide"
created: 2026-05-27
updated: '2026-05-27'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "guide"
status: "planned"
tags: ["robot", "webrtc", "sfu", "sensor", "integration", "guide"]
history:
- "2026-05-27 danya.kim <danya.kim@thundersoft.com>: split robot team WebRTC send guide from dev-server deployment runbook"
- "2026-05-27 danya.kim <danya.kim@thundersoft.com>: made guide self-contained for robot team sharing"
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: sync robot team WebRTC guide with current gateway and sensor API contract'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: clarify recorder storage boundaries for event and control data channels'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: align robot team WebRTC guide with canonical REST token contract and no publisher token'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: remove publisher token from robot team P0 guide'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: update robot team guide to role-specific SFU endpoint'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: add deployed dev-server connection, robot tokens, and TURN credentials for robot team test'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: add live dev-server connection details and TURN credentials'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: clarify public-only robot team address contract'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: record verified dev-server WebRTC send test results and clarify RTC checks'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: expand documented TURN relay range for dev-server testing'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: refresh verification timestamp after expanded TURN range deployment'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: keep robot team sharing guide on canonical schema only'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: normalize canonical guide history wording'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: add robot team signaling Q&A'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: make robot team test slots self-service'
---

# Robot Team WebRTC Send Test Guide

## 1. 목적

이 문서는 로봇팀이 실제 Robot Gateway/Publisher에서 관제 서버로 WebRTC 영상과 센서 데이터를 송신할 때 필요한 연동 절차를 정의한다.

현재 관제 시스템은 임무 단위 WebRTC room을 만들고, 로봇은 해당 room에 publisher로 접속한다. 관제 UI와 recorder는 같은 room에 subscriber로 접속해 로봇의 영상, 오디오, telemetry, sensor 데이터를 수신한다.

이 문서 하나만 보고 테스트를 진행할 수 있도록 서버 주소, REST API, WebRTC signaling, media track, DataChannel, 통과 기준을 함께 정리한다.

테스트에서 확인하려는 것:

- 로봇이 관제 서버에 heartbeat를 보낼 수 있는가
- 로봇이 자신에게 배정된 active mission을 조회할 수 있는가
- 로봇이 mission room에 WebRTC publisher로 접속할 수 있는가
- 관제 서버가 로봇의 RGB/Thermal/Audio track을 수신할 수 있는가
- 관제 서버가 telemetry/spatial DataChannel 메시지를 수신하고 저장할 수 있는가
- 관제 UI에서 영상, 위치, 센서값, 녹화 상태를 확인할 수 있는가

이 테스트에서 제외하는 것:

- 로봇 장치 내부 camera/sensor/ROS/GStreamer 구현 방식
- 제어 명령 송신과 control ACK 정책
- HTTPS/WSS 운영화
- 장기 운영 인증/권한 정책

## 2. 테스트 서버 접속 정보

현재 임시 개발서버는 배포 완료 상태다.

```text
serverUrl: http://192.168.20.12:18080
operatorUi: http://192.168.20.12:18080
recorderHealthUrl: http://192.168.20.12:18082/healthz
missionCode: 각 테스트자가 직접 생성
missionStatus: active로 시작 후 테스트
```

현재 관제팀 재현 결과:

```text
verifiedAt: 2026-05-27 18:01 KST
verifiedWith: Android Robot app 2 devices
robot-001: heartbeat OK, mission OK, WebSocket joined, relay ICE CONNECTED/COMPLETED
robot-002: heartbeat OK, mission OK, WebSocket joined, relay ICE CONNECTED/COMPLETED
app-server SFU: mission-001 room, robotCount 2, recorderCount 1
recorder-worker: iceState connected, trackCount 6, dataChannelCount 4, appendFailedCount 0
recording files: rgb.h264, thermal.h264, audio.ogg, telemetry.jsonl created per robot
```

`robot-001`, `robot-002`, `mission-001`은 관제팀이 Android Robot app 2대로 검증한 baseline slot이다. 여러 명이 동시에 테스트할 때는 이 공용 slot을 재사용하지 않는다.

로봇팀 개별 테스트 원칙:

- 테스트 담당자 또는 장비마다 새 robot을 만든다.
- WebRTC publisher 인스턴스 1개당 robot 1개와 `robotToken` 1개를 사용한다.
- 테스트 시나리오마다 새 mission을 만들고 active 상태로 시작한다.
- 같은 robot을 여러 active mission에 동시에 배정하지 않는다.
- 생성한 `robotCode`, `robotToken`, `missionCode`는 각자 테스트 로그에 남긴다.

`robotToken`은 Bearer token이다. 이 값은 임시 개발서버 테스트용이며, 테스트 종료 또는 재배포 시 교체될 수 있다. 현재 임시 개발서버의 robot/mission 생성 API는 로봇팀 병렬 테스트 편의를 위해 열려 있다. 운영 환경에서는 별도 발급/권한 정책으로 바뀔 수 있다.

TURN 서버는 다음 값으로 기동 중이다.

```json
{
  "iceTransportPolicy": "relay",
  "iceServers": [
    {
      "urls": ["turn:192.168.20.12:3478?transport=udp"],
      "username": "robot-center-turn",
      "credential": "rc-turn-2026-0527",
      "credentialType": "password"
    }
  ]
}
```

로봇 구현에서는 TURN 값을 하드코딩하지 말고 `GET /api/robot-gateway/mission` 응답의 `turnServers`를 그대로 `RTCPeerConnection.iceServers`에 넣는다. 위 값은 네트워크 디버깅과 수동 테스트를 위한 현재 서버 값이다.

이 문서의 주소는 모두 로봇팀 단말에서 접근하는 public address 기준이다. Docker 내부에서 쓰는 `app-server`, `turn` 같은 service DNS는 관제 서버 내부 구현값이며 로봇팀 구현에 사용하지 않는다.

방화벽/네트워크 확인 대상:

| Purpose | Address |
| --- | --- |
| REST API / Web UI / WebSocket signaling | `192.168.20.12:18080/tcp` |
| TURN allocation | `192.168.20.12:3478/udp`, `192.168.20.12:3478/tcp` |
| TURN relay port range | `192.168.20.12:49160-49300/udp`, `192.168.20.12:49160-49300/tcp` |
| Recorder health check | `192.168.20.12:18082/tcp` |

### 2.1 개인 테스트용 robot과 mission 생성

아래 명령은 테스트 담당자가 자기 robot과 mission을 새로 만드는 절차다. `TEST_OWNER`는 충돌을 피할 수 있게 본인 이름, 장비명, 날짜 등을 넣는다.

```bash
SERVER_URL='http://192.168.20.12:18080'
TEST_OWNER='robot-team-your-name-jetson-01'
```

Robot 생성:

```bash
ROBOT_RESPONSE="$(
  curl -fsS -X POST "$SERVER_URL/api/robots" \
    -H 'Content-Type: application/json' \
    -d '{
      "displayName": "'"$TEST_OWNER"'",
      "modelName": "Jetson WebRTC Publisher"
    }'
)"

ROBOT_CODE="$(
  printf '%s' "$ROBOT_RESPONSE" \
    | python3 -c 'import json,sys; print(json.load(sys.stdin)["robot"]["robotCode"])'
)"

ROBOT_TOKEN="$(
  printf '%s' "$ROBOT_RESPONSE" \
    | python3 -c 'import json,sys; print(json.load(sys.stdin)["connectionInfo"]["robotToken"])'
)"

printf 'ROBOT_CODE=%s\nROBOT_TOKEN=%s\n' "$ROBOT_CODE" "$ROBOT_TOKEN"
```

Robot 생성 응답에서 `connectionInfo.robotToken`이 내려온다. 이 token이 이후 heartbeat, mission 조회, WebSocket signaling에 모두 쓰인다.

Mission 생성:

```bash
MISSION_RESPONSE="$(
  curl -fsS -X POST "$SERVER_URL/api/missions" \
    -H 'Content-Type: application/json' \
    -d '{
      "name": "'"$TEST_OWNER"' WebRTC send test",
      "missionType": "mountain_rescue",
      "siteNote": "Robot team self-service WebRTC send test",
      "robotCode": "'"$ROBOT_CODE"'",
      "robotCodes": ["'"$ROBOT_CODE"'"]
    }'
)"

MISSION_CODE="$(
  printf '%s' "$MISSION_RESPONSE" \
    | python3 -c 'import json,sys; print(json.load(sys.stdin)["mission"]["missionCode"])'
)"

printf 'MISSION_CODE=%s\n' "$MISSION_CODE"
```

Mission 시작:

```bash
curl -fsS -X POST "$SERVER_URL/api/missions/$MISSION_CODE/start" \
  | python3 -m json.tool
```

테스트 중 사용할 값:

```bash
printf 'SERVER_URL=%s\nROBOT_CODE=%s\nROBOT_TOKEN=%s\nMISSION_CODE=%s\n' \
  "$SERVER_URL" "$ROBOT_CODE" "$ROBOT_TOKEN" "$MISSION_CODE"
```

생성 확인:

```bash
curl -fsS "$SERVER_URL/api/robot-gateway/mission" \
  -H "Authorization: Bearer $ROBOT_TOKEN" \
  | python3 -m json.tool
```

정상이라면 `missionStatus`는 `active`, `missionCode`는 방금 생성한 값, `sfu.signalingUrl`은 `ws://192.168.20.12:18080/sfu/robot/ws?room={MISSION_CODE}` 형태다.

### 2.2 빠른 접속 확인

아래 예시는 2.1에서 생성한 `ROBOT_TOKEN`을 그대로 사용한다.

```bash
SERVER_URL='http://192.168.20.12:18080'
ROBOT_TOKEN='<2.1에서 생성한 robotToken>'
```

Heartbeat:

```bash
curl -fsS -X POST "$SERVER_URL/api/robot-gateway/heartbeat" \
  -H "Authorization: Bearer $ROBOT_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "state": "online",
    "batteryPercent": 82,
    "networkQuality": "good"
  }' | python3 -m json.tool
```

Active mission 조회:

```bash
curl -fsS "$SERVER_URL/api/robot-gateway/mission" \
  -H "Authorization: Bearer $ROBOT_TOKEN" \
  | python3 -m json.tool
```

현재 정상 응답이면 `missionStatus`는 `active`, `missionCode`는 2.1에서 생성한 값, `sfu.signalingUrl`은 `ws://192.168.20.12:18080/sfu/robot/ws?room={missionCode}`로 내려온다.

RTC 설정 확인:

```bash
curl -fsS "$SERVER_URL/api/rtc-config" | python3 -m json.tool
```

현재 정상 응답이면 `signalingUrl`은 `ws://192.168.20.12:18080/sfu/operator/ws`, `turnServers[0].urls[0]`은 `turn:192.168.20.12:3478?transport=udp`, `iceTransportPolicy`는 `relay`다.

## 3. 전체 연결 구조

```text
Robot Gateway/Publisher
  -> REST heartbeat
  -> REST active mission lookup
  -> WebRTC signaling WebSocket
  -> app-server internal SFU
  -> Browser 관제 UI / recorder-worker
```

역할:

| Component | Role |
| --- | --- |
| Robot Gateway/Publisher | media track과 DataChannel을 publish |
| app-server | REST API, robot token 검증, mission 관리, WebRTC signaling, SFU |
| Browser 관제 UI | operator subscriber, 선택한 robotCode의 live stream 표시 |
| recorder-worker | recorder subscriber, 모든 robotCode의 media/data 저장 |
| TURN | relay-only ICE 경로 |

WebRTC room 규칙:

- room id는 `missionCode`다.
- `roomId`는 `missionCode`와 같아야 한다.
- `robotCode`는 room id에 합치지 않는다.
- 같은 mission room에 여러 robot publisher가 동시에 들어올 수 있다.
- robot identity는 REST token으로 mission 조회 시 검증하고, SFU publish 단계에서는 active mission assignment를 서버가 다시 확인한다.

## 4. 로봇팀 구현 책임

로봇팀 Robot Gateway/Publisher는 다음 순서로 동작한다.

```text
1. heartbeat 송신
2. active mission 조회
3. mission 응답의 signalingUrl / turnServers 사용
4. mission room에 WebRTC publisher로 join
5. offer 송신, answer 수신
6. relay ICE candidate 사용
7. media track publish
8. DataChannel publish
```

관제팀이 제공하는 Android/Python Mock Robot은 테스트용 sample client다. 실제 로봇 구현은 로봇팀 코드베이스에서 담당한다.

## 5. REST API Reference

Base URL:

```text
http://192.168.20.12:18080
```

Security:

| Scheme | Header | Format |
| --- | --- | --- |
| Bearer token | `Authorization` | `Bearer {robotToken}` |

Common response content type:

```text
application/json
```

Common error response body:

```json
{
  "error": "error message"
}
```

### 5.1 `sendRobotHeartbeat`

로봇이 관제 서버에 자신이 살아 있고 연결 가능한 상태임을 알리는 API다.

```yaml
method: POST
path: /api/robot-gateway/heartbeat
auth: Bearer token required
contentType: application/json
```

Request headers:

| Name | Required | Value |
| --- | --- | --- |
| `Authorization` | Yes | `Bearer {robotToken}` |
| `Content-Type` | Yes | `application/json` |

Request body schema:

| Field | Type | Required | Allowed values / Example | Description |
| --- | --- | --- | --- | --- |
| `state` | string | No | `online`, `offline`, `fault`, other non-empty value | 생략 시 서버는 `online`으로 처리. `offline`, `fault` 외의 non-empty 값은 현재 서버에서 `online`으로 정규화 |
| `batteryPercent` | integer | No | `0`-`100` | 배터리 퍼센트. 현재 heartbeat 저장/판단에는 사용하지 않지만 요청 DTO에 포함 |
| `networkQuality` | string | No | `good`, `normal`, `poor`, `unknown` | 로봇 측 네트워크 상태. 현재 heartbeat 저장/판단에는 사용하지 않지만 요청 DTO에 포함 |
| `sentAt` | string(date-time) | No | `2026-05-27T05:00:00.000Z` | 로봇이 heartbeat를 보낸 시각. 현재 heartbeat 저장/판단에는 사용하지 않지만 요청 DTO에 포함 |

Request example:

```json
{
  "state": "online",
  "batteryPercent": 82,
  "networkQuality": "good",
  "sentAt": "2026-05-27T05:00:00.000Z"
}
```

`200 OK` response body schema:

| Field | Type | Values / Example | Description |
| --- | --- | --- | --- |
| `robotId` | string(uuid) | `8f0e4c69-8c9b-40ef-a3fe-7b8a7ad9a111` | 관제 서버 내부 robot id |
| `robotCode` | string | `robot-123` | heartbeat가 반영된 robot code |
| `status` | string | `online`, `offline`, `fault` | 서버가 저장한 장치 상태. `state` 정규화 결과 |
| `serverTime` | string(date-time) | `2026-05-27T05:00:00.120Z` | 서버 응답 시각 |

`200 OK` response example:

```json
{
  "robotId": "8f0e4c69-8c9b-40ef-a3fe-7b8a7ad9a111",
  "robotCode": "robot-123",
  "status": "online",
  "serverTime": "2026-05-27T05:00:00.120Z"
}
```

Error responses:

| Status | Meaning |
| --- | --- |
| `400` | JSON body 파싱 실패 또는 요청 값 오류 |
| `401` | Bearer token이 없거나 유효하지 않음 |
| `404` | token에 해당하는 robot이 등록되어 있지 않음 |

Client behavior:

- active mission 여부와 무관하게 주기적으로 heartbeat를 보낸다.
- 실패하면 2s, 5s, 10s 수준의 backoff로 재시도한다.
- `401` 또는 `404`는 token 발급 상태를 관제팀과 확인한다.

### 5.2 `getActiveMissionForRobot`

로봇이 현재 수행할 active mission과 WebRTC 연결 정보를 조회하는 API다.

```yaml
method: GET
path: /api/robot-gateway/mission
auth: Bearer token required
```

Request headers:

| Name | Required | Value |
| --- | --- | --- |
| `Authorization` | Yes | `Bearer {robotToken}` |

Request example:

```yaml
method: GET
url: http://192.168.20.12:18080/api/robot-gateway/mission
headers:
  Authorization: Bearer {robotToken}
```

`200 OK` response body when no active mission exists:

```json
{
  "missionId": null,
  "missionStatus": "none"
}
```

`200 OK` response body when active mission exists:

```json
{
  "missionId": "a8c2d4e1-25ef-4720-8d8c-2f4f5d0a1001",
  "missionCode": "mission-123",
  "missionStatus": "active",
  "robotCode": "robot-123",
  "roomId": "mission-123",
  "sfu": {
    "signalingUrl": "ws://192.168.20.12:18080/sfu/robot/ws?room=mission-123",
    "iceTransportPolicy": "relay"
  },
  "turnServers": [
    {
      "urls": ["turn:192.168.20.12:3478?transport=udp"],
      "username": "robot-center-turn",
      "credential": "rc-turn-2026-0527"
    }
  ],
  "tracks": ["track.video_1", "track.video_2", "track.audio_1", "track.audio_2"],
  "dataChannels": ["channel.telemetry", "channel.spatial", "channel.event", "channel.control"],
  "videoPolicy": {
    "mode": "robot_defined"
  }
}
```

Active mission response schema:

| Field | Type | Values / Example | Description |
| --- | --- | --- | --- |
| `missionId` | string(uuid) | `a8c2d4e1-25ef-4720-8d8c-2f4f5d0a1001` | 관제 서버 내부 mission id |
| `missionCode` | string | `mission-123` | 사람이 읽는 mission code이며 WebRTC room id |
| `missionStatus` | string | `active`, `none` | gateway mission 조회 응답에서는 `active`이면 WebRTC publish 가능. `none`은 active mission 없음. 내부 mission lifecycle 값은 `ready`, `active`, `ended` |
| `robotCode` | string | `robot-123` | token으로 인증된 robot code. 로깅과 WebRTC publisher identity에 사용 |
| `roomId` | string | `mission-123` | WebRTC room id. `missionCode`와 같아야 함 |
| `sfu.signalingUrl` | string | `ws://192.168.20.12:18080/sfu/robot/ws?room=mission-123` | Robot publisher WebSocket URL. client가 재구성하지 않고 그대로 사용 |
| `sfu.iceTransportPolicy` | string | `relay` | 현재 `relay`만 사용 |
| `turnServers[].urls` | string[] | `["turn:192.168.20.12:3478?transport=udp"]` | TURN URL 목록 |
| `turnServers[].username` | string | `robot-center-turn` | TURN username |
| `turnServers[].credential` | string | `rc-turn-2026-0527` | TURN password |
| `tracks` | string[] | `track.video_1`, `track.video_2`, `track.audio_1`, `track.audio_2` | canonical media track slot 목록 |
| `dataChannels` | string[] | `channel.telemetry`, `channel.spatial`, `channel.event`, `channel.control` | canonical DataChannel label 목록 |
| `videoPolicy.mode` | string | `robot_defined` | 해상도/FPS는 로봇 송신 설정을 따름 |

Error responses:

| Status | Meaning |
| --- | --- |
| `401` | Bearer token이 없거나 유효하지 않음 |
| `404` | token에 해당하는 robot이 등록되어 있지 않음 |

Client behavior:

- `missionStatus=none`이면 WebRTC publish를 시작하지 않는다.
- active mission이 없을 때도 heartbeat와 mission 조회를 계속 재시도한다.
- active mission이 오면 `sfu.signalingUrl`, `turnServers`를 그대로 사용한다.
- `roomId`는 `missionCode`와 같아야 한다.

## 6. WebRTC Signaling

Robot publisher는 mission 조회 응답의 `sfu.signalingUrl`로 WebSocket 연결한다.

WebSocket request:

```yaml
method: GET
url: "{sfu.signalingUrl}"
auth: Bearer token required
headers:
  Authorization: Bearer {robotToken}
```

위 `headers`는 WebSocket upgrade HTTP request header다. URL query나 JSON payload에 `Authorization`을 넣지 않는다.

예시는 다음과 같다.

```yaml
method: GET
url: ws://192.168.20.12:18080/sfu/robot/ws?room=mission-123
headers:
  Authorization: Bearer {robotToken}
```

Robot publisher는 mission 응답의 `sfu.signalingUrl`을 그대로 사용한다. URL query를 client가 다시 조립하지 않는다. `role`, `robotCode` query parameter는 붙이지 않는다. robot identity는 `Authorization` header의 `robotToken`으로 서버가 판단한다.

연결 흐름:

```text
Robot
-> WebSocket connect
-> joined 수신
-> offer 송신
-> answer 수신
-> ICE candidate 교환
-> media/data publish
```

P0는 SDP 내부 codec line을 고정 계약으로 보지 않는다. 우선 H.264 video와 Opus audio를 기대하지만, 실제 지원 codec과 SDP는 테스트 중 관측해 확정한다.

Signaling message 개요:

| Message | Direction | Meaning |
| --- | --- | --- |
| `joined` | server -> robot | WebSocket room join 완료 |
| `peer-present` / `peer-joined` | server -> robot | 같은 room peer 존재 알림 |
| `offer` | robot -> server | robot publisher SDP offer |
| `answer` | server -> robot | SFU SDP answer |
| `candidate` | both | ICE candidate |
| `publish-error` | server -> robot | active mission assignment 검증 실패 |

WebRTC signaling 인증은 REST와 같은 `robotToken` 하나만 사용한다. app-server는 robot token으로 robotCode를 확인하고 active mission assignment를 기준으로 publish를 검증한다.

offer payload는 최소 다음 값을 포함한다.

```json
{
  "type": "offer",
  "payload": {
    "type": "offer",
    "sdp": "..."
  }
}
```

ICE candidate payload는 browser/Pion 표준 candidate 필드를 사용한다.

```json
{
  "type": "candidate",
  "payload": {
    "candidate": "...",
    "sdpMid": "0",
    "sdpMLineIndex": 0
  }
}
```

## 7. Media Track

로봇팀 구현은 아래 canonical track slot만 사용한다.

| Slot | Kind | Expected value | Required |
| --- | --- | --- | --- |
| `track.video_1` | video | RGB 또는 주 영상. H.264 우선 테스트 | Yes |
| `track.video_2` | video | Thermal 또는 보조 영상. H.264 우선 테스트 | Recommended |
| `track.audio_1` | audio | Audio. Opus 우선 테스트 | Optional |
| `track.audio_2` | audio | Reserved secondary audio slot | Optional |

권장:

- `track.video_1`에는 주 RGB 카메라를 송신한다.
- `track.video_2`에는 thermal 또는 보조 영상을 송신한다.
- audio가 없으면 `track.audio_1`은 생략 가능하다.
- 영상 codec은 우선 H.264로 테스트한다.

## 8. DataChannel

로봇팀 구현은 아래 canonical DataChannel label만 사용한다.

| Label | Expected messageType | 용도 |
| --- | --- | --- |
| `channel.telemetry` | `telemetry` | GPS, battery, 환경값 같은 저속 상태. recorder-worker가 sensor API로 저장 |
| `channel.spatial` | `spatial` 또는 domain-specific type | IMU, odometry, point cloud descriptor. recorder-worker가 sensor API로 저장 |
| `channel.event` | `event` 또는 domain-specific type | alarm, fault, detection, mission event. 현재 recorder-worker 저장 대상은 아님 |
| `channel.control` | reserved | reserved control/ack side channel. 현재 recorder-worker 저장 대상은 아님 |

`channel.telemetry` 예시:

```json
{
  "messageId": "uuid",
  "messageType": "telemetry",
  "sequence": 102,
  "sentAt": "2026-05-27T05:00:00.000Z",
  "descriptors": [
    {
      "sensorId": "telemetry.position_1",
      "displayName": "GPS",
      "sensorType": "position",
      "valueType": "object",
      "enabled": true
    }
  ],
  "samples": [
    {
      "sensorId": "telemetry.position_1",
      "sequence": 102,
      "values": {
        "latitude": 37.402183,
        "longitude": 127.106812,
        "accuracyMeter": 3.5
      }
    }
  ]
}
```

`robotCode`, `missionId`, `missionCode`, `channelRole`은 로봇 payload에 넣지 않는다. 서버가 WebRTC publisher identity, room, DataChannel label에서 주입한다.

Telemetry envelope schema:

| Field | Type | Required | Values / Example | Description |
| --- | --- | --- | --- | --- |
| `messageId` | string | Recommended | UUID 또는 `robot-123-telemetry-102` | 메시지 추적 id |
| `messageType` | string | Recommended | `telemetry` | telemetry channel 기본 타입 |
| `sequence` | integer | Recommended | `102` | DataChannel message 증가값 |
| `sentAt` | string(date-time) | Recommended | `2026-05-27T05:00:00.000Z` | 로봇 송신 시각 |
| `descriptors` | array | No | SensorDescriptor list | 센서 metadata |
| `samples` | array | No | SensorSample list | 실제 측정값 |

SensorDescriptor schema:

| Field | Type | Required | Values / Example | Description |
| --- | --- | --- | --- | --- |
| `sensorId` | string | Yes | `telemetry.position_1`, `telemetry.battery_1`, `spatial.imu_1` | robot 내부에서 안정적으로 쓰는 sensor id |
| `displayName` | string | Recommended | `GPS`, `Battery`, `IMU` | UI 표시 이름 |
| `sensorType` | string | Recommended | `position`, `battery`, `environment`, `imu`, `odometry`, `point_cloud`, `gas`, `event` | 센서 계열 |
| `valueType` | string | Recommended | `number`, `boolean`, `string`, `vector`, `object`, `object_ref` | sample 값 형태 |
| `unit` | string | No | `percent`, `celsius`, `ppm`, `m`, `m/s` | 표시 단위 |
| `sampleRateHz` | number | No | `1`, `5`, `0.2` | canonical sampling rate |
| `enabled` | boolean | No | `true`, `false` | UI/저장 대상으로 활성화할지 여부 |
| `metadata` | object | No | `{ "frameId": "base_link" }` | frameId, axes 같은 부가 정보 |

SensorSample schema:

| Field | Type | Required | Values / Example | Description |
| --- | --- | --- | --- | --- |
| `sensorId` | string | Yes | `telemetry.position_1` | descriptor의 sensorId와 매칭 |
| `sequence` | integer | Recommended | `102` | sample sequence |
| `timestamp` | string(date-time) | No | `2026-05-27T05:00:00.000Z` | sample 측정 시각. sample `sentAt`보다 우선 |
| `sentAt` | string(date-time) | No | `2026-05-27T05:00:00.000Z` | sample 송신 시각. 없으면 envelope `sentAt` 사용 |
| `quality` | string | No | `good`, `normal`, `poor`, `unknown` | 로봇 측 품질값. 현재 typed sample field로는 저장하지 않음 |
| `values` | any | Recommended | `{ "latitude": 37.402183 }` | 실제 측정값. `valueType`에 맞는 JSON number/string/boolean/object/vector 형태 |
| `objectKey` | string | No | `missions/.../point_cloud.bin` | object storage 참조가 필요할 때 |

Position `values` 권장 필드:

| Field | Type | Required | Example | Description |
| --- | --- | --- | --- | --- |
| `latitude` | number | Yes | `37.402183` | WGS84 latitude |
| `longitude` | number | Yes | `127.106812` | WGS84 longitude |
| `altitudeMeter` | number | No | `42.5` | 고도 meter |
| `accuracyMeter` | number | No | `3.5` | 위치 정확도 meter |
| `headingDegree` | number | No | `90` | 진행 방향. 0-360 degree |

권장:

- `channel.telemetry`는 1Hz 수준의 저속 상태값부터 시작한다.
- GPS 위치는 `telemetry.position_1` 같은 안정적인 `sensorId`를 사용한다.
- `sequence`는 DataChannel message 순서를 추적할 수 있게 증가시킨다.
- `sentAt`은 로봇 송신 시각이다.
- descriptor는 매번 보내도 되고, sensor 구성이 바뀔 때 다시 보내도 된다.

## 9. 통과 기준

Robot gateway:

- heartbeat 성공
- mission 조회가 `missionStatus=active` 반환
- `roomId == missionCode`
- WebSocket signaling 연결 성공

SFU/WebRTC:

- `/api/system/status`의 `sfuRooms`에 mission room 표시
- 해당 room의 `robotCount`가 송신 로봇 수와 일치
- published tracks에 `robotCode:track.video_1` 표시
- Robot publisher ICE state가 `connected` 또는 `completed`
- `GET /api/missions/{missionCode}/live-status`에서 robot별 `stream.state=streaming`
- recorder-worker health에서 `iceState=connected`, 해당 robot의 track/data 수신 시각 확인
- recorder-worker health에서 `appendFailedCount=0`

Sensor/UI:

- `sensor-latest`에 robotCode별 sensor 목록 표시
- GPS/position sample이 있으면 관제 UI 위치 영역에 표시
- RGB/Thermal 영상이 live 화면에 표시
- recording 상태가 `recording` 또는 기대 상태로 표시
- recorder runtime 또는 object storage에 robot별 recording artifact가 생성됨

## 10. 장애 대응

| 증상 | 우선 확인 |
| --- | --- |
| heartbeat 401 | robotToken 누락/만료/불일치 |
| mission 조회가 `missionStatus=none` | active mission 없음, robot assignment 누락 |
| WebSocket 400 | mission 응답의 `sfu.signalingUrl`을 사용하지 않았거나 필수 query 누락 |
| `publish-error` | robot이 active mission에 배정되지 않았거나 room이 missionCode와 다름 |
| ICE `failed` | mission 응답의 `turnServers` 사용 여부, UDP 3478, relay port range, 방화벽, relay candidate 생성 여부 |
| room은 보이나 영상 없음 | track publish, codec negotiation, track label/order |
| 영상은 보이나 센서 없음 | DataChannel label, open 상태, payload envelope |
| 센서는 오나 위치 없음 | position sensorId/value shape |
| recorder가 idle | recorder-worker join, active recording target, ICE state |
| recorder는 connected인데 저장 실패 | recorder health의 `appendFailedCount`, `lastAppendError`, recording artifact 생성 여부 |

## 11. 로봇팀이 관제팀에 공유할 로그

문제가 발생하면 다음 로그를 함께 공유한다.

- heartbeat request/response status
- mission lookup request/response body
- WebSocket open/close code와 close reason
- 송신한 offer SDP의 media section 요약
- 수신한 answer SDP의 media section 요약
- ICE gathering/connection state 변화
- TURN candidate 생성 여부
- publish한 track id, stream id, codec, resolution, fps
- 열린 DataChannel label 목록
- 마지막으로 송신한 DataChannel payload 예시

## 12. 테스트 결과 기록

관제팀과 로봇팀은 테스트 후 다음 정보를 남긴다.

- 테스트 일시
- robotCode
- missionCode, missionId
- 로봇팀 publisher 버전
- heartbeat 결과
- signaling 연결 결과
- ICE state
- track 수와 DataChannel 수
- RGB/Thermal 표시 여부
- sensor-latest 저장 여부
- recorder chunk 생성 여부
- 실패 로그와 재현 절차

## 13. 질문과 답변

### Q1. endpoint 주소는 무엇인가?

로봇팀 테스트에서 직접 호출하는 endpoint는 아래와 같다.

| Purpose | Protocol | Endpoint |
| --- | --- | --- |
| Robot 생성 | HTTP | `POST http://192.168.20.12:18080/api/robots` |
| Mission 생성 | HTTP | `POST http://192.168.20.12:18080/api/missions` |
| Mission 시작 | HTTP | `POST http://192.168.20.12:18080/api/missions/{missionCode}/start` |
| Heartbeat | HTTP | `POST http://192.168.20.12:18080/api/robot-gateway/heartbeat` |
| Active mission 조회 | HTTP | `GET http://192.168.20.12:18080/api/robot-gateway/mission` |
| Robot WebRTC signaling | WebSocket | mission 조회 응답의 `sfu.signalingUrl` |

`missionCode=mission-123` 예시의 signaling URL은 `ws://192.168.20.12:18080/sfu/robot/ws?room=mission-123` 형태다. 로봇 구현은 이 값을 직접 조립하지 말고 mission 조회 응답의 `sfu.signalingUrl`을 그대로 사용한다.

### Q2. WebSocket인가, HTTP POST인가?

둘 다 사용하지만 역할이 다르다.

| Step | Transport |
| --- | --- |
| robot 생성 | HTTP POST |
| mission 생성 | HTTP POST |
| mission 시작 | HTTP POST |
| heartbeat | HTTP POST |
| active mission 조회 | HTTP GET |
| WebRTC offer/answer/candidate signaling | WebSocket |
| 영상/오디오 | WebRTC media track |
| 센서/telemetry | WebRTC DataChannel |

영상과 센서 데이터를 HTTP POST로 계속 업로드하는 구조가 아니다. HTTP는 테스트 slot 생성, 로봇 상태, mission 조회까지만 사용하고, 송신 본류는 WebRTC 연결 이후 media track과 DataChannel로 보낸다.

### Q3. Jetson이 offer를 보내는 방식인가?

Yes. Jetson 쪽 Robot Gateway/Publisher가 WebSocket에 접속한 뒤 SDP offer를 서버로 보낸다.

서버는 `joined`와 `peer-present`/`peer-joined` 메시지를 보낸다. Robot publisher는 SFU peer가 확인되면 `offer` 메시지를 보낸다.

```json
{
  "type": "offer",
  "payload": {
    "targetPeerId": "sfu",
    "type": "offer",
    "sdp": "v=0..."
  }
}
```

`targetPeerId`는 생략 가능하지만, 명시할 경우 `sfu`를 사용한다.

### Q4. answer JSON 형식은 무엇인가?

서버가 Robot publisher에게 보내는 answer는 아래 형식이다.

```json
{
  "type": "answer",
  "payload": {
    "room": "mission-123",
    "fromRole": "sfu",
    "fromPeerId": "sfu",
    "targetPeerId": "peer_xxxxxxxxxxxxxxxx",
    "type": "answer",
    "sdp": "v=0...",
    "robotCode": "robot-123"
  }
}
```

Robot publisher는 `payload.type`과 `payload.sdp`로 remote description을 설정한다. `room`, `fromRole`, `fromPeerId`, `targetPeerId`, `robotCode`는 라우팅/로깅용 metadata다.

### Q5. ICE candidate는 어떻게 교환하는가?

현재 테스트 서버는 `iceTransportPolicy=relay` 기준이다. Robot publisher는 mission 응답의 `turnServers`를 `RTCPeerConnection`에 넣고 TURN relay candidate를 사용한다.

권장 방식은 ICE gathering 완료 후 relay candidate가 포함된 SDP offer를 보내는 것이다. 이 경우 서버 answer SDP에도 ICE 정보가 포함되므로 별도 candidate 메시지 없이 연결될 수 있다.

trickle ICE를 구현한 경우에는 아래 WebSocket 메시지를 보낼 수 있다.

```json
{
  "type": "candidate",
  "payload": {
    "targetPeerId": "sfu",
    "candidate": "candidate:...",
    "sdpMid": "0",
    "sdpMLineIndex": 0
  }
}
```

현재 P0 테스트에서는 로봇 쪽에서 relay candidate가 생성됐는지, 최종 ICE state가 `connected` 또는 `completed`인지가 핵심이다.

### Q6. 인증 token이 필요한가?

Yes. REST와 Robot WebSocket 모두 같은 `robotToken`을 Bearer token으로 보낸다.

```http
Authorization: Bearer {robotToken}
```

token은 HTTP header에 넣는다. URL query, WebSocket message payload, DataChannel payload에 넣지 않는다.

### Q7. `robot_id`, `session_id`, `room_id`가 필요한가?

| Identifier | Robot team input 여부 | 기준 |
| --- | --- | --- |
| `robot_id` | No | 관제 서버 내부 UUID다. 로봇팀 구현 입력값으로 쓰지 않는다. |
| `robotCode` | 직접 입력 최소화 | token으로 서버가 판단한다. 화면/로그 식별용으로만 다룬다. WebSocket query에 붙이지 않는다. |
| `session_id` | No | 클라이언트가 만들 필요 없다. WebSocket join 후 서버가 `peerId`를 내려준다. |
| `room_id` | Yes, but response-driven | room id는 `missionCode`와 같은 값이다. 직접 새로 정하지 말고 mission 응답의 `roomId` 또는 `sfu.signalingUrl`을 사용한다. |

정리하면, 테스트 시작 전 필요한 고정 입력값은 `serverUrl`이다. `POST /api/robots`로 자기 테스트용 `robotCode`와 `robotToken`을 만들고, `POST /api/missions`와 start API로 자기 테스트용 `missionCode`를 만든다. Robot publisher 실행 시에는 `serverUrl`과 `robotToken`을 보관하면 되고, active mission이 있으면 서버가 `robotCode`, `missionCode`, `roomId`, `sfu.signalingUrl`, `turnServers`를 응답으로 내려준다.
