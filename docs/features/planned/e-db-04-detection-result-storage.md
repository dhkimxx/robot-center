---
title: "E-DB-04 탐지 결과 저장"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "e_db", "e-db-04", "planned"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: classify feature document by implementation status'
---

# E-DB-04 탐지 결과 저장

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `E-DB-04` |
| 중분류 | E. DB/저장 시스템 |
| 기능명 | 탐지 결과 저장 |
| 우선순위 | P0 |
| 난이도 | 중 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

perception/detection event metadata, confidence, position, media timestamp 연계 저장

## 원본 XLSX 상세 설명

객체 탐지 결과(클래스/바운딩박스/신뢰도/좌표/타임스탬프), 3D 위치(월드좌표계 변환), 탐지 이미지 스냅샷 연관 저장, 탐지 추적(Tracking) ID 관리, 탐지 이벤트 히스토리.

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/stable/data-storage.md`
- `docs/stable/go-gorm-persistence.md`
