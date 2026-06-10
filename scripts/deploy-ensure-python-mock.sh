#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$1"
APP_SERVER_URL="$2"
MISSION_CODE="$3"
ROBOT_CODE="$4"

cd "$ROOT_DIR"
source scripts/common.sh

python3 - "$APP_SERVER_URL" "$MISSION_CODE" "$ROBOT_CODE" <<'PY'
import json
import sys
import urllib.request

base_url, mission_code, robot_code = sys.argv[1:4]
with urllib.request.urlopen(f"{base_url}/api/v1/operator/missions", timeout=10) as response:
    missions = json.load(response).get("missions", [])

target = next((mission for mission in missions if mission.get("missionCode") == mission_code), None)
if not target:
    raise SystemExit(f"mission not found: {mission_code}")
if target.get("status") != "active":
    raise SystemExit(f"mission is not active: {mission_code} status={target.get('status')}")
robot_codes = set(target.get("robotCodes") or [])
if target.get("robotCode"):
    robot_codes.add(target.get("robotCode"))
if robot_code not in robot_codes:
    raise SystemExit(f"robot {robot_code} is not assigned to mission {mission_code}")
PY

PYTHON_VENV_DIR="$ROOT_DIR/.runtime/python-mock-robot-venv"
PYTHON_BIN="$PYTHON_VENV_DIR/bin/python"
if [[ ! -x "$PYTHON_BIN" ]]; then
  python3 -m venv "$PYTHON_VENV_DIR"
  "$PYTHON_BIN" -m pip install --upgrade pip >/dev/null
  "$PYTHON_BIN" -m pip install -r "$ROOT_DIR/apps/mock-robot-python/requirements.txt" >/dev/null
fi

ROBOT_TOKEN="$(curl -fsS "$APP_SERVER_URL/api/v1/operator/robots/$ROBOT_CODE/connection-info" | /usr/bin/python3 -c 'import json,sys; print(json.load(sys.stdin)["connectionInfo"]["robotToken"])')"
SESSION_SUFFIX="${ROBOT_CODE#robot-}"
SESSION_NAME="robot-center-pyrobot-$SESSION_SUFFIX"
LEGACY_SESSION_NAME="robot-center-pyrobot-$ROBOT_CODE"
TOKEN_ENV_FILE="$ROOT_DIR/.runtime/$SESSION_NAME.env"

list_existing_mock_pids() {
  local pid command robot_code_value
  ps -axo pid=,command= \
    | awk '$0 ~ /mock_robot[.]py/ || $0 ~ /runpy[.]run_path[(]"mock_robot[.]py"/ { print $1 }' \
    | while IFS= read -r pid; do
      [[ -n "$pid" ]] || continue
      command="$(ps -p "$pid" -o command= 2>/dev/null || true)"
      if [[ "$command" == *"--robot-code"* && "$command" == *"$ROBOT_CODE"* ]]; then
        printf '%s\n' "$pid"
        continue
      fi
      if [[ -r "/proc/$pid/environ" ]]; then
        robot_code_value="$(tr '\0' '\n' < "/proc/$pid/environ" | sed -n 's/^ROBOT_CODE=//p' | head -n 1)"
        if [[ "$robot_code_value" == "$ROBOT_CODE" ]]; then
          printf '%s\n' "$pid"
        fi
      fi
    done
}

stop_screen_session "$SESSION_NAME"
if [[ "$LEGACY_SESSION_NAME" != "$SESSION_NAME" ]]; then
  stop_screen_session "$LEGACY_SESSION_NAME"
fi

EXISTING_MOCK_PIDS="$(list_existing_mock_pids | sort -u || true)"
if [[ -n "$EXISTING_MOCK_PIDS" ]]; then
  printf '%s\n' "$EXISTING_MOCK_PIDS" | xargs kill >/dev/null 2>&1 || true
  sleep 1
  EXISTING_MOCK_PIDS="$(list_existing_mock_pids | sort -u || true)"
  if [[ -n "$EXISTING_MOCK_PIDS" ]]; then
    printf '%s\n' "$EXISTING_MOCK_PIDS" | xargs kill -9 >/dev/null 2>&1 || true
  fi
fi

{
  printf 'APP_SERVER_URL=%q\n' "$APP_SERVER_URL"
  printf 'ROBOT_CODE=%q\n' "$ROBOT_CODE"
  printf 'ROBOT_TOKEN=%q\n' "$ROBOT_TOKEN"
} > "$TOKEN_ENV_FILE"
chmod 600 "$TOKEN_ENV_FILE"

start_screen_session "$SESSION_NAME" "cd '$ROOT_DIR/apps/mock-robot-python' && while true; do set -a; . '$TOKEN_ENV_FILE'; set +a; '$PYTHON_BIN' -c 'import os, runpy, sys; sys.argv = [\"mock_robot.py\", \"--server-url\", \"$APP_SERVER_URL\", \"--robot-code\", \"$ROBOT_CODE\", \"--robot-token\", os.environ[\"ROBOT_TOKEN\"], \"--rgb-width\", \"1280\", \"--rgb-height\", \"720\", \"--thermal-width\", \"640\", \"--thermal-height\", \"480\", \"--fps\", \"15\"]; runpy.run_path(\"mock_robot.py\", run_name=\"__main__\")'; exit_code=\$?; printf '[%s] %s mock exited with %s; restarting in 5s\n' \"\$(date -u +%Y-%m-%dT%H:%M:%SZ)\" '$ROBOT_CODE' \"\$exit_code\"; sleep 5; done"
printf 'python mock ensured: mission=%s robot=%s session=%s\n' "$MISSION_CODE" "$ROBOT_CODE" "$SESSION_NAME"
