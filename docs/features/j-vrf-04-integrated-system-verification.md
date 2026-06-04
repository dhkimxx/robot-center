---
title: "J-VRF-04 통합 시스템 검증"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "j_vrf", "j-vrf-04"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
---

# J-VRF-04 통합 시스템 검증

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `J-VRF-04` |
| 중분류 | J. 검증/실증 |
| 기능명 | 통합 시스템 검증 |
| 우선순위 | P0 |
| 난이도 | 중 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

Python Mock Robot -> app-server/SFU -> Web UI -> recorder-worker -> PostgreSQL/MinIO E2E 검증

## 원본 XLSX 상세 설명

End-to-End 통합 테스트(로봇→엣지→서버→관제), 스트레스 테스트(최대부하/장시간 운용), 장애 복구 테스트(통신두절/전원차단/센서고장), 보안 취약점 점검, 사용성 평가(소방대원 대상).

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/guides/dev-server-docker-deployment.md`
- `docs/guides/robot-team-webrtc-send-test-guide.md`
