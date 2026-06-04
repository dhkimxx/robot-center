---
title: "A-SYS-06 시스템 모니터링 및 로깅 아키텍처"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "in-progress"
tags: ["feature", "a_sys", "a-sys-06", "in-progress"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: classify feature document by implementation status'
---

# A-SYS-06 시스템 모니터링 및 로깅 아키텍처

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `A-SYS-06` |
| 중분류 | A. 시스템 아키텍처 및 인터페이스 |
| 기능명 | 시스템 모니터링 및 로깅 아키텍처 |
| 우선순위 | P2 |
| 난이도 | 중 |

## 현재 기반 우선 원칙

본 문서는 진행 중인 기능이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

app-server, recorder-worker, PostgreSQL, MinIO, SFU room/peer 상태 관측과 운영 로그 설계

## 원본 XLSX 상세 설명

분산 로깅(ELK/Loki), 메트릭 수집(Prometheus), 분산 트레이싱(Jaeger), 알림 체계(PagerDuty/Slack 연동) 설계.

## 진행 중 분류 기준

현재 PoC에 일부 구현이 있으나 외부 계약, UI/운영 범위, 검증 기준이 계속 정리되는 기능이다.

완료 처리 전에는 관련 stable 문서와 실제 검증 결과를 함께 확인한다.

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/stable/architecture.md`
- `docs/stable/robot-interface.md`
- `docs/guides/dev-server-docker-deployment.md`
