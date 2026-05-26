---
title: "server-architecture"
created: 2026-05-22
updated: '2026-05-26'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>", "dhkimxx <dhkimxx@naver.com>"]
type: "design"
status: "stable"
tags: ["harness", "server", "go", "postgresql", "minio"]
history:
- "2026-05-22 danya.kim <danya.kim@thundersoft.com>: initial entry"
- '2026-05-22 danya.kim <danya.kim@thundersoft.com>: migrated server architecture harness into meta-docs managed structure'
- '2026-05-22 danya.kim <danya.kim@thundersoft.com>: specified GORM as the PostgreSQL ORM'
- '2026-05-22 danya.kim <danya.kim@thundersoft.com>: aligned recorder metadata flow and JSONL recording status with current implementation'
- '2026-05-22 dhkimxx <dhkimxx@naver.com>: aligned storage_objects, relay-only failure behavior, and staged GORM transition'
- '2026-05-23 dhkimxx <dhkimxx@naver.com>: aligned WebRTC room flow with mission-level multi-robot publishers'
- '2026-05-26 danya.kim <danya.kim@thundersoft.com>: updated PostgreSQL storage targets to current sensor descriptor/sample tables'
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: moved into docs/stable/harness lifecycle structure"
---
# Server Architecture Harness

이 문서는 서버 쪽 기본 아키텍처만 정의한다.

구체적인 REST JSON 포맷, DB schema, MinIO object key, WebRTC signaling message 포맷은 이 문서에서 정의하지 않는다.

## 기술 기준

| 영역 | 기준 |
| --- | --- |
| Backend language | Go |
| API server | app-server |
| Recording worker | recorder-worker |
| Relational database | PostgreSQL |
| ORM | GORM |
| Object storage | MinIO |
| WebRTC role | SFU signaling/media 책임은 app-server 쪽 서버 책임으로 둔다 |

## 서버 구성

```text
Browser UI
-> app-server
-> PostgreSQL
-> MinIO

Android/Python Mock Robots
-> app-server
-> SFU room

recorder-worker
-> app-server / SFU room
-> MinIO
-> app-server API
-> PostgreSQL metadata
```

## 구성요소 책임

### app-server

`app-server`는 P0 서버의 중심 프로세스다.

책임:

- REST API 제공
- Web UI static serving
- robot 등록과 연결 정보 발급
- mission 생성, 시작, 종료 관리
- robot gateway heartbeat와 active mission 조회 처리
- WebRTC signaling endpoint 제공
- SFU room/session 상태 관리
- Browser subscriber 연결 관리
- recorder-worker subscriber 연결 관리
- telemetry/sensor 최신 상태 조회와 저장 API 제공
- recording session/chunk/file 상태 조회 API 제공
- PostgreSQL repository 접근
- MinIO object URL 또는 presigned URL 발급 책임

비책임:

- recorder media muxing
- 장기 저장 파일 생성
- MinIO 자체 저장 엔진 역할
- PostgreSQL 자체 저장 엔진 역할

### recorder-worker

`recorder-worker`는 저장 전용 서버 프로세스다.

책임:

- active mission의 recording target 확인
- SFU room에 recorder subscriber로 연결
- media/data 흐름 수신
- recording chunk lifecycle 처리
- media 파일 생성
- data 파일 생성
- MinIO 업로드
- app-server API를 통한 recording metadata 갱신
- 저장 실패 상태 보고

비책임:

- Web UI serving
- 사용자 facing REST API 제공
- robot 등록 관리
- mission 생성/시작/종료 결정

### PostgreSQL

PostgreSQL은 구조화된 운영 데이터의 기준 저장소다.

저장 대상:

- users
- robots
- robot tokens
- missions
- mission robots
- robot sessions
- browser sessions
- recorder sessions
- streaming statuses
- sensor descriptors
- sensor samples
- recording sessions
- recording chunks
- storage object metadata
- events
- control commands and ACKs

책임:

- API 조회 기준 데이터 저장
- recording 상태와 metadata 저장
- replay 조회에 필요한 metadata 저장
- 운영 이벤트와 제어 감사 기록 저장

비책임:

- MP4, JSONL, manifest 같은 파일 본문 저장

### MinIO

MinIO는 파일성 artifact의 object storage다.

저장 대상:

- RGB/Audio media file
- Thermal media file
- sensor data file
- telemetry data file
- manifest file
- event snapshot
- generated report artifact

책임:

