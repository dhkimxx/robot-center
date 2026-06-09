#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

DEV_SERVER_SSH="${DEV_SERVER_SSH:-danya@galaxybook}"
DEV_SERVER_PATH="${DEV_SERVER_PATH:-/home/danya/robot-center-dev}"
DEV_SERVER_PUBLIC_URL="${DEV_SERVER_PUBLIC_URL:-http://192.168.20.12:18080}"
DEV_SERVER_RECORDER_URL="${DEV_SERVER_RECORDER_URL:-http://192.168.20.12:18082}"
SSH_OPTS="${SSH_OPTS:--o StrictHostKeyChecking=accept-new}"
COMPOSE_ARGS="--env-file .env.dev-server -f deploy/docker-compose.yml --profile docker-turn"

COMMIT_MESSAGE=""
NO_COMMIT=0
SKIP_LOCAL_CHECKS=0
SKIP_DEPLOY=0
SKIP_POST_DEPLOY_CHECKS=0
SKIP_WEB_ROUTE_CHECK=0
FORCE_ALL_CHECKS=0
LOG_SINCE="${LOG_SINCE:-5m}"
WEB_BUILD_RAN=0

usage() {
  cat <<'EOF'
Usage:
  ./scripts/deploy-verify.sh --commit-message "commit message"
  ./scripts/deploy-verify.sh --no-commit

Options:
  --commit-message MESSAGE  Stage all current changes, commit, push, deploy, and verify.
  --no-commit              Skip commit and push. Use when the branch is already pushed.
  --skip-local-checks      Skip local tests/build checks.
  --skip-deploy            Skip dev-server deployment and run the remaining checks.
  --skip-post-deploy-checks
                           Skip deployed API/UI/log checks. Use only for harness dry-runs.
  --skip-web-route-check   Skip the deployed /system route reachability check.
  --all-checks             Run backend and frontend checks even if changed files do not require them.
  --since DURATION         Remote log window for post-deploy scan. Default: 5m.
  -h, --help               Show this help.

Environment:
  SSHPASS                  Optional sshpass password. Do not write it into docs or git.
  DEV_SERVER_SSH           Default: danya@galaxybook
  DEV_SERVER_PATH          Default: /home/danya/robot-center-dev
  DEV_SERVER_PUBLIC_URL    Default: http://192.168.20.12:18080
  DEV_SERVER_RECORDER_URL  Default: http://192.168.20.12:18082
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --commit-message)
      COMMIT_MESSAGE="${2:-}"
      if [[ -z "$COMMIT_MESSAGE" ]]; then
        printf '%s\n' '--commit-message requires a value' >&2
        exit 1
      fi
      shift 2
      ;;
    --no-commit)
      NO_COMMIT=1
      shift
      ;;
    --skip-local-checks)
      SKIP_LOCAL_CHECKS=1
      shift
      ;;
    --skip-deploy)
      SKIP_DEPLOY=1
      shift
      ;;
    --skip-post-deploy-checks)
      SKIP_POST_DEPLOY_CHECKS=1
      shift
      ;;
    --skip-web-route-check)
      SKIP_WEB_ROUTE_CHECK=1
      shift
      ;;
    --all-checks)
      FORCE_ALL_CHECKS=1
      shift
      ;;
    --since)
      LOG_SINCE="${2:-}"
      if [[ -z "$LOG_SINCE" ]]; then
        printf '%s\n' '--since requires a value' >&2
        exit 1
      fi
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      printf 'unknown option: %s\n' "$1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ -n "$COMMIT_MESSAGE" && "$NO_COMMIT" == "1" ]]; then
  printf '%s\n' '--commit-message and --no-commit cannot be used together' >&2
  exit 1
fi

require_command() {
  local command_name="$1"
  if ! command -v "$command_name" >/dev/null 2>&1; then
    printf '%s command not found\n' "$command_name" >&2
    exit 1
  fi
}

print_step() {
  printf '\n==> %s\n' "$1"
}

git_has_changes() {
  ! git diff --quiet || ! git diff --cached --quiet || [[ -n "$(git ls-files --others --exclude-standard)" ]]
}

changed_files() {
  {
    git diff --name-only HEAD --
    git ls-files --others --exclude-standard
  } | sort -u
}

files_match() {
  local pattern="$1"
  grep -Eq "$pattern" <<<"$CHANGED_FILES"
}

