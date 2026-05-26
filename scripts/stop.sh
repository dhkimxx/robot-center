#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

printf 'stopping app processes...\n'
stop_local_processes

printf 'stopping docker services...\n'
docker compose -f "$COMPOSE_FILE" stop postgres minio turn >/dev/null 2>&1 || true

printf 'stopped\n'
