---
title: "dev-server-docker-deployment"
created: 2026-05-27
updated: '2026-06-09'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "runbook"
status: "planned"
tags: ["dev-server", "docker", "deployment", "ops", "runbook"]
history:
- "2026-05-27 danya.kim <danya.kim@thundersoft.com>: split deployment procedure from robot team WebRTC send guide"
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: rename SFU WebSocket env to base URL and document operator endpoint'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: separate public robot-facing and internal docker WebRTC addresses'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: document docker TURN NAT mapping and verified recorder runtime volume'
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: expand TURN relay range for repeated WebRTC reconnect tests'
- '2026-06-01 danya.kim <danya.kim@thundersoft.com>: update dev-server verification endpoints to role-based /api/v1 namespaces'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: reorganize docs directories by document purpose'
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: absorb multi-robot SFU regression checks into deployment verification'
- '2026-06-09 danya.kim <danya.kim@thundersoft.com>: standardize dev-server rsync deployment script'
- '2026-06-09 danya.kim <danya.kim@thundersoft.com>: document deploy artifact exclusions and local runtime cleanup'
- '2026-06-09 danya.kim <danya.kim@thundersoft.com>: document deploy verification harness'
- '2026-06-09 danya.kim <danya.kim@thundersoft.com>: document WebRTC smoke option for deploy verification'
- '2026-06-09 danya.kim <danya.kim@thundersoft.com>: document deploy verification summary report'
- '2026-06-09 danya.kim <danya.kim@thundersoft.com>: document browser smoke option for deploy verification'
---

# Dev Server Docker Deployment

## 1. 목적

이 문서는 관제팀이 임시 개발서버에 `robot-center` 관제 스택을 Docker로 배포하는 절차를 정의한다.

로봇팀 연동 절차와 WebRTC 송신 계약은 `docs/guides/robot-team-webrtc-send-test-guide.md`를 따른다.

## 2. 서버 기준

```text
host: 192.168.20.12
os: Ubuntu 24.04
deploy user: danya
deploy path: /home/danya/robot-center-dev
```

현재 사전 조사 기준:

- Docker Engine 설치됨
- Docker Compose plugin 설치됨
- `danya` 계정은 `docker` group에 포함됨
- UFW inactive
- CPU/RAM/Disk 여유 충분
- Node/npm은 설치되어 있지 않음
- 기존 `registry:2` container가 `5000/tcp`를 사용 중이며, 본 배포 포트와 충돌하지 않음

## 3. 공개 포트

| Port | Protocol | Purpose |
| --- | --- | --- |
| `18080` | TCP | Web UI / REST API / SFU WebSocket |
| `18082` | TCP | recorder-worker health, 관제팀 확인용 |
| `3478` | TCP/UDP | TURN |
| `49160-49300` | TCP/UDP | TURN relay range |
| `9000` | TCP | MinIO object download |
| `9001` | TCP | MinIO console, 필요 시 |

## 4. 배포 파일 준비

권장 경로:

```bash
mkdir -p /home/danya/robot-center-dev
cd /home/danya/robot-center-dev
```

현재 개발서버 배포는 로컬에서 web build 후 `rsync`로 서버 디렉터리를 동기화하는 방식을 표준으로 한다. 서버에는 Node/npm이 없으므로 web build는 로컬에서 수행한다.

권장 배포 명령:

```bash
cd /Users/dhkim/workspace/sst/robot-center
./scripts/deploy-dev-server.sh
```

password 방식 SSH 환경에서는 다음처럼 실행할 수 있다.

```bash
SSHPASS='...' ./scripts/deploy-dev-server.sh
```

수동 배포가 필요할 때만 로컬에서 `apps/web/dist`를 빌드한 뒤 서버에 복사한다.

배포 동기화 대상에서는 git metadata, secret env, 로컬 runtime, Playwright 산출물, `node_modules`, Android build output을 제외한다. 개발서버에는 Docker 실행에 필요한 소스, 배포 설정, 빌드된 web static 파일만 올린다.

### 4.1 배포검증 하네스

반복 작업에서 사용하는 `배포검증`은 로컬 확인, commit/push, 개발서버 배포, 배포 후 상태 점검을 하나의 흐름으로 묶은 관제팀 내부 절차다.

변경사항을 commit/push하고 개발서버까지 반영할 때:

