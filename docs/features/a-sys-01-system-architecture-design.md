---
title: "A-SYS-01 전체 시스템 아키텍처 설계"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "a_sys", "a-sys-01"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
---

# A-SYS-01 전체 시스템 아키텍처 설계

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `A-SYS-01` |
| 중분류 | A. 시스템 아키텍처 및 인터페이스 |
| 기능명 | 전체 시스템 아키텍처 설계 |
| 우선순위 | P0 |
| 난이도 | 중 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

로봇-관제 서버-SFU-recorder-worker-DB/MinIO-React UI 구성과 데이터 흐름 정의

## 원본 XLSX 상세 설명

로봇(사족보행)-엣지(Jetson)-서버-관제 전체 구조 설계. 구성요소 간 데이터 흐름도(DFD), 배포도, 컴포넌트 다이어그램 작성. 서브시스템 간 의존성 및 통신 경로 정의.

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/stable/architecture.md`
- `docs/stable/robot-interface.md`
- `docs/guides/dev-server-docker-deployment.md`
