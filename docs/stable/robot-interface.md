---
title: "robot-interface"
created: 2026-05-26
updated: '2026-05-28'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "design"
status: "stable"
tags: ["robot", "gateway", "webrtc", "sfu", "streaming", "mission"]
history:
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: mission scoped robot gateway and WebRTC interface 정리"
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: missionId UUID, roomId missionCode, streaming freshness updatedAt 기준 명시"
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: moved into docs/stable lifecycle structure"
- '2026-05-26 danya.kim <danya.kim@thundersoft.com>: simplified robot WebRTC connection by making SFU observed streams the live source of truth'
- '2026-05-26 danya.kim <danya.kim@thundersoft.com>: removed streaming-status gateway API from robot requirements'
- '2026-05-26 danya.kim <danya.kim@thundersoft.com>: removed streaming-status gateway API from robot interface'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: clarified robot device_state persistence and computed API status semantics'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: sync robot interface with current app-server gateway and sensor contracts'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: clarify sensor sample quality and recorder data channel storage boundaries'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: simplify robot external contract with REST token identity and no publisher token'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: remove publisher token from P0 robot contract'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: document robot WebSocket endpoint with robot token authorization'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: sync TURN public/internal terminology with current config'
- '2026-05-28 danya.kim <danya.kim@thundersoft.com>: clarify DataChannel open-before-send contract'
---

# Robot Gateway Interface

## 1. 문서 목적

Python Mock Robot, Android Mock Robot, 향후 실제 Robot Gateway가 관제센터에 연결하기 위한 최소 API, WebRTC, DataChannel 계약을 정의한다.

P0에서는 이 repo의 Python Mock Robot을 기본 로컬 검증 샘플로 사용하고, Android Mock Robot은 단말 검증용 샘플로 둔다. 둘 다 관제팀 테스트용 client이며 실제 로봇 제품 구현은 로봇팀 코드베이스에서 담당한다.

이 문서의 JSON과 track 값은 현재 mock/harness 연동 예시다. 확정되지 않은 DataChannel payload 세부 필드, track metadata, codec 세부 정책은 별도 schema가 생기기 전까지 고정 계약으로 간주하지 않는다.

## 2. 설계 원칙

- 로봇 등록 생성은 관제센터 UI/Backend에서 수행한다.
- Mock Robot 또는 실제 Robot Gateway는 발급받은 연결 정보를 입력받아 연결한다.
- 별도 CLI는 만들지 않는다.
- QR 등록은 P0에서 제외한다.
- gatewayVersion, capabilities, hardware fingerprint는 P0 필수값에서 제외한다.
- WebRTC 송출 스펙은 `robot_defined`로 둔다.
- 실제 송출 여부는 app-server 내부 SFU가 관측한 publisher/track/DataChannel 상태를 기준으로 판단한다.
- 관제 Live 화면의 통합 상태는 `GET /api/missions/{missionCode}/live-status`에서 확인한다.
- codec, 해상도, FPS, bitrate 같은 상세 송출 metadata 보고는 선택 기능으로 둔다.

## 3. 전체 연결 흐름

```text
1. 관제 UI에서 로봇 생성
2. Backend가 robotCode, robotToken, serverUrl 발급
3. 관제 UI가 Mock Robot / Robot Gateway 입력값을 표시
4. Mock Robot 또는 Robot Gateway에 값 입력
5. Mock Robot 또는 Robot Gateway가 heartbeat 호출
6. Backend가 로봇 online 표시
7. Mock Robot 또는 Robot Gateway가 mission 조회
8. active mission이면 mission room으로 SFU publish 시작
9. Browser와 Recorder가 SFU subscriber로 수신
```

## 4. Mock Robot / Robot Gateway에 입력할 값

```yaml
serverUrl: http://localhost:8080
robotCode: robot-001
robotToken: rb_poc_xxxxx
```

`serverUrl`은 app-server의 `APP_SERVER_PUBLIC_URL` 값이다. Mock Robot 또는 실제 Robot Gateway는 `serverUrl`을 기준으로 REST API를 호출하고, mission 응답에 포함된 SFU/TURN 설정으로 WebRTC publish를 시작한다. 같은 mission에 여러 Robot이 배정되더라도 각 Robot Gateway 인스턴스는 자기 `robotCode`와 `robotToken`으로 개별 실행한다.

