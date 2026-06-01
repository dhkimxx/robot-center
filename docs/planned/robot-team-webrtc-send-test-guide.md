---
title: "robot-team-webrtc-send-test-guide"
created: 2026-05-27
updated: '2026-06-01'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "guide"
status: "planned"
tags: ["robot", "webrtc", "sfu", "sensor", "integration", "guide"]
history:
- "2026-05-27 danya.kim <danya.kim@thundersoft.com>: split robot team WebRTC send guide from dev-server deployment runbook"
- "2026-05-27 danya.kim <danya.kim@thundersoft.com>: made guide self-contained for robot team sharing"
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: sync robot team WebRTC guide with current gateway and sensor API contract'
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
- '2026-05-28 danya.kim <danya.kim@thundersoft.com>: clarify DataChannel lifecycle'
- '2026-05-28 danya.kim <danya.kim@thundersoft.com>: keep the robot-team sharing guide focused on negotiation details'
- '2026-05-28 danya.kim <danya.kim@thundersoft.com>: simplify sensor contract to structured descriptors plus object values'
- '2026-05-28 danya.kim <danya.kim@thundersoft.com>: remove descriptor metadata and move gas channel alarm fields into sample values'
- '2026-05-28 danya.kim <danya.kim@thundersoft.com>: clarify telemetry-only payload schema contract for robot team send test'
- '2026-06-01 danya.kim <danya.kim@thundersoft.com>: separate robot-facing API namespace and remove internal diagnostics from robot guide'
- '2026-06-01 danya.kim <danya.kim@thundersoft.com>: clarify /api/v1/robot self-scope API boundary'
- '2026-06-01 danya.kim <danya.kim@thundersoft.com>: add Swagger API docs URL for robot team field reference'
---

# Robot Team WebRTC Send Test Guide

## 1. 목적

이 문서는 로봇팀이 실제 Robot Gateway/Publisher에서 관제 서버로 WebRTC 영상과 센서 데이터를 송신할 때 필요한 연동 절차를 정의한다.

현재 관제 시스템은 임무 단위 WebRTC room을 만들고, 로봇은 해당 room에 publisher로 접속한다. 로봇팀은 이 문서의 `/api/v1/robot/*` REST API와 Robot WebSocket signaling 계약만 구현하면 된다.

이 문서 하나만 보고 테스트를 진행할 수 있도록 서버 주소, REST API, WebRTC signaling, media track, DataChannel, 통과 기준을 함께 정리한다.

테스트에서 확인하려는 것:

- 로봇이 관제 서버에 heartbeat를 보낼 수 있는가
- 로봇이 자신에게 배정된 active mission을 조회할 수 있는가
- 로봇이 mission room에 WebRTC publisher로 접속할 수 있는가
- 관제 서버가 로봇의 RGB/Thermal/Audio track을 수신할 수 있는가
- 관제 서버가 telemetry DataChannel 메시지를 수신할 수 있는가
- 관제 UI에서 영상, 위치, 센서값을 확인할 수 있는가

이 테스트에서 제외하는 것:

- 로봇 장치 내부 camera/sensor/ROS/GStreamer 구현 방식
- 제어 명령 송신과 control ACK 정책
- HTTPS/WSS 운영화
- 장기 운영 인증/권한 정책

### 1.1 Robot API 경계

로봇 런타임이 직접 호출하는 API는 `/api/v1/robot/*` 하위로 제한한다.

```text
POST /api/v1/robot/heartbeat
GET  /api/v1/robot/mission
GET  /api/v1/robot/sfu/ws?room={missionCode}
```

이 API들은 모두 `Authorization: Bearer {robotToken}`으로 인증한다. 서버는 token으로 robot identity를 판단하므로 로봇 런타임은 request body, query, WebSocket message, DataChannel payload에 `robotCode`, `robotId`, `sessionId`, `roomId`를 넣지 않는다.

`/api/v1/robot/*` 응답은 인증된 자기 로봇이 지금 publish하는 데 필요한 정보만 포함한다.

