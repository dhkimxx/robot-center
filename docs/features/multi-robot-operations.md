---
title: "multi-robot-operations"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "design"
tags: ["feature", "multi-robot", "mission", "sfu", "operations"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: initial feature split from rescue robot functional reference"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: split rescue robot function candidates into feature docs'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: clarify feature/imported document priority against current implementation baseline'
---

# Multi-Robot Operations

## 목적

mission room 기반 다중 로봇 관제 구조를 유지하면서, 향후 운영 기능으로 확장할 수 있는 후보를 정리한다.

## 현재 기반 우선 원칙

본 문서는 다중 로봇 운영 확장 후보이며, 현재 구현된 app-server 내부 SFU와 live-status SSOT 구조를 대체하지 않는다.

현재 기준은 다음 구조다.

```text
missionCode room
  robot-001 publish once
  robot-002 publish once
  operator A selected robot-001
  operator B selected robot-002
  recorder receives mission room robots
```

## 운영 후보

- mission 안에서 area 분담
- robot별 담당 구역 표시
- shared map
- shared victim detection
- shared hazard event
- robot별 mission status
- robot별 stream/recording/connection 상태 분리

## Mission Robot Assignment 확장 후보

필드 후보:

- robotCode
- assignmentStatus
- assignedAreaId
- role
- priority
- startedAt
- completedAt
- lastMissionEventAt

role 후보:

- search
- relay
- scout
- recorder
- standby

## UI 후보

- mission overview에서 robot별 card/tile 표시
- 선택 로봇 중심 4분면 관제 유지
- multi-view는 별도 기능으로 분리
- shared event는 timeline에 mission-level event로 표시

## 우리 프로젝트 반영 순서

### P0

- 현재 selected robot 구독 구조를 유지한다.
- robot별 live-status가 서로 섞이지 않게 한다.

### P1

- assignment에 담당 구역/역할을 추가할지 검토한다.
- mission list에서 active robot count와 stale reason을 명확히 보여준다.

### P2

- shared map과 shared hazard event를 spatial/event channel과 연결한다.