- large object 저장
- recorder-worker upload target
- replay/download file source

비책임:

- mission 상태 판단
- robot 상태 판단
- recording metadata의 source of truth
- 사용자 권한 판단

## 서버 내부 레이어 기준

Go 서버는 다음 방향을 따른다.

```text
Controller / Handler
-> Service
-> Repository
-> GORM / MinIO client
-> PostgreSQL / MinIO
```

규칙:

- Handler는 request parsing과 response 변환 중심으로 둔다.
- Service는 비즈니스 흐름과 트랜잭션 경계를 관리한다.
- Repository는 저장소 접근만 담당한다.
- PostgreSQL 접근은 GORM을 기준으로 한다.
- 현재 P0 구현은 GORM 연결을 기준으로 열고, 일부 PostgreSQL/PostGIS 전용 조회는 raw SQL을 단계적으로 유지할 수 있다.
- GORM model은 persistence concern으로 두고 API response DTO와 분리한다.
- API 응답은 entity를 직접 반환하지 않고 response DTO로 변환한다.
- 외부 저장소 호출은 timeout과 명시적 에러 처리를 둔다.

## 데이터 흐름

### Control plane

```text
Browser UI
-> app-server REST API
-> Service
-> PostgreSQL
```

대상:

- robot 등록
- mission 관리
- recording 상태 조회
- system status 조회

### Robot gateway

```text
Android/Python Mock Robots
-> app-server robot gateway API
-> PostgreSQL
```

대상:

- heartbeat
- active mission 조회
- streaming status 보고

### WebRTC room

```text
Android/Python Mock Robots
-> app-server / mission SFU room
-> Browser subscribers
-> recorder-worker subscriber
```

서버 책임:

- missionCode 단위 room 관리
- publisher/subscriber 세션 관리
- signaling 연결 관리
- 같은 mission room의 robot publisher들을 `robotCode` 기준으로 구분
- Browser A/B와 Recorder subscriber의 독립성 유지

### Recording

```text
recorder-worker
-> app-server / mission SFU room subscribe
-> local chunk artifact
-> MinIO upload
-> app-server API
-> PostgreSQL metadata update
```

저장 기준:

- 파일 본문은 MinIO에 저장한다.
- 파일 metadata와 recording 상태는 PostgreSQL에 저장한다.
- MinIO object 단위 metadata의 기준 row는 `storage_objects`다.
- 같은 mission room 안에서도 recording chunk/file은 robotCode별로 생성하고 조회할 수 있어야 한다.
- `recording_chunks`는 chunk lifecycle과 manifest 연결을 관리한다.
- `recording_chunks.metadata`는 UI 호환과 chunk manifest 생성을 위한 요약 캐시로만 사용한다.
- manifest 파일이 업로드되면 `recording_chunks.manifest_object_id`가 `storage_objects.id`를 참조한다.
- recorder-worker는 PostgreSQL을 직접 갱신하지 않고 app-server API로 metadata 갱신을 요청한다.
- 현재 P0 구현은 sensor/telemetry DataChannel payload를 chunk 단위 JSONL snapshot으로 저장하고 MinIO에 업로드한다.
- UI는 MinIO object key를 직접 조합하지 않는다.
- app-server가 조회 결과와 파일 URL을 제공한다.

## 장애 격리 기준

| 장애 | 기대 동작 |
| --- | --- |
| Browser 연결 실패 | 해당 Browser subscriber만 재시도 |
| recorder-worker 실패 | 저장만 중단되고 Browser live 관제는 유지 |
| Robot publish 실패 | mission room의 media/data 수신 중단 |
| MinIO 업로드 실패 | recording chunk/file 상태를 실패로 기록 |
| PostgreSQL 실패 | API write/read 실패를 명시적 에러로 반환 |
| TURN 실패 | relay-only 정책에서는 WebRTC 연결 실패로 보고 재시도 또는 장애 처리 |

## 이 문서에서 정의하지 않는 항목

아래 항목은 이 server harness 문서의 범위 밖이다.

- REST endpoint별 request/response JSON
- WebRTC signaling message JSON
- DataChannel payload JSON
- media track 이름과 codec 정책
- PostgreSQL DDL 상세
- MinIO bucket policy
- MinIO object key 상세 규칙
- presigned URL 만료 시간 정책
- recorder chunk 상세 파일 포맷
- replay manifest schema

필요하면 각 항목은 별도 harness 문서나 schema로 분리해 정의한다.
