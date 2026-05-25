# AGENTS.md

이 문서는 `robot-center` 프로젝트에서 AI 코딩 에이전트가 따라야 할 작업 규칙이다.

목표는 빠른 임시 구현이 아니라, 사용자가 바로 실행하고 확인할 수 있는 상태까지 책임지는 것이다.

## 1. 기본 원칙

### 선 설계, 후 코딩

코드를 수정하기 전에 반드시 먼저 공유한다.

- 변경할 파일 목록
- 구현 단계
- 검증 방법
- 불확실한 부분과 가정

요구사항이 모호하면 추측해서 구현하지 말고 질문한다.

### 단순함 우선

요청받은 범위만 구현한다.

- 필요 없는 기능을 미리 만들지 않는다.
- 한 번만 쓰는 추상화를 만들지 않는다.
- 설정 가능성이 필요하지 않으면 설정화하지 않는다.
- 200줄로 작성한 코드가 50줄로 가능하면 다시 줄인다.

### 기존 구조 존중

먼저 현재 코드의 폴더 구조, 네이밍, 계층, 스타일을 읽고 따른다.

- 기존 패턴을 우선 사용한다.
- 새로운 구조는 기존 구조로 해결하기 어려울 때만 추가한다.
- 관련 없는 리팩터링은 하지 않는다.
- 발견한 별도 문제는 임의로 고치지 말고 보고한다.

### Tidy First

기능 구현 전에 변경을 쉽게 만드는 최소한의 정리를 먼저 한다.

단, 정리는 구현과 직접 관련된 범위로 제한한다.

### Cleanup

내 변경으로 생긴 불필요한 코드는 즉시 제거한다.

- 사용하지 않는 import
- 사용하지 않는 변수
- 죽은 함수
- 의미 없는 주석
- 중복된 로직

기존에 있던 무관한 dead code는 사용자 요청 없이 삭제하지 않는다.

### SRP & DRY

파일, 함수, 컴포넌트는 하나의 책임을 갖게 한다.

반복되는 로직은 함수, 유틸리티, 서비스 모듈로 분리한다.

### 명시적 네이밍

축약어보다 의도가 드러나는 이름을 사용한다.

가능하면 `verb + noun` 형태를 사용한다.

## 2. 백엔드 작업 규칙

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

## 3. 프론트엔드 작업 규칙

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

## 4. Git 규칙

- 사용자 허락 없이 commit 하지 않는다.
- 사용자 허락 없이 push 하지 않는다.
- destructive command를 사용하지 않는다.
- `git reset --hard`, `git checkout --` 등은 명시 요청 없이는 금지한다.
- 작업 트리에 사용자가 만든 변경이 있어도 되돌리지 않는다.
- 내 작업과 무관한 변경은 건드리지 않는다.

## 5. 실행 방식

### 검색과 파일 확인

파일 검색은 우선 `rg`, `rg --files`를 사용한다.

여러 파일을 독립적으로 읽을 수 있으면 병렬로 확인한다.

### 파일 수정

수동 파일 수정은 `apply_patch`를 사용한다.

간단한 파일 수정에 `cat > file`, Python 스크립트 작성 등 우회 방식을 사용하지 않는다.

### 중간 보고

긴 작업 중에는 사용자가 현재 상태를 알 수 있게 짧게 공유한다.

- 어떤 맥락을 확인 중인지
- 어떤 문제가 발견됐는지
- 다음에 무엇을 고칠지

## 6. 목표 기반 실행

작업은 검증 가능한 목표로 바꾼다.

예시:

```text
1. API 저장소를 PostgreSQL로 전환 -> verify: 재시작 후 데이터 유지 확인
2. Python Mock Robot 영상 송출 변경 -> verify: 브라우저 videoWidth/videoHeight 확인
3. 녹화 저장 구현 -> verify: MinIO object와 DB recording chunk 상태 확인
```

성공 조건이 약하면 먼저 성공 조건을 구체화한다.

## 7. 최종 사용자 검증 플로우

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
   verify: `./scripts/python-mock-robots-up.sh` 실행 완료
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

## 8. 문서화 규칙

단계형 구현 결과는 `docs/rebuild-results/` 아래에 남긴다.

기록할 내용:

- 구현 범위
- 검증 명령
- 실제 확인 결과
- 사용자 확인 URL
- Python Mock Robot 상태
- 남은 한계 또는 다음 단계

문서에는 임시 데이터와 실제 검증 데이터를 구분해서 쓴다.

## 9. 완료 보고 형식

완료 보고는 짧고 구체적으로 한다.

포함할 내용:

- 무엇을 바꿨는지
- 어떤 파일을 수정했는지
- 어떤 검증을 통과했는지
- 사용자가 어디서 확인하면 되는지
- 아직 남은 한계가 있는지

테스트하지 못한 항목은 성공한 것처럼 말하지 않는다.

## 10. 이 프로젝트의 현재 검증 기준

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
