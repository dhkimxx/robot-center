#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${DEV_SERVER_PUBLIC_URL:-http://192.168.20.12:18080}"
MISSION_CODE="${REPLAY_SMOKE_MISSION_CODE:-}"
ROBOT_CODE="${REPLAY_SMOKE_ROBOT_CODE:-}"
FILE_TYPES="${REPLAY_SMOKE_FILE_TYPES:-rgb_audio_mp4,thermal_mp4}"
TIMEOUT_SECONDS="${REPLAY_SMOKE_TIMEOUT_SECONDS:-60}"
SESSION_NAME="${REPLAY_SMOKE_SESSION:-rs-$$-$RANDOM}"
KEEP_OPEN=0
PWCLI_COMMAND=(npx -y --package @playwright/cli playwright-cli --session "$SESSION_NAME")
TEMP_FILES=()

usage() {
  cat <<'EOF'
Usage:
  ./scripts/deploy-verify-replay-smoke.sh --mission-code mission-054 --robot-code robot-042

Options:
  --base-url URL           Control server public URL. Default: DEV_SERVER_PUBLIC_URL or dev server.
  --mission-code CODE      Mission replay screen to verify. Defaults to a mission with playable files.
  --robot-code CODE        Robot replay target. Defaults to a robot with playable files.
  --file-types TYPES       Comma-separated video file types. Default: rgb_audio_mp4,thermal_mp4.
  --timeout-seconds N      Max wait for replay UI/video readiness. Default: 60.
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
    --file-types)
      FILE_TYPES="${2:-}"
      [[ -n "$FILE_TYPES" ]] || { printf '%s\n' '--file-types requires a value' >&2; exit 1; }
      shift 2
      ;;
    --timeout-seconds)
      TIMEOUT_SECONDS="${2:-}"
      [[ -n "$TIMEOUT_SECONDS" ]] || { printf '%s\n' '--timeout-seconds requires a value' >&2; exit 1; }
      shift 2
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

select_replay_target() {
  local target_file="$1"
  python3 - "$BASE_URL" "$MISSION_CODE" "$ROBOT_CODE" "$FILE_TYPES" >"$target_file" <<'PY'
import json
import sys
import urllib.parse
import urllib.request

base_url, mission_code, robot_code, file_types_arg = sys.argv[1:5]
file_types = [value.strip() for value in file_types_arg.split(",") if value.strip()]
if not file_types:
    raise SystemExit("replay smoke failed: no file types requested")

def fetch_json(path):
    with urllib.request.urlopen(f"{base_url}{path}", timeout=10) as response:
        return json.load(response)

def quote(value):
    return urllib.parse.quote(value, safe="")

def playable_files(recording):
    files = {}
    for file_entry in recording.get("files") or []:
        if file_entry.get("status") != "available":
            continue
        if not file_entry.get("url"):
            continue
        if not str(file_entry.get("contentType") or "").startswith("video/"):
            continue
        files[file_entry.get("type")] = file_entry
    return files

def find_candidate(recordings, required_mission, required_robot):
    for recording in recordings:
        if required_mission and recording.get("missionCode") != required_mission:
            continue
        if required_robot and recording.get("robotCode") != required_robot:
            continue
        files = playable_files(recording)
        if all(file_type in files for file_type in file_types):
            return recording, files
    return None, {}

if mission_code and not robot_code:
    summary = fetch_json(f"/api/v1/operator/missions/{quote(mission_code)}/recordings/summary")
    for robot in summary.get("robots", []):
        available_counts = robot.get("availableFileCounts") or {}
        if all((available_counts.get(file_type) or 0) > 0 for file_type in file_types):
            robot_code = robot.get("robotCode") or ""
            break

if mission_code:
    query = f"?limit=120"
    if robot_code:
        query += f"&robotCode={quote(robot_code)}"
    payload = fetch_json(f"/api/v1/operator/missions/{quote(mission_code)}/recordings/chunks{query}")
else:
    payload = fetch_json("/api/v1/operator/recordings")

recording, files = find_candidate(payload.get("recordings", []), mission_code, robot_code)
if not recording:
    target = f"mission={mission_code or '*'} robot={robot_code or '*'} fileTypes={','.join(file_types)}"
    raise SystemExit(f"replay smoke failed: no playable recording found for {target}")

print(json.dumps({
    "missionCode": recording.get("missionCode") or "",
    "robotCode": recording.get("robotCode") or "",
    "recordingSessionId": recording.get("recordingSessionId") or "",
    "chunkIndex": recording.get("chunkIndex"),
    "fileTypes": file_types,
    "files": {
        file_type: {
            "label": files[file_type].get("label") or file_type,
            "url": files[file_type].get("url") or ""
        }
        for file_type in file_types
    }
}))
PY
}

