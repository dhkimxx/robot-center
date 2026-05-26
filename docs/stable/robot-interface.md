---
title: "robot-interface"
created: 2026-05-26
updated: '2026-05-26'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "design"
status: "stable"
tags: ["robot", "gateway", "webrtc", "sfu", "streaming", "mission"]
history:
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: mission scoped robot gateway and WebRTC interface 정리"
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: missionId UUID, roomId missionCode, streaming freshness updatedAt 기준 명시"
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: moved into docs/stable lifecycle structure"
---

# Robot Gateway Interface

## 1. 문서 목적

Python Mock Robot, Android Mock Robot, 향후 실제 Robot Gateway가 관제센터에 연결하기 위한 최소 API, WebRTC, DataChannel 계약을 정의한다.

P0에서는 Python Mock Robot을 기본 로컬 검증 샘플로 사용하고, Android Mock Robot은 단말 검증용 샘플로 둔다.

이 문서의 JSON과 track 값은 현재 mock/harness 연동 예시다. 확정되지 않은 DataChannel payload 세부 필드, track metadata, codec 세부 정책은 별도 schema가 생기기 전까지 고정 계약으로 간주하지 않는다.

## 2. 설계 원칙

- 로봇 등록 생성은 관제센터 UI/Backend에서 수행한다.
- Mock Robot 또는 실제 Robot Gateway는 발급받은 연결 정보를 입력받아 연결한다.
- 별도 CLI는 만들지 않는다.
- QR 등록은 P0에서 제외한다.
- gatewayVersion, capabilities, hardware fingerprint는 P0 필수값에서 제외한다.
- WebRTC 송출 스펙은 `robot_defined`로 둔다.
- 실제 송출/수신/저장된 codec, 해상도, FPS, bitrate는 status/metadata로 남긴다.

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
9. Mock Robot 또는 Robot Gateway가 streaming status 보고
10. Browser와 Recorder가 SFU subscriber로 수신
```

## 4. Mock Robot / Robot Gateway에 입력할 값

```yaml
serverUrl: http://192.168.20.26:8080
robotCode: robot-001
robotToken: rb_poc_xxxxx
```

Mock Robot 또는 실제 Robot Gateway는 `serverUrl`을 기준으로 REST API를 호출하고, mission 응답에 포함된 SFU/TURN 설정으로 WebRTC publish를 시작한다. 같은 mission에 여러 Robot이 배정되더라도 각 Robot Gateway 인스턴스는 자기 `robotCode`와 `robotToken`으로 개별 실행한다.

P0 UI에서는 이 값을 복사하기 쉬운 형태로 표시한다.

## 5. 인증

Robot Gateway API는 bearer token을 사용한다.

```http
Authorization: Bearer {robotToken}
```

P0 token 정책:

- 로봇 생성 시 1회 발급
- UI에서 token 확인 가능
- token rotation/revoke는 P1

## 6. REST API

### 6.1 Heartbeat

로봇이 online 상태와 기본 상태를 보고한다.

```http
POST /api/robot-gateway/heartbeat
Authorization: Bearer {robotToken}
Content-Type: application/json
```

요청:

```json
{
  "robotCode": "robot-001",
  "state": "online",
  "batteryPercent": 82,
  "networkQuality": "good",
  "sentAt": "2026-05-18T08:00:00.000Z"
}
```

응답:

```json
{
  "robotId": "9d3c1e5d-0c41-4e4f-a21f-8b69f7c0a001",
  "robotCode": "robot-001",
  "status": "online",
  "serverTime": "2026-05-18T08:00:00.120Z"
}
```

### 6.2 Mission 조회

로봇이 현재 수행할 임무와 WebRTC 연결 정보를 조회한다.

```http
GET /api/robot-gateway/mission?robotCode=robot-001
Authorization: Bearer {robotToken}
```

active mission이 있을 때:

```json
{
  "missionId": "9d3c1e5d-0c41-4e4f-a21f-8b69f7c0a001",
  "missionCode": "mission-001",
  "missionStatus": "active",
  "roomId": "mission-001",
  "sfu": {
    "signalingUrl": "ws://192.168.20.26:8081/ws?room=mission-001&role=robot&robotCode=robot-001",
    "iceTransportPolicy": "relay"
  },
  "turnServers": [
    {
      "urls": ["turn:192.168.20.26:3478?transport=udp"],
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

Mission 단위 multi-robot 구조에서 `roomId`는 `missionCode`와 같아야 한다. `robotCode`는 room id에 합치지 않고 payload, status, recording metadata에서 별도로 유지한다. `missionId`는 DB UUID이고, `missionCode`는 사람이 읽는 코드이자 SFU room id다.

active mission이 없을 때:

```json
{
  "missionId": null,
  "missionStatus": "none"
}
```

### 6.3 Streaming Status

로봇이 WebRTC publish 상태와 실제 송출 스펙을 보고한다.

```http
POST /api/robot-gateway/streaming-status
Authorization: Bearer {robotToken}
Content-Type: application/json
```

요청 예시:

```json
{
  "robotCode": "robot-001",
  "missionId": "9d3c1e5d-0c41-4e4f-a21f-8b69f7c0a001",
  "roomId": "mission-001",
  "status": "streaming",
  "publishedTracks": [
    {
      "name": "track.video_1",
      "displayName": "RGB",
      "kind": "video",
      "codec": "h264",
      "width": 1280,
      "height": 720,
      "fps": 30,
      "bitrateKbps": 2500
    },
    {
      "name": "track.video_2",
      "displayName": "Thermal",
      "kind": "video",
      "codec": "h264",
      "width": 640,
      "height": 480,
      "fps": 15,
      "bitrateKbps": 800
    },
    {
      "name": "track.audio_1",
      "displayName": "Audio",
      "kind": "audio",
      "codec": "opus"
    }
  ],
  "publishedDataChannels": ["channel.telemetry", "channel.spatial", "channel.event", "channel.control"],
  "sentAt": "2026-05-18T08:01:00.000Z"
}
```

`publishedTracks`와 `publishedDataChannels`는 실제 mock/robot이 송출한 값을 보고하는 metadata다. 슬롯명은 `track.video_1`, `channel.telemetry` 같은 canonical role을 사용하고, 화면 표시 의미는 `displayName` 등 metadata로 분리한다.

Streaming status 수락 규칙:

- `status=streaming` 또는 `publishing`은 `missionId`가 active mission이고 해당 robot이 active assignment일 때만 수락한다.
- `roomId`는 반드시 mission 조회 응답의 `roomId`, 즉 `missionCode`와 같아야 한다.
- 이전 호환 room 형식인 `missionCode__robotCode`는 streaming status와 SFU publish 기준으로 사용하지 않는다.
- 서버는 로봇이 보낸 `sentAt`을 metadata로 보존하지만, 송출 freshness와 임무 생성 차단 판단은 서버가 status를 받은 `updatedAt` 기준 30초 window로 판단한다.
- `stopped`, `failed`, `stale` 같은 종료성 status는 현재 저장된 robot streaming row의 `missionId + roomId`가 보고값과 같을 때만 반영한다. 늦게 도착한 이전 임무의 stop 보고가 새 임무 송출 상태를 덮어쓰면 안 된다.

응답:

```json
{
  "accepted": true,
  "serverTime": "2026-05-18T08:01:00.100Z"
}
```

### 6.4 Control ACK

P1에서 구현한다. P0에서는 API/DB 모델만 준비할 수 있다.

```http
POST /api/robot-gateway/control-acks
Authorization: Bearer {robotToken}
```

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

P0 mock 예시 track:

| Slot | Codec | Display Metadata Example | Required |
| --- | --- | --- | --- |
| track.video_1 | H.264 | RGB / front camera | Yes |
| track.video_2 | H.264 | Thermal / synthetic thermal | Yes |
| track.audio_1 | Opus | microphone | Optional |

## 8. DataChannel

DataChannel은 역할별로 분리한다.

```text
channel.telemetry  descriptor/sample stream
channel.spatial    point cloud/space status or object reference
channel.event      alarm/fault/detection/mission event stream
channel.control    command ack/control side channel
```

DataChannel 메시지는 공통 envelope를 사용한다. 세부 payload schema는 아직 확정하지 않는다.

```json
{
  "messageId": "uuid",
  "schemaVersion": "1.0",
  "messageType": "telemetry.robotStatus",
  "robotCode": "robot-001",
  "missionId": "mission-001",
  "sequence": 1,
  "sentAt": "2026-05-18T08:00:00.000Z",
  "payload": {}
}
```

`robotCode`는 같은 mission room 안에서 Robot별 telemetry/event/spatial/control을 구분하는 필수 식별자이므로 유지해야 한다.

### 8.1 Telemetry

P0 telemetry는 `SensorDescriptor`와 `SensorSample` 개념을 따른다. 고정 필드만 전제로 하지 않고, UI는 descriptor를 보고 동적으로 렌더링할 수 있어야 한다.

```json
{
  "messageId": "uuid",
  "messageType": "telemetry",
  "channelRole": "channel.telemetry",
  "robotCode": "robot-001",
  "missionId": "mission-001",
  "sequence": 102,
  "sentAt": "2026-05-18T08:00:00.000Z",
  "descriptors": [
    {
      "sensorId": "telemetry.position_1",
      "kind": "position",
      "displayName": "GPS",
      "samplingRate": 1,
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

송신 주기:

- P0 기본 1Hz
- 실제 로봇 연동 시 협의

### 8.2 Spatial / Event / Control

- `channel.spatial`: 기본 자동 표시 대상이 아니다. `available`, `subscribed`, `paused`, `unsupported` 같은 상태를 먼저 표현한다.
- `channel.event`: telemetry와 별도 경로다. alarm/fault/detection/mission event처럼 발생 순서가 중요한 메시지를 보낸다.
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

`robots.status`는 장치 online/offline/fault 성격의 상태다. 임무 배정 상태는 `mission_robots.status`, WebRTC 송출 상태는 `streaming_statuses.status`에서 판단한다.

## 10. 재시도 정책

Mock Robot / Robot Gateway 기본 재시도:

- heartbeat 실패: 2초, 5초, 10초 backoff
- mission 조회 실패: 2초, 5초, 10초 backoff
- signaling 끊김: mission이 active이면 재접속
- ICE failed: PeerConnection 재생성
- publish 실패: streaming status를 `failed`로 보고

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
