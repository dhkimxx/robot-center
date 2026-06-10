#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$1"
COMPOSE_ARGS="$2"
LOG_SINCE="$3"

normalize_output() {
  tr '\n' ' ' | sed 's/[[:space:]][[:space:]]*/ /g; s/^ //; s/ $//'
}

cd "$ROOT_DIR"

set +e
# shellcheck disable=SC2086
log_output="$(docker compose $COMPOSE_ARGS logs --since "$LOG_SINCE" app-server recorder-worker 2>&1)"
log_exit=$?
set -e

if [[ "$log_exit" != "0" ]]; then
  printf 'docker log command failed'
  if [[ -n "$log_output" ]]; then
    printf ': %s' "$(printf '%s' "$log_output" | normalize_output)"
  fi
  printf '\n'
  exit 10
fi

suspicious="$(printf '%s\n' "$log_output" | grep -Ei '(^|[^[:alpha:]])(panic|fatal|error|failed|timeout|refused|denied)([^[:alpha:]]|$)' || true)"
if [[ -n "$suspicious" ]]; then
  printf '%s\n' "$suspicious"
  printf '%s\n' 'remote logs contain suspicious entries'
  exit 2
fi

printf 'no suspicious app-server/recorder-worker logs since %s\n' "$LOG_SINCE"
