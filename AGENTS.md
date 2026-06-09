# AGENTS.md

이 문서는 `robot-center` 프로젝트에서 AI 코딩 에이전트가 따라야 할 작업 규칙이다.

## 0. 프로젝트 역할 구분과 영향 범위

이 프로젝트의 실제 제품 책임은 관제 파트다. 실제 로봇 구현은 로봇팀이 별도 프로젝트/코드베이스에서 작업하며, 이 프로젝트의 역할이 아니다.

- 관제 파트: Go `app-server`, 내부 SFU, `recorder-worker`, PostgreSQL/MinIO 저장, React 관제 UI, 운영/검증 스크립트
- 관제 검증용 Mock: `apps/android-robot`, `apps/mock-robot-python`
- 외부 로봇 파트: 실제 Robot Gateway/Publisher, 로봇 장치 쪽 WebRTC 송출, 실제 센서/미디어 송신 구현

현재 사용자는 관제 파트를 담당하며 이 프로젝트를 관리한다. 따라서 별도 지시가 없으면 AI 에이전트는 작업의 기본 범위를 관제 파트로 잡는다.

관제 파트 작업 범위:

- `apps/server`: API, service, store, SFU, recording, config
- `apps/web`: 관제 UI, live/replay/mission/robot/system 화면
- `deploy`, `scripts`: 로컬/시연 실행과 검증 자동화
- `docs`: 관제-로봇 계약, 아키텍처, 저장소, 검증 기준
- `apps/android-robot`, `apps/mock-robot-python`: 관제팀이 만든 테스트용 Mock Robot client. 실제 로봇 파트 구현으로 간주하지 않는다.

외부 로봇 파트 영향 범위는 사전에 확인한다.

다음 항목을 변경할 때는 구현 전에 영향 범위를 명시하고, 외부 실제 로봇 쪽 수정 또는 로봇팀 확인이 필요한지 먼저 판단한다.

- Robot Gateway REST API: heartbeat, mission 조회, connection-info, token 정책
- WebRTC signaling 계약: role별 `/sfu/robot/ws`, `/sfu/operator/ws`, `/sfu/recorder/ws` endpoint와 signaling message type/payload
- Room/identity 규칙: `missionId`, `missionCode`, `roomId`, `robotCode`
- Media/DataChannel 계약: `track.video_1`, `track.video_2`, `track.audio_1`, `channel.telemetry`, `channel.spatial`, `channel.event`, `channel.control`
- Sensor/telemetry payload schema와 저장/표시 의미
- TURN/ICE 설정, relay-only 정책, SFU publish/subscribe 방향
- 녹화 metadata, object key, robotCode별 chunk/file 분리 규칙

외부 로봇 파트에 영향을 주는 변경은 다음 원칙을 따른다.

- `apps/android-robot`, `apps/mock-robot-python`은 관제 검증용 mock/harness로만 수정한다.
- Mock Robot 수정이 통과하더라도 실제 로봇 호환성이 자동 보장된다고 말하지 않는다.
- 실제 로봇 구현 변경이 필요해 보이면 이 repo에서 임의로 구현하지 말고, 관제-로봇 계약 변경 사항과 로봇팀 확인 필요 항목을 명시한다.
- 공유 계약을 바꾸면 `docs/stable/robot-interface.md`, `docs/stable/architecture.md`, 관련 checklist를 함께 확인한다.
- WebRTC, 센서, 위치, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot으로 실제 관제 흐름까지 검증한다.

## 1. 백엔드 작업 규칙

백엔드는 계층 방향을 지킨다.

```text
Controller -> Service -> Repository
```

규칙:

- Controller에서 비즈니스 로직을 과도하게 처리하지 않는다.
- Entity를 API 응답으로 직접 반환하지 않는다.
- Request/Response DTO를 분리한다.
- 에러는 명시적으로 처리한다.
- 공통 에러 포맷 또는 Global Exception Filter 흐름을 따른다.
- 외부 API 호출에는 타임아웃과 필요한 경우 재시도를 둔다.
- 데이터 변경 트랜잭션 범위는 최소화한다.
- 복잡한 비즈니스 로직에는 테스트 케이스를 함께 제안하거나 작성한다.

Go 서버 작업 시:

