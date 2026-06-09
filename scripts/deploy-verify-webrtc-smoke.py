#!/usr/bin/env python3
import argparse
import json
import sys
from datetime import datetime, timezone
from typing import Any
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen


REQUIRED_TRACKS = ("track.video_1", "track.video_2", "track.audio_1")
REQUIRED_DATA_CHANNELS = (
    "channel.telemetry",
    "channel.spatial",
    "channel.event",
    "channel.control",
)
CONNECTED_ICE_STATES = {"connected", "completed"}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Verify dev-server WebRTC smoke health.")
    parser.add_argument("--base-url", required=True, help="Dev server public base URL.")
    parser.add_argument("--mission-code", default="", help="Mission code to verify. Defaults to any active SFU room.")
    parser.add_argument("--robot-code", default="", help="Robot code to verify. Defaults to any robot in the mission.")
    parser.add_argument("--min-robots", type=int, default=1, help="Minimum passing robot publishers. Default: 1.")
    parser.add_argument("--freshness-seconds", type=int, default=120, help="Max age for media/data timestamps. Default: 120.")
    parser.add_argument("--require-recording", action="store_true", help="Require live recording state to be recording.")
    return parser.parse_args()


def fetch_json(url: str) -> dict[str, Any]:
    request = Request(url, headers={"Accept": "application/json"})
    try:
        with urlopen(request, timeout=10) as response:
            return json.loads(response.read().decode("utf-8"))
    except HTTPError as error:
        raise RuntimeError(f"{url} returned HTTP {error.code}") from error
    except URLError as error:
        raise RuntimeError(f"{url} is not reachable: {error.reason}") from error
    except json.JSONDecodeError as error:
        raise RuntimeError(f"{url} returned invalid JSON: {error}") from error


def parse_time(value: str | None) -> datetime | None:
    if not value:
        return None
    normalized = value.replace("Z", "+00:00")
    parsed = datetime.fromisoformat(normalized)
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=timezone.utc)
    return parsed.astimezone(timezone.utc)


def age_seconds(value: str | None, now: datetime) -> float | None:
    parsed = parse_time(value)
    if parsed is None:
        return None
    return (now - parsed).total_seconds()


def is_fresh(value: str | None, now: datetime, freshness_seconds: int) -> bool:
    age = age_seconds(value, now)
    return age is not None and age <= freshness_seconds


def track_is_present(tracks: list[str], robot_code: str, track_id: str) -> bool:
    expected = f"{robot_code}:{track_id}"
    return any(track == expected or track == track_id or track.endswith(f":{track_id}") for track in tracks)


def data_channel_by_label(publisher: dict[str, Any]) -> dict[str, dict[str, Any]]:
    states = publisher.get("dataChannelStates") or []
    return {
        str(item.get("label")): item
        for item in states
        if isinstance(item, dict) and item.get("label")
    }


def live_robots_by_code(live_status: dict[str, Any]) -> dict[str, dict[str, Any]]:
    robots = live_status.get("robots") or []
    return {
        str(robot.get("robotCode")): robot
        for robot in robots
        if isinstance(robot, dict) and robot.get("robotCode")
    }


def publisher_errors(
    publisher: dict[str, Any],
    now: datetime,
    freshness_seconds: int,
) -> list[str]:
    robot_code = str(publisher.get("robotCode") or "")
    errors: list[str] = []
    if not robot_code:
        errors.append("missing robotCode")
        return errors

    if publisher.get("state") != "publishing":
        errors.append(f"publisher state={publisher.get('state')}")

    ice_state = str(publisher.get("iceState") or "")
    if ice_state not in CONNECTED_ICE_STATES:
        errors.append(f"iceState={ice_state or 'empty'}")

    tracks = [str(track) for track in publisher.get("tracks") or []]
    missing_tracks = [track_id for track_id in REQUIRED_TRACKS if not track_is_present(tracks, robot_code, track_id)]
    if missing_tracks:
        errors.append(f"missing tracks={','.join(missing_tracks)}")

    channels = data_channel_by_label(publisher)
    missing_channels = []
    closed_channels = []
    for label in REQUIRED_DATA_CHANNELS:
        channel = channels.get(label)
        if channel is None:
            missing_channels.append(label)
            continue
        if channel.get("state") != "open":
            closed_channels.append(f"{label}:{channel.get('state')}")
    if missing_channels:
        errors.append(f"missing dataChannels={','.join(missing_channels)}")
    if closed_channels:
        errors.append(f"closed dataChannels={','.join(closed_channels)}")

    telemetry = channels.get("channel.telemetry") or {}
    if int(telemetry.get("messageCount") or 0) <= 0:
        errors.append("channel.telemetry has no messages")
    if not is_fresh(telemetry.get("lastMessageAt"), now, freshness_seconds):
        errors.append("channel.telemetry lastMessageAt is stale")

    return errors