P0 UI에서는 이 값을 복사하기 쉬운 형태로 표시한다.

## 5. 인증

Robot Gateway API는 bearer token을 사용한다.

```http
Authorization: Bearer {robotToken}
```

P0 token 정책:

- 로봇 생성 시 1회 발급
- UI에서 token 확인 가능
- token rotation은 관제 API `POST /api/robots/{robotCode}/connection-token`에서 수행
- Robot Gateway REST identity의 source of truth는 Bearer `robotToken`이다.
- `robotCode`는 운영자/로그 표시용으로 알고 있어도 되지만, REST heartbeat/mission 조회 요청에서 필수 식별자로 보내지 않는다.
- 전환 기간에 `robotCode` query/body가 들어오면 서버는 token의 robot과 일치하는지 확인하고, 다르면 `401`을 반환한다.

## 6. REST API

### 6.1 Heartbeat

로봇이 online 상태와 기본 상태를 보고한다.

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

Request body:

| Field | Type | Required | Values / Example | Description |
| --- | --- | --- | --- | --- |
| `state` | string | No | `online`, `offline`, `fault`, other non-empty value | 빈 값이면 `online`. `offline`, `fault` 외의 값은 현재 코드에서 `online`으로 정규화된다. |
| `batteryPercent` | integer | No | `0`-`100` | 현재 heartbeat 저장에는 사용하지 않지만 요청 DTO에 포함되어 있다. |
| `networkQuality` | string | No | `good`, `normal`, `poor`, `unknown` | 현재 heartbeat 저장에는 사용하지 않지만 요청 DTO에 포함되어 있다. |
| `sentAt` | string(date-time) | No | `2026-05-27T05:00:00.000Z` | 현재 heartbeat 저장에는 사용하지 않지만 요청 DTO에 포함되어 있다. |

Request example:

```json
{
  "state": "online",
  "batteryPercent": 82,
  "networkQuality": "good",
  "sentAt": "2026-05-18T08:00:00.000Z"
}
```

`200 OK` response:

```json
{
  "robotId": "9d3c1e5d-0c41-4e4f-a21f-8b69f7c0a001",
  "robotCode": "robot-001",
  "status": "online",
  "serverTime": "2026-05-18T08:00:00.120Z"
}
```

`status`는 DB의 `robots.device_state` 값이다. `state`가 빈 값이면 `online`, `offline`이면 `offline`, `fault`이면 `fault`, 그 외 non-empty 값은 `online`으로 저장된다.

### 6.2 Mission 조회

로봇이 현재 수행할 임무와 WebRTC 연결 정보를 조회한다.

```yaml
method: GET
path: /api/robot-gateway/mission
auth: Bearer token required
```

active mission이 있을 때:

```json
{
  "missionId": "9d3c1e5d-0c41-4e4f-a21f-8b69f7c0a001",
  "missionCode": "mission-001",
  "missionStatus": "active",
  "robotCode": "robot-001",
  "roomId": "mission-001",
  "sfu": {
    "signalingUrl": "ws://localhost:8080/sfu/robot/ws?room=mission-001",
    "iceTransportPolicy": "relay"
  },
  "turnServers": [
    {
      "urls": ["turn:127.0.0.1:3478?transport=udp"],
      "username": "robot",
      "credential": "robot-pass"
    }
  ],
  "tracks": ["track.video_1", "track.video_2", "track.audio_1", "track.audio_2"],
  "dataChannels": ["channel.telemetry", "channel.spatial", "channel.event", "channel.control"],
  "videoPolicy": {
    "mode": "robot_defined"
  }
}
```

Mission response field:

