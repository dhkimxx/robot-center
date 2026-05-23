# Appendix. UI Plan

## 1. 문서 목적

AI Web P0 React 관제 UI의 화면 구조, 기능 범위, 주요 상태, 컴포넌트 책임을 정의한다.

UI는 단순 WebRTC 데모 화면이 아니라 관제센터 운영 애플리케이션으로 설계한다.

## 2. UX 원칙

- 첫 화면은 임무 목록이다.
- 관제요원이 반복적으로 보는 정보는 밀도 있게 배치한다.
- 개발 기술명은 운영 UI에 노출하지 않는다.
- 시스템 상태는 별도 화면으로 분리한다.
- 실시간 영상, 지도, 이벤트, 제어 패널은 한 화면에서 동시에 확인 가능해야 한다.
- Loading, Error, Empty, Reconnecting, Offline 상태를 명시적으로 처리한다.
- PoC는 로봇 1대 중심이지만 UI 구조는 다중 로봇을 막지 않는다.

## 3. 정보 구조

```text
Login
Robots
Missions
Recordings
Events
System Status
```

대시보드 탭은 P0 범위에서 제외한다. 진행 임무, 로봇 상태, 녹화 상태는 임무 목록과 임무 관제 화면에서 확인한다.

실시간 관제는 별도 `Live 관제` 탭을 두지 않는다. 관제요원은 `Missions`에서 임무를 선택한 뒤 해당 임무의 관제 화면으로 진입한다.

## 4. 화면별 기능

### 4.1 Login

기능:

- ID/PW 로그인
- 로그인 실패 표시
- 세션 만료 처리

P0 seed user:

```text
operator / operator
```

### 4.2 Robots

목적:

- 로봇을 등록하고 Android Mock Robot 연결 정보를 확인한다.

기능:

- 로봇 목록
- 로봇 생성
- 로봇 상세
- 로봇 표시 이름/모델 수정
- 로봇 삭제
- robotCode 확인
- robotToken 확인 또는 재발급
- Android Mock 입력값 표시
- 마지막 heartbeat 시각
- 현재 상태 표시

로봇 생성 입력:

- displayName
- modelName
- memo

로봇 삭제 원칙:

- 녹화, 센서, 임무 이력 보존을 위해 물리 삭제하지 않는다.
- `archived_at` 기준으로 관제 목록에서 제외한다.
- 진행 중 또는 대기 중 임무에 배정된 로봇은 삭제할 수 없다.

생성 결과:

```text
serverUrl
robotCode
robotToken
```

### 4.3 Missions

목적:

- 임무를 생성하고 시작/종료한다.

기능:

- 임무 목록
- 상태 필터
- 시나리오 필터
- 임무 생성
- 로봇 배정
- 임무 시작
- 임무 종료
- 관제 화면 진입

임무 생성 입력:

- mission name
- mission type
- site note
- assigned robot

### 4.4 Mission Control

P0 핵심 화면이다.

진입 경로:

```text
임무 목록 -> 관제 -> 임무 관제 화면 -> 로봇 선택
```

레이아웃:

```text
┌─────────────────────────────────────────────────────────────┐
│ Mission Header                                               │
├───────────────────────────────────────────────┬─────────────┤
│ Robot Selector / RGB / Thermal / Map / Sensor │ Operation   │
│                                               │ Events      │
└───────────────────────────────────────────────┴─────────────┘
```

필수 패널:

- Mission Header
- Robot Connection Summary
- Robot Selector
- RGB Video
- Thermal Video
- Audio status
- GPS Map
- Sensor Summary
- Sensor Trend
- Event Timeline
- Recording Status
- Recording Playback
- Control Panel
- AI Situation Summary

Mission Header:

- mission name
- mission type
- status
- elapsed time
- assigned robot count
- recording status

Video Panel:

- RGB video
- Thermal video
- stream state
- current resolution/FPS/bitrate
- reconnecting/offline overlay

Map Panel:

- robot marker
- GPS accuracy circle
- heading
- trail polyline
- last updated time

Sensor Panel:

- battery
- network quality
- temperature
- humidity
- oxygen
- CO
- CH4

Event Timeline:

- event type
- severity
- occurredAt
- related robot
- related recording chunk

Recording Playback:

- 최근 재생 가능한 MP4를 관제 UI 모달에서 바로 재생
- RGB MP4와 Thermal MP4 선택 재생
- 외부 MinIO Console이나 새 탭으로 보내지 않음
- 아직 생성 중인 chunk는 `예정`으로 두고, 가장 최근 `available` 파일을 재생 대상으로 사용

Control Panel:

- E-Stop
- Return-to-Home
- PTZ controls
- command status
- ACK/NACK

AI Panel:

- current situation summary
- risk level
- recommended SOP candidates
- control draft, approval required

### 4.5 Recordings

목적:

- 임무별 녹화 세션을 조회한다.
- 세션 안의 chunk와 저장 파일 상태를 한 화면에서 확인한다.
- 저장된 media와 sensor/telemetry 데이터를 replay 화면으로 연결한다.

기능:

- mission filter
- robot filter
- robot grouped navigation
- recording session list
- session status
- chunk list
- chunk status
- media metadata
- modal video player
- metadata/download link
- replay 진입
- failed upload 표시

