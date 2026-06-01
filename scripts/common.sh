#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="$ROOT_DIR/deploy/docker-compose.yml"

APP_SESSION="robot-center-app"
RECORDER_SESSION="robot-center-recorder"
TURN_SESSION="robot-center-turn"
PYTHON_MOCK_SESSION_PREFIX="robot-center-pyrobot"
GSTREAMER_MOCK_SESSION_PREFIX="robot-center-gstrobot"

APP_PORT="${APP_PORT:-18080}"
RECORDER_PORT="${RECORDER_PORT:-18082}"
TURN_PORT="${TURN_PORT:-3478}"
TURN_RELAY_MIN_PORT="${TURN_RELAY_MIN_PORT:-49160}"
TURN_RELAY_MAX_PORT="${TURN_RELAY_MAX_PORT:-49180}"

POSTGRES_DB="${POSTGRES_DB:-robot_center}"
POSTGRES_USER="${POSTGRES_USER:-robot_center}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-robot_center}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"

MINIO_ENDPOINT="${MINIO_ENDPOINT:-http://127.0.0.1:9000}"
MINIO_BUCKET="${MINIO_BUCKET:-robot-center-poc}"
MINIO_ROOT_USER="${MINIO_ROOT_USER:-minioadmin}"
MINIO_ROOT_PASSWORD="${MINIO_ROOT_PASSWORD:-minioadmin}"

detect_host_ip() {
  if [[ -n "${HOST_IP:-}" ]]; then
    printf '%s\n' "$HOST_IP"
    return
  fi

  local detected_ip=""
  detected_ip="$(ipconfig getifaddr en0 2>/dev/null || true)"
  if [[ -z "$detected_ip" ]]; then
    detected_ip="$(ipconfig getifaddr en1 2>/dev/null || true)"
  fi
  if [[ -z "$detected_ip" ]]; then
    detected_ip="$(route -n get default 2>/dev/null | awk '/interface:/{print $2}' | xargs -I{} ipconfig getifaddr {} 2>/dev/null || true)"
  fi
  if [[ -z "$detected_ip" ]]; then
    detected_ip="127.0.0.1"
  fi
  printf '%s\n' "$detected_ip"
}

screen_exists() {
  if ! command -v screen >/dev/null 2>&1; then
    pid_session_running "$1"
    return
  fi
  local screen_output
  screen_output="$(screen -ls 2>/dev/null || true)"
  grep -q "[.]$1[[:space:]]" <<<"$screen_output"
}

runtime_pid_file() {
  local session_name="$1"
  printf '%s/.runtime/%s.pid\n' "$ROOT_DIR" "$session_name"
}

pid_session_running() {
  local session_name="$1"
  local pid_file
  pid_file="$(runtime_pid_file "$session_name")"
  [[ -s "$pid_file" ]] || return 1
  local pid
  pid="$(cat "$pid_file")"
  [[ -n "$pid" ]] || return 1
  kill -0 "$pid" >/dev/null 2>&1
}

stop_pid_session() {
  local session_name="$1"
  local pid_file
  pid_file="$(runtime_pid_file "$session_name")"
  [[ -s "$pid_file" ]] || return 0
  local pid
  pid="$(cat "$pid_file")"
  if [[ -n "$pid" ]] && kill -0 "$pid" >/dev/null 2>&1; then
    kill "$pid" >/dev/null 2>&1 || true
    sleep 1
    if kill -0 "$pid" >/dev/null 2>&1; then
      kill -9 "$pid" >/dev/null 2>&1 || true
    fi
  fi
  rm -f "$pid_file"
}

stop_screen_session() {
  local session_name="$1"
  if command -v screen >/dev/null 2>&1 && screen_exists "$session_name"; then
    screen -S "$session_name" -X quit || true
  fi
  stop_pid_session "$session_name"
}