| Field | Type | Description |
| --- | --- | --- |
| `missionId` | string(uuid) | DB UUID |
| `missionCode` | string | 사람이 읽는 코드이자 SFU room id |
| `missionStatus` | string | 현재 active mission 조회에서는 `active` 또는 `none`을 기대한다. mission lifecycle 값은 `ready`, `active`, `ended`다. |
| `robotCode` | string | token으로 인증된 robot code. 로깅과 WebRTC publisher identity에 사용한다. |
| `roomId` | string | WebRTC room id. 현재 코드 기준 `missionCode`와 같다. |
| `sfu.signalingUrl` | string | Robot Gateway가 재구성하지 않고 그대로 접속할 publisher WebSocket URL |
| `sfu.iceTransportPolicy` | string | 현재 `relay` |
| `turnServers` | array | 로봇 외부 계약은 `TURN_PUBLIC_URL`, `TURN_USERNAME`, `TURN_PASSWORD`에서 내려간다. Docker 내부 recorder/SFU용 `TURN_INTERNAL_URL`은 로봇 구현에 사용하지 않는다. |
| `tracks` | string[] | canonical media track slot |
| `dataChannels` | string[] | canonical DataChannel role |
| `videoPolicy.mode` | string | 현재 `robot_defined` |

Mission 단위 multi-robot 구조에서 `roomId`는 `missionCode`와 같아야 한다. `robotCode`는 room id에 합치지 않고 WebRTC publisher identity, status, recording metadata에서 서버가 유지한다. `missionId`는 DB UUID이고, `missionCode`는 사람이 읽는 코드이자 SFU room id다.

active mission이 없을 때:

```json
{
  "missionId": null,
  "missionStatus": "none"
}
```

### 6.3 Live 상태 판단

Robot Gateway는 별도 streaming-status REST API를 호출하지 않는다. Live 화면은 `GET /api/missions/{missionCode}/live-status`를 사용하며, app-server가 SFU observed stream, recorder-worker runtime, mission assignment, robot heartbeat를 합성해 상태를 제공한다.

로봇 쪽 필수 책임은 heartbeat, active mission 조회, WebRTC signaling 접속, track/DataChannel publish다. 송출 여부는 로봇이 REST로 보고하는 값이 아니라 app-server 내부 SFU가 실제로 수신한 publisher, media track, DataChannel 상태로 판단한다.

### 6.4 Control ACK

현재 Robot Gateway용 Control ACK REST API는 구현되어 있지 않다. `channel.control`은 reserved DataChannel role이며, 제어 명령/ACK 계약은 P1에서 별도 확정한다.

## 7. WebRTC publish

Mock Robot 또는 실제 Robot Gateway는 active mission 수신 후 mission room에 SFU publish한다.

```text
Mock Robot / Robot Gateway
  -> SFU signalingUrl 접속
  -> offer 생성
  -> track.video_1 / track.video_2 / track.audio_1 track 추가
  -> channel.telemetry / channel.event / channel.spatial / channel.control DataChannel 생성
  -> relay ICE candidate 사용
  -> publish 시작
```

WebSocket endpoint는 app-server의 `GET /sfu/robot/ws?room={missionCode}`다. Robot Gateway는 mission 응답의 `sfu.signalingUrl`을 그대로 사용하고, `Authorization: Bearer {robotToken}` header를 포함한다. URL query를 client가 재구성하지 않는다.

P0에서는 WebRTC signaling에 별도 publisher token을 요구하지 않는다. app-server는 robot token으로 robotCode를 resolve하고 active mission assignment를 기준으로 publish를 검증한다. 운영화 단계에서는 단기 publisher grant/token 추가를 검토한다.

P0 media track:

| Slot | Kind | Codec | Required | Description |
| --- | --- | --- | --- | --- |
| `track.video_1` | video | H.264 우선 테스트 | Yes | RGB 또는 주 영상 |
| `track.video_2` | video | H.264 우선 테스트 | Recommended | Thermal 또는 보조 영상. RGB/Thermal 송신 테스트에서는 필요 |
| `track.audio_1` | audio | Opus 우선 테스트 | Optional | microphone |
| `track.audio_2` | audio | Opus 우선 테스트 | Optional | reserved secondary audio |

서버는 track ID 또는 stream ID에서 canonical slot을 찾는다. canonical slot이 없으면 legacy keyword(`rgb`, `thermal`, `audio`)와 media kind 순서로 fallback한다.

## 8. DataChannel

DataChannel은 역할별로 분리한다.

```text
channel.telemetry  descriptor/sample stream
channel.spatial    point cloud/space status or object reference
channel.event      alarm/fault/detection/mission event stream
channel.control    command ack/control side channel
```

