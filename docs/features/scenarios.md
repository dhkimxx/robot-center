---
title: "scenarios"
created: 2026-05-26
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "design"
status: "planned"
tags: ["scenario", "rescue", "requirements", "robot", "ai"]
history:
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: moved into docs/planned lifecycle structure"
- '2026-05-26 danya.kim <danya.kim@thundersoft.com>: updated scenario maintenance reference to docs index'
- '2026-05-26 danya.kim <danya.kim@thundersoft.com>: removed dependency on docs index file'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: align scenario planning with stable implementation baseline'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: reorganize docs directories by document purpose'
---

# Scenarios

> Canonical document for rescue scenario requirements.

## 문서 목적

AI Web팀이 담당할 관제 웹, 서버 API, DB, 이벤트 처리, LLM/SOP AI 에이전트의 요구사항을 도출하기 위한 실증 시나리오 문서이다.

현재 1차 실증 시나리오는 다음 3가지로 정의한다.

- 산악조난
- 붕괴현장
- 지하시설

시나리오는 추후 추가될 수 있으므로, 모든 시나리오는 동일한 구조로 관리한다.

## 현재 기반 우선 원칙

본 문서는 시나리오 요구사항을 도출하기 위한 planned 문서이며, 구현 기준을 새로 정의하지 않는다.

구현 기준과 계약은 다음 stable 문서를 우선한다.

- `docs/stable/architecture.md`
- `docs/stable/robot-interface.md`
- `docs/stable/data-storage.md`

따라서 본 문서의 기능 후보가 stable 문서 또는 현재 코드와 충돌하면 stable/코드 기준을 우선한다. 특히 WebRTC signaling, track/DataChannel 이름, API namespace, 저장 규칙은 현재 구현 기반을 기준선으로 둔다.

## 공통 전제

### 사용자

- 지휘관: 임무 현황을 판단하고 구조 우선순위와 대응 지시를 결정한다.
- 관제요원: 로봇 상태, 영상, 센서, 탐지 결과를 모니터링하고 원격 제어를 수행한다.
- 구조대원: 현장 진입 또는 구조 활동을 수행하며 AI Web의 상황 요약과 SOP 제안을 참고한다.
- 관리자: 사용자, 권한, 로봇, 임무, 시스템 설정을 관리한다.

### 시스템 구성

- React Operator UI: 임무 현황, 로봇 상태, 영상, 센서, 탐지 결과, 알람, SOP 제안을 표시한다.
- Go app-server: actor별 API, 인증, 임무, 로봇 gateway, live status, SFU signaling, 이벤트, 탐지 결과, 로봇 제어 API를 제공한다.
- app-server 내부 Pion SFU: Robot stream을 mission room에 한 번 받아 Browser와 recorder-worker로 분배한다.
- recorder-worker: SFU subscriber로 붙어 영상, 센서, 이벤트 데이터를 수신하고 PostgreSQL/MinIO 저장을 담당한다.
- Live Status / DataChannel Relay: SFU observed stream, recorder runtime, heartbeat, DataChannel relay를 합성해 관제 UI 상태를 제공한다.
- Video Streaming: RGB/Thermal 영상 표시와 이벤트 시점 북마크를 지원한다.
- Sensor Ingestion: 가스, 온도, 습도, 위치, IMU, spatial summary 관련 데이터를 수신하고 저장한다.
- Event / Detection Processing 후보: 인명 탐지, 열원 탐지, 가스 이상치 등 탐지 결과를 이벤트로 구조화한다. P0 기준 별도 외부 서비스가 아니라 event DataChannel 또는 서버 처리 확장 후보로 둔다.
- LLM/SOP Agent: 상황 요약, 위험도 분석, SOP 매핑, 대응 시나리오 추천을 수행한다.
- Database/Storage: 임무, 로봇 상태, 센서 시계열, 이벤트, 탐지 결과, 영상 메타데이터를 저장한다.

### 현재 연동 합의 상태

Robot팀과 AI Web팀 사이의 세부 payload schema는 아직 확정되지 않았다.

현재 구현 기반으로 확정한 관제 측 기준은 다음과 같다.

- Robot은 `/api/v1/robot/*` REST API로 heartbeat와 active mission을 조회한다.
- Robot은 `/api/v1/robot/sfu/ws?room={missionCode}`로 app-server 내부 SFU에 WebRTC publish한다.
- Operator UI는 `/api/v1/operator/sfu/ws?room={missionCode}`로 mission room에 subscribe한다.
- recorder-worker는 `/api/v1/recorder/sfu/ws?room={missionCode}`로 mission room 전체를 subscribe한다.
- room id는 missionCode이다.
- Robot publisher identity는 robotToken으로 resolve된 robotCode를 사용한다.