run_local_checks() {
  print_step "local checks"
  require_command git
  git diff --check
  git diff --cached --check

  local should_check_scripts=0
  local should_check_server=0
  local should_check_web=0

  if [[ "$FORCE_ALL_CHECKS" == "1" ]]; then
    should_check_scripts=1
    should_check_server=1
    should_check_web=1
  else
    if files_match '(^scripts/.*[.]sh$|^deploy/)'; then
      should_check_scripts=1
    fi
    if files_match '(^apps/server/|^deploy/|^scripts/generate-server-swagger[.]sh$)'; then
      should_check_server=1
    fi
    if files_match '(^apps/web/|^package(-lock)?[.]json$|^pnpm-lock[.]yaml$|^yarn[.]lock$)'; then
      should_check_web=1
    fi
  fi

  if [[ "$should_check_scripts" == "1" ]]; then
    print_step "script syntax checks"
    while IFS= read -r script_path; do
      [[ -n "$script_path" ]] || continue
      bash -n "$script_path"
    done < <(find "$ROOT_DIR/scripts" -maxdepth 1 -name '*.sh' -type f | sort)
  fi

  if [[ "$should_check_server" == "1" ]]; then
    print_step "server checks"
    require_command go
    "$ROOT_DIR/scripts/generate-server-swagger.sh" --check
    (
      cd "$ROOT_DIR/apps/server"
      export GOTOOLCHAIN="${GOTOOLCHAIN:-go1.24.4}"
      export GOCACHE="${GOCACHE:-${TMPDIR:-/tmp}/robot-center-go-build}"
      mkdir -p "$GOCACHE"
      go test ./...
    )
  fi

  if [[ "$should_check_web" == "1" ]]; then
    print_step "web checks"
    require_command npm
    (
      cd "$ROOT_DIR/apps/web"
      npm test
      npm run build
    )
    WEB_BUILD_RAN=1
  fi

  if [[ "$should_check_scripts" != "1" && "$should_check_server" != "1" && "$should_check_web" != "1" ]]; then
    printf 'no backend/frontend/script checks required for current diff\n'
  fi
}

commit_and_push_if_requested() {
  require_command git
  if [[ -z "$COMMIT_MESSAGE" ]]; then
    if [[ "$NO_COMMIT" == "0" ]] && git_has_changes; then
      printf '%s\n' 'working tree has changes; pass --commit-message or --no-commit' >&2
      exit 1
    fi
    if [[ "$NO_COMMIT" == "1" && "$SKIP_DEPLOY" == "0" ]] && git_has_changes; then
      printf '%s\n' 'working tree has changes; --no-commit deployments require a clean tree' >&2
      exit 1
    fi
    if [[ "$NO_COMMIT" == "1" ]]; then
      print_step "commit/push skipped"
    fi
    return
  fi

  print_step "commit"
  git add -A
  if git diff --cached --quiet; then
    printf 'no staged changes; commit skipped\n'
  else
    git commit -m "$COMMIT_MESSAGE"
  fi

  local branch_name
  branch_name="$(git rev-parse --abbrev-ref HEAD)"
  if [[ "$branch_name" == "HEAD" ]]; then
    printf '%s\n' 'cannot push detached HEAD' >&2
    exit 1
  fi

  print_step "push"
  git push origin "$branch_name"
}

run_deploy() {
  if [[ "$SKIP_DEPLOY" == "1" ]]; then
    print_step "deployment skipped"
    return
  fi

  print_step "dev-server deploy"
  if [[ "$WEB_BUILD_RAN" == "1" ]]; then
    SKIP_WEB_BUILD=1 "$ROOT_DIR/scripts/deploy-dev-server.sh"
    return
  fi
  "$ROOT_DIR/scripts/deploy-dev-server.sh"
}

fetch_json() {
  local url="$1"
  local label="$2"
  local tmp_file
  tmp_file="$(mktemp)"
  curl -fsS "$url" >"$tmp_file"
  python3 - "$tmp_file" "$label" <<'PY'
import json
import sys

path = sys.argv[1]
label = sys.argv[2]
with open(path, "r", encoding="utf-8") as file:
    data = json.load(file)

status = data.get("status") or data.get("state") or data.get("serviceStatus")
if status is None and isinstance(data.get("summary"), dict):
    status = data["summary"].get("status")

print(f"{label}: json ok" + (f", status={status}" if status else ""))
PY
  rm -f "$tmp_file"
}

