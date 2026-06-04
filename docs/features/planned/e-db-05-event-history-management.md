---
title: "E-DB-05 이벤트 이력 관리"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "e_db", "e-db-05", "planned"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: classify feature document by implementation status'
---

# E-DB-05 이벤트 이력 관리

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `E-DB-05` |
| 중분류 | E. DB/저장 시스템 |
| 기능명 | 이벤트 이력 관리 |
| 우선순위 | P1 |
| 난이도 | 중 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

event log, alarm, control command, system fault, recording event 검색/필터링 구조

## 원본 XLSX 상세 설명

통합 이벤트 로그(탐지/알람/제어명령/시스템장애), 이벤트 분류/심각도 태깅, 고급 검색(기간/유형/로봇/위치별 필터링), 이벤트 연관 분석(전후 5분 컨텍스트), CSV/Excel 이벤트 내보내기.

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/stable/data-storage.md`
- `docs/stable/go-gorm-persistence.md`
