---
title: "mission-planning"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "design"
tags: ["feature", "mission", "rescue", "planning", "search"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: initial feature split from rescue robot functional reference"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: split rescue robot function candidates into feature docs'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: clarify mission planning workflow as extension over stable mission status'
---

# Mission Planning

## 목적

임무 생성 기능을 단순 `mission name + robot assignment`에서 탐색 임무 계획 중심으로 확장하기 위한 기능 후보를 정리한다.

## 현재 기반 우선 원칙

본 문서는 임무 계획 기능의 초기 후보이며, 현재 구현된 mission status를 대체하지 않는다.

현재 관제 서버의 mission 상태 기준은 stable/코드 기반을 우선한다.

```text
ready -> active -> ended
ready/active -> cancelled
```

아래의 draft/approval 흐름은 mission status 자체가 아니라, mission 생성 전 계획/승인 workflow 후보로 본다.

## 기능 후보

### 탐색 임무 Draft

임무 시작 전 관제자가 탐색 임무 초안을 생성한다.

후보 입력:

- mission name
- scenario
- assigned robot codes
- search area
- search method
- SOP profile
- mission priority
- operator id

초안은 바로 실행되지 않고, 검증과 승인 단계를 거친다.

### Search Area

| Area Type | 의미 | 적용 후보 |
| --- | --- | --- |
| `POLYGON` | 지도 위 다각형 탐색 영역 | 산악조난/붕괴현장 지도 UI |
| `WAYPOINT_ROUTE` | 순차 경유점 경로 | 로봇팀 waypoint 계약 이후 |
| `GRID` | 격자 기반 구역 탐색 | 수색 패턴 UI 고도화 |
| `GEOFENCE` | 진입/이탈 제한 구역 | 위험 구역, no-go zone |

검증 후보:

- `POLYGON`: 점 3개 이상, self-intersection 없음
- `WAYPOINT_ROUTE`: waypoint 2개 이상
- `GRID`: cell size와 boundary 필수
- `GEOFENCE`: mission area와 충돌 여부 확인

### Search Method

| Search Method | 활용 시나리오 |
| --- | --- |
| `AREA_SWEEP` | 일반 구역 수색 |
| `PARALLEL_SWEEP` | 평탄/개방 구역 균일 탐색 |
| `CREEPING_LINE` | 가능성이 한쪽으로 치우친 넓은 구역 |
| `EXPANDING_SQUARE` | 마지막 목격 위치 중심 확장 |
| `SECTOR_SEARCH` | 기준점 중심 반경 탐색 |
| `TRACKLINE_SEARCH` | 도로, 계곡, 터널, 통로 추적 |
| `CONTOUR_SEARCH` | 산악/사면 등고선 탐색 |
| `GRID_COVERAGE` | 격자 기반 정밀 탐색 |
| `FRONTIER_EXPLORATION` | 미지 영역 지도 확장 |
| `MANUAL_ASSISTED` | 관제자 보조 반자동 탐색 |

### Mission Planning Workflow 후보

```text
draft
  -> approval_requested
  -> approved
  -> ready
```

`emergency_stop`은 mission status가 아니라 제어/이벤트/안전 상태 후보로 둔다. 임무 상태 전이는 현재 기준인 `ended` 또는 `cancelled`와 연결해서 별도 설계한다.

## 우리 프로젝트 반영 순서

### P0

- 임무 생성 UI/API에 `searchArea`, `searchMethod`, `sopProfile` 추가 여부를 설계한다.
- 기존 `ready/active/ended/cancelled` 상태와 planning workflow의 호환 방식을 정리한다.

### P1

- 지도에서 polygon/waypoint를 입력하는 UI를 추가한다.
- Mission start 전에 validation/approval guard를 추가한다.

### P2

- Search plan을 실제 로봇 waypoint/search command 계약과 연결한다.
