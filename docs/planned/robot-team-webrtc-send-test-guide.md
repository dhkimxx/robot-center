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

## 2. 테스트 서버

```text
serverUrl: http://192.168.20.12:18080
```

관제팀은 테스트 전에 robot을 등록하고 active mission에 배정한다. 로봇팀에는 robot별로 다음 값을 별도 채널로 전달한다.

```yaml
serverUrl: http://192.168.20.12:18080
robotCode: robot-001
robotToken: rb_poc_xxxxx
```

`robotToken`은 Bearer token이다. 문서, git, 공개 채팅에 남기지 않는다.

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
- robot identity는 payload, status, recording metadata에서 `robotCode`로 유지한다.

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
| `robotCode` | string | Yes | `robot-001` | 관제팀이 발급한 robot code |
| `state` | string | No | `online`, `offline`, `fault`, other non-empty value | 생략 시 서버는 `online`으로 처리. `offline`, `fault` 외의 non-empty 값은 현재 서버에서 `online`으로 정규화 |
| `batteryPercent` | integer | No | `0`-`100` | 배터리 퍼센트. 현재 heartbeat 저장/판단에는 사용하지 않지만 요청 DTO에 포함 |
| `networkQuality` | string | No | `good`, `normal`, `poor`, `unknown` | 로봇 측 네트워크 상태. 현재 heartbeat 저장/판단에는 사용하지 않지만 요청 DTO에 포함 |
| `sentAt` | string(date-time) | No | `2026-05-27T05:00:00.000Z` | 로봇이 heartbeat를 보낸 시각. 현재 heartbeat 저장/판단에는 사용하지 않지만 요청 DTO에 포함 |

Request example:

```json
{
  "robotCode": "robot-001",
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
| `robotCode` | string | `robot-001` | heartbeat가 반영된 robot code |
| `status` | string | `online`, `offline`, `fault` | 서버가 저장한 장치 상태. `state` 정규화 결과 |
| `serverTime` | string(date-time) | `2026-05-27T05:00:00.120Z` | 서버 응답 시각 |

`200 OK` response example:

```json
{
  "robotId": "8f0e4c69-8c9b-40ef-a3fe-7b8a7ad9a111",
  "robotCode": "robot-001",
  "status": "online",
  "serverTime": "2026-05-27T05:00:00.120Z"
}
```

Error responses:

| Status | Meaning |
| --- | --- |
| `400` | JSON body 파싱 실패 또는 요청 값 오류 |
| `401` | `robotCode`와 token이 일치하지 않음 |
| `404` | robotCode가 등록되어 있지 않음 |

Client behavior:

- active mission 여부와 무관하게 주기적으로 heartbeat를 보낸다.
- 실패하면 2s, 5s, 10s 수준의 backoff로 재시도한다.
- `401` 또는 `404`는 token/robotCode 발급 상태를 관제팀과 확인한다.

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

Query parameters:

| Name | Type | Required | Allowed values / Example | Description |
| --- | --- | --- | --- | --- |
| `robotCode` | string | Yes | `robot-001` | 관제팀이 발급한 robot code |

Request example:

```yaml
method: GET
url: http://192.168.20.12:18080/api/robot-gateway/mission?robotCode=robot-001
headers:
  Authorization: Bearer rb_poc_xxxxx
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
  "missionCode": "mission-001",
  "missionStatus": "active",
  "roomId": "mission-001",
  "legacyRoomId": "mission-001__robot-001",
  "sfu": {
    "signalingUrl": "ws://192.168.20.12:18080/sfu/ws?room=mission-001&role=robot&robotCode=robot-001",
    "iceTransportPolicy": "relay"
  },
  "turnServers": [
    {
      "urls": ["turn:192.168.20.12:3478?transport=udp"],
      "username": "robot",
      "credential": "robot-pass"
    }
  ],
  "tracks": ["track.video_1", "track.video_2", "track.audio_1", "track.audio_2"],
  "dataChannels": ["channel.telemetry", "channel.spatial", "channel.event", "channel.control"],
  "legacyTracks": ["rgb", "thermal", "audio"],
  "legacyDataChannels": ["sensor", "telemetry"],
  "videoPolicy": {
    "mode": "robot_defined"
  }
}
```

Active mission response schema:

| Field | Type | Values / Example | Description |
| --- | --- | --- | --- |
| `missionId` | string(uuid) | `a8c2d4e1-25ef-4720-8d8c-2f4f5d0a1001` | 관제 서버 내부 mission id |
| `missionCode` | string | `mission-001` | 사람이 읽는 mission code이며 WebRTC room id |
| `missionStatus` | string | `active`, `none` | gateway mission 조회 응답에서는 `active`이면 WebRTC publish 가능. `none`은 active mission 없음. 내부 mission lifecycle 값은 `ready`, `active`, `ended` |
| `roomId` | string | `mission-001` | WebRTC room id. `missionCode`와 같아야 함 |
| `legacyRoomId` | string | `mission-001__robot-001` | 이전 호환용 room id. 신규 구현에서는 사용하지 않음 |
| `sfu.signalingUrl` | string | `ws://192.168.20.12:18080/sfu/ws?...` | Robot publisher WebSocket URL |
| `sfu.iceTransportPolicy` | string | `relay` | 현재 `relay`만 사용 |
| `turnServers[].urls` | string[] | `["turn:192.168.20.12:3478?transport=udp"]` | TURN URL 목록 |
| `turnServers[].username` | string | `robot` | TURN username |
| `turnServers[].credential` | string | `robot-pass` | TURN password |
| `tracks` | string[] | `track.video_1`, `track.video_2`, `track.audio_1`, `track.audio_2` | canonical media track slot 목록 |
| `dataChannels` | string[] | `channel.telemetry`, `channel.spatial`, `channel.event`, `channel.control` | canonical DataChannel label 목록 |
| `legacyTracks` | string[] | `rgb`, `thermal`, `audio` | 이전 호환용 label. 신규 구현에서는 사용하지 않음 |
| `legacyDataChannels` | string[] | `sensor`, `telemetry` | 이전 호환용 label. 신규 구현에서는 사용하지 않음 |
| `videoPolicy.mode` | string | `robot_defined` | 해상도/FPS는 로봇 송신 설정을 따름 |

