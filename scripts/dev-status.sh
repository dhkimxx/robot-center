#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/dev-common.sh"

HOST_IP="$(detect_host_ip)"

printf 'host ip: %s\n' "$HOST_IP"
printf 'ui: http://%s:%s\n\n' "$HOST_IP" "$APP_PORT"

printf 'screen sessions:\n'
for session_name in "$TURN_SESSION" "$APP_SESSION" "$RECORDER_SESSION"; do
  if screen_exists "$session_name"; then
    printf '  %-24s running\n' "$session_name"
  else
    printf '  %-24s stopped\n' "$session_name"
  fi
done
python_mock_sessions="$(list_python_mock_sessions)"
if [[ -n "$python_mock_sessions" ]]; then
  while IFS= read -r session_name; do
    [[ -z "$session_name" ]] && continue
    printf '  %-24s running\n' "$session_name"
  done <<<"$python_mock_sessions"
else
  printf '  %-24s stopped\n' "$PYTHON_MOCK_SESSION_PREFIX-*"
fi

printf '\ndocker services:\n'
docker compose -f "$COMPOSE_FILE" ps postgres minio turn 2>/dev/null || true

printf '\nhealth:\n'
for item in "app-server http://127.0.0.1:$APP_PORT/healthz" "recorder-worker http://127.0.0.1:$RECORDER_PORT/healthz" "system http://127.0.0.1:$APP_PORT/api/system/status"; do
  label="${item%% *}"
  url="${item#* }"
  if curl -fsS "$url" >/dev/null 2>&1; then
    printf '  %-16s ok\n' "$label"
  else
    printf '  %-16s down\n' "$label"
  fi
done

printf '\nrtc config:\n'
curl -fsS "http://127.0.0.1:$APP_PORT/api/rtc-config" 2>/dev/null | /usr/bin/python3 -m json.tool 2>/dev/null || printf '  unavailable\n'

printf '\nrecorder subscriber:\n'
curl -fsS "http://127.0.0.1:$RECORDER_PORT/healthz" 2>/dev/null | /usr/bin/python3 -c 'import json,sys; payload=json.load(sys.stdin); print(json.dumps(payload.get("subscriber", {}), ensure_ascii=False, indent=2))' 2>/dev/null || printf '  unavailable\n'

printf '\nandroid:\n'
if adb_has_device; then
  adb shell dumpsys window | grep -E 'mCurrentFocus|mFocusedApp' | sed 's/^/  /' || true
else
  printf '  adb device unavailable\n'
fi