def live_status_errors(
    robot_code: str,
    live_robot: dict[str, Any] | None,
    now: datetime,
    freshness_seconds: int,
    require_recording: bool,
) -> list[str]:
    if live_robot is None:
        return ["missing live-status robot row"]

    errors: list[str] = []
    connection = live_robot.get("connection") or {}
    stream = live_robot.get("stream") or {}
    recording = live_robot.get("recording") or {}

    if connection.get("state") != "online":
        errors.append(f"connection state={connection.get('state')}")
    if stream.get("state") != "streaming":
        errors.append(f"stream state={stream.get('state')}")
    if int(stream.get("trackCount") or 0) < len(REQUIRED_TRACKS):
        errors.append(f"live trackCount={stream.get('trackCount')}")
    if int(stream.get("dataChannelCount") or 0) < len(REQUIRED_DATA_CHANNELS):
        errors.append(f"live dataChannelCount={stream.get('dataChannelCount')}")
    if not is_fresh(stream.get("lastMediaAt") or stream.get("lastTrackAt"), now, freshness_seconds):
        errors.append("live lastMediaAt is stale")
    if not is_fresh(stream.get("lastDataAt"), now, freshness_seconds):
        errors.append("live lastDataAt is stale")
    if require_recording and recording.get("state") != "recording":
        errors.append(f"recording state={recording.get('state')}")

    return errors


def candidate_rooms(system_status: dict[str, Any], mission_code: str) -> list[dict[str, Any]]:
    rooms = [room for room in system_status.get("sfuRooms") or [] if isinstance(room, dict)]
    if mission_code:
        return [room for room in rooms if room.get("roomId") == mission_code]
    return [
        room
        for room in rooms
        if any((publisher.get("state") == "publishing") for publisher in (room.get("publishers") or []))
    ]


def verify_room(
    base_url: str,
    room: dict[str, Any],
    robot_code: str,
    min_robots: int,
    freshness_seconds: int,
    require_recording: bool,
) -> tuple[bool, list[str]]:
    mission_code = str(room.get("roomId") or "")
    if not mission_code:
        return False, ["SFU room has no roomId"]

    live_status = fetch_json(f"{base_url}/api/v1/operator/missions/{mission_code}/live-status")
    live_by_robot = live_robots_by_code(live_status)
    publishers = [publisher for publisher in room.get("publishers") or [] if isinstance(publisher, dict)]
    if robot_code:
        publishers = [publisher for publisher in publishers if publisher.get("robotCode") == robot_code]

    now = datetime.now(timezone.utc)
    passing: list[str] = []
    failures: list[str] = []
    for publisher in publishers:
        current_robot_code = str(publisher.get("robotCode") or "")
        errors = publisher_errors(publisher, now, freshness_seconds)
        errors.extend(
            live_status_errors(
                current_robot_code,
                live_by_robot.get(current_robot_code),
                now,
                freshness_seconds,
                require_recording,
            )
        )
        if errors:
            failures.append(f"{current_robot_code or 'unknown'}: {'; '.join(errors)}")
            continue
        passing.append(current_robot_code)

    if len(passing) >= min_robots:
        print(f"webrtc smoke: mission={mission_code} passed robots={','.join(passing)}")
        return True, []

    if not publishers:
        failures.append("no matching publisher")
    failures.insert(0, f"mission={mission_code} passing={len(passing)} required={min_robots}")
    return False, failures


def main() -> int:
    args = parse_args()
    base_url = args.base_url.rstrip("/")
    if args.min_robots < 1:
        print("--min-robots must be >= 1", file=sys.stderr)
        return 2

    try:
        system_status = fetch_json(f"{base_url}/api/v1/system/status")
        if system_status.get("status") != "ok":
            raise RuntimeError(f"system status is {system_status.get('status')}")

        rooms = candidate_rooms(system_status, args.mission_code)
        if not rooms:
            target = args.mission_code or "any active SFU room"
            raise RuntimeError(f"no SFU room found for {target}")

        all_failures: list[str] = []
        for room in rooms:
            passed, failures = verify_room(
                base_url,
                room,
                args.robot_code,
                args.min_robots,
                args.freshness_seconds,
                args.require_recording,
            )
            if passed:
                return 0
            all_failures.extend(failures)

        raise RuntimeError("\n".join(all_failures))
    except Exception as error:
        print(f"webrtc smoke failed: {error}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
