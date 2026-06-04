---
title: "robot-health-monitoring"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "design"
tags: ["feature", "robot", "health", "monitoring", "maintenance"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: initial feature split from rescue robot functional reference"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: split rescue robot function candidates into feature docs'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: clarify feature/imported document priority against current implementation baseline'
---

# Robot Health Monitoring

## 목적

로봇 장비 상태, live 품질, 서버/저장소 상태, 예방정비 후보를 하나의 운영 관점으로 정리한다.

## 현재 기반 우선 원칙

본 문서는 로봇 health 확장 후보이며, 현재 live 화면 상태 기준은 `/api/v1/operator/missions/{missionCode}/live-status` SSOT를 우선한다.

로봇 장비 online/offline, mission stream, recording 상태는 서로 다른 상태로 유지한다.

## 수집 항목 후보

| 항목 | 활용 |
| --- | --- |
| CPU/GPU/Memory | robot/system health |
| Temperature | 과열/성능 저하 판단 |
| Battery | 임무 지속 가능 시간 |
| Network | live 품질, reconnect 필요성 |
| FPS/Latency | 영상 품질 상태 |
| Storage | 녹화/재생 가능성 |
| Motor torque/vibration/noise | 예방정비 후보 |

## 상태 분리 원칙

```text
Robot online 상태 != Mission streaming 상태 != Recording 상태
```

구분:

- connection: heartbeat 또는 로봇 장비 상태
- stream: SFU observed publisher 상태
- recording: recorder runtime과 recording finalization 상태
- sensor: DataChannel payload freshness와 descriptor/sample 상태

## Alert Level 후보

| Level | 의미 |
| --- | --- |
| info | 정상 범위 알림 |
| warning | 성능 저하 가능 |
| critical | 임무 영향 가능 |

## 우리 프로젝트 반영 순서

### P0

- live-status SSOT로 connection/stream/recording 상태를 분리한다.
- robot detail은 online/offline과 lastSeen 중심으로 유지한다.

### P1

- sensor descriptor에 compute/network/battery/storage 계열 타입을 추가한다.
- system page에서 service health와 robot health를 혼동하지 않게 분리한다.

### P2

- trend 기반 maintenance recommendation을 AI Agent 후보로 연결한다.
