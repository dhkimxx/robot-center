---
title: "dev-server-docker-deployment"
created: 2026-05-27
updated: '2026-05-27'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "runbook"
status: "planned"
tags: ["dev-server", "docker", "deployment", "ops", "runbook"]
history:
- "2026-05-27 danya.kim <danya.kim@thundersoft.com>: split deployment procedure from robot team WebRTC send guide"
- '2026-05-27 danya.kim <danya.kim@thundersoft.com>: rename SFU WebSocket env to base URL and document operator endpoint'
---

# Dev Server Docker Deployment

## 1. 목적

이 문서는 관제팀이 임시 개발서버에 `robot-center` 관제 스택을 Docker로 배포하는 절차를 정의한다.

로봇팀 연동 절차와 WebRTC 송신 계약은 `docs/planned/robot-team-webrtc-send-test-guide.md`를 따른다.

## 2. 서버 기준

```text
host: 192.168.20.12
os: Ubuntu 24.04
deploy user: danya
deploy path: /home/danya/robot-center
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
| `49160-49180` | TCP/UDP | TURN relay range |
| `9000` | TCP | MinIO object download |
| `9001` | TCP | MinIO console, 필요 시 |

## 4. 배포 파일 준비

권장 경로:

```bash
mkdir -p /home/danya/robot-center
cd /home/danya/robot-center
```

코드는 `git clone`, `rsync`, archive upload 중 하나로 준비한다. 서버에는 Node/npm이 없으므로 web build는 다음 중 하나를 선택한다.

Option A: 로컬에서 `apps/web/dist`를 빌드한 뒤 서버로 복사한다.

```bash
cd /Users/dhkim/workspace/sst/robot-center/apps/web
npm ci
npm run build
```

Option B: 서버에 Node/npm을 설치하고 서버에서 빌드한다.

P0 임시 배포에서는 서버 패키지 설치를 줄이기 위해 Option A를 우선한다.

## 5. Env 파일

서버에서 `.env.dev-server`를 만든다. 이 파일은 git에 commit하지 않는다.

```bash
cat > .env.dev-server <<'EOF'
APP_ENV=development

APP_SERVER_HOST_PORT=18080
APP_SERVER_PUBLIC_URL=http://192.168.20.12:18080
SFU_WS_BASE_URL=ws://192.168.20.12:18080

RECORDER_WORKER_HOST_PORT=18082
RECORDER_WORKER_POLL_INTERVAL=5s
RECORDING_CHUNK_DURATION=10m

POSTGRES_HOST_PORT=5432
POSTGRES_DB=robot_center
POSTGRES_USER=robot_center
POSTGRES_PASSWORD=robot_center

MINIO_API_HOST_PORT=9000
MINIO_CONSOLE_HOST_PORT=9001
MINIO_BUCKET=robot-center-poc
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin

TURN_HOST_PORT=3478
TURN_DOCKER_EXTERNAL_IP=192.168.20.12
TURN_URL=turn:192.168.20.12:3478?transport=udp
TURN_USERNAME=robot
TURN_PASSWORD=robot-pass
TURN_REALM=robot-center.local
TURN_RELAY_MIN_PORT=49160
TURN_RELAY_MAX_PORT=49180
EOF
```

## 6. 서비스 기동

`turn` service는 `docker-turn` profile에 묶여 있으므로 profile을 켠다.

```bash
cd /home/danya/robot-center

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

## 7. Health 확인

서버 내부에서 확인:

```bash
curl -fsS http://127.0.0.1:18080/healthz | python3 -m json.tool
curl -fsS http://127.0.0.1:18080/api/system/status | python3 -m json.tool
curl -fsS http://127.0.0.1:18080/api/rtc-config | python3 -m json.tool
curl -fsS http://127.0.0.1:18082/healthz | python3 -m json.tool
```

외부 클라이언트에서 확인:

```bash
curl -fsS http://192.168.20.12:18080/healthz
curl -fsS http://192.168.20.12:18080/api/rtc-config
```

`/api/rtc-config`는 다음 public address를 반환해야 한다.

```text
signalingUrl: ws://192.168.20.12:18080/sfu/operator/ws
TURN URL: turn:192.168.20.12:3478?transport=udp
iceTransportPolicy: relay
```

## 8. 로그 확인

```bash
docker compose --env-file .env.dev-server -f deploy/docker-compose.yml --profile docker-turn logs -f app-server
docker compose --env-file .env.dev-server -f deploy/docker-compose.yml --profile docker-turn logs -f recorder-worker
docker compose --env-file .env.dev-server -f deploy/docker-compose.yml --profile docker-turn logs -f turn
```

## 9. 관제 데이터 준비

배포가 끝난 뒤 관제팀은 Web UI 또는 REST API로 테스트 robot과 mission을 준비한다.

Robot 생성:

```bash
curl -fsS -X POST http://127.0.0.1:18080/api/robots \
  -H 'Content-Type: application/json' \
  -d '{"displayName":"Robot Team Test 1","modelName":"Robot Team Gateway"}' \
  | python3 -m json.tool
```

Mission 생성:

```bash
curl -fsS -X POST http://127.0.0.1:18080/api/missions \
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
curl -fsS -X POST http://127.0.0.1:18080/api/missions/mission-001/start \
  | python3 -m json.tool
```

로봇팀에는 `serverUrl`, `robotCode`, `robotToken`만 전달한다. token은 문서, git, 공개 채팅에 남기지 않는다.

## 10. 종료와 초기화

테스트 mission 종료:

```bash
curl -fsS -X POST http://127.0.0.1:18080/api/missions/mission-001/end \
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

## 11. 배포 완료 기준

- `app-server`, `recorder-worker`, `turn`, `postgres`, `minio` container가 running/healthy
- `http://192.168.20.12:18080` UI 접속 가능
- `/api/system/status` OK
- `/api/rtc-config`가 `192.168.20.12` public URL을 반환
- robot 생성과 connection-info 조회 가능
- mission 생성과 start 가능
