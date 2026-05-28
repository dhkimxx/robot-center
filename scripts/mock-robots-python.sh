#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

MOCK_SESSION_PREFIX="robot-center-pyrobot"
PYTHON_MOCK_DIR="$ROOT_DIR/apps/mock-robot-python"
PYTHON_VENV_DIR="$ROOT_DIR/.runtime/python-mock-robot-venv"
BUNDLED_PYTHON="/Users/dhkim/.cache/codex-runtimes/codex-primary-runtime/dependencies/python/bin/python3"
if [[ -z "${PYTHON_BIN:-}" ]]; then
  if [[ -x "$BUNDLED_PYTHON" ]]; then
    PYTHON_BIN="$BUNDLED_PYTHON"
  else
    PYTHON_BIN="$(command -v python3)"
  fi
fi
APP_SERVER_URL="${APP_SERVER_URL:-}"
MOCK_ROBOT_COUNT="${MOCK_ROBOT_COUNT:-2}"
MOCK_SHARED_MISSION="${MOCK_SHARED_MISSION:-1}"
MOCK_MISSION_CODE="${MOCK_MISSION_CODE:-}"
MOCK_RGB_WIDTH="${MOCK_RGB_WIDTH:-1280}"
MOCK_RGB_HEIGHT="${MOCK_RGB_HEIGHT:-720}"
MOCK_THERMAL_WIDTH="${MOCK_THERMAL_WIDTH:-640}"
MOCK_THERMAL_HEIGHT="${MOCK_THERMAL_HEIGHT:-480}"
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

ensure_python_venv() {
  if [[ -x "$PYTHON_VENV_DIR/bin/python" ]]; then
    if ! "$PYTHON_VENV_DIR/bin/python" -c 'import sys; raise SystemExit(0 if sys.version_info >= (3, 10) else 1)' >/dev/null 2>&1; then
      rm -rf "$PYTHON_VENV_DIR"
    fi
  fi
  if [[ ! -x "$PYTHON_VENV_DIR/bin/python" ]]; then
    "$PYTHON_BIN" -m venv "$PYTHON_VENV_DIR"
  fi
  printf 'python: %s\n' "$("$PYTHON_VENV_DIR/bin/python" --version)"
  "$PYTHON_VENV_DIR/bin/python" -m pip install --upgrade pip >/dev/null
  "$PYTHON_VENV_DIR/bin/python" -m pip install -r "$PYTHON_MOCK_DIR/requirements.txt"
}

ensure_robot_count() {
  local robot_count
  robot_count="$(curl -fsS "$APP_SERVER_URL/api/robots" | json_get 'import json,sys; print(len(json.load(sys.stdin).get("robots", [])))')"
  while (( robot_count < MOCK_ROBOT_COUNT )); do
    curl -fsS -X POST "$APP_SERVER_URL/api/robots" \
      -H 'Content-Type: application/json' \
      -d '{"displayName":"Python Mock Robot","modelName":"Python Mock"}' >/dev/null
    robot_count=$((robot_count + 1))
  done
}

robot_code_at() {
  local index="$1"
  curl -fsS "$APP_SERVER_URL/api/robots" | json_get '
import json,sys
index = int(sys.argv[1])
robots = sorted(json.load(sys.stdin).get("robots", []), key=lambda item: item.get("robotCode", ""))
print(robots[index]["robotCode"])
' "$index"
}

connection_token_for() {
  local robot_code="$1"
  curl -fsS "$APP_SERVER_URL/api/robots/$robot_code/connection-info" | json_get '
import json,sys
print(json.load(sys.stdin)["connectionInfo"]["robotToken"])
'
}

active_mission_for_robot() {
  local robot_code="$1"
  curl -fsS "$APP_SERVER_URL/api/missions" | json_get '
import json,sys
robot_code = sys.argv[1]
missions = json.load(sys.stdin).get("missions", [])
for mission in missions:
    if mission.get("robotCode") == robot_code and mission.get("status") == "active":
        print(mission["missionCode"])
        break
' "$robot_code"
}

