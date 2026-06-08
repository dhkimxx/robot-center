---
title: "D-SRV-07 영상 저장 및 조회(VMS)"
created: 2026-06-04
updated: '2026-06-08'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "in-progress"
tags: ["feature", "d_srv", "d-srv-07", "in-progress"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: move implemented feature documentation to done catalog'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: classify feature document by implementation status'
- '2026-06-08 danya.kim <danya.kim@thundersoft.com>: update in-progress feature evidence from recent implementation'
- '2026-06-08 danya.kim <danya.kim@thundersoft.com>: record PoC recording stability validation'
---

# D-SRV-07 영상 저장 및 조회(VMS)

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `D-SRV-07` |
| 중분류 | D. 서버/관제 플랫폼 |
| 기능명 | 영상 저장 및 조회(VMS) |
| 우선순위 | P1 |
| 난이도 | 중 |

## 현재 기반 우선 원칙

본 문서는 진행 중인 기능이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

recorder-worker, MP4 muxing, MinIO 저장, recording session/chunk metadata, 종료 임무 replay 화면

## 원본 XLSX 상세 설명

전체 영상 스트림 녹화(NAS 저장 연동), 타임라인 기반 영상 검색/재생, 이벤트 연동 북마크(탐지 시점 이동), 영상 내보내기(클립 추출), 멀티채널 동기 재생.

## 최근 진행 반영

2026-06-08 기준으로 recording replay 안정화 작업이 추가 반영되었다.

반영된 범위:

- active recording chunk를 live status에서 우선 반영하도록 개선
- recording replay API와 DTO 정리
- mission replay 화면의 robot 목록/robot 이름 표시 개선
- replay chunk panel, replay helper, recording API 테스트 보강
- recorder-worker chunk lifecycle, media upload, media track writer 안정화

## 2026-06-08 PoC 안정성 검증 결과

오늘 기준으로 PoC 녹화 안정성 완료 조건은 다음 범위에서 검증되었다.

- Python Mock Robot 2대가 같은 mission room에 WebRTC publish
- recorder-worker가 mission room 전체 robot track/data 수신
- active mission의 recording chunk가 `recording` 상태로 생성
- 임무 종료 시 마지막 chunk가 `uploaded` 상태로 마감
- RGB MP4, Thermal MP4, Telemetry JSONL, manifest object가 MinIO에 저장
- replay API에서 robot별 최신 chunk와 재생 가능한 MP4 URL 반환
- 오래 열린 recording session을 재사용하지 않고 새 송출은 새 recording session의 `chunkIndex = 0`부터 시작

검증에 사용한 임무:

- `mission-017`
- `robot-001`, `robot-002`

검증 결과:

- `recordingChunkCount = 0`
- `uploadedChunkCount = 1` per robot
- RGB MP4 MinIO URL `200 OK`

아직 완료가 아닌 이유:

- 전체 VMS 기능 기준으로 이벤트 bookmark, clip export, 멀티채널 동기 재생은 아직 구현 범위가 열려 있다.
- 운영 환경에서 recorder-worker scale-out 시 중복 처리 방지와 장애 복구 정책을 더 다듬어야 한다.

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