Canonical DataChannel role:

| Label | Stored / Relayed Behavior |
| --- | --- |
| `channel.telemetry` | recorder-worker가 app-server `/api/sensor-samples`로 저장한다. |
| `channel.spatial` | recorder-worker가 app-server `/api/sensor-samples`로 저장한다. |
| `channel.event` | SFU relay/runtime 역할이다. 현재 recorder-worker 저장 대상은 아니다. |
| `channel.control` | reserved control/ack side channel이다. 현재 sensor API 저장 대상은 아니다. |

DataChannel lifecycle 계약:

- Robot Gateway는 offer 생성 전에 canonical DataChannel을 만든다.
- Robot Gateway는 `createDataChannel()` 직후 또는 `answer` 수신 직후에 payload를 send하지 않는다.
- Robot Gateway는 각 DataChannel의 OPEN callback 이후에만 payload를 send한다.
- SDK가 callback 대신 state polling을 사용하면 `readyState == open` 또는 동일 의미의 상태를 확인한 뒤 send한다.
- app-server SFU의 `lastDataAt`은 DataChannel open 시각이 아니라 실제 payload 수신 시각이다.
- recorder-worker의 `lastDataAt`은 SFU가 downstream DataChannel로 forward한 payload를 recorder가 수신한 시각이다.

Legacy DataChannel label fallback:

| Incoming label | Normalized label |
| --- | --- |
| `telemetry` | `channel.telemetry` |
| `sensor` | `channel.telemetry` |
| `spatial` | `channel.spatial` |
| `event` | `channel.event` |
| `control` | `channel.control` |

DataChannel 메시지는 공통 envelope를 사용한다. 현재 코드의 canonical sensor 저장 경로는 `descriptors` / `samples` 구조를 우선 처리하고, 둘 다 없고 `payload`가 있으면 `legacy.payload_1` sensor sample로 저장한다.

```json
{
  "messageId": "uuid",
  "messageType": "telemetry",
  "sequence": 1,
  "sentAt": "2026-05-18T08:00:00.000Z",
  "descriptors": [],
  "samples": []
}
```

`robotCode`, `missionId`, `missionCode`, `channelRole`은 로봇 payload에 넣지 않는다. 서버가 WebRTC publisher identity, room, DataChannel label에서 주입한다.

### 8.1 Telemetry

P0 telemetry는 `SensorDescriptor`와 `SensorSample` 개념을 따른다. 고정 필드만 전제로 하지 않고, UI는 descriptor를 보고 동적으로 렌더링할 수 있어야 한다. 로봇 payload는 측정값 중심으로 보내고 mission/robot/channel context는 서버가 채운다.

```json
{
  "messageId": "uuid",
  "messageType": "telemetry",
  "sequence": 102,
  "sentAt": "2026-05-18T08:00:00.000Z",
  "descriptors": [
    {
      "sensorId": "telemetry.position_1",
      "sensorType": "position",
      "valueType": "object",
      "displayName": "GPS",
      "sampleRateHz": 1,
      "enabled": true
    }
  ],
  "samples": [
    {
      "sensorId": "telemetry.position_1",
      "timestamp": "2026-05-18T08:00:00.000Z",
      "sequence": 102,
      "quality": "good",
      "values": {
        "latitude": 37.402183,
        "longitude": 127.106812,
        "accuracyMeter": 3.5
      }
    }
  ],
  "payload": {
    "batteryPercent": 82
  }
}
```

Sensor envelope:

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `messageId` | string | No | 메시지 추적 id |
| `messageType` | string | No | 예: `telemetry`, `spatial`, `event` |
| `sequence` | integer | No | envelope 기본 sequence |
| `sentAt` | string(date-time) | No | envelope 기본 송신 시각 |
| `descriptors` | array | No | SensorDescriptor list |
| `samples` | array | No | SensorSample list |
| `payload` | object | No | descriptors/samples가 없을 때 legacy sample로 저장 |

