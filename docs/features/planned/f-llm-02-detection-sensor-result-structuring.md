---
title: "F-LLM-02 탐지 및 센서 결과 구조화"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "f_llm", "f-llm-02", "planned"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: classify feature document by implementation status'
---

# F-LLM-02 탐지 및 센서 결과 구조화

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `F-LLM-02` |
| 중분류 | F. LLM/SOP AI 에이전트 |
| 기능명 | 탐지 및 센서 결과 구조화 |
| 우선순위 | P0 |
| 난이도 | 중 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

센서/이벤트/위치/영상 metadata를 AI Agent 입력으로 구조화

## 원본 XLSX 상세 설명

AI 탐지 결과 + 센서 데이터 통합 구조화(JSON Schema 정의), 실시간 상황 정보 정규화, 공간정보(GPS/3D 좌표/지도링크) 포함, 시간 윈도우별 상황 집계(최근 1분/5분/10분).

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/features/f-llm-01-llm-api-prompt-design.md`
