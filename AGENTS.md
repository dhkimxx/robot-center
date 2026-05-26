# AGENTS.md

이 문서는 `robot-center` 프로젝트에서 AI 코딩 에이전트가 따라야 할 작업 규칙이다.

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
