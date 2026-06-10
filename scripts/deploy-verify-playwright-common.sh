#!/usr/bin/env bash

TEMP_FILES=()
PWCLI_COMMAND=()

require_command() {
  local command_name="$1"
  if ! command -v "$command_name" >/dev/null 2>&1; then
    printf '%s command not found\n' "$command_name" >&2
    exit 1
  fi
}

configure_playwright_cli() {
  local session_name="$1"
  PWCLI_COMMAND=(npx -y --package @playwright/cli playwright-cli --session "$session_name")
}

json_string() {
  python3 -c 'import json, sys; print(json.dumps(sys.argv[1]))' "$1"
}

make_temp_file() {
  local file_path
  file_path="$(mktemp)"
  TEMP_FILES+=("$file_path")
  printf '%s\n' "$file_path"
}

pwcli() {
  "${PWCLI_COMMAND[@]}" "$@"
}

evaluate_raw() {
  local expression="$1"
  pwcli --raw eval "$expression"
}

close_playwright_browser() {
  if [[ "${KEEP_OPEN:-0}" == "1" ]]; then
    return
  fi
  pwcli close >/dev/null 2>&1 || true
}

cleanup_playwright_smoke() {
  local file_path
  for file_path in "${TEMP_FILES[@]-}"; do
    [[ -n "$file_path" ]] || continue
    rm -f "$file_path"
  done
  close_playwright_browser
}