ready_mission_for_robot() {
  local robot_code="$1"
  curl -fsS "$APP_SERVER_URL/api/missions" | json_get '
import json,sys
robot_code = sys.argv[1]
missions = json.load(sys.stdin).get("missions", [])
for mission in missions:
    if mission.get("robotCode") == robot_code and mission.get("status") == "ready":
        print(mission["missionCode"])
        break
' "$robot_code"
}

mission_field() {
  local field="$1"
  json_get '
import json,sys
field = sys.argv[1]
value = json.load(sys.stdin).get(field, "")
print("" if value is None else value)
' "$field"
}

select_shared_mission() {
  local mission_code="$1"
  shift || true
  curl -fsS "$APP_SERVER_URL/api/missions" | json_get '
import json,sys
mission_code = sys.argv[1]
required_robot_codes = set(sys.argv[2:])
missions = json.load(sys.stdin).get("missions", [])
candidate = None
if mission_code:
    candidate = next((mission for mission in missions if mission.get("missionCode") == mission_code), None)
else:
    for status in ("active", "ready"):
        for mission in missions:
            if mission.get("status") != status:
                continue
            robot_codes = set(mission.get("robotCodes") or [])
            if mission.get("robotCode"):
                robot_codes.add(mission.get("robotCode"))
            if required_robot_codes.issubset(robot_codes):
                candidate = mission
                break
        if candidate is not None:
            break
if candidate is not None:
    print(json.dumps(candidate))
' "$mission_code" "$@"
}

create_shared_mission() {
  local robot_codes=("$@")
  local payload
  payload="$(json_get '
import json,sys
robot_codes = [code for code in sys.argv[1:] if code]
print(json.dumps({
    "name": "Python Multi-Robot Mock Demo",
    "missionType": "mountain_rescue",
    "siteNote": "created by python mock script for shared mission room",
    "robotCode": robot_codes[0] if robot_codes else "",
    "robotCodes": robot_codes,
}))
' "${robot_codes[@]}")"
  curl -fsS -X POST "$APP_SERVER_URL/api/missions" \
    -H 'Content-Type: application/json' \
    -d "$payload" \
    | json_get 'import json,sys; print(json.dumps(json.load(sys.stdin)["mission"]))'
}

end_conflicting_active_missions() {
  local robot_codes=("$@")
  local mission_codes
  mission_codes="$(curl -fsS "$APP_SERVER_URL/api/missions" | json_get '
import json,sys
required_robot_codes = set(sys.argv[1:])
missions = json.load(sys.stdin).get("missions", [])
conflicting = []
for mission in missions:
    if mission.get("status") != "active":
        continue
    mission_robot_codes = set(mission.get("robotCodes") or [])
    if mission.get("robotCode"):
        mission_robot_codes.add(mission.get("robotCode"))
    if not mission_robot_codes.intersection(required_robot_codes):
        continue
    if required_robot_codes.issubset(mission_robot_codes):
        continue
    mission_code = mission.get("missionCode")
    if mission_code:
        conflicting.append(mission_code)
print("\n".join(sorted(set(conflicting))))
' "${robot_codes[@]}")"
  if [[ -z "$mission_codes" ]]; then
    return
  fi
  while IFS= read -r mission_code; do
    [[ -z "$mission_code" ]] && continue
    printf 'ending conflicting active mission: %s\n' "$mission_code"
    curl -fsS -X POST "$APP_SERVER_URL/api/missions/$mission_code/end" >/dev/null
  done <<<"$mission_codes"
}

start_mission_payload() {
  local mission_code="$1"
  curl -fsS -X POST "$APP_SERVER_URL/api/missions/$mission_code/start" \
    | json_get 'import json,sys; print(json.dumps(json.load(sys.stdin)["mission"]))'
}

ensure_shared_active_mission() {
  local robot_codes=("$@")
  local mission_payload
  mission_payload="$(select_shared_mission "$MOCK_MISSION_CODE" "${robot_codes[@]}")"
  if [[ -z "$mission_payload" ]]; then
    mission_payload="$(create_shared_mission "${robot_codes[@]}")"
  fi

  local mission_status
  mission_status="$(printf '%s' "$mission_payload" | mission_field status)"
  local mission_code
  mission_code="$(printf '%s' "$mission_payload" | mission_field missionCode)"
  case "$mission_status" in
    active)
      printf '%s\n' "$mission_payload"
      ;;
    ready)
      start_mission_payload "$mission_code"
      ;;
    *)
      printf 'shared mission %s is not startable: %s\n' "$mission_code" "$mission_status" >&2
      return 1
      ;;
  esac
}

