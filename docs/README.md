---
title: "docs"
created: 2026-06-04
updated: '2026-06-09'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "guide"
status: "stable"
tags: ["docs", "structure", "guide"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: define docs directory structure and priority"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: reorganize docs directories by document purpose'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: remove imported document category after distilling useful content'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: add feature code catalog to docs structure examples'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: describe features as feature-code based documents'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: document English ASCII feature filename examples'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: move implemented feature documentation to done catalog'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: align feature examples with done catalog'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: remove moved feature from examples'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: replace root done with feature status directories'
- '2026-06-08 danya.kim <danya.kim@thundersoft.com>: document monthly plan file convention'
- '2026-06-09 danya.kim <danya.kim@thundersoft.com>: mention deploy verification harness in guide structure'
---

# Docs

이 디렉토리는 `robot-center` 프로젝트 문서를 목적별로 관리한다.

문서 판단 우선순위는 다음과 같다.

```text
stable > guides > plans > features
```

단, 실제 코드와 stable 문서가 충돌하면 현재 동작하는 코드와 검증 결과를 먼저 확인하고 stable 문서를 갱신한다.

## Directory Rules

### `stable/`

현재 코드와 외부 계약의 source of truth다.

여기에 들어가는 문서는 로봇팀 또는 외부 이해관계자에게 공유 가능한 기준 문서여야 한다.

예:

- architecture
- robot interface
- data storage
- persistence convention

### `features/`

앞으로 만들 기능의 요구사항, 후보, 진행 상태를 기능 코드 단위로 정리한다.

여기 문서는 구현 기준이 아니라 backlog와 기능 진행 상태 관리용이다. 파일명은 원본 XLSX의 기능 코드와 기능명을 포함한다. 기능이 구현되고 계약이 확정되면 관련 stable 문서를 갱신하고 상태별 하위 디렉토리로 이동한다.

상태 디렉토리:

- `features/planned/`: 아직 구현 착수 전이거나 요구사항 후보인 기능
- `features/in-progress/`: PoC 구현은 일부 있으나 계약, 운영 범위, 검증 기준이 아직 열려 있는 기능
- `features/done/`: 현재 PoC 기준 구현 흐름이 확인된 기능

예:

- `d-srv-02-robot-status-remote-control-monitoring.md`
- `d-srv-05-event-alarm.md`
- `f-llm-03-situation-awareness-summary.md`
- `g-sec-01-user-authz-authn-management.md`

### `plans/`

특정 작업을 어떤 순서로 진행할지 정리한다.

작업 계획, 단계별 구현 계획, 보안/인증 전환 계획처럼 시간이 지나면 완료/폐기/수정될 수 있는 문서를 둔다.

월별 계획 문서는 다음 파일명으로 추가한다.

```text
docs/plans/YYYY-MM-plan.md
```

예:

- `docs/plans/2026-06-plan.md`
- `docs/plans/2026-07-plan.md`

월별 계획은 해당 월의 주차별 목표, 완료 기준, 제외 범위, 산출물을 포함한다.

### `guides/`

개발, 시연, 테스트, 검증 절차를 정리한다.

가이드는 실행 절차 문서이며, stable 계약을 대체하지 않는다.

예:

- 개발 서버 Docker 배포와 배포검증 하네스
- 로봇팀 WebRTC 송신 테스트
- multi-robot SFU 검증 checklist

## Maintenance Rules

- stable 문서는 현재 코드와 맞지 않으면 즉시 갱신한다.
- features 문서는 기능 후보와 진행 상태를 담되, stable과 충돌하는 계약을 확정 표현으로 쓰지 않는다.
- `features/done` 문서는 현재 구현 근거가 사라지거나 기능 범위가 다시 열리면 `features/in-progress` 또는 `features/planned`로 되돌린다.
- plans 문서는 작업 완료 후 필요하면 archive 또는 stable/features/guides로 정리한다.
- guides 문서는 명령어, endpoint, 실행 URL이 바뀌면 함께 갱신한다.
