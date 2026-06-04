---
title: "D-SRV-04 탐지 결과 시각화"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "d_srv", "d-srv-04", "planned"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: classify feature document by implementation status'
---

# D-SRV-04 탐지 결과 시각화

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `D-SRV-04` |
| 중분류 | D. 서버/관제 플랫폼 |
| 기능명 | 탐지 결과 시각화 |
| 우선순위 | P0 |
| 난이도 | 중 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

로봇/AI가 전송한 detection event를 영상 overlay, 지도 marker, timeline 형태로 표시할 수 있는 UI/API 구조

## 원본 XLSX 상세 설명

RGB 영상 위 바운딩박스/라벨 오버레이, 열화상 히트맵 오버레이, 3D 맵 위 탐지 위치 마커, 탐지 신뢰도(Confidence) 표시, 탐지 이력 타임라인 뷰.

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/stable/architecture.md`
- `docs/stable/robot-interface.md`
