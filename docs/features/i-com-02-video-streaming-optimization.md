---
title: "I-COM-02 영상 스트리밍 최적화"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "i_com", "i-com-02"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
---

# I-COM-02 영상 스트리밍 최적화

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `I-COM-02` |
| 중분류 | I. 통신 시스템 |
| 기능명 | 영상 스트리밍 최적화 |
| 우선순위 | P0 |
| 난이도 | 상 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

WebRTC/SFU 기반 RGB/Thermal/Audio/DataChannel 수신, selected robot subscribe, live-status 관측

## 원본 XLSX 상세 설명

WebRTC 기반 초저지연 영상 전송, SVC(Scalable Video Coding) - 채널 상태 적응, Simulcast(멀티비트레이트 동시 송출), FEC(Forward Error Correction) 패킷 손실 복구, 지터 버퍼 최적화.

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/stable/architecture.md`
- `docs/stable/robot-interface.md`
- `docs/guides/dev-server-docker-deployment.md`
