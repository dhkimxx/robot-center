---
title: "A-SYS-03 엣지-서버-관제 간 통신 인터페이스 정의"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "a_sys", "a-sys-03"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
---

# A-SYS-03 엣지-서버-관제 간 통신 인터페이스 정의

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `A-SYS-03` |
| 중분류 | A. 시스템 아키텍처 및 인터페이스 |
| 기능명 | 엣지-서버-관제 간 통신 인터페이스 정의 |
| 우선순위 | P0 |
| 난이도 | 중 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

Robot Gateway REST, role별 SFU WebSocket, WebRTC track/DataChannel 계약 문서화

## 원본 XLSX 상세 설명

Jetson ↔ 서버 간 REST API(제어/설정), gRPC(고속 데이터), MQTT(상태/이벤트) 프로토콜별 인터페이스 규격 정의. API 스펙 문서화(Swagger/OpenAPI).

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/stable/architecture.md`
- `docs/stable/robot-interface.md`
- `docs/guides/dev-server-docker-deployment.md`