- active mission이 없으면 `missionStatus=none`만 기준으로 재시도한다.
- active mission이 있으면 자기 로봇에 배정된 `missionCode`, `sfu.signalingUrl`, `turnServers`, `tracks`, `dataChannels`만 사용한다.
- 다른 robot, 다른 mission, 관제 UI 상태, 저장/녹화 상태, 내부 worker health 정보는 로봇 API 계약에 포함하지 않는다.

이 문서의 `/api/v1/operator/robots`, `/api/v1/operator/missions` 호출은 개발서버에서 테스트 slot을 직접 만들기 위한 편의 절차다. 실제 Robot Gateway/Publisher 런타임 루프는 `/api/v1/robot/*`만 호출한다.

## 2. 테스트 서버 접속 정보

현재 임시 개발서버는 배포 완료 상태다.

```text
serverUrl: http://192.168.20.12:18080
operatorUi: http://192.168.20.12:18080
apiDocs: http://192.168.20.12:18080/swagger/index.html
openapiJson: http://192.168.20.12:18080/swagger/openapi.json
missionCode: 각 테스트자가 직접 생성
missionStatus: active로 시작 후 테스트
```

Swagger UI와 OpenAPI JSON은 관제 서버 전체 API field reference다. 로봇팀 구현 계약은 이 문서의 `/api/v1/robot/*` REST/WebSocket API와 WebRTC/DataChannel 계약만 기준으로 삼는다.

현재 관제팀 재현 결과:

```text
verifiedAt: 2026-05-27 18:01 KST
verifiedWith: Android Robot app 2 devices
robot-001: heartbeat OK, mission OK, WebSocket joined, relay ICE CONNECTED/COMPLETED
robot-002: heartbeat OK, mission OK, WebSocket joined, relay ICE CONNECTED/COMPLETED
app-server SFU: mission-001 room, robotCount 2
관제 UI: RGB/Thermal/Audio, telemetry 수신 확인
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

로봇 구현에서는 TURN 값을 하드코딩하지 말고 `GET /api/v1/robot/mission` 응답의 `turnServers`를 그대로 `RTCPeerConnection.iceServers`에 넣는다. 위 값은 네트워크 디버깅과 수동 테스트를 위한 현재 서버 값이다.

이 문서의 주소는 모두 로봇팀 단말에서 접근하는 public address 기준이다. Docker 내부에서 쓰는 `app-server`, `turn` 같은 service DNS는 관제 서버 내부 구현값이며 로봇팀 구현에 사용하지 않는다.

방화벽/네트워크 확인 대상:

| Purpose | Address |
| --- | --- |
| REST API / Web UI / WebSocket signaling | `192.168.20.12:18080/tcp` |
| TURN allocation | `192.168.20.12:3478/udp`, `192.168.20.12:3478/tcp` |
| TURN relay port range | `192.168.20.12:49160-49300/udp`, `192.168.20.12:49160-49300/tcp` |

### 2.1 개인 테스트용 robot과 mission 생성

아래 명령은 테스트 담당자가 자기 robot과 mission을 새로 만드는 절차다. `TEST_OWNER`는 충돌을 피할 수 있게 본인 이름, 장비명, 날짜 등을 넣는다.

```bash
SERVER_URL='http://192.168.20.12:18080'
TEST_OWNER='robot-team-your-name-jetson-01'
```

Robot 생성:

```bash
ROBOT_RESPONSE="$(
  curl -fsS -X POST "$SERVER_URL/api/v1/operator/robots" \
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
  curl -fsS -X POST "$SERVER_URL/api/v1/operator/missions" \
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
curl -fsS -X POST "$SERVER_URL/api/v1/operator/missions/$MISSION_CODE/start" \
  | python3 -m json.tool
```

테스트 중 사용할 값:

```bash
printf 'SERVER_URL=%s\nROBOT_CODE=%s\nROBOT_TOKEN=%s\nMISSION_CODE=%s\n' \
  "$SERVER_URL" "$ROBOT_CODE" "$ROBOT_TOKEN" "$MISSION_CODE"
