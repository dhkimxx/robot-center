#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

APPLY=0
INCLUDE_PYTHON_VENV=0

for arg in "$@"; do
  case "$arg" in
    --apply)
      APPLY=1
      ;;
    --python-venv)
      INCLUDE_PYTHON_VENV=1
      ;;
    -h|--help)
      cat <<'USAGE'
Usage: ./scripts/clean-local-runtime.sh [--apply] [--python-venv]

Removes ignored local test artifacts:
- apps/server/.runtime/recordings/*
- root .runtime logs, pid files, screenshots, and temp env files
- Android mock build outputs

The Python mock robot virtualenv is kept by default. Add --python-venv to remove it too.
Without --apply this script only prints the targets.
USAGE
      exit 0
      ;;
    *)
      printf 'unknown argument: %s\n' "$arg" >&2
      exit 1
      ;;
  esac
done

remove_path() {
  local target="$1"
  if [[ ! -e "$target" ]]; then
    return
  fi
  if [[ "$APPLY" == "1" ]]; then
    rm -rf "$target"
    printf 'removed %s\n' "$target"
    return
  fi
  printf 'would remove %s\n' "$target"
}

remove_matching_files() {
  local directory="$1"
  local pattern="$2"
  [[ -d "$directory" ]] || return
  while IFS= read -r -d '' target; do
    remove_path "$target"
  done < <(find "$directory" -maxdepth 1 -type f -name "$pattern" -print0)
}

printf 'local runtime cleanup mode: %s\n' "$([[ "$APPLY" == "1" ]] && printf apply || printf dry-run)"

remove_path "$ROOT_DIR/apps/server/.runtime/recordings"
mkdir -p "$ROOT_DIR/apps/server/.runtime/recordings"

remove_matching_files "$ROOT_DIR/.runtime" "*.log"
remove_matching_files "$ROOT_DIR/.runtime" "*.pid"
remove_matching_files "$ROOT_DIR/.runtime" "*.png"
remove_matching_files "$ROOT_DIR/.runtime" "*.env"

remove_path "$ROOT_DIR/apps/android-robot/.gradle"
remove_path "$ROOT_DIR/apps/android-robot/build"
remove_path "$ROOT_DIR/apps/android-robot/app/build"

if [[ "$INCLUDE_PYTHON_VENV" == "1" ]]; then
  remove_path "$ROOT_DIR/.runtime/python-mock-robot-venv"
fi

if [[ "$APPLY" != "1" ]]; then
  printf '\nRun with --apply to delete these ignored local artifacts.\n'
fi