Error responses:

| Status | Meaning |
| --- | --- |
| `400` | query parameter 오류 |
| `401` | `robotCode`와 token이 일치하지 않음 |
| `404` | robotCode가 등록되어 있지 않음 |

Client behavior:

- `missionStatus=none`이면 WebRTC publish를 시작하지 않는다.
- active mission이 없을 때도 heartbeat와 mission 조회를 계속 재시도한다.
- active mission이 오면 `sfu.signalingUrl`과 `turnServers`를 그대로 사용한다.
- `roomId`는 `missionCode`와 같아야 하며, 신규 구현은 `legacyRoomId`를 사용하지 않는다.

## 6. WebRTC Signaling

Robot publisher는 mission 조회 응답의 `sfu.signalingUrl`로 WebSocket 연결한다.

WebSocket URL:

```text
ws://192.168.20.12:18080/sfu/ws?room={missionCode}&role=robot&robotCode={robotCode}
```

Query parameters:

| Parameter | Value | Meaning |
| --- | --- | --- |
| `room` | `{missionCode}` | mission room id |
| `role` | `robot` | publisher role |
| `robotCode` | `{robotCode}` | robot identity |

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

신규 구현은 canonical track slot을 우선 사용한다.

| Slot | Kind | Expected value | Required |
| --- | --- | --- | --- |
| `track.video_1` | video | RGB 또는 주 영상. H.264 우선 테스트 | Yes |
| `track.video_2` | video | Thermal 또는 보조 영상. H.264 우선 테스트 | Recommended |
| `track.audio_1` | audio | Audio. Opus 우선 테스트 | Optional |
| `track.audio_2` | audio | Reserved secondary audio slot | Optional |

