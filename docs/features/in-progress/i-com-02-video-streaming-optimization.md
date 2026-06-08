---
title: "I-COM-02 영상 스트리밍 최적화"
created: 2026-06-04
updated: '2026-06-08'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "in-progress"
tags: ["feature", "i_com", "i-com-02", "in-progress"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: move implemented feature documentation to done catalog'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: classify feature document by implementation status'
- '2026-06-08 danya.kim <danya.kim@thundersoft.com>: update in-progress feature evidence from recent implementation'
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

본 문서는 진행 중인 기능이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

WebRTC/SFU 기반 RGB/Thermal/Audio/DataChannel 수신, selected robot subscribe, live-status 관측

## 원본 XLSX 상세 설명

WebRTC 기반 초저지연 영상 전송, SVC(Scalable Video Coding) - 채널 상태 적응, Simulcast(멀티비트레이트 동시 송출), FEC(Forward Error Correction) 패킷 손실 복구, 지터 버퍼 최적화.

## 최근 진행 반영

2026-06-08 기준으로 SFU subscriber offer 처리와 live video 표시 보강이 반영되었다.

반영된 범위:

- SFU subscriber offer 생성/정리 흐름 개선
- robot reconnect 이후 subscriber track 재협상 흐름 보강
- live video panel에 영상 해상도/프레임 등 metrics overlay 추가
- canonical media track 기반 수신 흐름 유지

아직 완료가 아닌 이유:

- 원 기능 범위의 SVC, Simulcast, FEC, 지터 버퍼 최적화는 아직 구현되지 않았다.
- 현재 구현은 PoC 관제 송출 안정화와 관측성 보강 중심이다.
- 실제 로봇/네트워크 환경에서 bitrate adaptation 정책 검증이 필요하다.

## 진행 중 분류 기준

현재 PoC에 일부 구현이 있으나 외부 계약, UI/운영 범위, 검증 기준이 계속 정리되는 기능이다.

완료 처리 전에는 관련 stable 문서와 실제 검증 결과를 함께 확인한다.

## 유지보수 메모

- 기능 범위 변경 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/stable/architecture.md`
- `docs/stable/robot-interface.md`
- `docs/guides/dev-server-docker-deployment.md`