WebRTC 기반 연동은 다음 canonical slot을 기준으로 한다.

- Media Track: `track.video_1`, `track.video_2`, `track.audio_1`, `track.audio_2`
- DataChannel: `channel.telemetry`, `channel.spatial`, `channel.event`, `channel.control`
- Signaling: role별 WebSocket endpoint, SDP 교환, ICE Candidate 교환, select-robot, 재연결

권장 WebRTC 참여 주체는 다음 4개 역할이다.

- Robot Peer: Jetson/로봇 측 WebRTC peer. canonical Media Track과 DataChannel을 publish한다.
- app-server 내부 SFU: Robot stream을 mission room에 한 번 받아 Browser와 recorder-worker로 분배한다.
- recorder-worker: SFU subscriber로 붙어 영상, 센서, 이벤트 데이터를 수신하고 PostgreSQL/MinIO 저장을 담당한다.
- Browser Peer: 관제 화면의 WebRTC peer. 선택한 robot bundle의 실시간 영상/음성/상태를 수신한다.

SFU와 Recorder Peer가 필요한 이유는 다음과 같다.

- 관제 브라우저가 닫혀 있어도 로봇 데이터 수신, 이벤트 생성, 영상 저장이 가능해야 한다.
- VMS 녹화, 스냅샷, 탐지 이벤트 북마크를 서버에서 안정적으로 생성해야 한다.
- LLM/SOP Agent와 Event / Detection Processing 후보가 실시간 데이터를 서버 내부에서 사용할 수 있어야 한다.
- 다수 관제 브라우저가 같은 로봇 스트림을 볼 수 있어야 한다.
- Robot은 한 번만 publish하고, subscriber 증가 부하는 SFU가 담당해야 한다.
- 제어 명령의 권한 확인, 감사 로그, ACK/NACK 추적을 서버에서 일관되게 처리해야 한다.

WebRTC만으로는 메시지 구조, 전송 주기, 재전송 정책, 우선순위, 타임스탬프 기준, 좌표계 기준이 자동으로 정의되지 않는다. 해당 항목은 별도 인터페이스 계약 문서에서 정의해야 한다.

### 공통 성공 기준

- 로봇 상태와 주요 센서 데이터가 관제 화면에 실시간으로 표시된다.
- 인명 탐지 또는 위험 이벤트 발생 시 알람이 생성되고 이력이 저장된다.
- 지휘관이 상황 요약과 SOP 제안을 확인할 수 있다.
- 관제요원이 E-Stop, 이동 명령, PTZ 제어 등 핵심 제어를 수행할 수 있다.
- 임무 종료 후 탐지 결과, 센서 변화, 이벤트 이력, 주요 판단 근거가 조회 가능하다.

## 시나리오 1. 산악조난

### 목적

산악 지형에서 실종자 또는 조난자를 탐색하고, 위치 후보와 생존 가능성을 관제 시스템에서 빠르게 판단한다.

### 시작 조건

- 수색 대상자의 마지막 목격 지점 또는 추정 수색 구역이 등록되어 있다.
- 로봇이 산악 지형 주행 가능한 상태로 배치되어 있다.
- 관제 시스템에 임무가 생성되어 있고 담당 지휘관과 관제요원이 배정되어 있다.

### 정상 흐름

1. 관제요원이 산악조난 임무를 생성하고 수색 구역을 설정한다.
2. 로봇이 지정 구역으로 이동하며 RGB/Thermal 영상과 위치, 배터리, 통신 상태를 전송한다.
3. AI가 RGB/Thermal 기반으로 사람 후보, 열원 후보, 움직임 후보를 탐지한다.
4. 탐지 후보가 발생하면 관제 화면에 위치, 영상 스냅샷, 신뢰도, 탐지 시간이 표시된다.
5. LLM/SOP Agent가 최근 탐지 결과와 주변 환경 정보를 바탕으로 상황을 요약한다.
6. 지휘관이 후보 위치를 확인하고 구조대원 투입 또는 추가 탐색 지시를 결정한다.
7. 임무 종료 후 탐지 이력, 이동 경로, 영상 북마크, 주요 이벤트가 보고서 형태로 정리된다.

### 예외 흐름

