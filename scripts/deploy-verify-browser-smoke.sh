#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${DEV_SERVER_PUBLIC_URL:-http://192.168.20.12:18080}"
MISSION_CODE="${BROWSER_SMOKE_MISSION_CODE:-}"
ROBOT_CODE="${BROWSER_SMOKE_ROBOT_CODE:-}"
TIMEOUT_SECONDS="${BROWSER_SMOKE_TIMEOUT_SECONDS:-60}"
REQUIRE_RECORDING="${BROWSER_SMOKE_REQUIRE_RECORDING:-0}"
SESSION_NAME="${BROWSER_SMOKE_SESSION:-bs-$$-$RANDOM}"
KEEP_OPEN=0
PWCLI_COMMAND=(npx -y --package @playwright/cli playwright-cli --session "$SESSION_NAME")
TEMP_FILES=()

usage() {
  cat <<'EOF'
Usage:
  ./scripts/deploy-verify-browser-smoke.sh --mission-code mission-054 [--robot-code robot-042]

Options:
  --base-url URL           Control server public URL. Default: DEV_SERVER_PUBLIC_URL or dev server.
  --mission-code CODE      Mission control screen to verify. Defaults to a live SFU room.
  --robot-code CODE        Robot to select. Defaults to a streaming robot from live-status.
  --timeout-seconds N      Max wait for rendered videos. Default: 60.
  --require-recording      Require selected robot recording state to be recording.
  --session NAME           Playwright CLI session name. Keep it short on macOS.
  --keep-open              Keep browser open after the smoke check.
  -h, --help               Show this help.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --base-url)
      BASE_URL="${2:-}"
      [[ -n "$BASE_URL" ]] || { printf '%s\n' '--base-url requires a value' >&2; exit 1; }
      shift 2
      ;;
    --mission-code)
      MISSION_CODE="${2:-}"
      [[ -n "$MISSION_CODE" ]] || { printf '%s\n' '--mission-code requires a value' >&2; exit 1; }
      shift 2
      ;;
    --robot-code)
      ROBOT_CODE="${2:-}"
      [[ -n "$ROBOT_CODE" ]] || { printf '%s\n' '--robot-code requires a value' >&2; exit 1; }
      shift 2
      ;;
    --timeout-seconds)
      TIMEOUT_SECONDS="${2:-}"
      [[ -n "$TIMEOUT_SECONDS" ]] || { printf '%s\n' '--timeout-seconds requires a value' >&2; exit 1; }
      shift 2
      ;;
    --require-recording)
      REQUIRE_RECORDING=1
      shift
      ;;
    --session)
      SESSION_NAME="${2:-}"
      [[ -n "$SESSION_NAME" ]] || { printf '%s\n' '--session requires a value' >&2; exit 1; }
      PWCLI_COMMAND=(npx -y --package @playwright/cli playwright-cli --session "$SESSION_NAME")
      shift 2
      ;;
    --keep-open)
      KEEP_OPEN=1
      shift
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

BASE_URL="${BASE_URL%/}"

require_command() {
  local command_name="$1"
  if ! command -v "$command_name" >/dev/null 2>&1; then
    printf '%s command not found\n' "$command_name" >&2
    exit 1
  fi
}

json_string() {
  python3 -c 'import json, sys; print(json.dumps(sys.argv[1]))' "$1"
}

pwcli() {
  "${PWCLI_COMMAND[@]}" "$@"
}

close_browser() {
  if [[ "$KEEP_OPEN" == "1" ]]; then
    return
  fi
  pwcli close >/dev/null 2>&1 || true
}

cleanup() {
  local file_path
  for file_path in "${TEMP_FILES[@]}"; do
    [[ -n "$file_path" ]] || continue
    rm -f "$file_path"
  done
  close_browser
}

select_target_from_status() {
  local system_status_file live_status_file
  system_status_file="$(mktemp)"
  live_status_file="$(mktemp)"
  TEMP_FILES+=("$system_status_file" "$live_status_file")

  curl -fsS "$BASE_URL/api/v1/system/status" >"$system_status_file"
  if [[ -z "$MISSION_CODE" ]]; then
    MISSION_CODE="$(python3 - "$system_status_file" <<'PY'
import json
import sys

payload = json.load(open(sys.argv[1], encoding="utf-8"))
for room in payload.get("sfuRooms", []):
    if any(publisher.get("state") == "publishing" for publisher in room.get("publishers", [])):
        print(room.get("roomId", ""))
        break
PY
)"
  fi
  if [[ -z "$MISSION_CODE" ]]; then
    printf '%s\n' 'browser smoke failed: no live mission room found' >&2
    return 1
  fi

  curl -fsS "$BASE_URL/api/v1/operator/missions/$MISSION_CODE/live-status" >"$live_status_file"
  if [[ -z "$ROBOT_CODE" ]]; then
    ROBOT_CODE="$(python3 - "$live_status_file" "$REQUIRE_RECORDING" <<'PY'
import json
import sys

payload = json.load(open(sys.argv[1], encoding="utf-8"))
require_recording = sys.argv[2] == "1"
for robot in payload.get("robots", []):
    if robot.get("stream", {}).get("state") != "streaming":
        continue
    if require_recording and robot.get("recording", {}).get("state") != "recording":
        continue
    print(robot.get("robotCode", ""))
    break
PY
)"
  fi
  if [[ -z "$ROBOT_CODE" ]]; then
    printf 'browser smoke failed: no streaming robot found for %s\n' "$MISSION_CODE" >&2
    return 1
  fi
}

evaluate_raw() {
  local expression="$1"
  pwcli --raw eval "$expression"
}

