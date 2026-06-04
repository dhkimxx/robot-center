---
title: "safety-failover"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "design"
tags: ["feature", "safety", "failover", "emergency-stop", "webrtc"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: initial feature split from rescue robot functional reference"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: split rescue robot function candidates into feature docs'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: clarify feature/imported document priority against current implementation baseline'
---

# Safety And Failover

## 목적

관제 시스템이 실제 재난 현장 운영에 가까워질 때 필요한 안전 우선순위, 장애 격리, 통신 단절 대응 후보를 정리한다.

## 현재 기반 우선 원칙

본 문서는 안전/장애 대응 확장 후보이며, 현재 구현된 제어 정책을 대체하지 않는다.

현재 `channel.control`은 예약 채널이고, 실제 제어 명령 schema, 권한, ACK/NACK, audit log는 별도 설계 후 구현한다.

## 안전 우선순위

```text
Human Safety
  > Robot Safety
  > Mission Continuity
  > Performance
```

## Mission 시작 전 Checklist 후보

- battery 상태
- sensor 상태
- storage 상태
- network 상태
- RGB/Thermal 출력
- 위치/GPS/SLAM 표시
- emergency stop 경로
- operator / safety officer / mission commander 역할

## Emergency Stop Trigger 후보

- operator stop
- collision risk
- gas hazard danger
- motor failure
- control authority violation

## 통신 상태 후보

| 상태 | 의미 | 후보 동작 |
| --- | --- | --- |
| `connected` | 정상 연결 | 실시간 관제/제어/event streaming |
| `degraded` | 품질 저하 | Thermal 우선, RGB bitrate 감소, point cloud drop/downsample |
| `disconnected` | 연결 단절 | reconnect attempt, local autonomous, critical local save |

## 장애 격리 원칙

```text
Single Failure != Whole Mission Failure
Remote Disconnect != Mission Stop
```

| 장애 | 기대 동작 후보 |
| --- | --- |
| RGB failure | Thermal 유지 |
| Thermal failure | RGB/Audio 보조 판단 |
| AI detector failure | 영상 송출 유지, detector 재시작 또는 fallback |
| WebRTC disconnect | robot local autonomous/local save 유지, reconnect 시도 |
| DB/storage failure | local save 또는 retry queue 유지 |
| SLAM drift | recovery/replan/safe mode |
| gas sensor failure | gas unknown event |

## Field Test 중단 조건 후보

- emergency stop 실패
- 통신 단절 후 local autonomous 전환 실패
- critical event loss 발생
- 제어 권한 검증 실패
- sensor/SLAM 장애로 safe mode 진입 실패
- gas danger, cliff/drop, blocked terrain에서 정지/우회 실패
- safety officer 또는 mission commander의 중단 결정

## 우리 프로젝트 반영 순서

### P0

- live-status에 connection/stream/recording 상태와 reason을 분리해 표현한다.
- 관제 UI에서 live 장애와 recorder/storage 장애를 섞지 않는다.

### P1

- event timeline에 priority event를 추가한다.
- Mission start 전 checklist UI 후보를 구체화한다.

### P2

- E-Stop command, audit log, 권한 검증을 control command 구현과 함께 설계한다.