```

생성 확인:

```bash
curl -fsS "$SERVER_URL/api/v1/robot/mission" \
  -H "Authorization: Bearer $ROBOT_TOKEN" \
  | python3 -m json.tool
```

정상이라면 `missionStatus`는 `active`, `missionCode`는 방금 생성한 값, `sfu.signalingUrl`은 `ws://192.168.20.12:18080/api/v1/robot/sfu/ws?room={MISSION_CODE}` 형태다.

### 2.2 빠른 접속 확인

아래 예시는 2.1에서 생성한 `ROBOT_TOKEN`을 그대로 사용한다.

```bash
SERVER_URL='http://192.168.20.12:18080'
ROBOT_TOKEN='<2.1에서 생성한 robotToken>'
```

Heartbeat:

```bash
curl -fsS -X POST "$SERVER_URL/api/v1/robot/heartbeat" \
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
curl -fsS "$SERVER_URL/api/v1/robot/mission" \
  -H "Authorization: Bearer $ROBOT_TOKEN" \
  | python3 -m json.tool
```

현재 정상 응답이면 `missionStatus`는 `active`, `missionCode`는 2.1에서 생성한 값, `sfu.signalingUrl`은 `ws://192.168.20.12:18080/api/v1/robot/sfu/ws?room={missionCode}`로 내려온다.

## 3. 전체 연결 구조

```text
Robot Gateway/Publisher
  -> REST heartbeat
  -> REST active mission lookup
  -> WebRTC signaling WebSocket
  -> app-server internal SFU
  -> Browser 관제 UI
```

역할:

| Component | Role |
| --- | --- |
| Robot Gateway/Publisher | media track과 DataChannel을 publish |
| app-server | REST API, robot token 검증, mission 관리, WebRTC signaling, SFU |
| Browser 관제 UI | operator subscriber, 선택한 robotCode의 live stream 표시 |
| TURN | relay-only ICE 경로 |

관제 서버 내부 저장/녹화 진단 API는 로봇팀 구현 계약에 포함하지 않는다. 로봇팀은 `/api/v1/robot/*` 응답과 WebSocket signaling 메시지만 기준으로 삼는다.

WebRTC room 규칙:

- room id는 `missionCode`다.
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

실제 로봇 구현은 로봇팀 코드베이스에서 담당한다. 이 문서는 관제 서버와 Robot Gateway/Publisher 사이의 외부 연동 계약만 정의한다.

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
path: /api/v1/robot/heartbeat
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
| `robotCode` | string | `robot-123` | heartbeat가 반영된 robot code |
| `status` | string | `online`, `offline`, `fault` | 서버가 저장한 장치 상태. `state` 정규화 결과 |
| `serverTime` | string(date-time) | `2026-05-27T05:00:00.120Z` | 서버 응답 시각 |

`200 OK` response example:

```json
{
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
path: /api/v1/robot/mission
auth: Bearer token required
```

Request headers:

| Name | Required | Value |
| --- | --- | --- |
| `Authorization` | Yes | `Bearer {robotToken}` |

Request example:

```yaml
method: GET
url: http://192.168.20.12:18080/api/v1/robot/mission
headers:
  Authorization: Bearer {robotToken}
```

`200 OK` response body when no active mission exists:

```json
{
  "missionStatus": "none",
  "serverTime": "2026-06-01T01:30:00.000Z"
}
```

`200 OK` response body when active mission exists:

```json
{
  "missionCode": "mission-123",
  "missionStatus": "active",
  "serverTime": "2026-06-01T01:30:00.000Z",
  "sfu": {
    "signalingUrl": "ws://192.168.20.12:18080/api/v1/robot/sfu/ws?room=mission-123",
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
  "dataChannels": ["channel.telemetry", "channel.spatial", "channel.event", "channel.control"]
}
```

Active mission response schema:

