---
title: "D-SRV-01 관제 서버 기본 기능"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "done"
tags: ["feature", "d_srv", "d-srv-01", "done"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: move implemented feature documentation to done catalog'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: classify feature document by implementation status'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: align done feature wording with nested directory'
---

# D-SRV-01 관제 서버 기본 기능

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `D-SRV-01` |
| 중분류 | D. 서버/관제 플랫폼 |
| 기능명 | 관제 서버 기본 기능 |
| 우선순위 | P0 |
| 난이도 | 중 |

## 현재 기반 우선 원칙

본 문서는 현재 PoC 기준 구현 완료로 분류된 기능이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

Go app-server, REST API, DTO, service/repository, system health/status API

## 원본 XLSX 상세 설명

REST API 서버 구축(FastAPI/Spring Boot), 사용자 인증(JWT/OAuth2), RBAC 기반 권한 관리(지휘관/대원/관리자), API Rate Limiting, Swagger API 문서 자동화.

## 완료 분류 기준

현재 PoC 코드에서 관제팀 구현 범위가 동작하는 기능으로 확인되어 `docs/features/done`으로 이동했다.

확인 근거: Go app-server, health/system API, robot/mission/recording/sensor API 기본 기능이 구현됨.

세부 계약은 `docs/stable/*`, 실행 검증 절차는 `docs/guides/*`를 우선한다.

## 유지보수 메모

- 기능 범위 변경 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/stable/architecture.md`
- `docs/stable/robot-interface.md`
