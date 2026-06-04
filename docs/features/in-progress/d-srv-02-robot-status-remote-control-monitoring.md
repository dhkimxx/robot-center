---
title: "D-SRV-02 로봇 상태 및 원격 제어 모니터링"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "in-progress"
tags: ["feature", "d_srv", "d-srv-02", "in-progress"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: classify feature document by implementation status'
---

# D-SRV-02 로봇 상태 및 원격 제어 모니터링

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `D-SRV-02` |
| 중분류 | D. 서버/관제 플랫폼 |
| 기능명 | 로봇 상태 및 원격 제어 모니터링 |
| 우선순위 | P0 |
| 난이도 | 상 |

## 현재 기반 우선 원칙

본 문서는 진행 중인 기능이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

robot online/offline, active mission, live stream, control 예약 채널, 향후 제어 승인 UI 기반

## 원본 XLSX 상세 설명

실시간 로봇 상태 대시보드(배터리/온도/통신/위치), 원격 제어(Teleoperation) UI - 전진/후진/회전/정지, 카메라 PTZ 제어(팬/틸트/줌), 조이스틱/게임패드 HID 연동.

## 진행 중 분류 기준

현재 PoC에 일부 구현이 있으나 외부 계약, UI/운영 범위, 검증 기준이 계속 정리되는 기능이다.

완료 처리 전에는 관련 stable 문서와 실제 검증 결과를 함께 확인한다.

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/stable/architecture.md`
- `docs/stable/robot-interface.md`
