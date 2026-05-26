---
title: "docs-index"
created: 2026-05-26
updated: 2026-05-26
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "guide"
status: "stable"
tags: ["docs", "index", "architecture", "webrtc"]
history:
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: stable/planned/history 문서 구조 도입"
---

# Robot Center Docs

이 문서 폴더는 현재 기준, 계획, 작업 이력을 분리해 관리한다.

## 디렉토리 기준

| Directory | 의미 | 사용 기준 |
| --- | --- | --- |
| `docs/stable/` | 현재 구현과 검증이 일치하는 기준 문서 | Robot팀 공유, WebRTC/DB/API 기준 확인 |
| `docs/planned/` | 앞으로 진행할 사항, 미확정 설계 | AI Agent, 시나리오, 제품/연동 후보 |
| `docs/history/` | 작업 기록, 과거 판단, 리뷰/검증 결과 | 변경 이유와 검증 근거 추적 |

## 빠른 진입

| 목적 | 문서 |
| --- | --- |
| 전체 아키텍처 | `docs/stable/architecture.md` |
| Robot Gateway / WebRTC 연결 계약 | `docs/stable/robot-interface.md` |
| DB / MinIO / 저장 구조 | `docs/stable/data-storage.md` |
| 검증된 SFU harness 기준 | `docs/stable/harness/20260522-harness-index.md` |
| AI Agent 계획 | `docs/planned/ai-agent.md` |
| 실증 시나리오 후보 | `docs/planned/scenarios.md` |
| 앞으로 할 일 | `docs/planned/roadmap.md` |
| 작업 로그 | `docs/history/rebuild-results/` |

## 현재 안정 기준

```text
Python Mock Robot
-> app-server / Robot Gateway
-> app-server internal Pion SFU
-> Browser operator subscriber
-> recorder-worker subscriber
-> PostgreSQL / MinIO
```

핵심 규칙:

- `missionId`는 DB UUID다.
- `roomId`는 `missionCode`이며 SFU room id다.
- `robotCode`는 room id에 합치지 않고 payload/status/recording metadata로 유지한다.
- Robot은 mission room에 한 번만 publish한다.
- Browser는 같은 mission room에서 선택한 robot만 subscribe한다.
- Recorder는 같은 mission room의 모든 robot을 subscribe한다.
- Streaming freshness는 서버 `updatedAt` 기준으로 판단한다.

## 문서 작성 규칙

- 현재 구현과 검증이 맞으면 `stable`에 둔다.
- 구현 전이거나 합의 전이면 `planned`에 둔다.
- 날짜별 작업 결과, 리뷰, 검증 로그는 `history`에 둔다.
- 모든 문서는 가능하면 YAML frontmatter에 `type`, `status`, `tags`, `history`를 유지한다.