- 통신 품질 저하: 영상 품질을 낮추고 센서/상태 데이터 우선 전송 모드로 전환한다.
- 배터리 부족: 관제 화면에 복귀 권고 알람을 표시하고 Return-to-Home 명령을 제안한다.
- 탐지 신뢰도 낮음: 후보 이벤트로 분류하고 추가 관찰 또는 PTZ 줌인을 요청한다.
- 경사/장애물 주행 실패: 로봇 상태 이벤트를 생성하고 우회 경로 또는 수동 제어를 제안한다.

### 필요한 화면

- 임무 생성/설정 화면
- 로봇 실시간 상태 패널
- 지도 기반 위치/경로 화면
- RGB/Thermal 영상 멀티뷰
- 탐지 후보 타임라인
- 이벤트/알람 패널
- LLM 상황 요약 및 SOP 제안 패널
- 임무 결과 조회 화면

### 필요한 데이터

- 임무 정보: 임무 ID, 유형, 수색 구역, 시작/종료 시간, 담당자
- 로봇 상태: 위치, 배터리, 온도, 통신 품질, 주행 상태
- 영상 메타데이터: 채널, 타임스탬프, 해상도, 북마크
- 탐지 결과: 클래스, 신뢰도, 바운딩박스, 스냅샷, 위치, 탐지 시간
- 이벤트: 탐지, 통신 저하, 배터리 부족, 장애물, 제어 명령
- LLM 입력: 최근 이벤트, 탐지 후보, 위치, 센서 요약, 임무 맥락
- LLM 출력: 상황 요약, 위험도, 권고 조치, SOP 매핑 결과

### Robot팀 연동 지점

- WebRTC Media Track을 통한 RGB/Thermal 영상 스트림 제공 방식 정의
- WebRTC DataChannel을 통한 로봇 위치, 이동 경로, 배터리, 통신 품질, 주행 상태 메시지 정의
- 탐지 결과를 로봇 측에서 생성할지, AI Web 측에서 영상 기반 후처리할지 역할 분담 정의
- 이동 명령, 정지 명령, PTZ 제어 명령의 DataChannel 메시지 구조와 ACK/NACK 정책 정의
- 통신 저하 시 영상 품질 조정과 상태 데이터 우선 전송 기준 정의

### AI Web 핵심 요구사항

- 산악 수색 구역과 로봇 경로를 지도 위에서 확인할 수 있어야 한다.
- 인명/열원 후보 이벤트를 탐지 신뢰도와 함께 타임라인으로 제공해야 한다.
- 통신 저하 상황에서도 핵심 상태와 이벤트 데이터가 우선 표시되어야 한다.
- 지휘관이 후보 위치별 구조 우선순위를 판단할 수 있도록 요약 정보를 제공해야 한다.

## 시나리오 2. 붕괴현장

### 목적

건물 붕괴 또는 매몰 현장에서 로봇이 진입 가능한 구역을 탐색하고, 조난자 후보와 위험 요소를 식별한다.

### 시작 조건

- 붕괴 현장 임무가 생성되어 있다.
- 로봇이 붕괴 잔해, 협소 공간, 장애물 주변 탐색을 수행할 수 있는 상태이다.
- 지휘관은 현장 위험도가 높아 구조대원 직접 진입 전 로봇 선탐색을 요구한다.

### 정상 흐름

1. 관제요원이 붕괴현장 임무를 생성하고 진입 지점과 탐색 우선 구역을 설정한다.
2. 로봇이 현장 내부로 진입하며 RGB/Thermal, LiDAR, IMU, 가스/온습도 데이터를 전송한다.
3. 3D LiDAR 기반 포인트클라우드 또는 맵 정보가 관제 화면에 표시된다.
4. AI가 인명 후보, 열원 후보, 부분 가려짐 후보, 위험 온도, 가스 이상치를 탐지한다.
5. 탐지 이벤트가 발생하면 영상, 3D 위치, 센서값, 신뢰도가 함께 표시된다.
6. LLM/SOP Agent가 붕괴, 가스, 고온, 접근성 정보를 종합해 위험도와 대응 절차를 제안한다.
7. 지휘관은 구조대원 진입 여부, 우회 경로, 추가 탐색, 철수 여부를 결정한다.

### 예외 흐름

- 가스 농도 위험: 고위험 알람을 생성하고 SOP 기반 대피 또는 접근 제한 권고를 표시한다.
- 온도 급상승: 화재 확산 가능성 이벤트로 분류하고 위험도 상향을 제안한다.
- 통신 두절: 로봇 로컬 저장 모드와 자율 복귀 또는 정지 상태를 표시한다.
- 로봇 전도/고착: 긴급 상태 이벤트를 생성하고 수동 복구 절차를 안내한다.
- 탐지 위치 불확실: 2D 영상 좌표와 3D 맵 좌표의 신뢰도를 분리 표시한다.

