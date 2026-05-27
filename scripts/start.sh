#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

HOST_IP="$(detect_host_ip)"
PUBLIC_URL="${PUBLIC_URL:-http://$HOST_IP:$APP_PORT}"
SFU_WS_PUBLIC_URL="${SFU_WS_URL:-ws://$HOST_IP:$APP_PORT/sfu/ws}"
TURN_PUBLIC_IP_VALUE="${TURN_PUBLIC_IP:-$HOST_IP}"
TURN_URL_VALUE="${TURN_URL:-turn:$HOST_IP:$TURN_PORT?transport=udp}"

printf 'host ip: %s\n' "$HOST_IP"
printf 'public url: %s\n' "$PUBLIC_URL"
printf 'sfu ws url: %s\n' "$SFU_WS_PUBLIC_URL"
printf 'turn url: %s\n' "$TURN_URL_VALUE"
printf 'starting postgres/minio...\n'
docker compose -f "$COMPOSE_FILE" up -d postgres minio
printf 'waiting for postgres...\n'
for _ in $(seq 1 60); do
  if docker compose -f "$COMPOSE_FILE" exec -T postgres pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
if ! docker compose -f "$COMPOSE_FILE" exec -T postgres pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1; then
  printf 'postgres not ready\n' >&2
  exit 1
fi

printf 'waiting for minio...\n'
for _ in $(seq 1 60); do
  if docker compose -f "$COMPOSE_FILE" exec -T minio mc ready local >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
if ! docker compose -f "$COMPOSE_FILE" exec -T minio mc ready local >/dev/null 2>&1; then
  printf 'minio not ready\n' >&2
  exit 1
fi
docker compose -f "$COMPOSE_FILE" stop turn >/dev/null 2>&1 || true
stop_local_processes

if [[ "${SKIP_WEB_BUILD:-0}" != "1" ]]; then
  printf 'building web...\n'
  (cd "$ROOT_DIR/apps/web" && npm run build)
fi

printf 'starting host TURN...\n'
start_screen_session "$TURN_SESSION" "cd '$ROOT_DIR/apps/server' && \
TURN_PUBLIC_IP='$TURN_PUBLIC_IP_VALUE' \
TURN_HOST_PORT='$TURN_PORT' \
TURN_RELAY_MIN_PORT='$TURN_RELAY_MIN_PORT' \
TURN_RELAY_MAX_PORT='$TURN_RELAY_MAX_PORT' \
TURN_REALM='robot-center.local' \
TURN_USERNAME='robot' \
TURN_PASSWORD='robot-pass' \
go run ./cmd/turn-server"

printf 'starting app-server...\n'
start_screen_session "$APP_SESSION" "cd '$ROOT_DIR/apps/server' && \
APP_ENV='development' \
APP_SERVER_HTTP_ADDR=':$APP_PORT' \
APP_SERVER_PUBLIC_URL='$PUBLIC_URL' \
WEB_STATIC_DIR='$ROOT_DIR/apps/web/dist' \
RECORDER_WORKER_URL='http://127.0.0.1:$RECORDER_PORT' \
POSTGRES_HOST='127.0.0.1' \
POSTGRES_PORT='$POSTGRES_PORT' \
POSTGRES_DB='$POSTGRES_DB' \
POSTGRES_USER='$POSTGRES_USER' \
POSTGRES_PASSWORD='$POSTGRES_PASSWORD' \
MINIO_ENDPOINT='$MINIO_ENDPOINT' \
MINIO_BUCKET='$MINIO_BUCKET' \
SFU_WS_URL='$SFU_WS_PUBLIC_URL' \
TURN_URL='$TURN_URL_VALUE' \
TURN_USERNAME='robot' \
TURN_PASSWORD='robot-pass' \
go run ./cmd/app-server"

