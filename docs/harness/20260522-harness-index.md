---
title: "harness-index"
created: 2026-05-22
updated: '2026-05-23'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>", "dhkimxx <dhkimxx@naver.com>"]
type: "guide"
tags: ["harness", "webrtc", "server", "sfu"]
history:
- "2026-05-22 danya.kim <danya.kim@thundersoft.com>: initial entry"
- '2026-05-22 danya.kim <danya.kim@thundersoft.com>: migrated harness docs into meta-docs managed structure'
- '2026-05-22 danya.kim <danya.kim@thundersoft.com>: linked Go GORM persistence harness document'
- '2026-05-22 danya.kim <danya.kim@thundersoft.com>: clarified recorder chunk file format scope'
- '2026-05-22 dhkimxx <dhkimxx@naver.com>: added relay-only, storage_objects, and test/regression agent criteria'
- '2026-05-23 dhkimxx <dhkimxx@naver.com>: added mission-level multi-robot SFU checklist'
- '2026-05-23 dhkimxx <dhkimxx@naver.com>: marked selective subscribe signaling as internal PoC provisional'
---
# Harness Index

이 문서는 지금까지 시험으로 확정한 harness 문서의 진입점이다.

목표는 구현 코드나 상세 인터페이스 계약을 고정하는 것이 아니라, 이후 팀 간 인터페이스를 정하기 전에 공유할 수 있는 검증된 서버/SFU 흐름 기준을 남기는 것이다.

## 현재 정의

| 문서 | 역할 |
| --- | --- |
| `docs/harness/20260522-server-architecture.md` | 서버 기본 아키텍처 정의 |
| `docs/harness/20260522-go-gorm-persistence.md` | Go/GORM persistence 작성 기준 |
| `docs/harness/20260522-webrtc-sfu-topology.md` | WebRTC SFU 송/수신 구조 정의 |
| `docs/harness/20260523-multi-robot-sfu-checklist.md` | mission 단위 multi-robot SFU selective subscribe 검증 체크리스트 |

## 확정 범위

- Robot publisher들이 같은 `missionCode` room에 WebRTC publish 한다.
- Robot publisher는 subscriber 수와 무관하게 한 번만 publish 한다.
- Browser operator subscriber A/B가 같은 mission room에서 live 관제용으로 subscribe하고, 내부 PoC 기준 선택한 robotCode 하나만 수신한다.
- Recorder subscriber가 같은 mission room에서 저장용으로 모든 robotCode를 subscribe 한다.
- Robot별 데이터와 녹화 chunk는 `robotCode`로 구분한다.
- TURN은 ICE relay 인프라로만 사용한다.
- 현재 PoC/시연 기준 WebRTC peer connection은 relay-only 정책을 사용한다.
- Browser와 Recorder는 서로 독립적인 subscriber다.
- Recorder 장애가 Browser 관제를 막지 않는 구조를 목표로 한다.
- Browser 장애가 Recorder 저장을 막지 않는 구조를 목표로 한다.
- 업로드 완료된 녹화 파일의 metadata 기준 row는 `storage_objects`다.
- Selective subscribe signaling은 내부 PoC 임시 동작이며 장기 API/schema 계약으로 보지 않는다.

## 작업 에이전트 기준

계속 작업할 때는 역할을 4개 관점으로 나눠 검토한다.

| 역할 | 확인 대상 |
| --- | --- |
| Architecture / Integration | harness 문서, 레이어 구조, 도메인 경계 |
| Realtime Platform | WebRTC, SFU, TURN, recorder subscriber |
| Demo QA | 사용자가 바로 확인할 수 있는 시연 상태 |
| Test / Regression | Go test, API 회귀, DB/MinIO 정합성 |

## 이 문서에서 정의하지 않는 범위

아래 항목은 이 harness 문서들의 범위 밖이다.

- signaling message JSON schema
- selective subscribe request/response schema
- REST API request/response schema
- DataChannel 이름
- DataChannel payload JSON
- media track 이름
- media track metadata format
- codec, bitrate, resolution, framerate 정책
- recorder chunk 상세 파일 포맷
- replay manifest schema

상세 인터페이스는 필요할 때 별도 문서나 schema로 정의한다.