```bash
cd /Users/dhkim/workspace/sst/robot-center
SSHPASS='...' ./scripts/deploy-verify.sh --commit-message "Describe change"
```

이미 commit/push가 끝난 변경을 다시 배포하고 확인할 때:

```bash
cd /Users/dhkim/workspace/sst/robot-center
SSHPASS='...' ./scripts/deploy-verify.sh --no-commit
```

`--no-commit`으로 실제 배포할 때는 작업 트리가 깨끗해야 한다. 미커밋 변경이 있으면 하네스가 배포를 중단한다.

하네스가 수행하는 기본 확인:

- `git diff --check`
- 변경 범위에 따른 script syntax, server test, Swagger 최신성, web test/build
- `scripts/deploy-dev-server.sh`를 통한 개발서버 Docker 배포
- `healthz`, `/api/v1/system/status`, `/swagger/openapi.json`, `/api/v1/operator/rtc-config`, `/system` route 확인
- 최근 `app-server`, `recorder-worker` 로그의 panic/fatal/error 계열 확인

하네스는 secret을 저장하지 않는다. SSH password가 필요한 환경에서는 `SSHPASS` 환경변수로만 주입한다.

WebRTC 관련 변경은 API 기반 smoke 확인을 추가한다.

```bash
SSHPASS='...' ./scripts/deploy-verify.sh \
  --no-commit \
  --webrtc-smoke \
  --webrtc-smoke-mission mission-054 \
  --webrtc-smoke-min-robots 1
```

`--webrtc-smoke`는 개발서버의 `/api/v1/system/status`와 mission `live-status`를 읽어서 다음을 확인한다.

- SFU publisher가 `publishing`이고 ICE state가 `connected` 또는 `completed`
- `track.video_1`, `track.video_2`, `track.audio_1` 수신
- `channel.telemetry`, `channel.spatial`, `channel.event`, `channel.control` open
- telemetry message와 live media/data timestamp가 최근 값
- live-status 기준 robot connection이 `online`, stream이 `streaming`

녹화 상태까지 함께 확인하려면 `--webrtc-smoke-require-recording`을 추가한다.

관제 화면 렌더링이나 영상 표시 경로를 바꾼 경우에는 브라우저 기반 smoke 확인을 추가한다.

```bash
SSHPASS='...' ./scripts/deploy-verify.sh \
  --no-commit \
  --browser-smoke \
  --browser-smoke-mission mission-054 \
  --browser-smoke-robot robot-042 \
  --browser-smoke-require-recording
```

`--browser-smoke`는 Playwright CLI로 mission control 화면을 열고 다음을 확인한다.

- 선택된 로봇이 요청한 `robotCode`와 일치
- RGB/Thermal `<video>`가 `MediaStream`을 가지고 있고 해상도와 `readyState`가 유효
- 센서 패널과 연결 상태가 화면에 표시됨
- `--browser-smoke-require-recording` 지정 시 녹화 중 상태가 화면에 표시됨

하네스는 마지막에 요약 리포트를 한 번 출력한다. 성공 시에는 단계별 상태와 확인 URL, API/WebRTC 상세 결과가 남는다.

```text
==> deploy verification summary
result:            ok
branch:            main
commit:            c0de684
localChecks:       ok
commitStep:        skipped
pushStep:          skipped
deploy:            skipped
postDeploy:        skipped
logScan:           skipped
webrtcSmoke:       ok
browserSmoke:      ok
ui:                http://192.168.20.12:18080
recorder:          http://192.168.20.12:18082/healthz
details:
  - webrtc smoke: mission=mission-054 passed robots=robot-042,robot-043,robot-045
  - browser smoke: mission=mission-054 robot=robot-042 rgb=640x360 readyState=4; thermal=640x360 readyState=4
```

실패 시에는 `result=failed`와 `failedStep`을 먼저 보고, 아래 `details`에서 실패 원인을 확인한다.

```text
==> deploy verification summary
result:            failed
failedStep:        WebRTC smoke
webrtcSmoke:       failed
details:
  - WebRTC smoke: failed
  - webrtc smoke failed: no SFU room found for mission-999
```

브라우저에서 실제 영상이 보이는지, 녹화 파일이 재생되는지, Android/Python/GStreamer Mock Robot을 새로 기동하는 회귀 확인은 변경 성격에 따라 하네스 이후 별도로 수행한다.

