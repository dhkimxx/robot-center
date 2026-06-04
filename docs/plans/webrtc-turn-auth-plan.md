---
title: "webrtc-turn-auth-plan"
created: 2026-05-27
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "roadmap"
status: "planned"
tags: ["webrtc", "turn", "sfu", "auth", "operator", "recorder", "security"]
history:
- "2026-05-27 danya.kim <danya.kim@thundersoft.com>: initial plan for TURN, operator WebSocket, and recorder WebSocket authentication"
- "2026-06-01 danya.kim <danya.kim@thundersoft.com>: update robot-facing signaling path to /api/v1/robot/sfu/ws"
- '2026-06-01 danya.kim <danya.kim@thundersoft.com>: clarify Robot API token scope and returned data boundary'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: reorganize docs directories by document purpose'
---

# WebRTC / TURN Authentication Plan

## 1. 목적

WebRTC signaling과 TURN 사용 경계를 역할별로 분리하고 인증 정책을 정리한다.

현재 정리 중인 Robot WebSocket은 다음 계약을 기준으로 한다.

```http
GET /api/v1/robot/sfu/ws?room={missionCode}
Authorization: Bearer {robotToken}
```

이 문서는 Robot WebSocket 인증이 이미 별도 작업으로 진행 중이라는 전제에서, 남은 인증 범위를 다룬다.

- Operator WebSocket 인증
- Recorder WebSocket 인증
- TURN credential 발급/검증 정책

목표:

- 로봇팀이 구현할 Robot Gateway 스펙에 추후 변경이 가지 않도록 한다.
- operator, recorder 인증을 나중에 붙여도 endpoint와 연결 계약이 바뀌지 않게 한다.
- TURN credential은 client가 하드코딩하지 않고 서버 응답값을 opaque secret으로 사용하게 한다.
- 임시 개발서버용 static credential과 운영형 time-limited credential의 경계를 명확히 한다.

비목표:

- 사용자 계정/권한 전체 설계
- HTTPS/WSS 인증서 운영 구성
- mTLS 기반 service mesh 구성
- 실제 로봇 장치 구현

## 2. 현재 기준

### 2.1 WebSocket endpoint

역할별 endpoint로 분리한다.

```http
GET /api/v1/robot/sfu/ws?room={missionCode}
GET /api/v1/operator/sfu/ws?room={missionCode}
GET /api/v1/recorder/sfu/ws?room={missionCode}
```

Robot:

- `Authorization: Bearer {robotToken}` 필수
- token으로 robotCode resolve
- room mission에 해당 robot이 active assignment인지 검증
- `/api/v1/robot/*`는 token으로 인증된 자기 로봇에 필요한 정보만 반환
- 다른 robot, 다른 mission, operator/recorder/internal runtime 정보는 Robot API 응답에 포함하지 않음

Operator:

- 현재는 인증 없음
- 향후 사용자 세션 또는 operator WS ticket으로 인증

Recorder:

- 현재는 인증 없음
- 향후 service token으로 인증

### 2.2 TURN

Docker Compose는 coturn을 사용한다.

```yaml
turn:
  image: coturn/coturn:4.6.3
  command:
    - --lt-cred-mech
    - --realm=${TURN_REALM:-robot-center.local}
    - --user=${TURN_USERNAME:-robot}:${TURN_PASSWORD:-robot-pass}
```

현재 방식은 shared static credential이다.

```json
{
  "urls": ["turn:192.168.20.12:3478?transport=udp"],
  "username": "robot",
  "credential": "server-managed-secret"
}
```

이 방식은 임시 개발서버 테스트에는 가능하지만 운영형 계약으로는 부족하다.

## 3. 인증 원칙

### 3.1 Robot Gateway 영향 최소화

로봇팀은 다음만 지키면 된다.

- Robot WebSocket은 `sfu.signalingUrl` 그대로 연결한다.
- Robot WebSocket에는 `Authorization: Bearer {robotToken}`을 넣는다.
- TURN 설정은 mission response의 `turnServers`를 그대로 `RTCPeerConnection.iceServers`에 넣는다.
- TURN username/credential을 하드코딩하거나 장기 캐시하지 않는다.