| Field | Type | Values / Example | Description |
| --- | --- | --- | --- |
| `missionCode` | string | `mission-123` | 사람이 읽는 mission code이며 WebRTC room id |
| `missionStatus` | string | `active`, `none` | gateway mission 조회 응답에서는 `active`이면 WebRTC publish 가능. `none`은 active mission 없음. 내부 mission lifecycle 값은 `ready`, `active`, `ended` |
| `serverTime` | string(date-time) | `2026-06-01T01:30:00.000Z` | 서버 응답 시각 |
| `sfu.signalingUrl` | string | `ws://192.168.20.12:18080/api/v1/robot/sfu/ws?room=mission-123` | Robot publisher WebSocket URL. client가 재구성하지 않고 그대로 사용 |
| `sfu.iceTransportPolicy` | string | `relay` | 현재 `relay`만 사용 |
| `turnServers[].urls` | string[] | `["turn:192.168.20.12:3478?transport=udp"]` | TURN URL 목록 |
| `turnServers[].username` | string | `robot-center-turn` | TURN username |
| `turnServers[].credential` | string | `rc-turn-2026-0527` | TURN password |
| `tracks` | string[] | `track.video_1`, `track.video_2`, `track.audio_1`, `track.audio_2` | canonical media track slot 목록 |
| `dataChannels` | string[] | `channel.telemetry`, `channel.spatial`, `channel.event`, `channel.control` | canonical DataChannel label 목록 |

Error responses:

| Status | Meaning |
| --- | --- |
| `401` | Bearer token이 없거나 유효하지 않음 |
| `404` | token에 해당하는 robot이 등록되어 있지 않음 |

Client behavior:

- `missionStatus=none`이면 WebRTC publish를 시작하지 않는다.
- active mission이 없을 때도 heartbeat와 mission 조회를 계속 재시도한다.
- active mission이 오면 `sfu.signalingUrl`, `turnServers`를 그대로 사용한다.
- `robotCode`, `robotId`, `sessionId`, `roomId`를 request query나 signaling payload에 별도로 넣지 않는다.

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
url: ws://192.168.20.12:18080/api/v1/robot/sfu/ws?room=mission-123
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
| `channel.telemetry` | `telemetry` | GPS, battery, 가스 같은 저속 상태. 이번 테스트에서 payload schema 확정 |
| `channel.spatial` | `spatial` 또는 domain-specific type | IMU, odometry, point cloud descriptor. 이번 테스트에서는 label 예약 |
| `channel.event` | `event` 또는 domain-specific type | alarm, fault, detection, mission event. 이번 테스트에서는 label 예약 |
| `channel.control` | reserved | reserved control/ack side channel. 이번 테스트에서는 label 예약 |

### 8.1 Payload schema 확정 범위

현재 로봇팀 송신 테스트에서 확정된 payload schema는 `channel.telemetry`의 sensor envelope 구조다.

| Label | Negotiation status | Payload schema status | 이번 테스트 필수 여부 |
| --- | --- | --- | --- |
| `channel.telemetry` | 확정 | 확정. `descriptors` / `samples` / `values` 구조 사용 | Yes |
| `channel.spatial` | DataChannel label 예약 | 미확정. mock은 IMU/odometry 예시를 보내지만 실제 로봇 계약으로 고정하지 않음 | No |
| `channel.event` | DataChannel label 예약 | 미확정. alarm/fault/detection/mission event taxonomy는 별도 협의 필요 | No |
| `channel.control` | DataChannel label 예약 | 미확정. command/ack/권한/감사 정책은 별도 협의 필요 | No |

따라서 이번 로봇팀 송신 테스트에서 가스, GPS, battery 같은 센서 측정값은 `channel.telemetry`로 보낸다. `channel.spatial`, `channel.event`, `channel.control`은 offer/DataChannel negotiation에 포함할 수 있지만, payload 세부 schema는 이 문서에서 확정 계약으로 보지 않는다.

`messageType`은 payload subtype 식별자이며 1차 라우팅 기준이 아니다. 현재 1차 라우팅 기준은 DataChannel label이다.

권장 `messageType`:

| Label | Recommended messageType |
| --- | --- |
| `channel.telemetry` | `telemetry` 또는 `telemetry.*` |
| `channel.spatial` | `spatial.*` |
| `channel.event` | `event.*` |
| `channel.control` | `control.*` |

