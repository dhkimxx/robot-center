# AI Web Documentation

이 문서 폴더는 AI Web 관제센터 P0 기준으로 정리한다.

과거 PoC 단계별 로그와 오래된 rebuild phase 로그는 현재 구현 기준과 섞이지 않도록 제거했다. 필요한 과거 이력은 git history에서 확인한다.

## 빠른 진입

| 목적 | 먼저 볼 문서 |
| --- | --- |
| 제품 범위와 수용 기준 확인 | `docs/ai-web-prd.md` |
| 현재 구현/검증 상태 확인 | `docs/rebuild-results/control-center-live-refactor-2026-05-25.md` |
| 서버/SFU 기준 확인 | `docs/harness/20260522-harness-index.md` |
| 전체 구조와 런타임 구성 확인 | `docs/appendix/architecture.md` |
| Robot Gateway / Mock Robot 연동 확인 | `docs/appendix/robot-interface.md` |
| 관제 UI 화면 범위 확인 | `docs/appendix/ui-plan.md` |
| DB schema, MinIO object, replay 구조 확인 | `docs/appendix/data-storage.md` |
| AI Agent 범위 확인 | `docs/appendix/ai-agent.md` |
| 실증 시나리오 확인 | `docs/appendix/scenarios.md` |
| 주요 설계 결정 확인 | `docs/appendix/decisions.md` |

## 현재 기준

최신 검증 기준은 `control-center-live-refactor-2026-05-25` 결과와
`docs/harness/20260522-harness-index.md`의 mission room SFU 기준이다.

```text
Robot Gateway / Mock Robot
-> app-server / robot gateway
-> TURN relay-only WebRTC path
-> app-server internal Pion SFU
-> Browser 관제 UI
-> recorder-worker
-> PostgreSQL / MinIO
```

확인된 상태:

- 여러 mock robot이 같은 `roomId=missionCode` mission room에 publish
- Browser와 recorder-worker가 SFU subscriber로 독립 수신
- RGB/Thermal/Audio track 송수신
- sensor/telemetry DataChannel 수신과 DB 적재
- MP4, JSONL, manifest MinIO 업로드
- 업로드 완료 object metadata를 `storage_objects`에 기록
- manifest object를 `recording_chunks.manifest_object_id`에 연결
- 관제 UI에서 녹화 목록과 파일 URL 조회

## 핵심 문서

| 문서 | 역할 |
| --- | --- |
| `docs/ai-web-prd.md` | 전체 시스템 P0 PRD |
| `docs/appendix/architecture.md` | Docker Compose, app-server, recorder-worker, Robot Gateway 기준 전체 아키텍처 |
| `docs/appendix/robot-interface.md` | Robot Gateway / Mock Robot REST/WebRTC 연동 계약 |
| `docs/appendix/ui-plan.md` | React 관제 UI 화면과 기능 범위 |
| `docs/appendix/data-storage.md` | PostgreSQL/PostGIS 테이블과 MinIO object key, replay 규칙 |
| `docs/appendix/ai-agent.md` | Go/Eino 기반 AI Agent 기능 명세 |
| `docs/appendix/scenarios.md` | 산악조난, 붕괴현장, 지하시설 실증 시나리오 |
| `docs/appendix/decisions.md` | 주요 설계 결정 기록 |

## Harness 문서

| 문서 | 역할 |
| --- | --- |
| `docs/harness/20260522-harness-index.md` | harness 문서 진입점 |
| `docs/harness/20260522-server-architecture.md` | 서버 기본 아키텍처 기준 |
| `docs/harness/20260522-webrtc-sfu-topology.md` | WebRTC SFU 송수신 구조 기준 |
| `docs/harness/20260522-go-gorm-persistence.md` | Go/GORM persistence 작성 기준 |

## 검증 결과

| 문서 | 내용 |
| --- | --- |
| `docs/rebuild-results/control-center-live-refactor-2026-05-25.md` | 최신 관제 UI, live connection lifecycle, Python mock robot 검증 결과 |

## 문서 원칙

- PRD에는 제품 목표, 범위, 수용 기준을 둔다.
- 상세 API, schema, UI, object key는 appendix로 분리한다.
- 구현 기준과 충돌하는 과거 로그는 유지하지 않는다.
- 최신 구현 결과와 실제 검증 결과는 `docs/rebuild-results/`에 남긴다.
