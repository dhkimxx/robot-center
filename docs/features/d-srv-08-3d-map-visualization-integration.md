---
title: "D-SRV-08 3D 맵 생성 및 시각화 연동"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "d_srv", "d-srv-08"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
---

# D-SRV-08 3D 맵 생성 및 시각화 연동

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `D-SRV-08` |
| 중분류 | D. 서버/관제 플랫폼 |
| 기능명 | 3D 맵 생성 및 시각화 연동 |
| 우선순위 | P1 |
| 난이도 | 상 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

로봇이 제공하는 spatial/point cloud object 참조를 수신하고 향후 WebGL/지도 패널로 표시할 수 있는 구조

## 원본 XLSX 상세 설명

LiDAR SLAM 결과 3D 포인트클라우드 실시간 수신, WebGL/Three.js 기반 3D 렌더링, 탐지 위치/로봇 궤적 3D 오버레이, 단면 뷰/탑뷰/자유시점 전환, 붕괴 건물 내부 구조 가시화.

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/stable/architecture.md`
- `docs/stable/robot-interface.md`