관제 서버는 현재 `channel.telemetry`의 `descriptors` / `samples` / `values` 구조를 저장 및 Live UI 표시 검증 대상으로 본다. 나머지 채널은 label과 negotiation 확인용으로만 다루며, payload schema를 확정하려면 관제팀과 별도 합의한다.

### 8.2 DataChannel lifecycle

DataChannel 생성과 전송 시작 조건:

```text
1. Robot은 offer 생성 전에 필요한 DataChannel을 생성한다.
2. Robot은 offer를 WebSocket으로 보낸다.
3. Robot은 answer를 받아 remote description으로 적용한다.
4. ICE / DTLS / SCTP / DataChannel negotiation이 끝난다.
5. 각 DataChannel이 OPEN 상태가 된다.
6. Robot은 해당 DataChannel OPEN 이후에만 JSON payload를 send한다.
```

주의:

- `createDataChannel()` 직후에는 send하지 않는다.
- `answer` 수신 직후에는 send하지 않는다.
- ICE state가 `connected`가 됐더라도 DataChannel OPEN callback 전이면 send하지 않는다.
- SDK가 open callback을 제공하면 callback 이후 전송한다.
- SDK가 callback 대신 state polling을 사용하면 `readyState == open` 또는 동일 의미의 상태를 확인한 뒤 전송한다.

상태 판단 기준:

| 관측값 | 의미 |
| --- | --- |
| SFU `detected` | 로봇 offer의 DataChannel/SCTP 협상이 서버에 도달함 |
| SFU `open` | 로봇 publisher와 SFU 사이 DataChannel이 실제 OPEN됨 |
| SFU `lastMessageAt` | SFU가 로봇 DataChannel payload를 실제 수신함 |

로봇팀은 Robot Publisher local open callback과 SFU publisher 구간 상태만 기준으로 본다. 관제 내부 subscriber/downstream 상태는 로봇팀 판정 기준이 아니다.

### 8.3 `channel.telemetry` payload schema

`channel.telemetry` 예시:

