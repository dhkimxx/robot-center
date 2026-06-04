---
title: "F-LLM-01 LLM API 연동 및 프롬프트 설계"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "f_llm", "f-llm-01"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
---

# F-LLM-01 LLM API 연동 및 프롬프트 설계

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `F-LLM-01` |
| 중분류 | F. LLM/SOP AI 에이전트 |
| 기능명 | LLM API 연동 및 프롬프트 설계 |
| 우선순위 | P0 |
| 난이도 | 중 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

관제센터 AI Agent용 모델 연동, Eino 기반 workflow 후보, prompt/domain context 설계

## 원본 XLSX 상세 설명

LLM 모델 선정 및 API 연동(GPT-4o/Claude/Gemini 중 최적 선택), 소방재난 도메인 특화 시스템 프롬프트 설계, Few-shot 예제 구축, 토큰 사용량 최적화, 멀티턴 대화 컨텍스트 관리.

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/features/f-llm-01-llm-api-prompt-design.md`
