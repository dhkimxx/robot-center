# AI Web Product Requirements

## 1. 문서 목적

AI Web은 구조/재난 현장에서 로봇이 수집한 영상, 음성, 센서, 위치 데이터를 관제센터에서 실시간으로 확인하고, 임무 운영, 녹화, 이력 조회, 제어 승인, AI Agent 보조 판단을 수행하는 웹 기반 관제 시스템이다.

이 문서는 기존 WebRTC 아이디어 검증 PoC를 넘어, Android Mock Robot을 로봇 샘플로 사용하는 전체 시스템 P0 구현 범위를 정의한다.

기존 WebRTC PoC 소스는 제거하고, 검증 결과 문서만 이력으로 유지한다. 제품형 애플리케이션은 새 구조로 재작성한다.

## 2. 목표

P0 목표:

- Android Mock Robot을 실제 Robot Gateway 샘플로 취급한다.
- 로봇 등록, heartbeat, mission 조회, WebRTC publish까지 하나의 연결 플로우로 동작한다.
- React 관제 UI에서 로봇 등록, 임무 생성, 실시간 관제, 녹화 상태, 저장 이력을 확인한다.
- app-server가 로봇/임무/녹화/제어/AI Agent API와 SFU의 기준점이 된다.
- app-server/SFU는 Robot publish once 구조를 유지하고 Browser와 recorder-worker로 fan-out한다.
- recorder-worker는 SFU subscriber로 붙어 MP4/JSONL을 만들고 MinIO와 PostgreSQL에 저장 결과를 남긴다.
- 서버 측 구성요소는 Docker Compose로 `up/down` 가능해야 한다.

## 3. 비목표

P0에서 제외한다.

- 실제 Jetson/ROS/GStreamer 연동
- 실제 소방 SOP 전체 RAG 구축
- 완성형 VMS 검색/재생 시스템
- 다중 기관 사용자 관리, SSO, MFA
- 장기 보관/백업/복구 자동화
- 고도화된 SLAM/point cloud viewer
- 실제 로봇 자율주행 제어
- Android 앱의 배포/업데이트 체계

단, 데이터 모델과 API는 이후 확장이 가능하게 둔다.

## 4. 사용자 역할

| 역할 | 설명 |
| --- | --- |
| 관제요원 | 임무 생성, 로봇 모니터링, 녹화 확인, 제어 요청을 수행한다. |
| 지휘관 | 상황 요약, 이벤트, 위험도, SOP 제안을 확인하고 대응 판단을 내린다. |
| 관리자 | 로봇 등록 정보, 사용자, 시스템 설정을 관리한다. |
| Robot Gateway | Android Mock 또는 실제 로봇 측 클라이언트. 관제센터 API와 WebRTC에 접속한다. |

## 5. P0 시스템 구성

```text
Android Mock Robot
  -> app-server REST API
       - robot registration
       - heartbeat
       - mission polling
       - streaming status
  -> app-server/SFU WebRTC publish
       - rgb h264
       - thermal h264
       - audio opus
       - sensor DataChannel
       - telemetry DataChannel

React Operator UI
  -> app-server REST API
       - robots
       - missions
       - recordings
       - events
       - control commands
  -> app-server/SFU WebRTC subscribe
       - live media
       - live DataChannel

recorder-worker
  -> app-server/SFU WebRTC subscribe
  -> MP4 muxing
  -> MinIO upload
  -> PostgreSQL metadata write

Docker Compose
  -> app-server
       - Go REST API
       - SFU signaling/media
       - Web static serving
  -> recorder-worker
  -> turn
  -> postgres/postgis
  -> minio
```

## 6. 핵심 플로우

### 6.1 로봇 등록

```text
1. 관제요원이 UI에서 로봇을 생성한다.
2. app-server가 robotCode, robotToken, serverUrl을 발급한다.
3. UI는 Android Mock Robot에 입력할 연결 정보를 표시한다.
4. Android Mock Robot이 연결 정보를 입력받는다.
5. Android Mock Robot이 heartbeat를 호출한다.
6. app-server가 로봇을 online으로 표시한다.
```