이 원칙을 지키면 TURN static credential에서 ephemeral credential로 바뀌어도 로봇팀 구현 변경이 없다.

### 3.2 Browser Operator는 Authorization header에 의존하지 않는다

브라우저의 native `WebSocket` API는 임의의 `Authorization` header 설정이 불가능하다.

따라서 operator 인증은 다음 중 하나로 설계한다.

우선안:

- app-server와 same-origin cookie 기반 session
- `HttpOnly`, `Secure`, `SameSite=Lax` 또는 배포 구조에 맞는 SameSite 설정
- `/api/v1/operator/sfu/ws` handshake에서 session cookie 검증

대안:

- REST API로 짧은 수명의 operator WS ticket 발급
- `/api/v1/operator/sfu/ws?room={missionCode}&ticket={ticket}` 형태로 연결
- query ticket은 로그 노출 위험이 있으므로 TTL을 짧게 두고 기본안으로 쓰지 않는다.

### 3.3 Recorder는 service token을 사용한다

recorder-worker는 server-side client이므로 WebSocket handshake에 `Authorization` header를 넣을 수 있다.

```http
GET /api/v1/recorder/sfu/ws?room={missionCode}
Authorization: Bearer {recorderServiceToken}
```

service token은 `.env` 또는 secret manager로 관리한다.

## 4. 목표 계약

### 4.1 Operator WebSocket

P0 내부 테스트:

```http
GET /api/v1/operator/sfu/ws?room={missionCode}
Cookie: operator session
```

초기에는 operator session이 없으므로 인증을 비활성화할 수 있다. 단, endpoint와 path는 지금 확정한다.

향후 인증 활성화 시:

- room query는 필수
- `role`, `robotCode` query는 금지
- session이 없으면 `401`
- mission read 권한이 없으면 `403`
- 연결 후 기존 `select-robot` 메시지로 선택 robot을 지정한다.

### 4.2 Recorder WebSocket

P0 내부 테스트:

```http
GET /api/v1/recorder/sfu/ws?room={missionCode}
```

향후 인증 활성화 시:

```http
GET /api/v1/recorder/sfu/ws?room={missionCode}
Authorization: Bearer {recorderServiceToken}
```

규칙:

- room query는 필수
- `role`, `robotCode` query는 금지
- token이 없거나 틀리면 `401`
- service token은 recorder-worker 전용이다.

### 4.3 TURN credential

임시 개발서버 P0:

```json
{
  "urls": ["turn:192.168.20.12:3478?transport=udp"],
  "username": "server-issued-static-username",
  "credential": "server-issued-static-password",
  "credentialType": "password"
}
```

운영형 목표:

```json
{
  "urls": ["turn:192.168.20.12:3478?transport=udp"],
  "username": "1769489700:robot:robot-001",
  "credential": "base64-hmac-sha1-secret",
  "credentialType": "password",
  "expiresAt": "2026-05-27T05:15:00.000Z"
}
```

규칙:

- `username`은 `{expiresUnix}:{role}:{subject}` 형태다.
- `credential`은 coturn shared secret 기반 HMAC 결과다.
- `expiresAt`은 client가 재조회/재연결 판단에 사용할 수 있도록 내려준다.
- Robot Gateway는 credential 생성 방식을 알 필요가 없다.

## 5. 구현 단계

### Phase 1. 현재 static TURN credential 명확화

목표:

- 임시 개발서버용 static TURN credential을 명시적으로 관리한다.
- 기본값 `robot-pass`에 의존하지 않는다.

변경 대상:

- `.env.example`
- `.env.dev-server` 운용 가이드
- `deploy/docker-compose.yml`
- `docs/guides/dev-server-docker-deployment.md`
- `docs/guides/robot-team-webrtc-send-test-guide.md`
- `docs/stable/robot-interface.md`

작업:

