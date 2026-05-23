# Appendix. Robot Gateway Interface

## 1. 문서 목적

Android Mock Robot과 향후 실제 Robot Gateway가 관제센터에 연결하기 위한 최소 API, WebRTC, DataChannel 계약을 정의한다.

P0에서는 Android Mock Robot을 로봇 샘플로 사용한다.

이 문서의 JSON과 track 값은 현재 mock/harness 연동 예시다. 확정되지 않은 DataChannel payload 세부 필드, track metadata, codec 세부 정책은 별도 schema가 생기기 전까지 고정 계약으로 간주하지 않는다.

## 2. 설계 원칙

- 로봇 등록 생성은 관제센터 UI/Backend에서 수행한다.
- Android Mock Robot은 발급받은 연결 정보를 입력받아 연결한다.
- 별도 CLI는 만들지 않는다.
- QR 등록은 P0에서 제외한다.
- gatewayVersion, capabilities, hardware fingerprint는 P0 필수값에서 제외한다.
- WebRTC 송출 스펙은 `robot_defined`로 둔다.
- 실제 송출/수신/저장된 codec, 해상도, FPS, bitrate는 status/metadata로 남긴다.

## 3. 전체 연결 흐름

```text
1. 관제 UI에서 로봇 생성
2. Backend가 robotCode, robotToken, serverUrl 발급
3. 관제 UI가 Android Mock 입력값을 표시
4. Android Mock Robot에 값 입력
5. Android Mock Robot이 heartbeat 호출
6. Backend가 로봇 online 표시
7. Android Mock Robot이 mission 조회
8. active mission이면 mission room으로 SFU publish 시작
9. Android Mock Robot이 streaming status 보고
10. Browser와 Recorder가 SFU subscriber로 수신
```

## 4. Android Mock에 입력할 값

```yaml
serverUrl: http://192.168.20.26:8080
robotCode: robot-001
robotToken: rb_poc_xxxxx
```

Android Mock Robot은 `serverUrl`을 기준으로 REST API를 호출하고, mission 응답에 포함된 SFU/TURN 설정으로 WebRTC publish를 시작한다. 같은 mission에 여러 Robot이 배정되더라도 각 Android Mock은 자기 `robotCode`와 `robotToken`으로 개별 실행한다.

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
  "missionId": "mission-001",
  "missionCode": "mission-001",
  "missionStatus": "active",
  "roomId": "mission-001",
  "sfu": {
    "signalingUrl": "ws://192.168.20.26:8081/ws?room=mission-001&role=robot",
    "iceTransportPolicy": "relay"
  },
  "turnServers": [
    {
      "urls": ["turn:192.168.20.26:3478?transport=udp"],
      "username": "robot",
      "credential": "robot-pass"
    }
  ],
  "tracks": ["rgb", "thermal", "audio"],
  "dataChannels": ["sensor", "telemetry"],
  "videoPolicy": {
    "mode": "robot_defined"
  }
}
```

Mission 단위 multi-robot 구조에서 `roomId`는 `missionCode`와 같아야 한다. `robotCode`는 room id에 합치지 않고 payload, status, recording metadata에서 별도로 유지한다.

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
  "missionId": "mission-001",
  "roomId": "mission-001",
  "status": "streaming",
  "publishedTracks": [
    {
      "name": "rgb",
      "kind": "video",
      "codec": "h264",
      "width": 1280,
      "height": 720,
      "fps": 30,
      "bitrateKbps": 2500
    },
    {
      "name": "thermal",
      "kind": "video",
      "codec": "h264",
      "width": 640,
      "height": 480,
      "fps": 15,
      "bitrateKbps": 800
    },
    {
      "name": "audio",
      "kind": "audio",
      "codec": "opus"
    }
  ],
  "publishedDataChannels": ["sensor", "telemetry"],
  "sentAt": "2026-05-18T08:01:00.000Z"
}
```

`publishedTracks`와 `publishedDataChannels`는 실제 mock/robot이 송출한 값을 보고하는 metadata다. 이 문서의 값은 예시이며, 세부 codec/해상도/FPS 정책은 별도 검증 기준에서 확정한다.

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

Android Mock Robot은 active mission 수신 후 mission room에 SFU publish한다.

```text
Android Mock Robot
  -> SFU signalingUrl 접속
  -> offer 생성
  -> RGB/Thermal/Audio track 추가
  -> sensor/telemetry DataChannel 생성
  -> relay ICE candidate 사용
  -> publish 시작
```

P0 mock 예시 track:

| Track | Codec | Source | Required |
| --- | --- | --- | --- |
| rgb | H.264 | Android rear/front camera | Yes |
| thermal | H.264 | synthetic thermal | Yes |
| audio | Opus | Android microphone | Optional |

## 8. DataChannel

DataChannel 메시지는 공통 envelope를 사용한다.

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

`robotCode`는 같은 mission room 안에서 Robot별 telemetry/sensor를 구분하는 필수 식별자이므로 유지해야 한다.

### 8.1 Telemetry

P0 telemetry는 GPS 위치와 로봇 상태를 포함한다.

```json
{
  "messageId": "uuid",
  "schemaVersion": "1.0",
  "messageType": "telemetry.robotStatus",
  "robotCode": "robot-001",
  "missionId": "mission-001",
  "sequence": 102,
  "sentAt": "2026-05-18T08:00:00.000Z",
  "payload": {
    "batteryPercent": 82,
    "networkQuality": "good",
    "position": {
      "coordinateType": "gps",
      "latitude": 37.402183,
      "longitude": 127.106812,
      "altitudeMeter": 42.1,
      "accuracyMeter": 3.5,
      "headingDegree": 93.4
    }
  }
}
```

송신 주기:

- P0 기본 1Hz
- 실제 로봇 연동 시 협의

### 8.2 Sensor

P0 sensor는 환경 센서 샘플을 포함한다.

```json
{
  "messageId": "uuid",
  "schemaVersion": "1.0",
  "messageType": "sensor.environment",
  "robotCode": "robot-001",
  "missionId": "mission-001",
  "sequence": 103,
  "sentAt": "2026-05-18T08:00:01.000Z",
  "payload": {
    "temperatureCelsius": 24.5,
    "humidityPercent": 48.2,
    "oxygenPercent": 20.8,
    "coPpm": 3.1,
    "ch4Ppm": 0.0
  }
}
```

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
| `online` | heartbeat 성공, mission 대기 |
| `assigned` | active 또는 ready mission에 배정됨 |
| `streaming` | WebRTC publish 중 |
| `reconnecting` | 재접속 중 |
| `fault` | 오류 상태 |

## 10. 재시도 정책

Android Mock Robot 기본 재시도:

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
