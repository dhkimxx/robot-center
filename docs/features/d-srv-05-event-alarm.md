---
title: "D-SRV-05 이벤트 및 알람 기능"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "d_srv", "d-srv-05"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
---

# D-SRV-05 이벤트 및 알람 기능

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `D-SRV-05` |
| 중분류 | D. 서버/관제 플랫폼 |
| 기능명 | 이벤트 및 알람 기능 |
| 우선순위 | P0 |
| 난이도 | 중 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

탐지, 위험 센서, 통신 장애, recorder 상태 이벤트를 생성/표시/이력화하는 구조

## 원본 XLSX 상세 설명

탐지 이벤트(인명발견/가스누출/위험온도) 자동 생성, 알람 등급(주의/경고/위험) 분류, Push 알림(웹/모바일), 음성 알람 출력, 알람 확인/해제/이력 관리.

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/stable/architecture.md`
- `docs/stable/robot-interface.md`