서버는 일부 legacy label fallback을 지원하지만, 로봇팀 연동 기준은 canonical slot이다.

권장:

- `track.video_1`에는 주 RGB 카메라를 송신한다.
- `track.video_2`에는 thermal 또는 보조 영상을 송신한다.
- audio가 없으면 `track.audio_1`은 생략 가능하다.
- 영상 codec은 우선 H.264로 테스트한다.

## 8. DataChannel

신규 구현은 canonical DataChannel label을 우선 사용한다.

| Label | Expected messageType | 용도 |
| --- | --- | --- |
| `channel.telemetry` | `telemetry` | GPS, battery, 환경값 같은 저속 상태. recorder-worker가 sensor API로 저장 |
| `channel.spatial` | `spatial` 또는 domain-specific type | IMU, odometry, point cloud descriptor. recorder-worker가 sensor API로 저장 |
| `channel.event` | `event` 또는 domain-specific type | alarm, fault, detection, mission event. 현재 recorder-worker 저장 대상은 아님 |
| `channel.control` | reserved | reserved control/ack side channel. 현재 recorder-worker 저장 대상은 아님 |

Legacy label fallback:

| Incoming label | Server normalized label |
| --- | --- |
| `telemetry` | `channel.telemetry` |
| `sensor` | `channel.telemetry` |
| `spatial` | `channel.spatial` |
| `event` | `channel.event` |
| `control` | `channel.control` |

`channel.telemetry` 예시:

```json
{
  "messageId": "uuid",
  "messageType": "telemetry",
  "channelRole": "channel.telemetry",
  "robotCode": "robot-001",
  "missionId": "a8c2d4e1-25ef-4720-8d8c-2f4f5d0a1001",
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

`robotCode`는 payload에도 유지한다. 같은 mission room 안에서 여러 robot publisher를 구분하는 필수 식별자다.

Telemetry envelope schema:

| Field | Type | Required | Values / Example | Description |
| --- | --- | --- | --- | --- |
| `messageId` | string | Recommended | UUID 또는 `robot-001-telemetry-102` | 메시지 추적 id |
| `messageType` | string | Recommended | `telemetry` | telemetry channel 기본 타입 |
| `channelRole` | string | Recommended | `channel.telemetry` | 송신 DataChannel label과 일치 권장 |
| `robotCode` | string | Yes | `robot-001` | 관제팀이 발급한 robot code |
| `missionId` | string(uuid) | Yes | `a8c2d4e1-25ef-4720-8d8c-2f4f5d0a1001` | mission 조회 응답의 `missionId`. sensor 저장 경로에서는 필수 |
| `sequence` | integer | Recommended | `102` | robotCode/channel별 증가값 |
| `sentAt` | string(date-time) | Recommended | `2026-05-27T05:00:00.000Z` | 로봇 송신 시각 |
| `descriptors` | array | No | SensorDescriptor list | 센서 metadata |
| `samples` | array | No | SensorSample list | 실제 측정값 |
| `payload` | object | No | `{ "batteryPercent": 82 }` | `descriptors`/`samples`가 없을 때 legacy sample로 저장 |

SensorDescriptor schema:

| Field | Type | Required | Values / Example | Description |
| --- | --- | --- | --- | --- |
| `sensorId` | string | Yes | `telemetry.position_1`, `telemetry.battery_1`, `spatial.imu_1` | robot 내부에서 안정적으로 쓰는 sensor id |
| `displayName` | string | Recommended | `GPS`, `Battery`, `IMU` | UI 표시 이름 |
| `sensorType` | string | Recommended | `position`, `battery`, `environment`, `imu`, `odometry`, `point_cloud`, `gas`, `event` | 센서 계열 |
| `kind` | string | No | `position`, `battery`, `gas` | legacy alias. `sensorType`이 없을 때 서버가 사용 |
| `valueType` | string | Recommended | `number`, `boolean`, `string`, `vector`, `object`, `object_ref` | sample 값 형태 |
| `unit` | string | No | `percent`, `celsius`, `ppm`, `m`, `m/s` | 표시 단위 |
| `sampleRateHz` | number | No | `1`, `5`, `0.2` | canonical sampling rate |
| `samplingRate` | number | No | `1`, `5`, `0.2` | legacy alias. `sampleRateHz`가 없을 때 서버가 사용 |
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
| `values` | any | Recommended | `{ "latitude": 37.402183 }` | 실제 측정값. 서버는 `objectValue`로 저장 |
| `objectValue` | object | No | `{ "latitude": 37.402183 }` | `values` 대신 사용할 수 있는 object 값 |
| `vectorValue` | object | No | `{ "x": 1.0, "y": 2.0, "z": 3.0 }` | vector 값 |
| `numericValue` | number | No | `82` | 단일 숫자 값일 때 사용 가능 |
| `textValue` | string | No | `nominal` | 단일 문자열 값일 때 사용 가능 |
| `boolValue` | boolean | No | `true` | 단일 boolean 값일 때 사용 가능 |
| `objectKey` | string | No | `missions/.../point_cloud.bin` | object storage 참조가 필요할 때 |
| `rawPayload` | object | No | `{ "source": "robot-gateway" }` | sample 원문 일부 |

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
- `sequence`는 robotCode별 channel 안에서 증가시킨다.
- `sentAt`은 로봇 송신 시각이다.
- descriptor는 매번 보내도 되고, sensor 구성이 바뀔 때 다시 보내도 된다.

## 9. 관제팀 확인 기준

관제팀은 다음 API와 UI로 로봇 송신을 확인한다.

System:

```bash
curl -fsS http://192.168.20.12:18080/api/system/status | python3 -m json.tool
```

Mission live-status:

```bash
curl -fsS http://192.168.20.12:18080/api/missions/{missionCode}/live-status \
  | python3 -m json.tool
