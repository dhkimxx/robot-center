---
title: robot-legacy-schema-cleanup-plan
created: '2026-05-27'
updated: '2026-05-27'
author: danya.kim <danya.kim@thundersoft.com>
editors:
- danya.kim <danya.kim@thundersoft.com>
type: roadmap
tags:
- robot
- legacy
- schema
- sensor
- webrtc
- cleanup
history:
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: 최초 작성'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: expand plan for legacy robot data schema cleanup'
---

# Robot Legacy Schema Cleanup Plan

## 1. 목적

로봇팀에 아직 외부 스펙을 공유하기 전이므로, 관제 서버 안에 남아 있는 legacy 호환 데이터 구조와 schema fallback을 정리한다.

목표는 로봇팀이 구현할 외부 계약을 하나의 canonical 구조로 고정하고, 이후 운영화 인증/저장/관제 UI 작업에서 legacy 분기 때문에 생기는 혼선을 줄이는 것이다.

## 2. Canonical 계약

로봇팀에 공유하는 신규 계약은 다음만 허용한다.

| 영역 | Canonical |
| --- | --- |
| REST identity | `Authorization: Bearer {robotToken}` |
| Mission lookup | `GET /api/robot-gateway/mission` |
| Robot WebSocket | `GET /sfu/robot/ws?room={missionCode}` + Bearer robot token |
| Media track slot | `track.video_1`, `track.video_2`, `track.audio_1`, `track.audio_2` |
| DataChannel label | `channel.telemetry`, `channel.spatial`, `channel.event`, `channel.control` |
| Sensor envelope | `messageId`, `messageType`, `sequence`, `sentAt`, `descriptors`, `samples` |
| Sensor descriptor | `sensorId`, `displayName`, `sensorType`, `valueType`, `unit`, `sampleRateHz`, `enabled`, `metadata` |
| Sensor sample | `sensorId`, `sequence`, `timestamp`, `sentAt`, `quality`, `values`, `objectKey` |

## 3. 정리 대상

### 3.1 REST identity 전환 경로

현재 정리 대상:

- heartbeat/mission 요청에서 token 외 `robotCode`를 같이 보내는 전환 경로
- 에러 메시지와 테스트에 남아 있는 "전환용 robotCode" 설명

목표:

- 로봇 요청 identity는 Bearer token 하나로만 결정한다.
- `robotCode`는 로봇팀 설정값, 응답값, 로그/화면 표시값으로만 사용한다.

확인 파일:

- `apps/server/internal/api/robot_gateway_handlers.go`
- `apps/server/internal/api/control_plane_flow_test.go`
- `docs/stable/robot-interface.md`

### 3.2 WebRTC publisher identity fallback

현재 정리 대상:

- robot WebSocket peer에 이미 robotCode가 있는데도 offer payload의 `robotCode`를 fallback으로 읽는 흐름

목표:

- `/sfu/robot/ws?room={missionCode}` handshake의 Bearer token으로 robotCode를 확정한다.
- offer payload에는 SDP만 둔다.

확인 파일:

- `apps/server/internal/sfu/publisher_session.go`
- `apps/server/internal/sfu/publisher_orchestration.go`
- `apps/server/internal/api/sfu_handlers.go`

### 3.3 Media track role fallback

현재 정리 대상:

- track id/stream id에 `rgb`, `thermal`, `audio` 같은 semantic keyword가 있을 때 canonical slot으로 추론하는 fallback
- media kind 순서만 보고 `track.video_1`, `track.video_2`를 배정하는 fallback

목표:

- robot publisher는 track id 또는 stream id에서 canonical slot을 명시한다.
- 서버는 canonical slot이 없는 track을 reject하거나 `unknown`으로 격리한다.

확인 파일:

- `apps/server/internal/sfu/stream_roles.go`
- `apps/server/internal/recording/subscriber_helpers.go`
- `apps/server/internal/recording/media_track_writer.go`

### 3.4 DataChannel label fallback

현재 정리 대상:

- `telemetry`, `sensor`, `spatial`, `event`, `control` 같은 짧은 label을 `channel.*`로 normalize하는 fallback
- recorder가 `sensor`/`telemetry` alias를 저장 대상으로 받는 흐름

목표:

- robot publisher는 `channel.*` label만 생성한다.
- canonical label이 아니면 app-server/recorder에서 명시적으로 reject하거나 저장 대상에서 제외한다.

확인 파일:

- `apps/server/internal/sfu/stream_roles.go`
- `apps/server/internal/recording/subscriber_helpers.go`
- `apps/server/internal/recording/data_channel_queue.go`
- `apps/server/internal/recording/app_server_client.go`

### 3.5 Sensor envelope legacy payload

현재 정리 대상:

- `payload`만 들어온 메시지를 `legacy.payload_1` sample로 저장하는 fallback
- `legacy.payload_1`, `sensorType=legacy` 저장 데이터

목표:

- sensor 저장은 `descriptors`/`samples`만 허용한다.
- 최소 telemetry도 descriptor/sample 구조를 사용한다.

