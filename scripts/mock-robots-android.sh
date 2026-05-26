#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

ANDROID_DIR="$ROOT_DIR/apps/android-robot"
APP_SERVER_URL="${APP_SERVER_URL:-}"
APK_PATH="$ANDROID_DIR/app/build/outputs/apk/debug/app-debug.apk"
ROBOT_PACKAGE="com.sst.robotcenter.androidrobot"
ROBOT_ACTIVITY="com.sst.robotcenter.androidrobot/.MainActivity"

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

require_adb() {
  if ! command -v adb >/dev/null 2>&1; then
    printf 'adb command not found\n' >&2
    exit 1
  fi
}

list_adb_devices() {
  adb devices | awk 'NR > 1 && $2 == "device" {print $1}'
}

ensure_robot_count() {
  local required_count="$1"
  local robot_count
  robot_count="$(curl -fsS "$APP_SERVER_URL/api/robots" | json_get 'import json,sys; print(len(json.load(sys.stdin).get("robots", [])))')"
  while (( robot_count < required_count )); do
    curl -fsS -X POST "$APP_SERVER_URL/api/robots" \
      -H 'Content-Type: application/json' \
      -d '{"displayName":"Android Mock Robot","modelName":"Android Mock"}' >/dev/null
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

select_mission_for_robots() {
  curl -fsS "$APP_SERVER_URL/api/missions" | json_get '
import json,sys
required = set(sys.argv[1:])
missions = json.load(sys.stdin).get("missions", [])
for status in ("active", "ready"):
    for mission in missions:
        if mission.get("status") != status:
            continue
        robot_codes = set(mission.get("robotCodes") or [])
        if mission.get("robotCode"):
            robot_codes.add(mission["robotCode"])
        if required.issubset(robot_codes):
            print(json.dumps(mission))
            raise SystemExit
' "$@"
}

mission_field() {
  local field="$1"
  json_get '
import json,sys
field = sys.argv[1]
payload = json.load(sys.stdin)
value = payload.get(field, "")
print("" if value is None else value)
' "$field"
}

create_shared_mission() {
  local payload
  payload="$(json_get '
import json,sys
robot_codes = [code for code in sys.argv[1:] if code]
print(json.dumps({
    "name": "Android Mock Robot Demo",
    "missionType": "mountain_rescue",
    "siteNote": "created by android mock script",
    "robotCode": robot_codes[0] if robot_codes else "",
    "robotCodes": robot_codes,
}))
' "$@")"
  curl -fsS -X POST "$APP_SERVER_URL/api/missions" \
    -H 'Content-Type: application/json' \
    -d "$payload" \
    | json_get 'import json,sys; print(json.dumps(json.load(sys.stdin)["mission"]))'
}

ensure_shared_active_mission() {
  local mission_payload
  mission_payload="$(select_mission_for_robots "$@")"
  if [[ -z "$mission_payload" ]]; then
    mission_payload="$(create_shared_mission "$@")"
  fi

  local mission_status
  mission_status="$(printf '%s' "$mission_payload" | mission_field status)"
  local mission_code
  mission_code="$(printf '%s' "$mission_payload" | mission_field missionCode)"
  if [[ "$mission_status" == "ready" ]]; then
    curl -fsS -X POST "$APP_SERVER_URL/api/missions/$mission_code/start" \
      | json_get 'import json,sys; print(json.dumps(json.load(sys.stdin)["mission"]))'
    return
  fi
  if [[ "$mission_status" == "active" ]]; then
    printf '%s\n' "$mission_payload"
    return
  fi
  printf 'mission %s is not startable: %s\n' "$mission_code" "$mission_status" >&2
  exit 1
}

build_android_apk() {
  printf 'building Android Mock Robot APK...\n'
  (cd "$ANDROID_DIR" && ./gradlew --no-daemon :app:assembleDebug)
}

start_android_robot() {
  local adb_serial="$1"
  local robot_code="$2"
  local robot_token="$3"
  printf 'starting %s on %s\n' "$robot_code" "$adb_serial"
  adb -s "$adb_serial" install -r "$APK_PATH" >/dev/null
  adb -s "$adb_serial" shell settings put global stay_on_while_plugged_in 7 >/dev/null 2>&1 || true
  adb -s "$adb_serial" shell am force-stop "$ROBOT_PACKAGE" >/dev/null 2>&1 || true
  adb -s "$adb_serial" shell am start -n "$ROBOT_ACTIVITY" \
    --es serverUrl "$APP_SERVER_URL" \
    --es robotCode "$robot_code" \
    --es robotToken "$robot_token" \
    --ez autoConnect true >/dev/null
}

require_app_server_url
require_adb
wait_for_http "$APP_SERVER_URL/healthz" "app-server"

mapfile -t adb_devices < <(list_adb_devices)
if (( ${#adb_devices[@]} == 0 )); then
  printf 'adb device unavailable\n' >&2
  exit 1
fi

ensure_robot_count "${#adb_devices[@]}"
robot_codes=()
for index in "${!adb_devices[@]}"; do
  robot_codes+=("$(robot_code_at "$index")")
done

mission_payload="$(ensure_shared_active_mission "${robot_codes[@]}")"
mission_code="$(printf '%s' "$mission_payload" | mission_field missionCode)"

build_android_apk
for index in "${!adb_devices[@]}"; do
  robot_code="${robot_codes[$index]}"
  robot_token="$(connection_token_for "$robot_code")"
  start_android_robot "${adb_devices[$index]}" "$robot_code" "$robot_token"
done

printf '\nready\n'
printf 'control server: %s\n' "$APP_SERVER_URL"
printf 'mission: %s\n' "$mission_code"
printf 'robots: %s\n' "${robot_codes[*]}"