target_value() {
  local target_file="$1"
  local key="$2"
  python3 - "$target_file" "$key" <<'PY'
import json
import sys

payload = json.load(open(sys.argv[1], encoding="utf-8"))
value = payload
for part in sys.argv[2].split("."):
    value = value.get(part, {}) if isinstance(value, dict) else {}
print("" if value is None or isinstance(value, dict) else value)
PY
}

target_file_types() {
  local target_file="$1"
  python3 - "$target_file" <<'PY'
import json
import sys

payload = json.load(open(sys.argv[1], encoding="utf-8"))
for file_type in payload.get("fileTypes", []):
    print(file_type)
PY
}

evaluate_raw() {
  local expression="$1"
  pwcli --raw eval "$expression"
}

select_robot_in_browser() {
  local robot_code_json deadline click_result
  robot_code_json="$(json_string "$ROBOT_CODE")"
  deadline=$((SECONDS + TIMEOUT_SECONDS))
  while (( SECONDS <= deadline )); do
    click_result="$(evaluate_raw "(() => {
      const robotCode = $robot_code_json;
      const option = document.querySelector('[data-testid=\"mission-replay-robot-option\"][data-robot-code=\"' + robotCode + '\"]')
        || Array.from(document.querySelectorAll('button')).find((button) => (button.innerText || '').includes(robotCode));
      if (!option) {
        return { ok: false, reason: 'robot option not found', robotCode };
      }
      option.click();
      return { ok: true, robotCode };
    })()")"
    if python3 -c 'import json, sys; raise SystemExit(0 if json.load(sys.stdin).get("ok") else 1)' <<<"$click_result"; then
      return 0
    fi
    sleep 2
  done
  printf 'replay smoke failed: robot option selection failed: %s\n' "$click_result" >&2
  return 1
}

click_playback_file() {
  local target_file="$1"
  local file_type="$2"
  local file_type_json label_json chunk_index_json session_id_json click_result
  file_type_json="$(json_string "$file_type")"
  label_json="$(json_string "$(target_value "$target_file" "files.$file_type.label")")"
  chunk_index_json="$(json_string "$(target_value "$target_file" "chunkIndex")")"
  session_id_json="$(json_string "$(target_value "$target_file" "recordingSessionId")")"

  click_result="$(evaluate_raw "(() => {
    const fileType = $file_type_json;
    const fileLabel = $label_json;
    const chunkIndex = $chunk_index_json;
    const sessionId = $session_id_json;
    const buttons = Array.from(document.querySelectorAll('button')).filter((button) => (button.innerText || '').trim() === '재생');
    const fileTypeMatches = (button) => {
      const entry = button.closest('[data-testid=\"recording-object-entry\"]') || button.closest('div');
      const entryText = entry?.innerText || '';
      return button.dataset.fileType === fileType
        || entry?.dataset.fileType === fileType
        || (fileLabel && entryText.includes(fileLabel))
        || (fileType === 'rgb_audio_mp4' && entryText.includes('RGB'))
        || (fileType === 'thermal_mp4' && entryText.includes('Thermal'));
    };
    const chunkMatches = (button) => {
      const chunk = button.closest('[data-testid=\"mission-replay-chunk\"]');
      if (!chunk) {
        return true;
      }
      if (sessionId && chunk.dataset.recordingSessionId !== sessionId) {
        return false;
      }
      if (chunkIndex !== '' && String(chunk.dataset.chunkIndex) !== String(chunkIndex)) {
        return false;
      }
      return true;
    };
    const button = buttons.find((candidate) => fileTypeMatches(candidate) && chunkMatches(candidate))
      || buttons.find(fileTypeMatches)
      || null;
    if (!button) {
      return { ok: false, reason: 'playback button not found', fileType, count: buttons.length };
    }
    button.click();
    return { ok: true, fileType };
  })()")"
  if ! python3 -c 'import json, sys; raise SystemExit(0 if json.load(sys.stdin).get("ok") else 1)' <<<"$click_result"; then
    printf 'replay smoke failed: playback click failed: %s\n' "$click_result" >&2
    return 1
  fi
}