kill_listeners_on_port() {
  local port="$1"
  local pids
  pids="$(lsof -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null || true)"
  pids="$pids"$'\n'"$(lsof -tiUDP:"$port" 2>/dev/null || true)"
  pids="$(printf '%s\n' "$pids" | awk 'NF' | sort -u)"
  if [[ -z "$pids" ]]; then
    return
  fi

  printf '%s\n' "$pids" | xargs kill >/dev/null 2>&1 || true
  sleep 1
  pids="$(printf '%s\n' "$pids" | awk 'NF' | xargs -I{} sh -c 'kill -0 "$1" 2>/dev/null && printf "%s\n" "$1"' sh {} || true)"
  if [[ -n "$pids" ]]; then
    printf '%s\n' "$pids" | xargs kill -9 >/dev/null 2>&1 || true
  fi
}

stop_local_processes() {
	stop_screen_session "$APP_SESSION"
	stop_screen_session "$RECORDER_SESSION"
	stop_screen_session "$TURN_SESSION"
	stop_python_mock_sessions
	stop_gstreamer_mock_sessions
	kill_listeners_on_port "$APP_PORT"
	kill_listeners_on_port "$RECORDER_PORT"
	kill_listeners_on_port "$TURN_PORT"
}

list_python_mock_sessions() {
  {
    if command -v screen >/dev/null 2>&1; then
      screen -ls 2>/dev/null || true
    fi
  } | awk -v prefix="$PYTHON_MOCK_SESSION_PREFIX" '
    $1 ~ "[.]" prefix {
      split($1, parts, ".")
      print parts[2]
    }
  '
  list_pid_sessions "$PYTHON_MOCK_SESSION_PREFIX"
}

stop_python_mock_sessions() {
  local session_name
  while IFS= read -r session_name; do
    [[ -z "$session_name" ]] && continue
    stop_screen_session "$session_name"
  done < <(list_python_mock_sessions)
}

list_gstreamer_mock_sessions() {
  {
    if command -v screen >/dev/null 2>&1; then
      screen -ls 2>/dev/null || true
    fi
  } | awk -v prefix="$GSTREAMER_MOCK_SESSION_PREFIX" '
    $1 ~ "[.]" prefix {
      split($1, parts, ".")
      print parts[2]
    }
  '
  list_pid_sessions "$GSTREAMER_MOCK_SESSION_PREFIX"
}

stop_gstreamer_mock_sessions() {
  local session_name
  while IFS= read -r session_name; do
    [[ -z "$session_name" ]] && continue
    stop_screen_session "$session_name"
  done < <(list_gstreamer_mock_sessions)
}

start_screen_session() {
  local session_name="$1"
  local command="$2"
  local log_file="$ROOT_DIR/.runtime/$session_name.log"
  local pid_file
  pid_file="$(runtime_pid_file "$session_name")"

  stop_screen_session "$session_name"
  mkdir -p "$ROOT_DIR/.runtime"
  : > "$log_file"
  if command -v screen >/dev/null 2>&1; then
    screen -dmS "$session_name" bash -lc "$command >> '$log_file' 2>&1"
    return
  fi
  nohup bash -lc "$command >> '$log_file' 2>&1" >/dev/null 2>&1 &
  printf '%s\n' "$!" > "$pid_file"
}

list_pid_sessions() {
  local prefix="$1"
  local runtime_dir="$ROOT_DIR/.runtime"
  [[ -d "$runtime_dir" ]] || return
  find "$runtime_dir" -maxdepth 1 -type f -name "$prefix*.pid" -exec basename {} .pid \; 2>/dev/null
}

wait_for_http() {
  local url="$1"
  local label="$2"
  local attempts="${3:-60}"

  for _ in $(seq 1 "$attempts"); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      printf '%s ready: %s\n' "$label" "$url"
      return 0
    fi
    sleep 1
  done

  printf '%s not ready: %s\n' "$label" "$url" >&2
  return 1
}
