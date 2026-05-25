# Mission Room Multi-Robot SFU 전환 결과

## 구현 범위

- mission room 기준을 `missionCode`로 전환했다.
- mission 생성/조회는 `robotCodes` 배열을 지원한다.
- 기존 `robotCode`는 단일 로봇 호환 필드로 유지한다.
- app-server 내부 Pion SFU는 mission room 안에서 robot별 publisher를 관리한다.
- recorder-worker는 mission room 하나를 subscribe하고 robotCode별 media/data를 분리 저장한다.
- Web UI는 mission 중심 관제 화면에서 robot tile 여러 개를 표시하고, 선택 robot의 RGB/Thermal/위치/센서를 표시한다.
- Python/Android mock 문서와 Python mock 실행 스크립트를 mission room 기준으로 맞췄다.
- harness 문서를 mission room multi-publisher/subscriber 구조로 갱신했다.

## 검증 명령

```bash
cd /Users/dhkim/workspace/sst/robot-center/apps/server
go test ./...
go vet ./...
```

```bash
cd /Users/dhkim/workspace/sst/robot-center/apps/web
npm run build
```

```bash
cd /Users/dhkim/workspace/sst/robot-center
python3 -m py_compile apps/mock-robot-python/mock_robot.py
bash -n scripts/python-mock-robots-up.sh
./scripts/dev-up.sh
MOCK_ROBOT_COUNT=2 MOCK_SHARED_MISSION=1 ./scripts/python-mock-robots-up.sh
./scripts/dev-status.sh
```

## 실제 확인 결과

- app-server health: OK
- recorder-worker health: OK
- PostgreSQL/MinIO Docker service: running
- TURN: host process running
- Python mock robot:
  - `robot-center-pyrobot-001`: running
  - `robot-center-pyrobot-002`: running
- active mission:
  - `mission-007`
  - assigned robots: `robot-001`, `robot-002`
- SFU room:
  - roomId: `mission-007`
  - robot publishers: 2
  - browser operator subscriber: 1
  - recorder subscriber: 1
  - published tracks: `robot-001` 3개, `robot-002` 3개
- Web UI:
  - `mission-007 / 로봇 2대` 표시 확인
  - mission room subscribe 후 연결 `2/2대` 표시 확인
  - `robot-001`, `robot-002` tile 선택 확인
  - RGB 1280x720, Thermal 640x480 영상 재생 확인
  - GPS/센서 데이터 수신 확인
- Recorder/Storage:
  - `mission-007 / robot-001` uploaded chunk 1개 확인
  - `mission-007 / robot-002` uploaded chunk 1개 확인
  - MinIO object key가 robotCode별로 분리됨 확인

## 사용자 확인 URL

- Web UI: `http://172.30.1.5:18080`
- MinIO console: `http://127.0.0.1:9001`
- 검증 스크린샷: `/tmp/robot-center-multi-robot-sfu.png`

## 남은 한계

- Android Mock Robot은 이번 최종 실행에서 ADB device unavailable 상태라 직접 실행 검증하지 못했다.
- DB에는 이전 시연에서 남은 active mission room이 일부 존재한다. 현재 검증 대상은 `mission-007`이다.
- JSON/DataChannel/track/codec 세부 포맷은 이번 요구사항에서 확정하지 않았으므로 기존 포맷에 robotCode 식별만 보강했다.