ensure_active_mission_for_robot() {
  local robot_code="$1"
  local active_mission_code
  active_mission_code="$(active_mission_for_robot "$robot_code")"
  if [[ -n "$active_mission_code" ]]; then
    printf '%s\n' "$active_mission_code"
    return
  fi

  local ready_mission_code
  ready_mission_code="$(ready_mission_for_robot "$robot_code")"
  if [[ -z "$ready_mission_code" ]]; then
    ready_mission_code="$(curl -fsS -X POST "$APP_SERVER_URL/api/missions" \
      -H 'Content-Type: application/json' \
      -d '{"name":"Python Mock Demo","missionType":"mountain_rescue","siteNote":"created by python mock script","robotCode":"'"$robot_code"'"}' \
      | json_get 'import json,sys; print(json.load(sys.stdin)["mission"]["missionCode"])')"
  fi
  curl -fsS -X POST "$APP_SERVER_URL/api/missions/$ready_mission_code/start" >/dev/null
  printf '%s\n' "$ready_mission_code"
}

start_mock_robot() {
  local index="$1"
  local robot_code="$2"
  local robot_token="$3"
  local session_name
  session_name="$(printf '%s-%03d' "$MOCK_SESSION_PREFIX" "$((index + 1))")"
  start_screen_session "$session_name" "cd '$PYTHON_MOCK_DIR' && \
while true; do \
  '$PYTHON_VENV_DIR/bin/python' mock_robot.py \
    --server-url '$APP_SERVER_URL' \
    --robot-code '$robot_code' \
    --robot-token '$robot_token' \
    --rgb-width '$MOCK_RGB_WIDTH' \
    --rgb-height '$MOCK_RGB_HEIGHT' \
    --thermal-width '$MOCK_THERMAL_WIDTH' \
    --thermal-height '$MOCK_THERMAL_HEIGHT' \
    --fps '$MOCK_FPS'; \
  exit_code=\$?; \
  printf '[%s] mock robot process exited with %s; restarting in %ss\n' \"\$(date -u +%Y-%m-%dT%H:%M:%SZ)\" \"\$exit_code\" '$MOCK_RESTART_DELAY_SECONDS'; \
  sleep '$MOCK_RESTART_DELAY_SECONDS'; \
done"
}

require_app_server_url
wait_for_http "$APP_SERVER_URL/healthz" "app-server"
ensure_python_venv
ensure_robot_count

printf 'starting python mock robots against %s\n' "$APP_SERVER_URL"
shared_mission_code=""
if [[ "$MOCK_SHARED_MISSION" == "1" ]]; then
  shared_robot_codes=()
  for index in $(seq 0 "$((MOCK_ROBOT_COUNT - 1))"); do
    shared_robot_codes+=("$(robot_code_at "$index")")
  done
  end_conflicting_active_missions "${shared_robot_codes[@]}"
  shared_mission_payload="$(ensure_shared_active_mission "${shared_robot_codes[@]}")"
  shared_mission_code="$(printf '%s' "$shared_mission_payload" | mission_field missionCode)"
  printf 'shared mission: %s / room %s\n' "$shared_mission_code" "$shared_mission_code"
fi

for index in $(seq 0 "$((MOCK_ROBOT_COUNT - 1))"); do
  robot_code="$(robot_code_at "$index")"
  robot_token="$(connection_token_for "$robot_code")"
  if [[ "$MOCK_SHARED_MISSION" == "1" ]]; then
    mission_code="$shared_mission_code"
    start_mock_robot "$index" "$robot_code" "$robot_token"
  else
    mission_code="$(ensure_active_mission_for_robot "$robot_code")"
    start_mock_robot "$index" "$robot_code" "$robot_token"
  fi
  printf '  %s -> %s\n' "$robot_code" "$mission_code"
done

printf '\nready\n'
printf 'control server: %s\n' "$APP_SERVER_URL"
printf 'logs: screen -r %s-001 / screen -r %s-002\n' "$MOCK_SESSION_PREFIX" "$MOCK_SESSION_PREFIX"
