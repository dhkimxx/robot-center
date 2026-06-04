---
title: "D-SRV-06 원격 관제(긴급 제어) 기능"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "d_srv", "d-srv-06", "planned"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: classify feature document by implementation status'
---

# D-SRV-06 원격 관제(긴급 제어) 기능

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `D-SRV-06` |
| 중분류 | D. 서버/관제 플랫폼 |
| 기능명 | 원격 관제(긴급 제어) 기능 |
| 우선순위 | P0 |
| 난이도 | 중 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

control channel, 권한, 승인, 감사 로그를 고려한 관제 명령 UI/서버 경계. 직접 제어 자동화는 제외

## 원본 XLSX 상세 설명

긴급 정지(E-Stop) - 즉시 모든 모터 정지, 수동 경로 지정(Waypoint) 및 자동 주행, 귀환(Return-to-Home) 명령, 특정 영역 출입 제한(Geo-fence) 설정, 제어 명령 로깅 및 감사.

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/stable/architecture.md`
- `docs/stable/robot-interface.md`
