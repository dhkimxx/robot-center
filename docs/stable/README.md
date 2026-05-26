---
title: "stable-index"
created: 2026-05-26
updated: 2026-05-26
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "guide"
status: "stable"
tags: ["stable", "index"]
history:
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: stable 문서 인덱스 작성"
---

# Stable Docs

현재 코드와 검증 결과가 일치하는 기준 문서다.

| 문서 | 역할 |
| --- | --- |
| `architecture.md` | app-server, SFU, recorder-worker, 저장 흐름 전체 기준 |
| `robot-interface.md` | Robot Gateway, Mock Robot, WebRTC publish/status 계약 |
| `data-storage.md` | PostgreSQL/GORM, sensor, recording, MinIO 저장 기준 |
| `harness/20260522-harness-index.md` | 검증된 server/SFU harness 진입점 |

`stable` 문서를 바꿀 때는 구현 또는 실제 검증 결과도 함께 확인해야 한다.
