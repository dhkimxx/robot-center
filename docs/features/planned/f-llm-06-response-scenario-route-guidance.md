---
title: "F-LLM-06 대응 시나리오 추천 및 경로 안내"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "f_llm", "f-llm-06", "planned"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: classify feature document by implementation status'
---

# F-LLM-06 대응 시나리오 추천 및 경로 안내

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `F-LLM-06` |
| 중분류 | F. LLM/SOP AI 에이전트 |
| 기능명 | 대응 시나리오 추천 및 경로 안내 |
| 우선순위 | P1 |
| 난이도 | 상 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

지도/공간 정보 기반 접근/철수/대안 시나리오 권고. 실제 robot navigation 제어는 외부 로봇팀 영역

## 원본 XLSX 상세 설명

AI 기반 구조 경로 추천(위험도/접근성/시간 최적화), 3D 맵 위 경로 시각화, 대안 시나리오 비교(최단/최안전/최다인원 커버), 진입/철수 타이밍 권고, 실시간 상황 변화에 따른 경로 재계획.

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/features/f-llm-01-llm-api-prompt-design.md`
