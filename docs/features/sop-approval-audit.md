---
title: "sop-approval-audit"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "design"
tags: ["feature", "sop", "approval", "audit", "ai-agent"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: initial feature split from rescue robot functional reference"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: split rescue robot function candidates into feature docs'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: clarify feature/imported document priority against current implementation baseline'
---

# SOP, Approval, Audit

## 목적

SOP와 AI Agent를 직접 제어 주체가 아닌 추천/검증 보조 주체로 두고, 최종 결정과 감사 기록은 관제 운영 플로우에 남기는 방향을 정리한다.

## 현재 기반 우선 원칙

본 문서는 SOP/승인/audit 확장 후보이며, 현재 구현된 mission/control 상태를 대체하지 않는다.

AI Agent는 planned 기능이고, 현재 기준에서는 추천/초안 생성까지만 허용하며 실제 제어 실행은 operator 승인과 서버 검증 이후 단계로 둔다.

## 핵심 원칙

```text
SOP Agent -> MissionSetupRecommendation
Mission Core -> validation / approval
Operator or Mission Commander -> final decision
Robot Control -> approved command only
```

AI/SOP Agent는 Mission Start 또는 ControlCommand를 직접 생성하지 않는다.

## SOP Profile 후보

| SOP Profile | 기본 방향 | 제한 조건 후보 |
| --- | --- | --- |
| `mountain_missing_person` | Thermal 우선 수색, 넓은 구역 탐색 | 경사 제한, 배터리 장시간 정책 |
| `collapsed_structure` | 협소/잔해 구역 탐색 | 저속, 잔해 위험 alert, 수동 보조 |
| `tunnel_gas_risk` | 가스/산소 위험 우선 | gas danger 시 stop/retreat, 통신 저하 경고 |

## Role 후보

| Role | 권한 후보 |
| --- | --- |
| Observer | 영상/상태 조회 |
| Operator | 임무 생성 요청, 승인된 관제 조작 |
| Safety Officer | E-Stop, Mission Pause |
| Mission Commander | Mission Start/Stop/Resume 승인 |
| Admin | 사용자/권한/시스템 설정 |
| SOP Agent | Recommendation 생성 |

## Audit 후보

감사 로그 대상:

- Mission 생성
- Search area 변경
- Search method 선택
- SOP profile 적용
- Mission approval request
- Mission start/end/resume
- Control command
- Emergency stop

Control command audit 필드 후보:

- commandId
- missionId / missionCode
- robotCode
- operatorId
- commandType
- issuedAt
- result
- ack/nack reason

## 우리 프로젝트 반영 순서

### P0

- SOP Profile은 선택 필드/문서 수준으로만 준비한다.
- AI Agent 문서에 recommendation-only 원칙을 유지한다.

### P1

- Mission draft/approval을 구현할 때 Mission Commander 승인 상태를 추가한다.
- Mission 생성/승인 audit log 모델을 설계한다.

### P2

- Control command 구현 시 role guard, ACK/NACK, audit log를 필수 요구사항으로 둔다.