표시:

- 로봇별 세션 수
- 로봇별 최근 녹화 시각
- 세션 시작/종료 시각
- 세션 상태
- chunk 개수
- 저장 완료 파일 수
- chunk index
- chunk start/end
- chunk duration
- RGB + Audio MP4 상태
- Thermal MP4 상태
- Sensor JSONL 상태
- Telemetry/GPS JSONL 상태
- Manifest JSON 상태

숨김:

- bucket 이름
- object key
- MinIO Console URL
- recorder-worker 내부 처리 상세

파일 상태:

- `recording`: 현재 수집 중
- `planned`: 저장 예정 또는 아직 생성 전
- `available`: app-server가 접근 URL을 발급할 수 있음
- `failed`: 저장 실패

기본 정렬:

- 좌측 로봇 목록은 최근 녹화 시각 내림차순
- 로봇 선택 시 우측 세션 목록은 해당 로봇의 최근 녹화 시각 내림차순
- 한 세션 안의 chunk도 최신 chunk가 먼저 보이도록 chunk index 내림차순

Replay 화면:

- chunk timeline
- RGB video player
- Thermal video player
- GPS path replay
- sensor chart replay
- event marker overlay
- chunk metadata

Replay 원칙:

- UI는 MinIO object key를 직접 조합하지 않는다.
- UI는 app-server에서 recording metadata와 URL을 받아 재생한다.
- P0 개발 환경 URL은 MinIO API URL을 사용할 수 있다.
- P0는 MP4 직접 재생으로 시작한다.
- `video/mp4` 파일은 `열기` 대신 `재생` 버튼을 제공하고, 선택 시 관제 UI 모달 플레이어를 연다.
- 임무 관제 화면의 녹화 저장 패널은 `재생` 액션을 제공하고, 선택 시 관제 UI 모달 플레이어를 연다.
- manifest와 JSONL 같은 비디오가 아닌 파일은 보조 `보기/열기` 액션으로 둔다.
- 운영 URL은 app-server가 발급한 presigned URL을 기본으로 한다.
- 브라우저 호환성 문제가 확인되면 replay derivative 또는 HLS/fMP4를 P1로 검토한다.

### 4.6 Events

목적:

- 이벤트 상세와 근거 데이터를 확인한다.

기능:

- event list
- severity filter
- event detail
- related video chunk
- related sensor values
- related control command
- map position

### 4.7 System Status

목적:

- 운영자가 시스템 구성요소 상태를 확인한다.

표시:

- app-server
- TURN
- recorder-worker
- PostgreSQL
- MinIO

주의:

- 이 화면 외의 운영 화면에는 PostgreSQL, MinIO, Go, Pion 같은 개발 기술명을 노출하지 않는다.

## 5. 프론트엔드 구조

권장 구조:

```text
apps/web/src
  api/
    authApi.ts
    robotApi.ts
    missionApi.ts
    recordingApi.ts
    eventApi.ts
    rtcApi.ts
  components/
    layout/
    video/
    map/
    sensor/
    timeline/
    control/
    status/
  pages/
    LoginPage.tsx
    RobotsPage.tsx
    MissionsPage.tsx
    MissionControlPage.tsx
    RecordingsPage.tsx
    EventsPage.tsx
    SystemStatusPage.tsx
  hooks/
    useMissionControl.ts
    useRobotTelemetry.ts
    useSfuSubscriber.ts
  state/
    authStore.ts
  styles/
```

원칙:

- API 호출은 `api/*Api.ts`에 둔다.
- WebRTC 연결은 hook/service로 분리한다.
- 화면 컴포넌트는 데이터 조회, 상태 표시, 사용자 액션에 집중한다.
- map/video/sensor/control은 재사용 컴포넌트로 분리한다.

## 6. P0 API 연동

UI가 필요한 API:

- `POST /api/auth/login`
- `GET /api/robots`
- `POST /api/robots`
- `GET /api/robots/{robotId}`
- `GET /api/robots/{robotId}/connection-info`
- `GET /api/missions`
- `POST /api/missions`
- `POST /api/missions/{missionId}/start`
- `POST /api/missions/{missionId}/end`
- `GET /api/missions/{missionId}/live`
- `GET /api/missions/{missionId}/rtc-config`
- `GET /api/recordings`
- `GET /api/recordings/{recordingId}/chunks`
- `GET /api/events`
- `POST /api/control-commands`
- `GET /api/system/status`

## 7. 상태 처리

공통 상태:

- Loading
- Empty
- Error
- Offline
- Reconnecting
- Streaming
- Recording
- Failed

임무 관제 화면은 부분 장애를 허용한다.

예:

- 영상은 끊겼지만 sensor는 수신 중
- sensor는 끊겼지만 영상은 수신 중
- recorder-worker 실패지만 Browser live는 정상
- app-server API 일시 실패지만 기존 SFU stream은 가능한 범위에서 유지

## 8. 첫 구현 우선순위

1. Layout shell, navigation
2. Robot registration 화면
3. Mission list/create/start 화면
4. Mission Control 화면
5. Recording list 화면
6. System status 화면
7. Event detail 화면
