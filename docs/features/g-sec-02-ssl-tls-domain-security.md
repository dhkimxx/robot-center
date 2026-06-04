---
title: "G-SEC-02 SSL/TLS 및 도메인 보안"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "g_sec", "g-sec-02"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
---

# G-SEC-02 SSL/TLS 및 도메인 보안

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `G-SEC-02` |
| 중분류 | G. 보안 |
| 기능명 | SSL/TLS 및 도메인 보안 |
| 우선순위 | P0 |
| 난이도 | 중 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

운영 배포 시 HTTPS/WSS, TLS 설정, public endpoint 보안 기준

## 원본 XLSX 상세 설명

모든 웹 트래픽 HTTPS 강제(자동 Let's Encrypt 갱신), TLS 1.3 적용, HSTS 설정, 도메인 등록/갱신 자동화, DNSSEC 설정.

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/stable/architecture.md`
- `docs/plans/webrtc-turn-auth-plan.md`