```

Latest sensor:

```bash
curl -fsS 'http://192.168.20.12:18080/api/sensor-latest?missionId={missionId}&robotCode={robotCode}' \
  | python3 -m json.tool
```

Recorder runtime:

```bash
curl -fsS http://192.168.20.12:18082/healthz | python3 -m json.tool
```

UI:

```text
http://192.168.20.12:18080
```

## 10. 통과 기준

Robot gateway:

- heartbeat 성공
- mission 조회가 `missionStatus=active` 반환
- `roomId == missionCode`
- WebSocket signaling 연결 성공

SFU/WebRTC:

- `/api/system/status`의 `sfuRooms`에 mission room 표시
- 해당 room의 `robotCount`가 송신 로봇 수와 일치
- published tracks에 `robotCode:track.video_1` 표시
- `GET /api/missions/{missionCode}/live-status`에서 robot별 `stream.state=streaming`
- recorder-worker health에서 해당 robot의 track/data 수신 시각 확인

Sensor/UI:

- `sensor-latest`에 robotCode별 sensor 목록 표시
- GPS/position sample이 있으면 관제 UI 위치 영역에 표시
- RGB/Thermal 영상이 live 화면에 표시
- recording 상태가 `recording` 또는 기대 상태로 표시

## 11. 장애 대응

| 증상 | 우선 확인 |
| --- | --- |
| heartbeat 401 | robotCode와 robotToken mismatch |
| mission 조회가 `missionStatus=none` | active mission 없음, robot assignment 누락 |
| WebSocket 400 | `room`, `role`, `robotCode` query parameter 누락 |
| `publish-error` | robot이 active mission에 배정되지 않았거나 room이 missionCode와 다름 |
| ICE `failed` | TURN URL, UDP 3478, relay port range, 방화벽 |
| room은 보이나 영상 없음 | track publish, codec negotiation, track label/order |
| 영상은 보이나 센서 없음 | DataChannel label, open 상태, payload envelope |
| 센서는 오나 위치 없음 | position sensorId/value shape |
| recorder가 idle | recorder-worker join, active recording target, ICE state |

## 12. 로봇팀이 관제팀에 공유할 로그

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

## 13. 테스트 결과 기록

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