## 5. Env 파일

서버에서 `.env.dev-server`를 만든다. 이 파일은 git에 commit하지 않는다.

```bash
cat > .env.dev-server <<'EOF'
APP_ENV=development

APP_SERVER_HOST_PORT=18080
APP_SERVER_PUBLIC_URL=http://192.168.20.12:18080
APP_SERVER_INTERNAL_URL=http://app-server:8080
SFU_WS_PUBLIC_BASE_URL=ws://192.168.20.12:18080
SFU_WS_INTERNAL_BASE_URL=ws://app-server:8080

RECORDER_WORKER_HOST_PORT=18082
RECORDER_WORKER_INTERNAL_URL=http://recorder-worker:8082
RECORDER_WORKER_POLL_INTERVAL=5s
RECORDING_CHUNK_DURATION=10m
RECORDING_MEDIA_IDLE_TIMEOUT=2m

POSTGRES_HOST_PORT=15432
POSTGRES_DB=robot_center
POSTGRES_USER=robot_center
POSTGRES_PASSWORD=robot_center_dev

MINIO_API_HOST_PORT=19000
MINIO_CONSOLE_HOST_PORT=19001
MINIO_INTERNAL_URL=http://minio:9000
MINIO_PUBLIC_URL=http://192.168.20.12:19000
MINIO_BUCKET=robot-center
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin_dev

TURN_HOST_PORT=3478
TURN_DOCKER_EXTERNAL_IP=192.168.20.12
TURN_DOCKER_RELAY_IP=172.28.0.10
DOCKER_BRIDGE_SUBNET=172.28.0.0/16
TURN_PUBLIC_URL=turn:192.168.20.12:3478?transport=udp
TURN_INTERNAL_URL=turn:turn:3478?transport=udp
TURN_USERNAME=robot-center-turn
TURN_PASSWORD=rc-turn-2026-0527
TURN_REALM=robot-center.dev
TURN_RELAY_MIN_PORT=49160
TURN_RELAY_MAX_PORT=49300
EOF
```

주소 경계:

- 로봇팀/브라우저에는 `192.168.20.12` public address만 공유한다.
- Docker 내부 WebRTC peer(app-server SFU, recorder-worker)는 `turn`/`app-server` service DNS를 사용한다.
- `TURN_PUBLIC_URL`과 `TURN_INTERNAL_URL`은 같은 coturn을 가리키지만, 접속 경로가 다르므로 합치지 않는다.
- Docker bridge 안의 coturn은 NAT 뒤에 있으므로 `TURN_DOCKER_EXTERNAL_IP/TURN_DOCKER_RELAY_IP`를 `public/private` 매핑으로 둔다.
- `TURN_DOCKER_RELAY_IP`는 compose default network 안에서 고정되는 coturn container IP다. `DOCKER_BRIDGE_SUBNET`과 충돌하면 둘을 같이 바꾼다.

## 6. 서비스 기동

일반 배포는 `scripts/deploy-dev-server.sh`를 사용한다. `turn` service는 `docker-turn` profile에 묶여 있으므로 수동 기동 시에도 profile을 켠다.

```bash
cd /home/danya/robot-center-dev

docker compose \
  --env-file .env.dev-server \
  -f deploy/docker-compose.yml \
  --profile docker-turn \
  up -d --build postgres minio turn app-server recorder-worker
```

상태 확인:

```bash
docker compose \
  --env-file .env.dev-server \
  -f deploy/docker-compose.yml \
  --profile docker-turn \
  ps
```

## 7. 로컬 산출물 정리

로컬에서 WebRTC/녹화 테스트를 반복하면 `apps/server/.runtime/recordings`가 크게 늘어난다. 삭제 전 dry-run으로 대상을 확인한다.

```bash
./scripts/clean-local-runtime.sh
```

문제가 없으면 실제 삭제를 실행한다.

```bash
./scripts/clean-local-runtime.sh --apply
```

Python Mock Robot 가상환경까지 다시 만들고 싶을 때만 `--python-venv`를 추가한다.

## 8. Health 확인

서버 내부에서 확인:

```bash
curl -fsS http://127.0.0.1:18080/healthz | python3 -m json.tool
curl -fsS http://127.0.0.1:18080/api/v1/system/status | python3 -m json.tool
curl -fsS http://127.0.0.1:18080/api/v1/operator/rtc-config | python3 -m json.tool
curl -fsS http://127.0.0.1:18082/healthz | python3 -m json.tool
```