- 기존 `internal` 패키지 경계를 존중한다.
- API handler, domain, store, recording, sfu 책임을 섞지 않는다.
- PostgreSQL 작업은 schema/migration과 store 구현을 함께 확인한다.
- WebRTC signaling, recorder, storage 흐름은 실제 실행 상태로 검증한다.

Swagger 문서 작업 시:

- API endpoint, request/response DTO, 인증 방식, status code, query/path/body parameter를 바꾸면 handler의 Swagger 주석을 함께 갱신한다.
- Swagger 설명은 한국어로 작성하고, Robot/Operator/Recorder/System 역할 기준으로 책임을 명확히 분리한다.
- `/swagger/doc.json`을 canonical 문서로 보고, `/swagger/openapi.json`은 동일 생성 문서의 호환 alias로만 취급한다.
- 수동 OpenAPI builder를 다시 만들지 않는다. 문서 소스는 handler 주석과 generated `apps/server/internal/api/swaggerdocs`로 단일화한다.
- Swagger 주석 수정 후 `./scripts/generate-server-swagger.sh`를 실행해 generated docs를 갱신한다.
- 완료 전 `./scripts/generate-server-swagger.sh --check` 또는 `GOTOOLCHAIN=go1.24.4 go test ./internal/api -run TestGeneratedSwaggerDocsAreFresh -count=1`로 생성 문서 최신성을 확인한다.

## 2. 프론트엔드 작업 규칙

프론트엔드는 실제 관제 도구처럼 동작해야 한다.

규칙:

- 컴포넌트는 작고 재사용 가능하게 나눈다.
- Global state는 최소화한다.
- Props drilling이 깊어지면 composition 또는 Context API를 검토한다.
- API 호출은 컴포넌트 안에 직접 쓰지 않고 API/service 모듈로 분리한다.
- 비동기 작업에는 Loading, Error, Empty 상태를 둔다.
- Tailwind CSS 유틸리티를 우선 사용한다.
- 복잡한 class 조합에는 `clsx`, `tailwind-merge` 사용을 검토한다.

UI 원칙:

- 관제/운영 화면은 데스크톱 앱처럼 밀도 있고 실용적으로 만든다.
- 불필요한 설명 문구를 줄이고, 상태와 조작을 우선한다.
- 화면에 PostgreSQL, MinIO, Golang 같은 구현 세부사항을 노출하지 않는다.
- 사용자가 실제 임무, 로봇, 영상, 위치, 센서, 녹화 상태를 바로 파악하게 한다.
- 텍스트가 버튼, 카드, 패널 밖으로 넘치지 않게 한다.
- 화면 변경 후 브라우저에서 직접 확인한다.

## 3. Git 규칙

- 사용자 허락 없이 commit 하지 않는다.
- 사용자 허락 없이 push 하지 않는다.
- destructive command를 사용하지 않는다.
- `git reset --hard`, `git checkout --` 등은 명시 요청 없이는 금지한다.
- 작업 트리에 사용자가 만든 변경이 있어도 되돌리지 않는다.
- 내 작업과 무관한 변경은 건드리지 않는다.

## 4. 중간 보고

긴 작업 중에는 사용자가 현재 상태를 알 수 있게 짧게 공유한다.

- 어떤 맥락을 확인 중인지
- 어떤 문제가 발견됐는지
- 다음에 무엇을 고칠지

## 5. 목표 기반 실행

작업은 검증 가능한 목표로 바꾼다.

예시:

```text
1. API 저장소를 PostgreSQL로 전환 -> verify: 재시작 후 데이터 유지 확인
2. Python Mock Robot 영상 송출 변경 -> verify: 브라우저 videoWidth/videoHeight 확인
3. 녹화 저장 구현 -> verify: MinIO object와 DB recording chunk 상태 확인
```

성공 조건이 약하면 먼저 성공 조건을 구체화한다.

## 6. 최종 사용자 검증 플로우

이 프로젝트에서 작업은 테스트 통과만으로 완료되지 않는다.

**사용자가 바로 실제 동작을 확인할 수 있어야 완료다.**

완료 보고 전 반드시 확인한다.