- `TURN_USERNAME`, `TURN_PASSWORD`는 서버별 secret으로 설정한다고 명시한다.
- 문서 예시에서 실제 password처럼 보이는 값을 제거한다.
- 로봇팀 가이드에는 `turnServers`를 그대로 사용하라고만 안내한다.
- `TURN credential은 임시 개발서버 테스트용 static credential`이라고 명시한다.

검증:

- mission response에 `turnServers`가 내려온다.
- Python Mock Robot이 `turnServers` 그대로 사용해 ICE relay 연결한다.
- 문서에 `robot-pass`를 로봇팀 공유용 secret처럼 노출하지 않는다.

### Phase 2. Recorder WebSocket service token 추가

목표:

- `/api/v1/recorder/sfu/ws`를 service token으로 보호한다.
- recorder-worker만 recorder role로 join할 수 있게 한다.

변경 대상:

- `apps/server/internal/config/config.go`
- `apps/server/internal/api/sfu_handlers.go`
- `apps/server/internal/recording/subscriber.go`
- `deploy/docker-compose.yml`
- `.env.example`
- `scripts/start.sh`

환경변수:

```env
RECORDER_WS_TOKEN=server-managed-recorder-token
```

작업:

- app-server config에 `RecorderWebSocketToken` 추가
- recorder-worker config에 같은 token 추가
- `/api/v1/recorder/sfu/ws` handler에서 `Authorization: Bearer {RECORDER_WS_TOKEN}` 검증
- recorder-worker WebSocket dial에 Authorization header 추가

검증:

- token 없음: `401`
- 잘못된 token: `401`
- 올바른 token: recorder join 성공
- recorder-worker health에서 subscriber room 상태 정상
- RGB/Thermal track과 telemetry/spatial DataChannel 수신

### Phase 3. Operator WebSocket session 인증 설계/도입

목표:

- `/api/v1/operator/sfu/ws`를 관제 사용자 session으로 보호한다.
- 향후 사용자/권한 체계가 들어와도 endpoint 변경 없이 확장한다.

전제:

- 현재 프로젝트에 정식 operator login/session이 없다면 이 phase는 인증 hook만 준비하고 비활성화할 수 있다.

변경 대상:

- `apps/server/internal/api/sfu_handlers.go`
- operator auth/session package
- `apps/web/src/domains/live/liveConnectionClient.js`
- 배포 CORS/cookie 설정

환경변수:

```env
OPERATOR_WS_AUTH_MODE=disabled|session
```

작업:

- `disabled`일 때는 현재처럼 operator join 허용
- `session`일 때는 WebSocket handshake cookie를 검증
- `role`, `robotCode` query는 계속 금지
- mission read 권한이 생기면 room별 권한 검증 추가

검증:

- `OPERATOR_WS_AUTH_MODE=disabled`: 기존 UI live 동작
- `OPERATOR_WS_AUTH_MODE=session`: session 없으면 `401`
- session 있으면 operator join 성공
- `select-robot` 메시지와 track/data 수신 동작 유지

### Phase 4. TURN ephemeral credential 도입

목표:

- static TURN username/password를 time-limited credential로 전환한다.
- 로봇팀 구현 변경 없이 mission/rtc config 응답값만 바뀌게 한다.

변경 대상:

- `deploy/docker-compose.yml`
- `apps/server/internal/config/config.go`
- `apps/server/internal/api/robot_gateway_handlers.go`
- `apps/server/internal/api/system_handlers.go`
- `apps/server/internal/recording/subscriber.go`
- 필요 시 recorder-worker용 RTC config fetch/generation

환경변수:

```env
TURN_AUTH_MODE=static|ephemeral
TURN_AUTH_SECRET=server-managed-turn-secret
TURN_CREDENTIAL_TTL=30m
```

coturn command 변경:

```yaml
- --lt-cred-mech
- --use-auth-secret
- --static-auth-secret=${TURN_AUTH_SECRET}
- --realm=${TURN_REALM:-robot-center.local}
```

credential 생성:

```text
username = "{expiresUnix}:{role}:{subject}"
credential = base64(hmac-sha1(TURN_AUTH_SECRET, username))
```

role별 subject:

| Role | Subject |
| --- | --- |
| robot | `robotCode` |
| operator | session user id 또는 session id |
| recorder | recorder worker id |