```json
{
  "messageId": "uuid",
  "messageType": "telemetry",
  "descriptors": [
    {
      "sensorId": "telemetry.position_1",
      "label": "GPS",
      "sensorType": "position",
      "enabled": true
    },
    {
      "sensorId": "telemetry.gas.channel_1",
      "label": "CO",
      "sensorType": "gas",
      "unit": "ppm",
      "enabled": true
    }
  ],
  "samples": [
    {
      "sensorId": "telemetry.position_1",
      "timestamp": "2026-05-27T05:00:00.000Z",
      "values": {
        "latitude": 37.402183,
        "longitude": 127.106812,
        "accuracyMeter": 3.5
      }
    },
    {
      "sensorId": "telemetry.gas.channel_1",
      "timestamp": "2026-05-27T05:00:00.000Z",
      "values": {
        "concentration": 13.0,
        "scale_code": 1,
        "alarm_code": 0,
        "alarm": "normal",
        "low_alarm": 10.0,
        "high_alarm": 15.0,
        "valid": true
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
| `descriptors` | array | Conditional | SensorDescriptor list | 센서 식별/표시 schema. 새 `sensorId`를 처음 보낼 때는 필수 |
| `samples` | array | No | SensorSample list | 실제 측정값 |

SensorDescriptor schema:

| Field | Type | Required | Values / Example | Description |
| --- | --- | --- | --- | --- |
| `sensorId` | string | Yes | `telemetry.position_1`, `telemetry.battery_1`, `telemetry.gas.channel_1`, `spatial.imu_1` | robot 내부에서 안정적으로 쓰는 sensor id. descriptor/sample 매칭용 식별자이며 화면 해석 키로 쓰지 않는다 |
| `label` | string | Recommended | `GPS`, `Battery`, `CO` | 사람이 읽는 채널 label. 같은 `sensorType` 안에서 표시 전략의 보조 키로 사용 |
| `sensorType` | string | Yes | `position`, `battery`, `imu`, `odometry`, `point_cloud`, `gas` | 센서 계열. 관제 UI의 1차 해석 전략 선택 키. 누락/오타는 서버가 거절 |
| `unit` | string | No | `percent`, `celsius`, `ppm`, `m`, `m/s` | 표시 단위 |
| `enabled` | boolean | No | `true`, `false` | UI/저장 대상으로 활성화할지 여부 |

SensorSample schema:

| Field | Type | Required | Values / Example | Description |
| --- | --- | --- | --- | --- |
| `sensorId` | string | Yes | `telemetry.position_1` | descriptor의 sensorId와 매칭 |
| `timestamp` | string(date-time) | Recommended | `2026-05-27T05:00:00.000Z` | sample 측정 시각 |
| `values` | object | Recommended | `{ "latitude": 37.402183 }` | 실제 측정값. 모든 sensorType에서 object로 통일 |
| `objectKey` | string | No | `missions/.../point_cloud.bin` | object storage 참조가 필요할 때 |

Position `values` 권장 필드:

| Field | Type | Required | Example | Description |
| --- | --- | --- | --- | --- |
| `latitude` | number | Yes | `37.402183` | WGS84 latitude |
| `longitude` | number | Yes | `127.106812` | WGS84 longitude |
| `altitudeMeter` | number | No | `42.5` | 고도 meter |
| `accuracyMeter` | number | No | `3.5` | 위치 정확도 meter |
| `headingDegree` | number | No | `90` | 진행 방향. 0-360 degree |

Gas module `values` 권장 필드:

| Field | Type | Required | Example | Description |
| --- | --- | --- | --- | --- |
| `concentration` | number | Yes | `13.0` | 측정값. TEMP/HUM도 장비 원본 필드명을 유지한다 |
| `scale_code` | number | No | `1` | 장비 scale code 원본값 |
| `alarm_code` | number | No | `0` | 장비 alarm code 원본값. 현재 관제 UI는 해석하지 않음 |
| `alarm` | string | No | `normal` | 장비 alarm 문자열. 현재 관제 UI는 해석하지 않음 |
| `low_alarm` | number | No | `10.0` | 장비 원본 하한 alarm 기준. 현재 관제 UI는 해석하지 않음 |
| `high_alarm` | number | No | `15.0` | 장비 원본 상한 alarm 기준. 현재 관제 UI는 해석하지 않음 |
| `valid` | boolean | No | `true` | 장비 원본 valid flag. 현재 관제 UI는 해석하지 않음 |

가스 채널 구성:

```text
channel name  -> descriptor.label
concentration -> sample.values.concentration
unit          -> descriptor.unit
scale_code    -> sample.values.scale_code
alarm_code    -> sample.values.alarm_code
alarm         -> sample.values.alarm
low_alarm     -> sample.values.low_alarm
high_alarm    -> sample.values.high_alarm
valid         -> sample.values.valid
```

`TEMP`, `HUM` 채널도 같은 가스 모듈의 5/6번 채널이므로 `sensorType`은 `gas`로 통일한다.

현재 관제 UI는 가스 모듈 descriptor의 `label`, `unit`과 sample `values.concentration`만 표시한다. `alarm_code`, `alarm`, `low_alarm`, `high_alarm`, `valid`는 저장/전달만 하고 경고 상태 계산에는 사용하지 않는다.

권장:

- `channel.telemetry`는 1Hz 수준의 저속 상태값부터 시작한다.
- GPS 위치는 `telemetry.position_1` 같은 안정적인 `sensorId`를 사용한다.
- descriptor는 매번 보내도 되고, sensor 구성이 바뀔 때 다시 보내도 된다.
- `sensorType`은 관제 UI 해석 전략 선택 키이며 descriptor 필수값이다. 새 타입이 필요하면 관제팀과 먼저 이름을 맞춘다.
- `unknown`은 관제 내부 fallback용 예약값이다. 로봇팀 payload에서 보내지 않는다.

## 9. 통과 기준

Robot gateway:

- heartbeat 성공
- mission 조회가 `missionStatus=active` 반환
- mission 응답의 `sfu.signalingUrl`로 WebSocket 연결
- WebSocket signaling 연결 성공

SFU/WebRTC:

- Robot publisher ICE state가 `connected` 또는 `completed`
- Robot publisher local DataChannel `channel.telemetry` open callback 발생
- Robot publisher에서 RGB/Thermal/Audio track 송신 시작
- Robot publisher에서 telemetry 첫 payload send 성공

관제팀 확인:

- 관제 UI에서 RGB/Thermal 영상 표시
- GPS/position sample이 있으면 관제 UI 위치 영역에 표시
- telemetry sample이 있으면 관제 UI 센서 영역에 표시

## 10. 장애 대응

| 증상 | 우선 확인 |
| --- | --- |
| heartbeat 401 | robotToken 누락/만료/불일치 |
| mission 조회가 `missionStatus=none` | active mission 없음, robot assignment 누락 |
| WebSocket 400 | mission 응답의 `sfu.signalingUrl`을 사용하지 않았거나 필수 query 누락 |
| `publish-error` | robot이 active mission에 배정되지 않았거나 room이 missionCode와 다름 |
| ICE `failed` | mission 응답의 `turnServers` 사용 여부, UDP 3478, relay port range, 방화벽, relay candidate 생성 여부 |
| WebSocket은 연결됐지만 영상 없음 | track publish, codec negotiation, track label/order |
| 영상은 보이나 센서 없음 | DataChannel label, open 상태, payload envelope |
| 센서는 오나 위치 없음 | position sensorId/value shape |

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
- DataChannel별 open callback 발생 여부와 발생 시각
- DataChannel별 첫 send 시각과 send 직전 state
- 마지막으로 송신한 DataChannel payload 예시

## 12. 테스트 결과 기록

관제팀과 로봇팀은 테스트 후 다음 정보를 남긴다.

- 테스트 일시
- robotCode
- missionCode
- 로봇팀 publisher 버전
- heartbeat 결과
- signaling 연결 결과
- ICE state
- track 수와 DataChannel 수
- RGB/Thermal 표시 여부
- telemetry send 여부
- 실패 로그와 재현 절차

## 13. 질문과 답변

### Q1. endpoint 주소는 무엇인가?

로봇팀 테스트에서 직접 호출하는 endpoint는 아래와 같다.

| Purpose | Protocol | Endpoint |
| --- | --- | --- |
| Robot 생성 | HTTP | `POST http://192.168.20.12:18080/api/v1/operator/robots` |
| Mission 생성 | HTTP | `POST http://192.168.20.12:18080/api/v1/operator/missions` |
| Mission 시작 | HTTP | `POST http://192.168.20.12:18080/api/v1/operator/missions/{missionCode}/start` |
| Heartbeat | HTTP | `POST http://192.168.20.12:18080/api/v1/robot/heartbeat` |
| Active mission 조회 | HTTP | `GET http://192.168.20.12:18080/api/v1/robot/mission` |
| Robot WebRTC signaling | WebSocket | mission 조회 응답의 `sfu.signalingUrl` |

`missionCode=mission-123` 예시의 signaling URL은 `ws://192.168.20.12:18080/api/v1/robot/sfu/ws?room=mission-123` 형태다. 로봇 구현은 이 값을 직접 조립하지 말고 mission 조회 응답의 `sfu.signalingUrl`을 그대로 사용한다.

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
| `room_id` | No separate input | room id는 `missionCode`와 같은 값이며 `sfu.signalingUrl` query에 포함되어 내려온다. 직접 새로 정하지 않는다. |

정리하면, 테스트 시작 전 필요한 고정 입력값은 `serverUrl`이다. `POST /api/v1/operator/robots`로 자기 테스트용 `robotCode`와 `robotToken`을 만들고, `POST /api/v1/operator/missions`와 start API로 자기 테스트용 `missionCode`를 만든다. Robot publisher 실행 시에는 `serverUrl`과 `robotToken`을 보관하면 되고, active mission이 있으면 서버가 `missionCode`, `sfu.signalingUrl`, `turnServers`를 응답으로 내려준다.
