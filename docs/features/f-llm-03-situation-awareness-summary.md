---
title: "F-LLM-03 현장 상황 요약(Situation Awareness)"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "f_llm", "f-llm-03"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
---

# F-LLM-03 현장 상황 요약(Situation Awareness)

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `F-LLM-03` |
| 중분류 | F. LLM/SOP AI 에이전트 |
| 기능명 | 현장 상황 요약(Situation Awareness) |
| 우선순위 | P0 |
| 난이도 | 상 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

실시간 또는 replay 기반 임무 상황 요약, 시간대별 변화 요약, operator briefing

## 원본 XLSX 상세 설명

실시간 구조화 데이터를 자연어 요약으로 변환, "N층 M호 앞 통로에서 열원 2건 탐지, CO농도 상승 중" 등 구체적 상황 기술, 시간에 따른 상황 변화 추이 요약, 다국어 요약 지원(한/영).

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/features/f-llm-01-llm-api-prompt-design.md`
