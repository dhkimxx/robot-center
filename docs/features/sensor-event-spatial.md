---
title: "sensor-event-spatial"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "design"
tags: ["feature", "sensor", "event", "spatial", "datachannel"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: initial feature split from rescue robot functional reference"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: split rescue robot function candidates into feature docs'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: clarify feature/imported document priority against current implementation baseline'
---

# Sensor, Event, Spatial Data

## 목적

로봇별 센서 수가 고정되지 않는 구조를 전제로, `SensorDescriptor / SensorSample`와 WebRTC DataChannel 역할을 기능 관점에서 정리한다.

## 현재 기반 우선 원칙

본 문서는 센서/이벤트/공간 데이터 확장 후보이며, 현재 canonical DataChannel과 data-storage 기준을 우선한다.

현재 확정된 채널 이름은 `channel.telemetry`, `channel.spatial`, `channel.event`, `channel.control`이며, payload 세부 schema는 stable 계약과 코드 구현을 기준으로 점진 확장한다.

## DataChannel 역할

| Channel | 후보 역할 |
| --- | --- |
| `channel.telemetry` | GPS, battery, network, gas, temperature, humidity 같은 저속 상태 |
| `channel.spatial` | IMU, odometry, point cloud reference, terrain analysis |
| `channel.event` | alarm, fault, detection, mission event |
| `channel.control` | command/ack 전용 후보 |

## SensorDescriptor 후보

필드 후보:

- sensorId
- displayName
- sensorType
- valueType
- unit
- channelRole
- sampleRateHz
- metadata
- firstSeenAt
- lastSeenAt

sensor type 후보:

- position
- imu
- odometry
- gas
- temperature
- humidity
- battery
- network
- point_cloud
- terrain
- custom

## SensorSample 후보

필드 후보:

- sensorId
- messageId
- sequence
- sentAt
- receivedAt
- numericValue
- textValue
- boolValue
- vectorValue
- objectValue
- objectKey
- rawPayload

## Priority Event 후보

| Event | Priority | 의미 |
| --- | --- | --- |
| `VICTIM_CANDIDATE_DETECTED` | critical | 인명 후보 탐지 |
| `GAS_HAZARD_DETECTED` | critical | 가스 위험 |
| `NETWORK_DISCONNECTED` | high | 통신 단절 |
| `SLAM_DRIFT_DETECTED` | high | 위치 추정 신뢰도 저하 |
| `EMERGENCY_STOP` | critical | 긴급 정지 |
| `TERRAIN_ANALYZED` | normal | 지형 분석 결과 |
| `SEARCH_DRIVE_PROFILE_SELECTED` | normal | 탐색/주행 정책 선택 |

## Terrain/Spatial 후보

Terrain class 후보:

- `FLAT_OPEN`
- `MILD_SLOPE`
- `STEEP_SLOPE`
- `ROUGH_RUBBLE`
- `NARROW_PASSAGE`
- `OBSTACLE_DENSE`
- `CLIFF_OR_DROP`
- `UNKNOWN`

Traversability 후보:

- `PASSABLE`
- `CAUTION`
- `REPLAN_REQUIRED`
- `BLOCKED`

## 우리 프로젝트 반영 순서

### P0

- 현재 sensor latest UI가 descriptor/sample 구조를 동적으로 표시하는지 점검한다.
- 지도는 `sensorType=position` 우선으로 표시한다.

### P1

- event timeline에 priority와 event type을 표준화한다.
- Python Mock Robot이 scenario별 sensor/event payload를 보낼 수 있게 확장한다.

### P2

- point cloud/terrain analysis는 MinIO object reference와 replay 표시 후보로 연결한다.
