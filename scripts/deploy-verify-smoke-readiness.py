#!/usr/bin/env python3
import argparse
import json
import sys
import time
import urllib.parse
import urllib.request
from typing import Any


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Wait for a dev-server smoke target to become ready.")
    parser.add_argument("--base-url", required=True)
    parser.add_argument("--mission-code", required=True)
    parser.add_argument("--robot-code", required=True)
    parser.add_argument("--timeout-seconds", type=int, default=90)
    parser.add_argument("--require-recording", action="store_true")
    return parser.parse_args()


def fetch_json(base_url: str, path: str) -> dict[str, Any]:
    url = f"{base_url.rstrip('/')}{path}"
    with urllib.request.urlopen(url, timeout=10) as response:
        return json.load(response)


def quote(value: str) -> str:
    return urllib.parse.quote(value, safe="")


def find_by_key(items: list[dict[str, Any]], key: str, value: str) -> dict[str, Any] | None:
    return next((item for item in items if item.get(key) == value), None)


def evaluate_readiness(
    live_status: dict[str, Any],
    system_status: dict[str, Any],
    mission_code: str,
    robot_code: str,
    require_recording: bool,
) -> str:
    robot = find_by_key(live_status.get("robots", []), "robotCode", robot_code)
    if not robot:
        raise ValueError(f"robot not found: {robot_code}")

    connection = robot.get("connection") or {}
    stream = robot.get("stream") or {}
    recording = robot.get("recording") or {}
    errors: list[str] = []
    if connection.get("state") != "online":
        errors.append(f"connection={connection.get('state')}")
    if stream.get("state") != "streaming":
        errors.append(f"stream={stream.get('state')}")
    if int(stream.get("trackCount") or 0) < 3:
        errors.append(f"trackCount={stream.get('trackCount')}")
    if int(stream.get("dataChannelCount") or 0) < 4:
        errors.append(f"dataChannelCount={stream.get('dataChannelCount')}")
    if not stream.get("lastMediaAt") and not stream.get("lastTrackAt"):
        errors.append("lastMediaAt=empty")
    if not stream.get("lastDataAt"):
        errors.append("lastDataAt=empty")
    if require_recording and recording.get("state") != "recording":
        errors.append(f"recording={recording.get('state')}")

    room = find_by_key(system_status.get("sfuRooms", []), "roomId", mission_code)
    if not room:
        errors.append(f"sfuRoom=missing:{mission_code}")
    else:
        publisher = find_by_key(room.get("publishers", []), "robotCode", robot_code)
        if not publisher:
            errors.append(f"publisher=missing:{robot_code}")
        else:
            if publisher.get("state") != "publishing":
                errors.append(f"publisher={publisher.get('state')}")
            if publisher.get("iceState") not in {"connected", "completed"}:
                errors.append(f"iceState={publisher.get('iceState')}")
            if int(publisher.get("trackCount") or 0) < 3:
                errors.append(f"publisherTrackCount={publisher.get('trackCount')}")
            if int(publisher.get("dataChannelCount") or 0) < 4:
                errors.append(f"publisherDataChannelCount={publisher.get('dataChannelCount')}")
            validate_data_channels(publisher, errors)

    if errors:
        raise ValueError("; ".join(errors))
    return (
        f"smoke target ready: mission={live_status.get('missionCode')} robot={robot_code} "
        f"trackCount={stream.get('trackCount')} dataChannelCount={stream.get('dataChannelCount')} "
        f"recording={recording.get('state')}"
    )


def validate_data_channels(publisher: dict[str, Any], errors: list[str]) -> None:
    channels = {
        channel.get("label"): channel
        for channel in publisher.get("dataChannelStates", [])
        if isinstance(channel, dict) and channel.get("label")
    }
    for label in ("channel.telemetry", "channel.spatial", "channel.event", "channel.control"):
        channel = channels.get(label)
        if not channel:
            errors.append(f"{label}=missing")
        elif channel.get("state") != "open":
            errors.append(f"{label}={channel.get('state')}")
    telemetry = channels.get("channel.telemetry") or {}
    if int(telemetry.get("messageCount") or 0) <= 0:
        errors.append("channel.telemetry messageCount=0")
    if not telemetry.get("lastMessageAt"):
        errors.append("channel.telemetry lastMessageAt=empty")


def main() -> int:
    args = parse_args()
    deadline = time.monotonic() + max(args.timeout_seconds, 1)
    ready_count = 0
    last_error = "not evaluated"
    while time.monotonic() <= deadline:
        try:
            live_status = fetch_json(
                args.base_url,
                f"/api/v1/operator/missions/{quote(args.mission_code)}/live-status",
            )
            system_status = fetch_json(args.base_url, "/api/v1/system/status")
            summary = evaluate_readiness(
                live_status,
                system_status,
                args.mission_code,
                args.robot_code,
                args.require_recording,
            )
            ready_count += 1
            if ready_count >= 2:
                print(summary)
                return 0
        except Exception as error:
            ready_count = 0
            last_error = str(error)
        time.sleep(3)

    print(
        f"smoke target not ready after {args.timeout_seconds}s: "
        f"mission={args.mission_code} robot={args.robot_code}: {last_error}",
        file=sys.stderr,
    )
    return 1


if __name__ == "__main__":
    raise SystemExit(main())