check_rtc_config() {
  local tmp_file
  tmp_file="$(mktemp)"
  curl -fsS "$DEV_SERVER_PUBLIC_URL/api/v1/operator/rtc-config" >"$tmp_file"
  python3 - "$tmp_file" "$DEV_SERVER_PUBLIC_URL" <<'PY'
import json
import sys
from urllib.parse import urlparse

path = sys.argv[1]
public_url = sys.argv[2]
public_host = urlparse(public_url).hostname

with open(path, "r", encoding="utf-8") as file:
    data = json.load(file)

signaling_url = data.get("signalingUrl") or ""
ice_servers = data.get("iceServers") or []
turn_urls = []
for server in ice_servers:
    urls = server.get("urls") if isinstance(server, dict) else None
    if isinstance(urls, str):
        turn_urls.append(urls)
    elif isinstance(urls, list):
        turn_urls.extend(str(url) for url in urls)

if public_host and public_host not in signaling_url:
    raise SystemExit(f"rtc-config signalingUrl does not use public host {public_host}: {signaling_url}")
if public_host and not any(public_host in url for url in turn_urls):
    raise SystemExit(f"rtc-config TURN URL does not use public host {public_host}: {turn_urls}")

policy = data.get("iceTransportPolicy")
print(f"rtc-config: public urls ok, iceTransportPolicy={policy}")
PY
  rm -f "$tmp_file"
}

run_ssh_noninteractive() {
  if [[ -n "${SSHPASS:-}" ]]; then
    require_command sshpass
    # shellcheck disable=SC2086
    sshpass -e ssh $SSH_OPTS "$DEV_SERVER_SSH" "$@"
    return
  fi
  # shellcheck disable=SC2086
  ssh $SSH_OPTS -o BatchMode=yes "$DEV_SERVER_SSH" "$@"
}

scan_remote_logs() {
  print_step "remote log scan"
  if ! run_ssh_noninteractive "true" >/dev/null 2>&1; then
    printf 'remote log scan skipped: non-interactive SSH is not available\n'
    return
  fi

  local log_output
  log_output="$(run_ssh_noninteractive "set -euo pipefail
cd '$DEV_SERVER_PATH'
docker compose $COMPOSE_ARGS logs --since '$LOG_SINCE' app-server recorder-worker
" 2>&1 || true)"

  local suspicious
  suspicious="$(printf '%s\n' "$log_output" | grep -Ei '(^|[^[:alpha:]])(panic|fatal|error|failed|timeout|refused|denied)([^[:alpha:]]|$)' || true)"
  if [[ -n "$suspicious" ]]; then
    printf '%s\n' "$suspicious" >&2
    printf '%s\n' 'remote logs contain suspicious entries' >&2
    exit 1
  fi
  printf 'no suspicious app-server/recorder-worker logs in last %s\n' "$LOG_SINCE"
}

run_post_deploy_checks() {
  print_step "post-deploy checks"
  require_command curl
  require_command python3

  fetch_json "$DEV_SERVER_PUBLIC_URL/healthz" "app-server health"
  fetch_json "$DEV_SERVER_RECORDER_URL/healthz" "recorder health"
  fetch_json "$DEV_SERVER_PUBLIC_URL/api/v1/system/status" "system status"
  fetch_json "$DEV_SERVER_PUBLIC_URL/swagger/openapi.json" "swagger openapi"
  check_rtc_config

  if [[ "$SKIP_WEB_ROUTE_CHECK" != "1" ]]; then
    curl -fsS "$DEV_SERVER_PUBLIC_URL/system" >/dev/null
    printf 'web route: /system ok\n'
  fi

  scan_remote_logs
}

require_command git
CHANGED_FILES="$(changed_files)"

if [[ "$SKIP_LOCAL_CHECKS" != "1" ]]; then
  run_local_checks
else
  print_step "local checks skipped"
fi

commit_and_push_if_requested
run_deploy
if [[ "$SKIP_POST_DEPLOY_CHECKS" == "1" ]]; then
  print_step "post-deploy checks skipped"
else
  run_post_deploy_checks
fi

printf '\nready\n'
printf 'UI: %s\n' "$DEV_SERVER_PUBLIC_URL"
printf 'recorder health: %s/healthz\n' "$DEV_SERVER_RECORDER_URL"