```text
1. Server/API
   verify: health endpoint와 system status가 OK인지 확인
   verify: API 계약 변경 시 Swagger 주석과 generated docs가 최신인지 확인

2. Web UI
   verify: http://127.0.0.1:18080 접속 가능
   verify: 화면에 기대 상태가 표시됨

3. Python Mock Robot
   verify: `./scripts/mock-robots-python.sh` 실행 완료
   verify: robotCode, token, mission room이 맞음
   verify: connected 또는 streaming 상태

4. WebRTC
   verify: robot peer가 mission room에 join
   verify: browser/operator peer가 필요한 경우 같은 room에 join
   verify: ICE connected
   verify: RGB/Thermal/Audio track 상태 확인
   verify: sensor/telemetry DataChannel open

5. 사용자 가시 결과
   verify: RGB 영상 표시
   verify: Thermal 영상 표시
   verify: GPS 위치 표시
   verify: 센서값 표시
   verify: 녹화 상태 또는 MinIO 저장 결과 표시
```

기능과 무관한 항목은 생략할 수 있지만, 로봇 연결, 영상, 센서, WebRTC 관련 작업에서는 Python Mock Robot을 반드시 포함한다.

검증이 끝나면 사용자가 바로 볼 수 있는 상태로 남긴다.

- app-server 실행 유지
- recorder-worker 실행 유지
- 필요한 Docker Compose 서비스 실행 유지
- Python Mock Robot 실행 유지
- 관제 UI URL 제공

## 7. 완료 보고 형식

완료 보고는 짧고 구체적으로 한다.

포함할 내용:

- 무엇을 바꿨는지
- 어떤 파일을 수정했는지
- 어떤 검증을 통과했는지
- 사용자가 어디서 확인하면 되는지
- 아직 남은 한계가 있는지

테스트하지 못한 항목은 성공한 것처럼 말하지 않는다.

## 8. 이 프로젝트의 현재 검증 기준

현재 PoC/시연 기준의 핵심 흐름은 다음과 같다.

```text
Python Mock Robot
-> app-server / robot gateway
-> WebRTC signaling
-> TURN relay
-> Browser 관제 UI
-> telemetry/sensor 저장
-> recorder-worker
-> PostgreSQL / MinIO
```

따라서 WebRTC, 녹화, 위치, 센서, 로봇 등록을 건드리는 작업은 이 흐름 안에서 실제로 확인해야 한다.

## 9. 배포검증 명령

사용자가 `배포검증`이라고 요청하면 현재 의도된 변경사항에 대해 다음 플로우를 수행하라는 뜻이다. 이 요청은 해당 변경사항의 commit과 push 허가를 포함한다.

기본 플로우:

```text
1. git status와 diff로 변경 범위 확인
2. 변경 범위에 맞는 로컬 테스트/빌드/문서 생성 최신성 확인
3. 의미 있는 commit message로 commit
4. 현재 branch push
5. 개발서버 Docker 배포
6. 개발서버 health, system status, Swagger, UI route, 주요 로그 확인
7. 사용자가 확인할 수 있는 URL과 남은 리스크 보고
```

반복 실행은 가능한 한 하네스를 사용한다.

```bash
./scripts/deploy-verify.sh --commit-message "Describe change"
```

이미 commit/push가 끝난 변경을 재검증할 때는 commit/push를 건너뛴다.

```bash
./scripts/deploy-verify.sh --no-commit
```

WebRTC, SFU, live-status, recorder 연동에 영향을 주는 변경은 API 기반 smoke 확인을 추가한다.

```bash
./scripts/deploy-verify.sh --no-commit --webrtc-smoke --webrtc-smoke-mission mission-054
```

관제 live 화면, 영상 pane, 로봇 선택 UI처럼 브라우저 렌더링 경로에 영향을 주는 변경은 browser smoke 확인을 추가한다.

```bash
./scripts/deploy-verify.sh --no-commit --browser-smoke --browser-smoke-mission mission-054 --browser-smoke-robot robot-042
```

제한:

- `배포검증`은 destructive git 명령, force push, production 배포, 데이터 삭제 권한을 포함하지 않는다.
- SSH password 같은 secret은 `SSHPASS` 같은 환경변수로만 주입하고, 문서/코드/commit message에 남기지 않는다.
- 하네스가 커버하지 못하는 Mock Robot 장시간 구동, 특수 센서, 녹화 파일 재생 검증은 작업 성격에 맞게 별도로 수행한다.
- 실패하면 즉시 멈추고 어느 단계에서 실패했는지 보고한다.
