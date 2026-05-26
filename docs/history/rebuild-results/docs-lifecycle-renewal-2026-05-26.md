---
title: "docs-lifecycle-renewal"
created: 2026-05-26
updated: 2026-05-26
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "log"
status: "verified"
tags: ["docs", "stable", "planned", "history"]
history:
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: docs stable/planned/history renewal 기록"
---

# Docs Lifecycle Renewal

## 구현 범위

- `docs/stable/`를 현재 구현과 검증이 일치하는 기준 문서 위치로 정했다.
- `docs/planned/`를 앞으로 진행할 사항과 미확정 설계 위치로 정했다.
- `docs/history/`를 작업 기록, 리뷰 결과, 검증 로그 위치로 정했다.
- 기존 `appendix`, `harness`, `rebuild-results` 문서를 lifecycle 구조로 재배치했다.
- `docs/README.md`, 각 디렉토리별 `README.md`, `docs/planned/roadmap.md`를 추가했다.
- 문서 frontmatter에 `status`를 추가하고, 새 경로 기준으로 내부 링크를 갱신했다.

## 새 구조

```text
docs/
  stable/
  planned/
  history/
```

## 검증 명령

```bash
rg --files docs | sort
rg -n "legacy-doc-path-pattern" docs -S
uv run --project .agents/skills/meta-docs .agents/skills/meta-docs/doc_manager.py search --root /Users/dhkim/workspace/sst/robot-center --dir stable
uv run --project .agents/skills/meta-docs .agents/skills/meta-docs/doc_manager.py search --root /Users/dhkim/workspace/sst/robot-center --dir planned
uv run --project .agents/skills/meta-docs .agents/skills/meta-docs/doc_manager.py search --root /Users/dhkim/workspace/sst/robot-center --dir history
git diff --check -- docs
```

## 실제 확인 결과

- `rg --files docs`로 새 구조 파일 목록 확인.
- 과거 경로 문자열 잔존 없음.
- `meta-docs` search가 `stable`, `planned`, `history`에서 frontmatter 문서를 인식함.
- `git diff --check -- docs` 통과.

## 남은 한계

- 이전에 이미 삭제 상태였던 과거 로그, decision/UI plan/retired PRD 문서는 복구하지 않았다.
- 과거 삭제 문서를 git history에서 복구할지는 별도 결정이 필요하다.