### 필요한 화면

- 붕괴현장 임무 관제 화면
- RGB/Thermal 영상 멀티뷰
- 3D 맵/포인트클라우드 뷰
- 가스/온도/습도 센서 그래프
- 인명 탐지/열원 탐지 오버레이
- 위험 이벤트 알람 패널
- 원격 제어/E-Stop/PTZ 제어 패널
- SOP 추천 및 위험도 분석 패널

### 필요한 데이터

- 임무 정보: 진입 지점, 탐색 구역, 위험 구역, 담당자
- 로봇 상태: 자세, 속도, 배터리, 통신, 고착/전도 여부
- 센서 데이터: CO, CO2, H2S, CH4, O2, 온도, 습도, IMU
- LiDAR/맵 데이터: 포인트클라우드, 로봇 궤적, 탐지 위치
- 탐지 결과: 인명, 열원, 가려짐 후보, 신뢰도, 스냅샷, 좌표
- 제어 이력: 이동, 정지, E-Stop, PTZ, Waypoint, Return-to-Home
- LLM/SOP 데이터: 붕괴 현장 SOP, 위험도 기준, 최근 이벤트 요약

### Robot팀 연동 지점

- WebRTC Media Track을 통한 RGB/Thermal 영상 스트림 제공 방식 정의
- WebRTC DataChannel을 통한 LiDAR/SLAM 요약 데이터 또는 맵 참조 정보 전송 방식 정의
- RGB/Thermal 영상 타임스탬프와 탐지 좌표, 3D 위치의 동기화 기준 정의
- 가스/온도/습도 센서 샘플링 주기와 DataChannel 메시지 구조 정의
- E-Stop, Waypoint, Return-to-Home 명령의 우선순위, ACK/NACK, 실패 응답 기준 정의
- 통신 두절 시 로컬 저장 및 복구 업로드 정책 정의

### AI Web 핵심 요구사항

- 탐지 이벤트를 영상 좌표와 공간 좌표로 함께 추적할 수 있어야 한다.
- 고위험 이벤트는 일반 이벤트보다 높은 우선순위로 표시되어야 한다.
- 지휘관이 구조대원 진입 위험을 판단할 수 있도록 센서 추세와 SOP 근거를 제공해야 한다.
- 모든 제어 명령과 위험 이벤트는 감사 가능한 이력으로 저장되어야 한다.

## 시나리오 3. 지하시설

### 목적

지하 공동구, 전력구, 터널 등 GPS가 제한되는 시설에서 로봇이 내부 상태를 탐색하고 조난자, 유해가스, 고온, 통신 장애 위험을 관제한다.

### 시작 조건

- 지하시설 탐색 임무가 생성되어 있다.
- 시설 도면 또는 진입 경로 정보가 일부 등록되어 있다.
- GPS 대신 SLAM, Odometry, Waypoint 기반 위치 추정이 필요하다.

### 정상 흐름

1. 관제요원이 지하시설 임무를 생성하고 진입 지점과 목표 구역을 설정한다.
2. 로봇이 내부로 진입하며 영상, 센서, SLAM 기반 위치, 통신 품질을 전송한다.
3. 관제 화면은 도면 또는 3D 맵 위에 로봇 위치와 이동 경로를 표시한다.
4. AI가 인명 후보, 열원, 가스 이상치, 접근 불가 구역을 탐지한다.
5. 통신 품질이 낮아지면 시스템은 저대역폭 모드 또는 로컬 저장 모드 전환 여부를 표시한다.
6. LLM/SOP Agent가 지하시설 특성에 맞춰 질식, 폭발, 고온, 고립 위험을 요약한다.
7. 지휘관은 계속 탐색, 복귀, 구조대원 진입, 환기 또는 접근 제한 조치를 결정한다.

### 예외 흐름

- GPS 불가: SLAM/Odometry 기반 위치 신뢰도를 별도 표시한다.
- 통신 음영 구간: 영상 프레임을 낮추고 센서/상태 이벤트를 우선 전송한다.
- 가스 위험: 즉시 고위험 알람을 생성하고 SOP 기반 접근 제한을 제안한다.
- 맵 불일치: 도면과 SLAM 결과 차이를 표시하고 관제요원 확인이 필요함을 알린다.
- 복귀 경로 손실: 마지막 안정 위치와 로봇이 계산한 복귀 경로를 함께 표시한다.

