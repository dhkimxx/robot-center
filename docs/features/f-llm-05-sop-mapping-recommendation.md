---
title: "F-LLM-05 SOP(표준작전절차) 매핑 및 자동 제안"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "f_llm", "f-llm-05"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
---

# F-LLM-05 SOP(표준작전절차) 매핑 및 자동 제안

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `F-LLM-05` |
| 중분류 | F. LLM/SOP AI 에이전트 |
| 기능명 | SOP(표준작전절차) 매핑 및 자동 제안 |
| 우선순위 | P0 |
| 난이도 | 상 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

산악조난/붕괴현장/지하시설 SOP rule set과 상황 event 매칭, 대응 권고 표시

## 원본 XLSX 상세 설명

소방재난 SOP 데이터베이스 구축(구조/화재/붕괴/화학), 현재 상황과 SOP 자동 매칭(시맨틱 유사도 기반), 해당 SOP 단계별 가이드 제시. 목표 SOP 적합도 ≥ 90% 달성.

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/features/f-llm-01-llm-api-prompt-design.md`
