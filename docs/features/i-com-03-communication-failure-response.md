---
title: "I-COM-03 통신 장애 대응"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "i_com", "i-com-03"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
---

# I-COM-03 통신 장애 대응

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `I-COM-03` |
| 중분류 | I. 통신 시스템 |
| 기능명 | 통신 장애 대응 |
| 우선순위 | P0 |
| 난이도 | 중 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

SFU observed stream, heartbeat, ICE 상태 기반 송출/연결/녹화 상태 판단, 재연결 UX

## 원본 XLSX 상세 설명

자동 채널 전환(5G↔WiFi↔LTE), 통신 두절 감지 및 Edge 로컬모드 전환, 재연결 시 데이터 동기화(Delta Sync), 최저대역폭 모드(센서데이터 우선, 영상 프레임↓), 통신 품질 모니터링 대시보드.

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/stable/architecture.md`
- `docs/stable/robot-interface.md`
- `docs/guides/dev-server-docker-deployment.md`
