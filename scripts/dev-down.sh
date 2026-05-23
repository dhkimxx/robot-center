#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/dev-common.sh"

printf 'stopping app processes...\n'
stop_local_processes

if command -v adb >/dev/null 2>&1; then
  adb shell am force-stop "$ROBOT_PACKAGE" >/dev/null 2>&1 || true
fi

printf 'stopping docker services...\n'
docker compose -f "$COMPOSE_FILE" stop postgres minio turn >/dev/null 2>&1 || true

printf 'stopped\n'
