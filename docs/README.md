---
title: "docs"
created: 2026-06-04
updated: '2026-06-04'
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

앞으로 만들 기능의 요구사항 또는 후보를 기능 코드 단위로 정리한다.

여기 문서는 구현 기준이 아니라 backlog와 기능 설계 초안이다. 파일명은 원본 XLSX의 기능 코드와 기능명을 포함한다. 기능이 구현되고 계약이 확정되면 관련 stable 문서를 갱신한다.

예:

- `d-srv-03-video-sensor-integrated-dashboard.md`
- `d-srv-07-video-recording-replay-vms.md`
- `f-llm-03-situation-awareness-summary.md`
- `i-com-02-video-streaming-optimization.md`

### `plans/`

특정 작업을 어떤 순서로 진행할지 정리한다.

작업 계획, 단계별 구현 계획, 보안/인증 전환 계획처럼 시간이 지나면 완료/폐기/수정될 수 있는 문서를 둔다.

### `guides/`

개발, 시연, 테스트, 검증 절차를 정리한다.

가이드는 실행 절차 문서이며, stable 계약을 대체하지 않는다.

예:

- 개발 서버 Docker 배포
- 로봇팀 WebRTC 송신 테스트
- multi-robot SFU 검증 checklist

## Maintenance Rules

- stable 문서는 현재 코드와 맞지 않으면 즉시 갱신한다.
- features 문서는 기능 후보를 담되, stable과 충돌하는 계약을 확정 표현으로 쓰지 않는다.
- plans 문서는 작업 완료 후 필요하면 archive 또는 stable/features/guides로 정리한다.
- guides 문서는 명령어, endpoint, 실행 URL이 바뀌면 함께 갱신한다.