SensorDescriptor:

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `sensorId` | string | Yes | Robot 내부에서 안정적으로 쓰는 sensor id |
| `channelRole` | string | No | 없으면 envelope `channelRole` |
| `displayName` | string | No | 없으면 `sensorId` |
| `sensorType` | string | No | 없으면 `kind`, 그래도 없으면 sensorId에서 추론 |
| `kind` | string | No | legacy alias. `sensorType` 없을 때 사용 |
| `valueType` | string | No | 없으면 `object` |
| `unit` | string | No | 표시 단위 |
| `sampleRateHz` | number | No | canonical sampling rate |
| `samplingRate` | number | No | legacy alias. `sampleRateHz` 없을 때 사용 |
| `enabled` | boolean | No | descriptor 활성 여부 |
| `metadata` | object | No | frameId, axes 같은 부가 정보 |

SensorSample:

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `sensorId` | string | Yes | descriptor의 `sensorId`와 매칭 |
| `channelRole` | string | No | 없으면 envelope `channelRole` |
| `messageId` | string | No | 없으면 envelope `messageId` |
| `sequence` | integer | No | 없으면 envelope `sequence` |
| `timestamp` | string(date-time) | No | sample 측정 시각. `sentAt`보다 우선 |
| `sentAt` | string(date-time) | No | sample 송신 시각. 없으면 envelope `sentAt` |
| `quality` | string | No | 로봇 측 품질값. 현재 typed sample field로는 저장하지 않음 |
| `numericValue` | number | No | 단일 숫자 값 |
| `textValue` | string | No | 단일 문자열 값 |
| `boolValue` | boolean | No | 단일 boolean 값 |
| `vectorValue` | object | No | vector 값 |
| `objectValue` | object | No | object 값 |
| `values` | any | No | 있으면 `objectValue`로 저장 |
| `objectKey` | string | No | object storage 참조 |
| `rawPayload` | object | No | sample 원문 일부 |

송신 주기:

- P0 기본 1Hz
- 실제 로봇 연동 시 협의

### 8.2 Spatial / Event / Control

- `channel.spatial`: 기본 자동 표시 대상이 아니다. `available`, `subscribed`, `paused`, `unsupported` 같은 상태를 먼저 표현한다.
- `channel.event`: telemetry와 별도 경로다. alarm/fault/detection/mission event처럼 발생 순서가 중요한 메시지를 보낸다. 현재 recorder-worker 저장 대상은 아니다.
- `channel.control`: telemetry/event와 섞지 않는다. 향후 권한 체크, command validation, ack, audit log, rate limit이 붙을 자리다.

### 8.3 Relay

실시간 표시 경로:

```text
Robot DataChannel
  -> SFU application-level relay
  -> Browser DataChannel
```

저장 경로:

```text
Robot DataChannel
  -> SFU application-level relay
  -> Recorder DataChannel
  -> PostgreSQL / JSONL
```

Browser WebSocket은 실시간 센서 표시의 필수 경로가 아니다.

## 9. 상태 정의

| 상태 | 의미 |
| --- | --- |
| `offline` | heartbeat 없음 |
| `online` | heartbeat 성공 또는 최근 gateway 통신 성공 |
| `fault` | 오류 상태 |

DB의 `robots.device_state`는 장치 online/offline/fault 성격의 상태다. API의 `robot.status`는 `device_state + last_seen_at`을 합성한 현재 연결 상태로 내려간다. 임무 배정 상태는 `mission_robots.status`, WebRTC live 송출 상태는 app-server 내부 SFU의 observed stream 상태에서 판단한다.

## 10. 재시도 정책

Mock Robot / Robot Gateway 기본 재시도:

- heartbeat 실패: 2초, 5초, 10초 backoff
- mission 조회 실패: 2초, 5초, 10초 backoff
- signaling 끊김: mission이 active이면 재접속
- ICE failed: PeerConnection 재생성
- publish 실패: 연결 로그를 남기고 mission이 active이면 재시도

## 11. 실제 로봇 연동 시 협의 항목

Robot팀과 추후 확정한다.

- ROS1/ROS2 여부
- RGB camera topic
- thermal camera topic
- microphone/audio source
- H.264 GStreamer pipeline
- thermal H.264 지원 여부
- GPS/SLAM/Odometry 위치 source
- sensor topic 목록
- message timestamp 기준
- command 수신 방식
- controlAck schema
- 네트워크 환경
