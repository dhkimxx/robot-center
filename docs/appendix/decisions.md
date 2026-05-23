# Appendix. Design Decisions

## 목적

AI Web을 설계하면서 결정한 아키텍처와 범위 판단을 짧게 기록한다.

상세 요구사항은 `docs/ai-web-prd.md`를 따른다.

## D-000. 기존 PoC를 제품형 앱으로 직접 확장하지 않는다

결정:

- 기존 WebRTC PoC 소스와 과거 단계별 검증 문서는 제거한다.
- 제품형 P0 애플리케이션은 새 구조로 재작성한다.
- 새 구조는 app-server, recorder-worker, Android Mock Robot, Docker Compose, DB migration을 처음부터 역할별로 나눈다.

이유:

- 기존 PoC는 WebRTC 경로 검증에 최적화되어 있다.
- 로봇 등록, 임무, 저장 metadata, 관제 UI, 권한, 제어 감사 로그를 뒤늦게 붙이면 책임 경계가 흐려진다.
- Android Mock Robot을 실제 로봇 샘플로 쓰려면 REST 등록 플로우부터 앱 구조에 반영해야 한다.
- Docker Compose와 PostgreSQL/PostGIS/MinIO를 기준으로 한 제품형 실행 환경이 필요하다.

## D-011. P0 서버는 app-server 중심으로 묶고 recorder-worker만 분리한다

결정:

- P0에서는 REST API, SFU signaling, SFU media fan-out, React 정적 파일 서빙을 `app-server`로 묶는다.
- 녹화, MP4 muxing, MinIO upload, 저장 metadata write는 `recorder-worker` 프로세스로 분리한다.
- `recorder-worker`는 같은 Go 코드베이스와 Docker image를 사용할 수 있지만 compose service는 별도로 둔다.
- TURN, PostgreSQL/PostGIS, MinIO는 인프라 성격이 다르므로 별도 service로 유지한다.

이유:

- `backend`, `sfu`, `web`을 처음부터 모두 별도 서버로 쪼개면 P0 배포와 시연 복잡도가 커진다.
- API와 SFU room/session 상태는 P0에서 강하게 연결되어 있다.
- 녹화와 muxing은 CPU/IO 부하와 장애 특성이 달라 API/SFU 프로세스와 분리하는 편이 안전하다.
- 이후 운영 단계에서 app-server 내부 API와 SFU를 별도 서비스로 분리할 수 있다.

## D-001. TURN은 별도 애플리케이션으로 분리한다

결정:

- PoC에서도 TURN/STUN은 app-server 내부에 넣지 않는다.
- TURN은 `coturn` 또는 Go 기반 별도 TURN 서버로 둔다.

이유:

- TURN은 영상 대역폭을 그대로 relay하므로 부하 특성이 app-server와 다르다.
- TURN은 media 내용을 저장하거나 해석하지 않는다.
- app-server 재시작이 WebRTC relay 장애로 이어지면 안 된다.

## D-002. TURN은 fan-out 서버가 아니다

결정:

- `Robot -> TURN -> Browser | app-server` 구조로 media를 복제하는 방식은 사용하지 않는다.
- 다중 Browser와 recorder-worker 분배는 SFU가 담당한다.

이유:

- TURN은 NAT traversal과 relay allocation을 담당하는 인프라이다.
- TURN은 RTP/RTCP, track, subscriber, keyframe 요청을 관리하는 SFU가 아니다.

## D-003. SFU 중심 구조를 채택한다

결정:

```text
Robot Gateway
  -> WebRTC publish once
  -> SFU
       -> Browser
       -> recorder-worker
```

이유:

- Browser가 여러 개 붙어도 Robot은 한 번만 publish한다.
- recorder-worker가 Browser와 독립적으로 같은 track을 subscribe할 수 있다.
- Robot 리소스 낭비를 줄일 수 있다.

## D-004. LiveKit은 비교 후보로 두되, 장기적으로 custom Pion SFU를 우선 검토한다

결정:

- PoC 일정 리스크가 크면 LiveKit, Janus, mediasoup도 비교한다.
- 장기적으로는 custom Pion SFU를 우선 검토한다.

이유:

- Robot 쪽에 LiveKit SDK 의존성을 주고 싶지 않다.
- AI Web의 recorder-worker, DataChannel, 관제 특화 제어 요구를 직접 통제하고 싶다.
- 단, SFU 구현 난이도는 높으므로 PoC 범위는 작게 잡는다.

## D-005. recorder-worker는 app-server/SFU와 process를 분리한다

결정:

- SFU 내부에 저장/DB/AI 로직을 넣지 않는다.
- recorder-worker가 SFU subscriber로 붙어 저장을 담당한다.

이유:

- 저장 장애가 실시간 관제 fan-out에 영향을 주면 안 된다.
- SFU는 media plane, recorder-worker는 저장 pipeline으로 책임이 다르다.

## D-006. PoC 영상 정책은 robot_defined로 한다

결정:

- PoC에서는 관제센터가 해상도, FPS, bitrate를 강제하지 않는다.
- Robot Gateway가 가능한 스펙으로 publish한다.
- 실제 송출/수신/저장된 스펙을 metadata로 남긴다.

이유:

- 실제 로봇은 Jetson + ROS + GStreamer 기반이다.
- AI Web팀이 로봇팀의 안정 송출 스펙을 아직 확정할 수 없다.
- 저장/검색을 위해 실제값 기록은 필요하다.

## D-007. 미션 하나에 여러 로봇을 수용할 수 있게 설계한다

결정:

```text
1 mission = 1 SFU room
1 mission 안에 N robots
1 robot은 rgb / thermal / sensor publish
```

식별 기준:

```text
missionId + robotCode + trackName
```

이유:

- PoC는 로봇 1대로 시작하지만, 구조상 다중 로봇 관제가 목표이다.
- 저장 key와 DB 모델을 처음부터 다중 로봇에 맞춰두는 편이 낫다.

## D-008. Robot 등록은 REST API 조합으로 단순화한다

결정:

- 별도 CLI는 만들지 않는다.
- 관제센터가 `serverUrl`, `robotCode`, `robotToken`을 제공한다.
- Robot Gateway는 heartbeat, mission 조회, streaming status 보고만 구현한다.

이유:

- 로봇팀 연동 플로우를 최소화한다.
- GitLab Runner 등록 방식의 아이디어는 참고하되 CLI는 PoC에서 제외한다.

## D-009. AI Agent는 직접 제어 명령을 실행하지 않는다

결정:

- Agent는 SOP 근거 기반 제어 명령 초안을 만들 수 있다.
- 실제 명령 실행은 관제요원 또는 지휘관 승인 후 app-server가 수행한다.

이유:

- 제어 명령은 감사 로그와 권한 확인이 필요하다.
- E-Stop, PTZ, Waypoint, Return-to-Home을 LLM이 자동 실행하면 운영 리스크가 크다.

## D-010. PoC에서 제외할 운영 메타데이터는 뒤로 미룬다

결정:

- `gatewayVersion`, `capabilities`는 PoC 필수 payload에서 제외한다.
- 이후 운영 단계에서 optional metadata로 추가할 수 있다.

이유:

- 지금 목표는 연결, publish, subscribe, 저장 경로 검증이다.
- 로봇팀이 챙겨야 할 필드를 최소화한다.
