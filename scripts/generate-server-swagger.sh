#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVER_DIR="$ROOT_DIR/apps/server"
OUTPUT_DIR="$SERVER_DIR/internal/api/swaggerdocs"
GOTOOLCHAIN="${GOTOOLCHAIN:-go1.24.4}"
export GOTOOLCHAIN

run_swag_init() {
  local output_dir="$1"
  (
    cd "$SERVER_DIR"
    go run github.com/swaggo/swag/cmd/swag@v1.16.6 init \
      -g cmd/app-server/main.go \
      -o "$output_dir" \
      --parseDependency \
      --parseInternal
  )
}

if [[ "${1:-}" == "--check" ]]; then
  tmp_parent="$(mktemp -d)"
  trap 'rm -rf "$tmp_parent"' EXIT
  tmp_output="$tmp_parent/swaggerdocs"
  run_swag_init "$tmp_output" >/dev/null
  diff -ru "$OUTPUT_DIR" "$tmp_output"
  exit 0
fi

run_swag_init internal/api/swaggerdocs
