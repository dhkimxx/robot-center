#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

DEV_SERVER_SSH="${DEV_SERVER_SSH:-danya@galaxybook}"
DEV_SERVER_PATH="${DEV_SERVER_PATH:-/home/danya/robot-center-dev}"
DEV_SERVER_PUBLIC_URL="${DEV_SERVER_PUBLIC_URL:-http://192.168.20.12:18080}"
DEV_SERVER_RECORDER_URL="${DEV_SERVER_RECORDER_URL:-http://192.168.20.12:18082}"
SSH_OPTS="${SSH_OPTS:--o StrictHostKeyChecking=accept-new}"
COMPOSE_ARGS="--env-file .env.dev-server -f deploy/docker-compose.yml --profile docker-turn"

require_command() {
  local command_name="$1"
  if ! command -v "$command_name" >/dev/null 2>&1; then
    printf '%s command not found\n' "$command_name" >&2
    exit 1
  fi
}

run_ssh() {
  # shellcheck disable=SC2086
  if [[ -n "${SSHPASS:-}" ]]; then
    require_command sshpass
    sshpass -e ssh $SSH_OPTS "$DEV_SERVER_SSH" "$@"
    return
  fi
  ssh $SSH_OPTS "$DEV_SERVER_SSH" "$@"
}

run_rsync() {
  local remote_shell="ssh $SSH_OPTS"
  if [[ -n "${SSHPASS:-}" ]]; then
    require_command sshpass
    remote_shell="sshpass -e ssh $SSH_OPTS"
  fi
  # shellcheck disable=SC2086
  rsync -az --delete -e "$remote_shell" \
    --exclude='.git/' \
    --exclude='.env.dev-server' \
    --exclude='.runtime/' \
    --exclude='.agents/' \
    --exclude='.playwright-cli/' \
    --exclude='.playwright-mcp/' \
    --exclude='.DS_Store' \
    --exclude='node_modules/' \
    --exclude='apps/web/node_modules/' \
    --exclude='apps/web/.vite/' \
    --exclude='apps/android-robot/.gradle/' \
    --exclude='apps/android-robot/build/' \
    --exclude='apps/android-robot/app/build/' \
    --exclude='apps/server/.runtime/' \
    "$ROOT_DIR/" "$DEV_SERVER_SSH:$DEV_SERVER_PATH/"
}

require_command ssh
require_command rsync

if [[ "${SKIP_WEB_BUILD:-0}" != "1" ]]; then
  printf 'building web...\n'
  (cd "$ROOT_DIR/apps/web" && npm run build)
fi

printf 'ensuring remote path: %s:%s\n' "$DEV_SERVER_SSH" "$DEV_SERVER_PATH"
run_ssh "mkdir -p '$DEV_SERVER_PATH'"

printf 'syncing files...\n'
run_rsync

printf 'deploying docker services...\n'
run_ssh "set -euo pipefail
cd '$DEV_SERVER_PATH'
rm -rf server .DS_Store .playwright-cli .playwright-mcp
docker compose $COMPOSE_ARGS up -d --build postgres minio turn app-server recorder-worker
docker compose $COMPOSE_ARGS up -d --force-recreate app-server recorder-worker
docker compose $COMPOSE_ARGS ps
curl -fsS http://127.0.0.1:18080/healthz
curl -fsS http://127.0.0.1:18082/healthz >/dev/null
"

printf '\nready\n'
printf 'UI: %s\n' "$DEV_SERVER_PUBLIC_URL"
printf 'recorder health: %s/healthz\n' "$DEV_SERVER_RECORDER_URL"