외부 클라이언트에서 확인:

```bash
curl -fsS http://192.168.20.12:18080/healthz
curl -fsS http://192.168.20.12:18080/api/v1/operator/rtc-config
```

`/api/v1/operator/rtc-config`는 다음 public address를 반환해야 한다.

```text
signalingUrl: ws://192.168.20.12:18080/api/v1/operator/sfu/ws
TURN URL: turn:192.168.20.12:3478?transport=udp
iceTransportPolicy: relay
```

## 9. 로그 확인

```bash
docker compose --env-file .env.dev-server -f deploy/docker-compose.yml --profile docker-turn logs -f app-server
docker compose --env-file .env.dev-server -f deploy/docker-compose.yml --profile docker-turn logs -f recorder-worker
docker compose --env-file .env.dev-server -f deploy/docker-compose.yml --profile docker-turn logs -f turn
```

## 10. 관제 데이터 준비

배포가 끝난 뒤 관제팀은 Web UI 또는 REST API로 테스트 robot과 mission을 준비한다.

Robot 생성:

```bash
curl -fsS -X POST http://127.0.0.1:18080/api/v1/operator/robots \
  -H 'Content-Type: application/json' \
  -d '{"displayName":"Robot Team Test 1","modelName":"Robot Team Gateway"}' \
  | python3 -m json.tool
```

Mission 생성:

```bash
curl -fsS -X POST http://127.0.0.1:18080/api/v1/operator/missions \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Robot Team WebRTC Smoke",
    "missionType": "mountain_rescue",
    "siteNote": "temporary dev-server robot team send test",
    "robotCode": "robot-001"
  }' \
  | python3 -m json.tool
```

Mission 시작:

```bash
curl -fsS -X POST http://127.0.0.1:18080/api/v1/operator/missions/mission-001/start \
  | python3 -m json.tool
```

로봇팀에는 `serverUrl`, `robotCode`, `robotToken`만 전달한다. token은 문서, git, 공개 채팅에 남기지 않는다.

## 11. 종료와 초기화

테스트 mission 종료:

```bash
curl -fsS -X POST http://127.0.0.1:18080/api/v1/operator/missions/mission-001/end \
  | python3 -m json.tool
```

서비스만 중지:

```bash
docker compose \
  --env-file .env.dev-server \
  -f deploy/docker-compose.yml \
  --profile docker-turn \
  down
```

데이터까지 삭제해야 할 때만 volume을 삭제한다.

```bash
docker compose \
  --env-file .env.dev-server \
  -f deploy/docker-compose.yml \
  --profile docker-turn \
  down -v
```

## 12. 배포 완료 기준

- `app-server`, `recorder-worker`, `turn`, `postgres`, `minio` container가 running/healthy
- `http://192.168.20.12:18080` UI 접속 가능
- `/api/v1/system/status` OK
- `/swagger/index.html` 접속 가능
- `/swagger/openapi.json`에 `/api/v1/system/status`, `/api/v1/recorder/tick`, `/api/v1/operator/sensor-latest`, `/api/v1/operator/sfu/ws`, `/api/v1/recorder/sfu/ws` path 포함
- `/api/v1/operator/rtc-config`가 `192.168.20.12` public URL을 반환
- robot 생성과 connection-info 조회 가능
- mission 생성과 start 가능
- Android Robot 2대 기준 heartbeat, mission 조회, WebSocket join, relay ICE connected 확인
- mission room id가 `missionCode`이고 robot별 publisher가 subscriber 수와 무관하게 1개씩 유지됨
- operator A/B가 같은 mission room에서 서로 다른 robot을 선택해도 각자 선택한 robot의 RGB/Thermal/sensor만 표시됨
- operator 한 명의 robot 선택 변경/브라우저 종료가 다른 operator와 recorder-worker 수신을 끊지 않음
- recorder-worker가 같은 mission room의 모든 robot track/data를 수신하고 recording metadata를 robotCode별로 분리함
- recorder-worker health에서 `iceState=connected`, robot별 track/data 수신, append failure 0 확인
- recorder runtime volume(`/app/.runtime`)에 h264/ogg/jsonl 파일이 생성됨