P0에서는 별도 CLI, QR 등록, gatewayVersion, hardware fingerprint, token rotation은 제외한다.

### 6.2 임무 생성과 시작

```text
1. 관제요원이 임무를 생성한다.
2. 임무 유형을 선택한다.
3. 등록된 로봇을 배정한다.
4. 임무를 active로 전환한다.
5. Android Mock Robot이 mission 조회에서 active mission을 받는다.
6. Android Mock Robot이 SFU에 WebRTC publish를 시작한다.
7. app-server에 streaming status를 보고한다.
```

임무 유형:

- 산악조난
- 붕괴현장
- 지하시설

### 6.3 실시간 관제

```text
1. 관제요원이 임무 목록에서 관제할 임무를 선택한다.
2. UI가 해당 임무의 로봇 목록과 streaming status를 조회한다.
3. 관제요원이 임무 안에서 로봇을 선택한다.
4. UI가 app-server/SFU room에 subscriber로 접속한다.
5. RGB, Thermal, Audio track을 표시한다.
6. Sensor, GPS telemetry DataChannel 메시지를 표시한다.
7. 지도에 로봇 위치와 이동 경로를 표시한다.
8. 이벤트와 제어 이력을 타임라인으로 표시한다.
```

실시간 센서/GPS 표시는 WebRTC DataChannel을 사용한다. Browser WebSocket은 P0 필수 경로가 아니다.

### 6.4 녹화

```text
1. 임무가 active가 되면 recorder-worker가 app-server/SFU room에 subscriber로 접속한다.
2. RGB/Thermal/Audio track과 DataChannel을 수신한다.
3. 기본 10분 단위 chunk로 녹화한다.
4. chunk 종료 시 MP4 muxing을 수행한다.
5. MinIO에 media/jsonl/manifest를 업로드한다.
6. PostgreSQL에 storage metadata를 기록한다.
7. app-server가 storage metadata를 기준으로 파일 접근 URL을 생성한다.
8. UI에서 녹화 상태와 저장 결과를 조회한다.
```

녹화 기능의 사용자 단위는 MinIO object가 아니라 `녹화 세션`이다.

녹화 세션은 다음 정보를 가진다.

- 임무
- 로봇
- 녹화 시작/종료 시각
- 총 길이
- 상태: 녹화 중, 저장 중, 완료, 실패
- chunk 목록
- 포함 데이터: RGB, Thermal, Audio, Sensor, Telemetry/GPS, Manifest

관제 UI는 녹화 세션 목록을 먼저 보여준다. 세션을 펼치면 각 chunk와 파일 목록을 보여준다.

```text
녹화 세션
  -> chunk #0
     -> RGB + Audio MP4
     -> Thermal MP4
     -> Sensor JSONL
     -> Telemetry/GPS JSONL
     -> Manifest JSON
  -> chunk #1
     -> ...
```

UI에 노출하는 기본 정보:

- 세션 시작/종료 시각
- chunk index와 시간 범위
- 파일 종류
- 저장 상태
- 열기/다운로드 액션

UI에서 기본적으로 숨기는 정보:

- bucket 이름
- object key
- MinIO Console URL
- 내부 worker/API 구현 세부사항

관제 서비스에서 파일을 열 때는 MinIO Console URL을 사용하지 않는다.

```text
UI
-> app-server recording/replay API
-> app-server가 MinIO API object URL 또는 presigned URL 반환
-> UI가 반환된 URL로 media/manifest/sensor 파일 접근
```

P0 개발 환경에서는 MinIO API object URL을 사용할 수 있다.

```text
http://{host}:9000/{bucket}/{objectKey}
```

운영 환경에서는 bucket public 정책에 의존하지 않고 app-server가 만료 시간이 있는 presigned URL을 발급하는 것을 기본 정책으로 한다.

