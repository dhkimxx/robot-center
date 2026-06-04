---
title: "recording-replay-storage"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "design"
tags: ["feature", "recording", "replay", "storage", "minio"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: initial feature split from rescue robot functional reference"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: split rescue robot function candidates into feature docs'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: clarify feature/imported document priority against current implementation baseline'
---

# Recording, Replay, Storage

## 목적

관제 서버 recorder-worker, PostgreSQL, MinIO 저장 흐름과 향후 로봇 local save/sync 후보를 기능 관점에서 정리한다.

## 현재 기반 우선 원칙

본 문서는 녹화/리플레이 확장 후보이며, 현재 구현된 recorder-worker, recording finalization job, PostgreSQL/MinIO 저장 규칙을 우선한다.

로봇 local save/sync는 외부 로봇팀과 별도 계약이 필요한 후보이며, 현재 Robot Gateway 필수 계약으로 간주하지 않는다.

## 저장 대상 후보

- RGB MP4
- Thermal MP4
- Audio
- Telemetry/GPS JSONL
- Sensor JSONL
- Event log
- Mission log
- PointCloud object reference
- Terrain analysis object/reference
- Recording manifest

## Critical Event 보존 원칙

```text
Storage Full
  -> Low Priority 삭제
  -> Critical Event 유지
  -> Operator Alert
```

critical 후보:

- `EMERGENCY_STOP`
- `GAS_HAZARD_DETECTED`
- `VICTIM_CANDIDATE_DETECTED`
- `NETWORK_DISCONNECTED`
- `SLAM_DRIFT_DETECTED`

## Local Save / Sync Queue 후보

실제 로봇 연동 시 후보 흐름:

```text
Network Failure
  -> Robot Local Save
  -> Sync Queue
  -> Recovery 후 Control Center Sync
```

관제 서버 기준 후보:

- recorder-worker upload retry
- finalization job 중복 방지
- partially uploaded file 정합성 검증
- manifest와 media object 연결 검증
- replay에서 chunk/file 상태 오표시 방지

## Replay 후보

- 종료된 mission에서만 replay UI 표시
- robot별 recording session 그룹화
- 최신 chunk부터 표시
- MP4는 modal player로 재생
- event bookmark와 video timestamp 연결
- critical event 중심 필터

## 우리 프로젝트 반영 순서

### P0

- live 화면에서는 recording live status를 live-status SSOT 기준으로만 표시한다.
- replay는 종료된 mission 안에서 확인하는 현재 방향을 유지한다.

### P1

- critical event bookmark와 chunk timestamp를 연결한다.
- recording finalization 상태가 replay에 명확히 반영되게 한다.

### P2

- 로봇 local save/sync queue 계약을 로봇팀 인터페이스 후보로 협의한다.