확인 파일:

- `apps/server/internal/api/sensor_handlers.go`
- `apps/server/internal/store/postgres/sensor_repository.go`
- `apps/server/internal/recording/data_channel_queue.go`

### 3.6 Sensor field alias와 typed sample 혼재

현재 정리 대상:

- descriptor alias: `kind`, `samplingRate`
- sample value variants: `numericValue`, `textValue`, `boolValue`, `vectorValue`, `objectValue`, `rawPayload`
- `values`를 내부 저장용 typed field로 변환하는 과정이 문서마다 다르게 노출되는 문제

목표:

- 외부 로봇 계약은 `values` 하나로 단순화한다.
- 내부 DB DTO는 필요하면 유지하되 robot-facing schema에는 노출하지 않는다.
- alias 제거 전 기존 mock/client가 canonical field를 보내도록 먼저 수정한다.

확인 파일:

- `apps/server/internal/api/sensor_handlers.go`
- `apps/server/internal/api/dto/sensor_dto.go`
- `apps/android-robot`
- `apps/mock-robot` 또는 Python mock robot scripts
- `docs/stable/robot-interface.md`

## 4. 작업 순서

### Phase 1. 문서 분리

- 로봇팀 공유 문서에는 canonical 계약만 남긴다.
- legacy/fallback 설명은 이 계획 문서와 stable 내부 문서에만 둔다.
- `docs/stable/robot-interface.md`는 현재 코드 상태를 설명하되, "robot-facing canonical"과 "server compatibility"를 명확히 분리한다.

완료 기준:

- `docs/planned/robot-team-webrtc-send-test-guide.md`에서 legacy/fallback/alias 기반 구현 가이드가 사라진다.

### Phase 2. Mock client canonical화

- Android Robot app이 canonical track id, DataChannel label, descriptor/sample schema를 송신하는지 확인한다.
- Python Mock Robot도 동일 schema로 맞춘다.
- mock이 `payload`, `kind`, `samplingRate`, 짧은 DataChannel label을 보내지 않게 한다.

완료 기준:

- Android 2대와 Python mock 모두 canonical schema로 WebRTC/sensor 송신 성공.
- `sensor-latest`에 `legacy.payload_1` 신규 데이터가 생성되지 않는다.

### Phase 3. Server warning 계층 추가

- fallback이 사용되면 app-server/recorder log에 warning을 남긴다.
- fallback 사용 횟수를 health/status에서 확인할 수 있게 할지 검토한다.

완료 기준:

- canonical mock 테스트 중 fallback warning이 0이다.

### Phase 4. Fallback reject 전환

- DataChannel label fallback을 제거하거나 reject한다.
- media track role fallback을 제거하거나 `unknown` 격리로 바꾼다.
- sensor `payload` fallback과 `legacy.payload_1` 생성을 제거한다.
- REST/offer payload의 robotCode fallback을 제거한다.

완료 기준:

- canonical Android 2대 테스트 통과.
- 기존 fallback 입력은 명시적인 400, publish-error, 저장 제외 중 하나로 처리된다.

### Phase 5. DB/데이터 정리

- 기존 `legacy.payload_1` sample을 유지할지 삭제할지 결정한다.
- 삭제가 필요하면 migration 또는 admin cleanup script로 분리한다.
- 데모/테스트 데이터는 재생성 가능한 fixture로 남긴다.

완료 기준:

- 신규 테스트 데이터에 legacy sensor id가 없다.
- 기존 데이터 처리 방침이 문서화된다.

## 5. 검증 기준

필수 검증:

- `go test ./...`
- Android Robot app 2대 relay ICE connected/completed
- Python Mock Robot canonical schema 송신
- `/api/missions/mission-001/live-status`에서 robot 2대 `stream.state=streaming`
- recorder-worker `iceState=connected`, `appendFailedCount=0`
- `sensor-latest`에 canonical `sensorId`와 `values` 기반 sample 표시
- 신규 recording artifact 생성

회귀 확인:

- `/sfu/robot/ws?room={missionCode}`는 robot token 없으면 거부
- `/sfu/operator/ws`, `/sfu/recorder/ws`는 robotCode query를 받지 않음
- 짧은 DataChannel label과 `payload` only sensor body는 계획된 방식으로 거부 또는 저장 제외

## 6. 영향 범위

관제팀 영향:

- app-server SFU role normalization
- recorder-worker DataChannel 저장
- sensor API DTO와 저장 모델
- Android/Python mock client
- 관제 UI의 sensor-latest 표시
- 테스트 fixture와 문서

로봇팀 영향:

- 아직 외부 스펙 공유 전이면 영향 없음
- 공유 후 변경하면 robot publisher 구현 변경이 필요하므로, Phase 4 이후에는 robot-facing schema를 고정해야 한다

실제 로봇 코드베이스는 이 프로젝트 범위가 아니다. 이 문서는 관제 서버와 관제팀 테스트 client의 정리 계획만 다룬다.