P0 API 범위:

```text
GET /api/recordings
  -> 녹화 세션/청크 목록
  -> 각 파일의 label, type, status, contentType
  -> 접근 가능한 파일은 app-server가 생성한 url 포함

GET /api/recordings/{recordingId}
  -> P1 상세 조회 후보

GET /api/recordings/{recordingId}/replay
  -> P1 replay 타임라인 후보
```

P0에서는 RGB H.264 + Opus MP4와 Thermal H.264 MP4를 우선 재생 가능 파일로 제공한다.
MP4/JSONL 생성이 완료되기 전 파일은 UI에서 `예정` 또는 `녹화 중` 상태로 표시한다.

### 6.5 제어 명령

```text
Browser
  -> app-server API
  -> 권한 확인
  -> 감사 로그 저장
  -> command DataChannel
  -> Robot Gateway
  -> controlAck DataChannel
  -> app-server 저장
  -> Browser 표시
```

P0 명령:

- E-Stop
- Return-to-Home
- PTZ pan/tilt/zoom

AI Agent는 명령 초안을 만들 수 있지만 직접 실행하지 않는다.

## 7. 화면 범위

P0 화면:

| 화면 | 목적 |
| --- | --- |
| 로그인 | 사용자 세션 시작 |
| 로봇 관리 | 로봇 등록, 수정, 삭제, 연결 정보 확인, 토큰 재발급, heartbeat 상태 확인 |
| 임무 목록 | 임무 생성, 상태 필터, 임무별 관제 진입 |
| 임무 생성 | 시나리오, 현장 메모, 로봇 배정 |
| 임무 관제 | 임무 안 로봇 선택, 영상, 센서, GPS, 지도, 이벤트, 제어, AI 요약 |
| 녹화/기록 | 녹화 세션 목록, chunk별 파일 상태, app-server가 제공한 파일 URL 확인 |
| 이벤트 상세 | 이벤트 근거, 관련 영상/센서/제어 명령 확인 |
| 시스템 상태 | app-server, recorder-worker, TURN, DB, MinIO 상태 확인 |

상세 UI 구조는 `docs/appendix/ui-plan.md`를 따른다.

## 8. 데이터 요구사항

PostgreSQL/PostGIS:

- 사용자/권한
- 로봇 등록 정보
- 임무와 로봇 배정
- Robot heartbeat/session
- Browser/recorder-worker session
- telemetry/sensor 시계열
- GPS 위치와 경로
- 이벤트
- 제어 명령과 ACK
- 저장 object metadata

MinIO:

- RGB MP4 chunk
- Thermal MP4 chunk
- Audio OGG 또는 MP4 포함 audio
- sensor/telemetry JSONL
- recording manifest
- event snapshot
- report artifact

관제 UI는 MinIO object key나 Console URL을 직접 조합하지 않는다.

파일 접근 원칙:

```text
1. UI는 recording/replay API로 chunk와 object metadata를 조회한다.
2. app-server는 object key를 기준으로 MinIO API URL 또는 presigned URL을 생성한다.
3. UI는 app-server 응답의 `url`만 사용한다.
4. 운영 환경에서는 presigned URL 만료 시간과 권한 검사를 app-server가 책임진다.
```

상세 스키마와 object key는 `docs/appendix/data-storage.md`를 따른다.

## 9. WebRTC 요구사항

### 9.1 Room 모델

```text
1 mission = 1 SFU room
1 mission 안에 N robots
1 robot은 media/data track을 publish
Browser는 필요한 track을 subscribe
recorder-worker는 저장 대상 track을 subscribe
```

식별 기준:

```text
missionId + robotCode + trackName
```

### 9.2 Track

