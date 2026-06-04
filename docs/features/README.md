---
title: "features"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "guide"
status: "planned"
tags: ["features", "feature-code", "guide"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: reorganize features by feature code and title"
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: flatten feature documents directly under features directory"
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents"
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: rename feature files with English ASCII filenames"
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: move implemented feature documentation to done catalog"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: split features into planned in-progress and done directories'
---

# Features

이 디렉토리는 원본 XLSX 기능 ID를 기준으로 관제팀 기능 후보를 상태별로 관리한다.

```text
features/planned      아직 구현 착수 전이거나 요구사항 후보인 기능
features/in-progress  PoC 구현은 일부 있으나 계약, 운영 범위, 검증 기준이 아직 열려 있는 기능
features/done         현재 PoC 기준 구현 흐름이 확인된 기능
```

## 현재 기반 우선 원칙

features 문서는 backlog/진행상태 관리용이며, 현재 구현 기준이나 외부 로봇팀 계약을 새로 정의하지 않는다.

구현 기준은 `docs/stable/*`, 실행 절차는 `docs/guides/*`를 우선한다.

## Planned

| 기능 ID | 기능명 | 중분류 | 우선순위 | 난이도 | 문서 |
| --- | --- | --- | --- | --- | --- |
| `D-SRV-04` | 탐지 결과 시각화 | D. 서버/관제 플랫폼 | P0 | 중 | [d-srv-04-detection-result-visualization.md](planned/d-srv-04-detection-result-visualization.md) |
| `D-SRV-05` | 이벤트 및 알람 기능 | D. 서버/관제 플랫폼 | P0 | 중 | [d-srv-05-event-alarm.md](planned/d-srv-05-event-alarm.md) |
| `D-SRV-06` | 원격 관제(긴급 제어) 기능 | D. 서버/관제 플랫폼 | P0 | 중 | [d-srv-06-remote-control-emergency-command.md](planned/d-srv-06-remote-control-emergency-command.md) |
| `D-SRV-08` | 3D 맵 생성 및 시각화 연동 | D. 서버/관제 플랫폼 | P1 | 상 | [d-srv-08-3d-map-visualization-integration.md](planned/d-srv-08-3d-map-visualization-integration.md) |
| `D-SRV-10` | 현장 보고서 자동 생성 | D. 서버/관제 플랫폼 | P2 | 중 | [d-srv-10-field-report-generation.md](planned/d-srv-10-field-report-generation.md) |
| `E-DB-04` | 탐지 결과 저장 | E. DB/저장 시스템 | P0 | 중 | [e-db-04-detection-result-storage.md](planned/e-db-04-detection-result-storage.md) |
| `E-DB-05` | 이벤트 이력 관리 | E. DB/저장 시스템 | P1 | 중 | [e-db-05-event-history-management.md](planned/e-db-05-event-history-management.md) |
| `E-DB-07` | 데이터 백업/복구 및 감사 로그 | E. DB/저장 시스템 | P1 | 중 | [e-db-07-backup-recovery-audit-log.md](planned/e-db-07-backup-recovery-audit-log.md) |
| `F-LLM-01` | LLM API 연동 및 프롬프트 설계 | F. LLM/SOP AI 에이전트 | P0 | 중 | [f-llm-01-llm-api-prompt-design.md](planned/f-llm-01-llm-api-prompt-design.md) |
| `F-LLM-02` | 탐지 및 센서 결과 구조화 | F. LLM/SOP AI 에이전트 | P0 | 중 | [f-llm-02-detection-sensor-result-structuring.md](planned/f-llm-02-detection-sensor-result-structuring.md) |
| `F-LLM-03` | 현장 상황 요약(Situation Awareness) | F. LLM/SOP AI 에이전트 | P0 | 상 | [f-llm-03-situation-awareness-summary.md](planned/f-llm-03-situation-awareness-summary.md) |
| `F-LLM-04` | 위험도 분석 및 우선순위 판정 | F. LLM/SOP AI 에이전트 | P0 | 상 | [f-llm-04-risk-priority-assessment.md](planned/f-llm-04-risk-priority-assessment.md) |
| `F-LLM-05` | SOP(표준작전절차) 매핑 및 자동 제안 | F. LLM/SOP AI 에이전트 | P0 | 상 | [f-llm-05-sop-mapping-recommendation.md](planned/f-llm-05-sop-mapping-recommendation.md) |
| `F-LLM-06` | 대응 시나리오 추천 및 경로 안내 | F. LLM/SOP AI 에이전트 | P1 | 상 | [f-llm-06-response-scenario-route-guidance.md](planned/f-llm-06-response-scenario-route-guidance.md) |
| `F-LLM-07` | Hallucination 저감 및 품질 평가 | F. LLM/SOP AI 에이전트 | P0 | 상 | [f-llm-07-hallucination-reduction-quality-evaluation.md](planned/f-llm-07-hallucination-reduction-quality-evaluation.md) |
| `G-SEC-01` | 사용자 인증 및 권한 관리 | G. 보안 | P0 | 중 | [g-sec-01-user-authz-authn-management.md](planned/g-sec-01-user-authz-authn-management.md) |
| `G-SEC-02` | SSL/TLS 및 도메인 보안 | G. 보안 | P0 | 중 | [g-sec-02-ssl-tls-domain-security.md](planned/g-sec-02-ssl-tls-domain-security.md) |
| `G-SEC-04` | 영상/음성 데이터 암호화 | G. 보안 | P0 | 중 | [g-sec-04-video-audio-data-encryption.md](planned/g-sec-04-video-audio-data-encryption.md) |
| `G-SEC-05` | 접근 제어 및 이상 행위 탐지 | G. 보안 | P1 | 중 | [g-sec-05-access-control-anomaly-detection.md](planned/g-sec-05-access-control-anomaly-detection.md) |
| `J-VRF-06` | 성과 보고서 및 최종 평가 | J. 검증/실증 | P0 | 중 | [j-vrf-06-final-report-evaluation.md](planned/j-vrf-06-final-report-evaluation.md) |

## In Progress

| 기능 ID | 기능명 | 중분류 | 우선순위 | 난이도 | 문서 |
| --- | --- | --- | --- | --- | --- |
| `A-SYS-01` | 전체 시스템 아키텍처 설계 | A. 시스템 아키텍처 및 인터페이스 | P0 | 중 | [a-sys-01-system-architecture-design.md](in-progress/a-sys-01-system-architecture-design.md) |
| `A-SYS-03` | 엣지-서버-관제 간 통신 인터페이스 정의 | A. 시스템 아키텍처 및 인터페이스 | P0 | 중 | [a-sys-03-edge-server-control-interface.md](in-progress/a-sys-03-edge-server-control-interface.md) |
| `A-SYS-04` | 데이터 모델 및 저장 구조 설계 | A. 시스템 아키텍처 및 인터페이스 | P0 | 중 | [a-sys-04-data-model-storage-design.md](in-progress/a-sys-04-data-model-storage-design.md) |
| `A-SYS-06` | 시스템 모니터링 및 로깅 아키텍처 | A. 시스템 아키텍처 및 인터페이스 | P2 | 중 | [a-sys-06-system-monitoring-logging-architecture.md](in-progress/a-sys-06-system-monitoring-logging-architecture.md) |
| `D-SRV-02` | 로봇 상태 및 원격 제어 모니터링 | D. 서버/관제 플랫폼 | P0 | 상 | [d-srv-02-robot-status-remote-control-monitoring.md](in-progress/d-srv-02-robot-status-remote-control-monitoring.md) |
| `D-SRV-07` | 영상 저장 및 조회(VMS) | D. 서버/관제 플랫폼 | P1 | 중 | [d-srv-07-video-recording-replay-vms.md](in-progress/d-srv-07-video-recording-replay-vms.md) |
| `D-SRV-09` | 멀티 로봇 관제 | D. 서버/관제 플랫폼 | P1 | 상 | [d-srv-09-multi-robot-control.md](in-progress/d-srv-09-multi-robot-control.md) |
| `E-DB-01` | PostgreSQL DB 스키마 설계 및 구축 | E. DB/저장 시스템 | P0 | 중 | [e-db-01-postgresql-schema-design.md](in-progress/e-db-01-postgresql-schema-design.md) |
| `E-DB-02` | 로봇 상태 및 임무 이력 저장 | E. DB/저장 시스템 | P0 | 중 | [e-db-02-robot-status-mission-history-storage.md](in-progress/e-db-02-robot-status-mission-history-storage.md) |
| `E-DB-03` | 센서 로그 및 시계열 저장 | E. DB/저장 시스템 | P0 | 중 | [e-db-03-sensor-log-timeseries-storage.md](in-progress/e-db-03-sensor-log-timeseries-storage.md) |
| `E-DB-06` | 영상/스냅샷 저장 및 수명주기 관리 | E. DB/저장 시스템 | P1 | 중 | [e-db-06-media-snapshot-lifecycle-management.md](in-progress/e-db-06-media-snapshot-lifecycle-management.md) |
| `I-COM-02` | 영상 스트리밍 최적화 | I. 통신 시스템 | P0 | 상 | [i-com-02-video-streaming-optimization.md](in-progress/i-com-02-video-streaming-optimization.md) |
| `I-COM-03` | 통신 장애 대응 | I. 통신 시스템 | P0 | 중 | [i-com-03-communication-failure-response.md](in-progress/i-com-03-communication-failure-response.md) |
| `I-COM-04` | 멀티 로봇 통신 관리 | I. 통신 시스템 | P1 | 상 | [i-com-04-multi-robot-communication-management.md](in-progress/i-com-04-multi-robot-communication-management.md) |
| `J-VRF-04` | 통합 시스템 검증 | J. 검증/실증 | P0 | 중 | [j-vrf-04-integrated-system-verification.md](in-progress/j-vrf-04-integrated-system-verification.md) |

## Done

| 기능 ID | 기능명 | 중분류 | 우선순위 | 난이도 | 문서 |
| --- | --- | --- | --- | --- | --- |
| `D-SRV-01` | 관제 서버 기본 기능 | D. 서버/관제 플랫폼 | P0 | 중 | [d-srv-01-control-server-baseline.md](done/d-srv-01-control-server-baseline.md) |
| `D-SRV-03` | 영상/센서 통합 시각화 대시보드 | D. 서버/관제 플랫폼 | P0 | 중 | [d-srv-03-video-sensor-integrated-dashboard.md](done/d-srv-03-video-sensor-integrated-dashboard.md) |

## 상태 변경 기준

- `planned -> in-progress`: 코드, 계약 문서, 검증 스크립트 중 하나 이상이 실제 작업 대상이 된 경우
- `in-progress -> done`: 현재 PoC에서 사용자가 확인 가능한 기능 흐름과 관련 stable/guide 문서가 함께 정리된 경우
- `done -> in-progress`: 계약 변경, 기능 범위 재오픈, 회귀 버그로 완료 근거가 약해진 경우

## 제외 기준

다음 기능 코드는 현재 관제팀 repo 구현 범위가 아니므로 feature 파일로 만들지 않는다.

- `B-EDG-*`: Jetson/로봇 탑재 엣지 SW 구현
- `C-AI-*`: 온디바이스 AI 모델 학습/최적화 구현
- `H-RBT-*`: 로봇 HW/센서 직접 연동 구현
- `A-SYS-02`: 로봇-센서-엣지 내부 ICD
- `A-SYS-05`: 범용 개발환경/CI 운영 항목
- `G-SEC-03`, `G-SEC-06`: 장기 솔루션화 보안 항목
- `J-VRF-01`, `J-VRF-02`, `J-VRF-03`, `J-VRF-05`: 로봇/AI/현장 실증 주관 검증 항목
