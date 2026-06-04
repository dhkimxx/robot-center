---
title: "scenario-validation"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "design"
tags: ["feature", "scenario", "validation", "rescue", "test"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: initial feature split from rescue robot functional reference"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: split rescue robot function candidates into feature docs'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: clarify feature/imported document priority against current implementation baseline'
---

# Scenario Validation

## 목적

산악조난, 붕괴현장, 지하시설 시나리오를 실제 기능 검증 기준으로 연결한다.

## 현재 기반 우선 원칙

본 문서는 검증 시나리오 확장 후보이며, 현재 PoC 검증 기준은 Python Mock Robot, app-server 내부 SFU, React 관제 UI, recorder-worker, PostgreSQL/MinIO 흐름을 우선한다.

실제 Jetson/ROS/GStreamer 기반 검증은 외부 로봇팀 연동 이후 단계로 둔다.

## 3대 시나리오 후보

| 시나리오 | 기능 흐름 후보 |
| --- | --- |
| 산악조난 | 경사 감지, Thermal 우선 생존자 후보, 지형 기반 저속/우회, 임무 성공/복귀 |
| 붕괴현장 | rough rubble, 협소 공간, Thermal+Audio 융합, gas danger 시 safe mode/retreat |
| 지하시설 | GPS loss, 통신 음영, CO/가스 위험, narrow passage, E-Stop/lockdown 후보 |

## Acceptance 후보

| 영역 | 후보 기준 |
| --- | --- |
| Streaming | RGB/Thermal stream 표시, Thermal 기준 15 FPS 이상 목표 |
| Communication | reconnect 10초 이내 목표 |
| Safety | E-Stop 즉시 정지 |
| Storage | critical event loss 없음 |
| Mission Creation | search area, search method, SOP profile 기반 임무 생성 가능 |
| Terrain | 3D LiDAR 지형 분석 기반 탐색 주행 가능 |

## Failure Injection 후보

- WebRTC disconnect
- storage full
- sensor failure
- SLAM drift
- DB failure
- robot control timeout
- AI timeout
- GPU/CUDA OOM

## Python Mock Robot 확장 후보

- scenario option: mountain/collapsed/tunnel
- scenario별 sensor descriptor/sample 재생
- scenario별 priority event 재생
- RGB/Thermal stream 유지
- mission room에 robot 2대 이상 동시 publish
- recorder-worker 저장 확인

## 우리 프로젝트 반영 순서

### P0

- Python Mock Robot 2대 + browser 관제 + recorder 저장 검증을 유지한다.

### P1

- scenario별 event/sensor replay 옵션을 추가한다.
- UI event timeline과 replay bookmark를 검증한다.

### P2

- Jetson/ROS/GStreamer 연동 후 field-like acceptance 테스트로 확장한다.