| Track | Direction | Required |
| --- | --- | --- |
| rgb | Robot -> app-server/SFU -> Browser/recorder-worker | Yes |
| thermal | Robot -> app-server/SFU -> Browser/recorder-worker | Yes |
| audio | Robot -> app-server/SFU -> Browser/recorder-worker | Optional |
| sensor | Robot -> app-server/SFU -> Browser/recorder-worker | Yes |
| telemetry | Robot -> app-server/SFU -> Browser/recorder-worker | Yes |
| command | app-server/SFU -> Robot | P1 |
| controlAck | Robot -> app-server/Browser | P1 |

### 9.3 영상 정책

P0는 `robot_defined` 정책을 사용한다.

```text
videoPolicy.mode = robot_defined
```

Android Mock Robot 또는 실제 Robot Gateway가 가능한 스펙으로 송출하고, 실제 송출/수신/저장 metadata를 app-server에 보고한다.

## 10. Docker Compose 요구사항

서버 측 구성요소는 아래 명령으로 제어 가능해야 한다.

```bash
docker compose up -d
docker compose down
```

P0 compose service:

- `postgres`
- `minio`
- `turn`
- `app-server`
- `recorder-worker`

Android Mock Robot은 실제 단말에서 실행하므로 compose 대상이 아니다.

## 11. 수용 기준

P0 완료 기준:

- Docker Compose로 서버 측 구성요소가 기동된다.
- UI에서 로봇을 등록하고 연결 정보를 확인할 수 있다.
- Android Mock Robot이 등록 정보로 heartbeat를 보낸다.
- UI에서 로봇 online 상태를 볼 수 있다.
- UI에서 임무를 생성하고 로봇을 배정할 수 있다.
- 임무 시작 후 Android Mock Robot이 mission 조회를 통해 WebRTC publish를 시작한다.
- UI에서 RGB/Thermal/Audio와 sensor/GPS를 볼 수 있다.
- 지도에서 GPS 위치와 이동 경로를 볼 수 있다.
- recorder-worker가 10분 chunk 정책으로 녹화하고 MinIO에 MP4를 저장한다.
- PostgreSQL에 recording metadata와 telemetry/sensor metadata가 기록된다.
- UI에서 녹화 세션 목록, chunk별 파일 상태, app-server가 제공한 파일 URL을 조회할 수 있다.
- UI는 MinIO Console URL이 아니라 MinIO API URL 또는 presigned URL로 파일을 연다.
- E-Stop 명령 요청과 ACK/NACK 흐름을 감사 로그로 확인할 수 있다.

## 12. 구현 원칙

- 기존 PoC 소스와 과거 단계별 검증 문서는 유지하지 않는다. 현재 판단 기준은 PRD, appendix, harness, 최신 rebuild 결과 문서로 둔다.
- 제품형 앱은 새 폴더/새 구조로 재작성한다.
- `app-server`는 Go REST API와 SFU module을 함께 포함하되 내부 package 책임은 분리한다.
- `recorder-worker`는 같은 Go 코드베이스를 사용하되 별도 process로 실행한다.
- React UI는 화면/기능 단위 컴포넌트로 분리한다.
- API 호출은 UI 컴포넌트 내부가 아니라 API service 모듈로 분리한다.
- DB schema는 migration으로 관리한다.
- media plane과 control plane을 분리한다.

## 13. 참조 문서

| 문서 | 내용 |
| --- | --- |
| `docs/appendix/architecture.md` | 전체 시스템 아키텍처 |
| `docs/appendix/robot-interface.md` | Android Mock/Robot Gateway 연동 계약 |
| `docs/appendix/data-storage.md` | PostgreSQL/PostGIS/MinIO 규칙 |
| `docs/appendix/ui-plan.md` | UI 화면/기능 설계 |
| `docs/appendix/ai-agent.md` | Eino 기반 AI Agent 명세 |
| `docs/appendix/scenarios.md` | 실증 시나리오 |
| `docs/appendix/decisions.md` | 설계 결정 기록 |
| `docs/harness/20260522-harness-index.md` | 서버/SFU/persistence harness 진입점 |
| `docs/rebuild-results/server-layer-gorm-storage-2026-05-22.md` | 최신 구현 검증 결과 |