printf 'starting recorder-worker...\n'
start_screen_session "$RECORDER_SESSION" "cd '$ROOT_DIR/apps/server' && \
APP_ENV='development' \
RECORDER_WORKER_HTTP_ADDR=':$RECORDER_PORT' \
RECORDER_WORKER_POLL_INTERVAL='5s' \
RECORDING_CHUNK_DURATION='10m' \
APP_SERVER_PUBLIC_URL='http://127.0.0.1:$APP_PORT' \
POSTGRES_HOST='127.0.0.1' \
POSTGRES_PORT='$POSTGRES_PORT' \
POSTGRES_DB='$POSTGRES_DB' \
POSTGRES_USER='$POSTGRES_USER' \
POSTGRES_PASSWORD='$POSTGRES_PASSWORD' \
MINIO_ENDPOINT='$MINIO_ENDPOINT' \
MINIO_BUCKET='$MINIO_BUCKET' \
MINIO_ROOT_USER='$MINIO_ROOT_USER' \
MINIO_ROOT_PASSWORD='$MINIO_ROOT_PASSWORD' \
SFU_WS_URL='$SFU_WS_PUBLIC_URL' \
TURN_URL='$TURN_URL_VALUE' \
TURN_USERNAME='robot' \
TURN_PASSWORD='robot-pass' \
go run ./cmd/recorder-worker"

wait_for_http "http://127.0.0.1:$APP_PORT/healthz" "app-server"
wait_for_http "http://127.0.0.1:$RECORDER_PORT/healthz" "recorder-worker"

printf 'ensuring demo robot and active mission...\n'
robot_payload="$(curl -fsS 'http://127.0.0.1:'"$APP_PORT"'/api/robots')"
robot_code="$(printf '%s' "$robot_payload" | /usr/bin/python3 -c 'import json,sys; items=json.load(sys.stdin).get("robots", []); print(items[0]["robotCode"] if items else "")')"
if [[ -z "$robot_code" ]]; then
  robot_payload="$(curl -fsS -X POST 'http://127.0.0.1:'"$APP_PORT"'/api/robots' -H 'Content-Type: application/json' -d '{"displayName":"Bootstrap Robot","modelName":"PoC Bootstrap"}')"
  robot_code="$(printf '%s' "$robot_payload" | /usr/bin/python3 -c 'import json,sys; print(json.load(sys.stdin)["robot"]["robotCode"])')"
fi

mission_payload="$(curl -fsS 'http://127.0.0.1:'"$APP_PORT"'/api/missions')"
active_mission_code="$(printf '%s' "$mission_payload" | /usr/bin/python3 -c 'import json,sys; robot=sys.argv[1]; items=json.load(sys.stdin).get("missions", []); print(next((m["missionCode"] for m in items if m.get("robotCode")==robot and m.get("status")=="active"), ""))' "$robot_code")"
if [[ -z "$active_mission_code" ]]; then
  ready_mission_code="$(printf '%s' "$mission_payload" | /usr/bin/python3 -c 'import json,sys; robot=sys.argv[1]; items=json.load(sys.stdin).get("missions", []); print(next((m["missionCode"] for m in items if m.get("robotCode")==robot and m.get("status")=="ready"), ""))' "$robot_code")"
  if [[ -z "$ready_mission_code" ]]; then
    mission_payload="$(curl -fsS -X POST 'http://127.0.0.1:'"$APP_PORT"'/api/missions' -H 'Content-Type: application/json' -d '{"name":"P0 Integrated Demo","missionType":"mountain_rescue","siteNote":"created by dev-up","robotCode":"'"$robot_code"'"}')"
    ready_mission_code="$(printf '%s' "$mission_payload" | /usr/bin/python3 -c 'import json,sys; print(json.load(sys.stdin)["mission"]["missionCode"])')"
  fi
  curl -fsS -X POST 'http://127.0.0.1:'"$APP_PORT"'/api/missions/'"$ready_mission_code"'/start' >/dev/null
  active_mission_code="$ready_mission_code"
fi

printf 'demo robot: %s\n' "$robot_code"
printf 'active mission: %s\n' "$active_mission_code"

printf '\nready\n'
printf 'UI: %s\n' "$PUBLIC_URL"
printf 'MinIO API: http://%s:9000/%s\n' "$HOST_IP" "$MINIO_BUCKET"
printf 'MinIO console: http://127.0.0.1:9001\n'
printf 'Python Mock Robot: APP_SERVER_URL=%s %s\n' "$PUBLIC_URL" "$ROOT_DIR/scripts/mock-robots-python.sh"
printf 'status: %s\n' "$ROOT_DIR/scripts/status.sh"
