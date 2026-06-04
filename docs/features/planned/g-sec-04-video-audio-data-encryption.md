---
title: "G-SEC-04 영상/음성 데이터 암호화"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "g_sec", "g-sec-04", "planned"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: classify feature document by implementation status'
---

# G-SEC-04 영상/음성 데이터 암호화

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `G-SEC-04` |
| 중분류 | G. 보안 |
| 기능명 | 영상/음성 데이터 암호화 |
| 우선순위 | P0 |
| 난이도 | 중 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

WebRTC SRTP/DTLS, 저장 object 접근권한, 녹화 파일 접근 감사 정책

## 원본 XLSX 상세 설명

전송 구간 암호화(SRTP/DTLS), 저장 데이터 암호화(AES-256), 암호화 키 관리(KMS 연동), 영상 워터마킹(무단유출 추적), 접근 로그 기록.

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/stable/architecture.md`
- `docs/plans/webrtc-turn-auth-plan.md`