click_robot_option() {
  local robot_code_json
  robot_code_json="$(json_string "$ROBOT_CODE")"
  evaluate_raw "(() => {
    const toggle = document.querySelector('[data-testid=\"mission-robot-dropdown-toggle\"]')
      || Array.from(document.querySelectorAll('button')).find((button) => /robot-[A-Za-z0-9_-]+/.test(button.innerText || ''));
    if (!toggle) {
      return { ok: false, reason: 'dropdown toggle not found' };
    }
    toggle.click();
    const robotCode = $robot_code_json;
    const option = document.querySelector('[data-testid=\"mission-robot-option\"][data-robot-code=\"' + robotCode + '\"]')
      || Array.from(document.querySelectorAll('button')).find((button) => (button.innerText || '').includes(robotCode) && button !== toggle);
    if (!option) {
      return { ok: false, reason: 'robot option not found', robotCode };
    }
    option.click();
    return { ok: true, robotCode };
  })()"
}

select_robot_in_browser() {
  local deadline click_result
  deadline=$((SECONDS + TIMEOUT_SECONDS))
  while (( SECONDS <= deadline )); do
    click_result="$(click_robot_option)"
    if python3 -c 'import json, sys; raise SystemExit(0 if json.load(sys.stdin).get("ok") else 1)' <<<"$click_result"; then
      return 0
    fi
    sleep 2
  done
  printf 'browser smoke failed: robot option selection failed: %s\n' "$click_result" >&2
  return 1
}

read_browser_status() {
  evaluate_raw "(() => {
    const videoElements = Array.from(document.querySelectorAll('[data-testid=\"live-video\"]'));
    const videos = videoElements.length > 0
      ? videoElements
      : Array.from(document.querySelectorAll('video')).slice(0, 2);
    const byLabel = {};
    videos.forEach((video, index) => {
      const label = video.dataset.videoLabel || (index === 0 ? 'rgb' : index === 1 ? 'thermal' : 'video-' + index);
      const pane = video.closest('[data-testid=\"live-video-pane\"]') || video.parentElement;
      byLabel[label] = {
        currentTime: Number((video.currentTime || 0).toFixed(2)),
        hasStream: Boolean(video.srcObject),
        height: video.videoHeight || 0,
        paused: Boolean(video.paused),
        readyState: video.readyState || 0,
        text: pane?.innerText || '',
        width: video.videoWidth || 0
      };
    });
    const bodyText = document.body?.innerText || '';
    const selectedText = document.querySelector('[data-testid=\"mission-robot-dropdown-toggle\"]')?.innerText
      || Array.from(document.querySelectorAll('button')).find((button) => /robot-[A-Za-z0-9_-]+/.test(button.innerText || ''))?.innerText
      || '';
    return {
      bodyHasRecording: bodyText.includes('녹화 중') || selectedText.includes('녹화 중'),
      bodyHasSensor: bodyText.includes('센서'),
      bodyHasConnected: bodyText.includes('실시간 연결 연결됨') || bodyText.includes('연결됨'),
      selectedText,
      url: location.href,
      videos: byLabel
    };
  })()"
}

status_passes() {
  local status_file="$1"
  python3 - "$status_file" "$ROBOT_CODE" "$REQUIRE_RECORDING" <<'PY'
import json
import sys

status = json.load(open(sys.argv[1], encoding="utf-8"))
robot_code = sys.argv[2]
require_recording = sys.argv[3] == "1"
errors = []

selected_text = status.get("selectedText", "")
if robot_code not in selected_text:
    errors.append(f"selected robot is not {robot_code}: {selected_text!r}")

for label in ("rgb", "thermal"):
    video = status.get("videos", {}).get(label)
    if not video:
        errors.append(f"{label} video missing")
        continue
    if not video.get("hasStream"):
        errors.append(f"{label} video has no MediaStream")
    if int(video.get("width") or 0) <= 0 or int(video.get("height") or 0) <= 0:
        errors.append(f"{label} video has no dimensions: {video.get('width')}x{video.get('height')}")
    if int(video.get("readyState") or 0) < 2:
        errors.append(f"{label} video readyState={video.get('readyState')}")

if not status.get("bodyHasSensor"):
    errors.append("sensor panel not found")
if not status.get("bodyHasConnected"):
    errors.append("connected status not found")
if require_recording and not status.get("bodyHasRecording"):
    errors.append("recording status not found")

if errors:
    print("; ".join(errors))
    sys.exit(1)

video_parts = []
for label in ("rgb", "thermal"):
    video = status["videos"][label]
    video_parts.append(f"{label}={video['width']}x{video['height']} readyState={video['readyState']}")
print(f"browser smoke: mission={status.get('url', '').split('/missions/')[-1].split('/')[0]} robot={robot_code} {'; '.join(video_parts)}")
PY
}

require_command curl
require_command npx
require_command python3

trap cleanup EXIT

select_target_from_status

control_url="$BASE_URL/missions/$MISSION_CODE/control"
pwcli open "$BASE_URL/missions" >/dev/null
pwcli localstorage-set robot-center.selectedLiveTargetKey "$MISSION_CODE:$ROBOT_CODE" >/dev/null || true
pwcli goto "$control_url" >/dev/null

select_robot_in_browser

deadline=$((SECONDS + TIMEOUT_SECONDS))
last_status_file="$(mktemp)"
TEMP_FILES+=("$last_status_file")
summary="not evaluated"
while (( SECONDS <= deadline )); do
  read_browser_status >"$last_status_file"
  if summary="$(status_passes "$last_status_file" 2>&1)"; then
    printf '%s\n' "$summary"
    exit 0
  fi
  sleep 2
done

printf 'browser smoke failed after %ss: %s\n' "$TIMEOUT_SECONDS" "$summary" >&2
printf 'last browser status: %s\n' "$(tr '\n' ' ' < "$last_status_file")" >&2
exit 1