### 필요한 화면

- 지하시설 임무 관제 화면
- 도면/맵 기반 로봇 위치 화면
- 영상/열화상 멀티뷰
- 통신 품질 모니터링 패널
- 센서 시계열 그래프
- 가스/온도 위험 알람
- Waypoint 및 Return-to-Home 제어 패널
- LLM 위험 요약 및 SOP 제안 패널

### 필요한 데이터

- 임무 정보: 시설 유형, 진입 지점, 목표 구역, 도면 정보
- 위치 데이터: SLAM 좌표, Odometry, Waypoint, 위치 신뢰도
- 통신 데이터: RSSI, 지연시간, 패킷 손실, 전송 모드
- 센서 데이터: 가스 농도, 온도, 습도, 산소 농도
- 탐지 결과: 인명 후보, 열원, 위험 구역, 접근 불가 구간
- 이벤트: 통신 저하, 위치 신뢰도 저하, 가스 위험, 복귀 권고
- LLM/SOP 데이터: 지하시설 대응 SOP, 위험 기준, 최근 센서 추세

### Robot팀 연동 지점

- WebRTC Media Track을 통한 영상/열화상 스트림 제공 방식 정의
- WebRTC DataChannel을 통한 SLAM 좌표, Odometry, Waypoint, 위치 신뢰도 메시지 정의
- GPS 미사용 환경의 위치 좌표계와 SLAM/도면 좌표 변환 방식 정의
- 통신 품질 지표와 저대역폭 모드 전환 기준 정의
- Waypoint, Return-to-Home, 통신 두절 시 동작 정책 정의
- 가스/산소 센서 위험 임계값과 이벤트 발생 기준 정의

### AI Web 핵심 요구사항

- GPS 없이도 로봇의 상대 위치와 이동 경로를 이해할 수 있어야 한다.
- 통신 품질과 데이터 지연 상태를 관제요원이 즉시 판단할 수 있어야 한다.
- 가스, 산소, 온도 위험을 실시간 이벤트와 SOP 제안으로 연결해야 한다.
- 지하시설 특성상 복귀 경로와 마지막 안정 위치를 명확히 표시해야 한다.

## 시나리오 추가 규칙

추후 시나리오를 추가할 때는 아래 항목을 동일하게 작성한다.

- 목적
- 시작 조건
- 정상 흐름
- 예외 흐름
- 필요한 화면
- 필요한 데이터
- Robot팀 연동 지점
- AI Web 핵심 요구사항

시나리오가 추가되면 stable 아키텍처, Robot Gateway 계약, 저장 구조 문서에서 다음 항목을 함께 갱신한다.

- 신규 화면 또는 기존 화면 확장 여부
- 신규 이벤트 타입
- 신규 센서/탐지 데이터
- 신규 Robot팀 연동 계약
- LLM/SOP 입력 JSON Schema 변경 여부
- KPI 또는 검증 항목 변경 여부

## 인터페이스 계약 전 협의 필요 항목

WebRTC 기반 연동 구조와 canonical slot은 현재 구현 기준을 따른다. 다음 항목은 Robot팀과 별도 협의가 필요하다.

- Media Track 품질 metadata: 해상도, FPS, bitrate, codec profile, keyframe 정책
- `channel.telemetry` payload: SensorDescriptor/SensorSample 상세 schema, sampling rate, 단위, quality/status 표현
- `channel.spatial` payload: SLAM, odometry, point cloud summary, object reference, 대용량 데이터 전송 방식
- `channel.event` payload: alarm, fault, detection, mission event, robot event의 공통 필드와 severity 정책
- `channel.control` 정책: 명령 schema, 권한, ACK/NACK, timeout, retry, audit log, rate limit
- DataChannel 신뢰성 옵션: ordered, unordered, maxRetransmits 설정
- 메시지 공통 필드: messageId, timestamp, sequence, schemaVersion
- 시간 기준: Robot 장비 시간, 서버 수신 시간, NTP/PTP 동기화 기준
- 좌표계 기준: GPS, SLAM map, robot local, world 좌표 변환 방식
- 제어 명령 정책: E-Stop 우선순위, 중복 명령 처리, ACK/NACK, timeout, retry
- 장애 처리: 연결 끊김, 재연결, 로컬 저장, 지연 데이터 업로드 정책
- 보안: Operator/recorder 인증, TURN credential, 폐쇄망 실증 환경 구성
