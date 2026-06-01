#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

GSTREAMER_MOCK_DIR="$ROOT_DIR/apps/mock-robot-gstreamer"
GSTREAMER_MOCK_IMAGE="${GSTREAMER_MOCK_IMAGE:-robot-center-gstreamer-mock}"
GSTREAMER_MOCK_SESSION="${GSTREAMER_MOCK_SESSION:-robot-center-gstrobot-001}"
GSTREAMER_MOCK_CONTAINER="${GSTREAMER_MOCK_CONTAINER:-robot-center-gstrobot-001}"
APP_SERVER_URL="${APP_SERVER_URL:-}"
MOCK_GSTREAMER_ROBOT_CODE="${MOCK_GSTREAMER_ROBOT_CODE:-}"
MOCK_RGB_WIDTH="${MOCK_RGB_WIDTH:-640}"
MOCK_RGB_HEIGHT="${MOCK_RGB_HEIGHT:-360}"
MOCK_THERMAL_WIDTH="${MOCK_THERMAL_WIDTH:-640}"
MOCK_THERMAL_HEIGHT="${MOCK_THERMAL_HEIGHT:-360}"
MOCK_FPS="${MOCK_FPS:-15}"
MOCK_RESTART_DELAY_SECONDS="${MOCK_RESTART_DELAY_SECONDS:-5}"

json_get() {
  local code="$1"
  shift
  /usr/bin/python3 -c "$code" "$@"
}

require_app_server_url() {
  if [[ -z "$APP_SERVER_URL" ]]; then
    printf 'APP_SERVER_URL is required.\n' >&2
    printf 'example: APP_SERVER_URL=http://control-server.example:18080 %s\n' "$0" >&2
    exit 1
  fi
  APP_SERVER_URL="${APP_SERVER_URL%/}"
}

require_docker() {
  if ! command -v docker >/dev/null 2>&1; then
    printf 'docker command not found\n' >&2
    exit 1
  fi
}

create_robot() {
  curl -fsS -X POST "$APP_SERVER_URL/api/v1/operator/robots" \
    -H 'Content-Type: application/json' \
    -d '{"displayName":"GStreamer Mock Robot","modelName":"GStreamer webrtcbin"}' \
    | json_get 'import json,sys; print(json.load(sys.stdin)["robot"]["robotCode"])'
}

connection_token_for() {
  local robot_code="$1"
  curl -fsS "$APP_SERVER_URL/api/v1/operator/robots/$robot_code/connection-info" | json_get '
import json,sys
print(json.load(sys.stdin)["connectionInfo"]["robotToken"])
'
}

mission_code_for_robot_status() {
  local robot_code="$1"
  local status="$2"
  curl -fsS "$APP_SERVER_URL/api/v1/operator/missions" | json_get '
import json,sys
robot_code = sys.argv[1]
status = sys.argv[2]
missions = json.load(sys.stdin).get("missions", [])
for mission in missions:
    robot_codes = set(mission.get("robotCodes") or [])
    if mission.get("robotCode"):
        robot_codes.add(mission.get("robotCode"))
    if robot_code in robot_codes and mission.get("status") == status:
        print(mission["missionCode"])
        break
' "$robot_code" "$status"
}

create_mission_for_robot() {
  local robot_code="$1"
  curl -fsS -X POST "$APP_SERVER_URL/api/v1/operator/missions" \
    -H 'Content-Type: application/json' \
    -d '{"name":"GStreamer Mock Robot Demo","missionType":"mountain_rescue","siteNote":"created by gstreamer mock script","robotCode":"'"$robot_code"'"}' \
    | json_get 'import json,sys; print(json.load(sys.stdin)["mission"]["missionCode"])'
}

ensure_active_mission_for_robot() {
  local robot_code="$1"
  local mission_code
  mission_code="$(mission_code_for_robot_status "$robot_code" "active")"
  if [[ -n "$mission_code" ]]; then
    printf '%s\n' "$mission_code"
    return
  fi

  mission_code="$(mission_code_for_robot_status "$robot_code" "ready")"
  if [[ -z "$mission_code" ]]; then
    mission_code="$(create_mission_for_robot "$robot_code")"
  fi
  curl -fsS -X POST "$APP_SERVER_URL/api/v1/operator/missions/$mission_code/start" >/dev/null
  printf '%s\n' "$mission_code"
}

build_gstreamer_mock_image() {
  printf 'building GStreamer Mock Robot image: %s\n' "$GSTREAMER_MOCK_IMAGE"
  docker build -t "$GSTREAMER_MOCK_IMAGE" "$GSTREAMER_MOCK_DIR"
}

start_gstreamer_mock() {
  local robot_code="$1"
  local robot_token="$2"
  local env_file="$ROOT_DIR/.runtime/$GSTREAMER_MOCK_SESSION.env"
  mkdir -p "$ROOT_DIR/.runtime"
  printf 'ROBOT_TOKEN=%s\n' "$robot_token" > "$env_file"
  chmod 600 "$env_file"
  docker rm -f "$GSTREAMER_MOCK_CONTAINER" >/dev/null 2>&1 || true
  start_screen_session "$GSTREAMER_MOCK_SESSION" "while true; do \
  docker rm -f '$GSTREAMER_MOCK_CONTAINER' >/dev/null 2>&1 || true; \
  docker run --rm --network host --env-file '$env_file' --name '$GSTREAMER_MOCK_CONTAINER' '$GSTREAMER_MOCK_IMAGE' \
    --server-url '$APP_SERVER_URL' \
    --robot-code '$robot_code' \
    --rgb-width '$MOCK_RGB_WIDTH' \
    --rgb-height '$MOCK_RGB_HEIGHT' \
    --thermal-width '$MOCK_THERMAL_WIDTH' \
    --thermal-height '$MOCK_THERMAL_HEIGHT' \
    --fps '$MOCK_FPS'; \
  exit_code=\$?; \
  printf '[%s] gstreamer mock exited with %s; restarting in %ss\n' \"\$(date -u +%Y-%m-%dT%H:%M:%SZ)\" \"\$exit_code\" '$MOCK_RESTART_DELAY_SECONDS'; \
  sleep '$MOCK_RESTART_DELAY_SECONDS'; \
done"
}

require_app_server_url
require_docker
wait_for_http "$APP_SERVER_URL/healthz" "app-server"

robot_code="$MOCK_GSTREAMER_ROBOT_CODE"
if [[ -z "$robot_code" ]]; then
  robot_code="$(create_robot)"
fi
robot_token="$(connection_token_for "$robot_code")"
mission_code="$(ensure_active_mission_for_robot "$robot_code")"
build_gstreamer_mock_image
start_gstreamer_mock "$robot_code" "$robot_token"

printf '\nready\n'
printf 'control server: %s\n' "$APP_SERVER_URL"
printf 'mission: %s\n' "$mission_code"
printf 'robot: %s\n' "$robot_code"
if command -v screen >/dev/null 2>&1; then
  printf 'logs: screen -r %s\n' "$GSTREAMER_MOCK_SESSION"
else
  printf 'logs: tail -f %s/.runtime/%s.log\n' "$ROOT_DIR" "$GSTREAMER_MOCK_SESSION"
fi
