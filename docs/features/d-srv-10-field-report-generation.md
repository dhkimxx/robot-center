---
title: "D-SRV-10 현장 보고서 자동 생성"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "d_srv", "d-srv-10"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
---

# D-SRV-10 현장 보고서 자동 생성

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `D-SRV-10` |
| 중분류 | D. 서버/관제 플랫폼 |
| 기능명 | 현장 보고서 자동 생성 |
| 우선순위 | P2 |
| 난이도 | 중 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

임무 종료 후 영상, 센서, 이벤트, 로봇 이동/상태 요약을 보고서로 생성하는 후속 기능

## 원본 XLSX 상세 설명

임무 종료 후 자동 보고서 생성(PDF/HTML), 탐지 결과 요약(발견 인원/위치/시간), 센서 데이터 요약(가스농도/온도 변화), 로봇 이동 경로/거리/시간, 커스텀 템플릿 지원.

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/stable/architecture.md`
- `docs/stable/robot-interface.md`