응답 위치:

- Robot: `GET /api/v1/robot/mission`의 `turnServers`
- Operator: `GET /api/v1/operator/rtc-config`의 `iceServers`
- Recorder: recorder-worker config 또는 app-server에서 발급받은 recorder RTC config

검증:

- 유효 credential로 ICE relay 연결 성공
- 만료 credential로 새 TURN allocation 실패
- credential TTL 내 재연결 성공
- 로봇/브라우저/recorder 모두 relay candidate만 사용

## 6. 사이드이펙트와 대응

### 6.1 Browser WebSocket Authorization header 불가

영향:

- operator endpoint는 Bearer header 방식으로 설계하면 브라우저에서 바로 쓸 수 없다.

대응:

- operator는 cookie session을 기본안으로 둔다.
- ticket query 방식은 대안으로만 둔다.

### 6.2 TURN credential TTL

영향:

- credential이 너무 짧으면 ICE restart 또는 재연결 때 실패할 수 있다.

대응:

- 초기 TTL은 `30m` 이상으로 둔다.
- robot은 mission 재조회로 새 credential을 받는다.
- operator는 `/api/v1/operator/rtc-config` 재조회로 새 credential을 받는다.
- recorder는 tick/reconnect 전에 새 credential을 받는다.

### 6.3 Clock skew

영향:

- time-limited TURN username은 서버/클라이언트 시간 차이에 취약할 수 있다.

대응:

- credential 생성/검증은 서버와 coturn 기준이다.
- client는 `expiresAt`을 표시/재조회 힌트로만 사용한다.
- TTL에 여유를 둔다.

### 6.4 Secret rotation

영향:

- `TURN_AUTH_SECRET`, `RECORDER_WS_TOKEN` 회전 시 기존 연결 영향 가능.

대응:

- P1에서는 단일 secret으로 시작한다.
- 운영 전에는 active/next secret 두 개를 허용하는 rotation window를 설계한다.

### 6.5 로봇팀 영향

영향:

- TURN credential 방식 변경 자체는 로봇팀에 영향이 없어야 한다.

대응:

- 로봇팀 문서에는 `turnServers`를 그대로 사용하라고만 명시한다.
- username/credential 생성 규칙은 내부 문서에만 둔다.

## 7. 최종 검증 기준

TURN:

```text
verify: static mode에서 기존 Docker Compose TURN 연결 성공
verify: ephemeral mode에서 app-server가 role별 TURN credential 발급
verify: coturn이 ephemeral credential을 수락
verify: 잘못된 credential은 relay allocation 실패
verify: robot/operator/recorder 모두 relay ICE candidate로 연결
```

Operator WebSocket:

```text
verify: /api/v1/operator/sfu/ws?room={missionCode} role query 없이 연결
verify: robotCode query는 400
verify: auth disabled mode에서 UI live 연결 성공
verify: session mode에서 session 없으면 401
verify: session mode에서 session 있으면 select-robot 및 track/data 수신 성공
```

Recorder WebSocket:

```text
verify: /api/v1/recorder/sfu/ws?room={missionCode} role query 없이 연결
verify: token 없으면 401
verify: 잘못된 token이면 401
verify: 올바른 RECORDER_WS_TOKEN이면 join 성공
verify: recorder health에 lastTrackAt, lastDataAt, lastPersistedAt 반영
```

Robot 영향 확인:

```text
verify: Robot Gateway spec은 /api/v1/robot/sfu/ws + robotToken 유지
verify: Robot Gateway는 turnServers를 그대로 사용
verify: TURN static -> ephemeral 전환 시 Robot Gateway 코드 변경 없음
```

## 8. 권장 일정

1. P0 dev server: static TURN credential 유지, operator auth disabled, recorder service token 도입
2. P0 문서 공유 전: 로봇팀 가이드에 TURN credential opaque 사용 원칙 명시
3. P1: operator session auth 추가
4. P1: TURN ephemeral credential 추가
5. 운영 전: secret rotation과 HTTPS/WSS 구성 확정