read_playback_status() {
  evaluate_raw "(() => {
    const video = document.querySelector('[data-testid=\"recording-playback-video\"]')
      || Array.from(document.querySelectorAll('video')).find((candidate) => candidate.src && !candidate.srcObject);
    if (video && video.paused && video.readyState >= 1) {
      video.play?.().catch(() => {});
    }
    const modal = document.querySelector('[data-testid=\"recording-playback-modal\"]')
      || document.querySelector('[role=\"dialog\"]');
    return {
      currentTime: video ? Number((video.currentTime || 0).toFixed(2)) : 0,
      duration: video && Number.isFinite(video.duration) ? Number(video.duration.toFixed(2)) : 0,
      fileType: video?.dataset.fileType || modal?.dataset.fileType || '',
      height: video?.videoHeight || 0,
      modalFound: Boolean(modal),
      networkState: video?.networkState || 0,
      readyState: video?.readyState || 0,
      src: video?.currentSrc || video?.src || '',
      width: video?.videoWidth || 0
    };
  })()"
}

status_passes() {
  local status_file="$1"
  local file_type="$2"
  python3 - "$status_file" "$file_type" <<'PY'
import json
import sys

status = json.load(open(sys.argv[1], encoding="utf-8"))
file_type = sys.argv[2]
errors = []

if not status.get("modalFound"):
    errors.append("playback modal not found")
if file_type and status.get("fileType") and status.get("fileType") != file_type:
    errors.append(f"file type mismatch: {status.get('fileType')}")
if not status.get("src"):
    errors.append("video src missing")
if int(status.get("width") or 0) <= 0 or int(status.get("height") or 0) <= 0:
    errors.append(f"video has no dimensions: {status.get('width')}x{status.get('height')}")
if int(status.get("readyState") or 0) < 2:
    errors.append(f"video readyState={status.get('readyState')}")
if float(status.get("duration") or 0) <= 0:
    errors.append(f"video duration={status.get('duration')}")

if errors:
    print("; ".join(errors))
    sys.exit(1)

print(f"{file_type}={status['width']}x{status['height']} duration={status['duration']} readyState={status['readyState']}")
PY
}

close_playback_modal() {
  evaluate_raw "(() => {
    const button = Array.from(document.querySelectorAll('button')).find((candidate) => (candidate.innerText || '').trim() === '닫기');
    if (button) {
      button.click();
      return { ok: true };
    }
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    return { ok: false };
  })()" >/dev/null || true
}

verify_file_type() {
  local target_file="$1"
  local file_type="$2"
  local deadline last_status_file summary
  last_status_file="$(mktemp)"
  TEMP_FILES+=("$last_status_file")

  click_playback_file "$target_file" "$file_type"

  deadline=$((SECONDS + TIMEOUT_SECONDS))
  summary="not evaluated"
  while (( SECONDS <= deadline )); do
    read_playback_status >"$last_status_file"
    if summary="$(status_passes "$last_status_file" "$file_type" 2>&1)"; then
      printf '%s\n' "$summary"
      close_playback_modal
      return 0
    fi
    sleep 2
  done

  printf 'replay smoke failed after %ss for %s: %s\n' "$TIMEOUT_SECONDS" "$file_type" "$summary" >&2
  printf 'last playback status: %s\n' "$(tr '\n' ' ' < "$last_status_file")" >&2
  return 1
}

require_command npx
require_command python3

trap cleanup EXIT

target_file="$(mktemp)"
TEMP_FILES+=("$target_file")
select_replay_target "$target_file"

MISSION_CODE="$(target_value "$target_file" "missionCode")"
ROBOT_CODE="$(target_value "$target_file" "robotCode")"

control_url="$BASE_URL/missions/$MISSION_CODE/replay"
pwcli open "$BASE_URL/missions" >/dev/null
pwcli goto "$control_url" >/dev/null

select_robot_in_browser

summaries=()
while IFS= read -r file_type; do
  [[ -n "$file_type" ]] || continue
  summaries+=("$(verify_file_type "$target_file" "$file_type")")
done < <(target_file_types "$target_file")

joined_summary=""
for summary in "${summaries[@]}"; do
  if [[ -z "$joined_summary" ]]; then
    joined_summary="$summary"
  else
    joined_summary="$joined_summary; $summary"
  fi
done

printf 'replay smoke: mission=%s robot=%s %s\n' "$MISSION_CODE" "$ROBOT_CODE" "$joined_summary"
