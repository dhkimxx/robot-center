#!/usr/bin/env python3
import argparse
import json
import math
import os
import signal
import sys
import threading
import time
from datetime import datetime, timezone

import gi
import requests
import websocket

gi.require_version("Gst", "1.0")
gi.require_version("GstSdp", "1.0")
gi.require_version("GstWebRTC", "1.0")
from gi.repository import GLib, Gst, GstSdp, GstWebRTC  # noqa: E402


DATA_CHANNEL_LABELS = (
    "channel.telemetry",
    "channel.spatial",
    "channel.event",
    "channel.control",
)
INITIAL_RECONNECT_DELAY_SECONDS = 1
MAX_RECONNECT_DELAY_SECONDS = 30
HEARTBEAT_INTERVAL_SECONDS = 10
DATA_INTERVAL_SECONDS = 1
TRACK_ROLE_BY_MID = {
    "audio0": "track.audio_1",
    "video1": "track.video_1",
    "video2": "track.video_2",
}


def utc_now_iso() -> str:
    return datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")


def log(message: str) -> None:
    print(f"[{utc_now_iso()}] {message}", flush=True)


def summarize_sdp(name: str, sdp: str) -> None:
    lines = [line.strip() for line in sdp.replace("\r\n", "\n").split("\n") if line.strip()]
    interesting = []
    current_m = ""
    for line in lines:
        if line.startswith("a=group:BUNDLE"):
            interesting.append(line)
            continue
        if line.startswith("m="):
            current_m = line
            interesting.append(line)
            continue
        if current_m and (
            line.startswith("a=mid:")
            or line.startswith("a=sctp-port:")
            or line.startswith("a=setup:")
            or line.startswith("a=max-message-size:")
            or line == "a=bundle-only"
        ):
            interesting.append("  " + line)
    log(f"{name} SDP summary: {' | '.join(interesting)}")


def turn_url_for_gstreamer(turn_server: dict) -> str:
    urls = turn_server.get("urls") or []
    if isinstance(urls, str):
        urls = [urls]
    if not urls:
        return ""
    raw_url = urls[0]
    username = turn_server.get("username", "")
    credential = turn_server.get("credential", "")
    without_scheme = raw_url.removeprefix("turn:").removeprefix("turns:")
    if "?" in without_scheme:
        host_port, query = without_scheme.split("?", 1)
        return f"turn://{username}:{credential}@{host_port}?{query}"
    return f"turn://{username}:{credential}@{without_scheme}"


def normalize_media_track_msid(sdp: str) -> tuple[str, bool]:
    lines = [line for line in sdp.replace("\r\n", "\n").split("\n") if line]
    sections: list[list[str]] = []
    current_section: list[str] = []
    for line in lines:
        if line.startswith("m=") and current_section:
            sections.append(current_section)
            current_section = [line]
            continue
        current_section.append(line)
    if current_section:
        sections.append(current_section)

    changed = False
    normalized_sections = []
    for section in sections:
        mid = ""
        for line in section:
            if line.startswith("a=mid:"):
                mid = line.removeprefix("a=mid:").strip()
                break
        track_role = TRACK_ROLE_BY_MID.get(mid)
        if not track_role:
            normalized_sections.extend(section)
            continue

        rewritten = False
        normalized_section = []
        for line in section:
            if line.startswith("a=msid:") and not rewritten:
                normalized_section.append(f"a=msid:robot-gstreamer {track_role}")
                rewritten = True
                changed = True
                continue
            if line.startswith("a=ssrc:") and " msid:" in line:
                prefix, _ = line.split(" msid:", 1)
                normalized_section.append(f"{prefix} msid:robot-gstreamer {track_role}")
                changed = True
                continue
            normalized_section.append(line)
            if line.startswith("a=mid:") and not any(item.startswith("a=msid:") for item in section):
                normalized_section.append(f"a=msid:robot-gstreamer {track_role}")
                rewritten = True
                changed = True
        normalized_sections.extend(normalized_section)

    if not changed:
        return sdp, False
    return "\r\n".join(normalized_sections) + "\r\n", True


def robot_number(robot_code: str) -> int:
    suffix = "".join(char for char in robot_code if char.isdigit())
    if suffix:
        return int(suffix[-3:])
    return sum(ord(char) for char in robot_code) % 1000


class GStreamerMockRobot:
    def __init__(self, args: argparse.Namespace):
        self.server_url = args.server_url.rstrip("/")
        self.robot_code = args.robot_code
        self.robot_token = args.robot_token
        self.bundle_policy = args.bundle_policy
        self.rgb_width = args.rgb_width
        self.rgb_height = args.rgb_height
        self.thermal_width = args.thermal_width
        self.thermal_height = args.thermal_height
        self.fps = args.fps
        self.stop_event = threading.Event()
        self.http = requests.Session()
        self.http.headers.update({"Authorization": f"Bearer {self.robot_token}"})
        self.sequence = 0

    def stop(self) -> None:
        self.stop_event.set()

    def run(self) -> int:
        Gst.init(None)
        reconnect_delay = INITIAL_RECONNECT_DELAY_SECONDS
        while not self.stop_event.is_set():
            attempt = None
            try:
                self.send_heartbeat("online")
                mission = self.wait_for_active_mission()
                self.send_heartbeat("streaming")
                attempt = PublishAttempt(self, mission)
                attempt.run()
                reconnect_delay = INITIAL_RECONNECT_DELAY_SECONDS
                if not self.stop_event.is_set():
                    log("publish connection ended; retrying from mission lookup")
            except KeyboardInterrupt:
                self.stop()
            except Exception as exc:
                if not self.stop_event.is_set():
                    log(f"publish connection failed: {type(exc).__name__}: {exc}")
            finally:
                if attempt is not None:
                    attempt.cleanup()
                try:
                    self.send_heartbeat("online")
                except Exception as exc:
                    log(f"publish stopped heartbeat failed: {exc}")

            if self.stop_event.is_set():
                break
            log(f"reconnect retry in {reconnect_delay}s")
            self.stop_event.wait(reconnect_delay)
            reconnect_delay = min(reconnect_delay * 2, MAX_RECONNECT_DELAY_SECONDS)
        return 0

    def wait_for_active_mission(self) -> dict:
        while not self.stop_event.is_set():
            mission = self.get_json("/api/v1/robot/mission")
            if mission.get("missionStatus") == "active":
                log(f"mission active: {mission.get('missionCode')}")
                return mission
            log("waiting for active mission")
            self.stop_event.wait(2)
        raise KeyboardInterrupt

    def fetch_active_mission(self) -> dict:
        mission = self.get_json("/api/v1/robot/mission")
        if mission.get("missionStatus") == "active":
            return mission
        return {}

    def send_heartbeat(self, state: str) -> None:
        payload = {
            "state": state,
            "batteryPercent": self.current_battery_percent(),
            "networkQuality": "gstreamer-docker",
            "sentAt": utc_now_iso(),
        }
        self.post_json("/api/v1/robot/heartbeat", payload)

    def get_json(self, path: str) -> dict:
        response = self.http.get(self.server_url + path, timeout=10)
        response.raise_for_status()
        return response.json()

    def post_json(self, path: str, payload: dict) -> dict:
        response = self.http.post(self.server_url + path, json=payload, timeout=10)
        response.raise_for_status()
        return response.json()

    def next_sequence(self) -> int:
        self.sequence += 1
        return self.sequence

    def current_position(self, sequence: int) -> tuple[float, float]:
        robot_delta = (sum(ord(char) for char in self.robot_code) % 20) / 10000
        movement = math.sin(sequence / 10) / 10000
        return 37.5665 + robot_delta + movement, 126.9780 + robot_delta + movement

    def current_battery_percent(self) -> int:
        return max(30, 91 - (self.sequence // 15) % 35)

    def create_telemetry_payload(self) -> dict:
        sequence = self.next_sequence()
        latitude, longitude = self.current_position(sequence)
        return {
            "messageId": f"{self.robot_code}-gst-telemetry-{sequence}",
            "messageType": "telemetry",
            "descriptors": [
                {
                    "sensorId": "telemetry.position_1",
                    "sensorType": "position",
                    "label": "GPS",
                    "enabled": True,
                },
                {
                    "sensorId": "telemetry.gas.channel_1",
                    "sensorType": "gas",
                    "label": "CO",
                    "unit": "ppm",
                    "enabled": True,
                },
                {
                    "sensorId": "telemetry.gas.channel_2",
                    "sensorType": "gas",
                    "label": "O2",
                    "unit": "%Vol",
                    "enabled": True,
                },
                {
                    "sensorId": "telemetry.battery_1",
                    "sensorType": "battery",
                    "label": "Battery",
                    "unit": "percent",
                    "enabled": True,
                },
            ],
            "samples": [
                {
                    "sensorId": "telemetry.position_1",
                    "timestamp": utc_now_iso(),
                    "values": {
                        "latitude": latitude,
                        "longitude": longitude,
                        "altitudeMeter": 45.0 + (sequence % 8),
                        "accuracyMeter": 4.5,
                        "headingDegree": (sequence * 12) % 360,
                        "speedMeterPerSecond": 0.4 + (sequence % 4) * 0.1,
                    },
                },
                {
                    "sensorId": "telemetry.gas.channel_1",
                    "timestamp": utc_now_iso(),
                    "values": {
                        "concentration": 13.0 + sequence % 3,
                        "scale_code": 1,
                        "alarm_code": 0,
                        "alarm": "normal",
                        "low_alarm": 10.0,
                        "high_alarm": 15.0,
                        "valid": True,
                    },
                },
                {
                    "sensorId": "telemetry.gas.channel_2",
                    "timestamp": utc_now_iso(),
                    "values": {
                        "concentration": 21.1 + math.sin(sequence / 10) * 0.2,
                        "scale_code": 1,
                        "alarm_code": 0,
                        "alarm": "normal",
                        "low_alarm": 19.5,
                        "high_alarm": 23.5,
                        "valid": True,
                    },
                },
                {
                    "sensorId": "telemetry.battery_1",
                    "timestamp": utc_now_iso(),
                    "values": {"batteryPercent": self.current_battery_percent()},
                },
            ],
        }

class PublishAttempt:
    def __init__(self, robot: GStreamerMockRobot, mission: dict):
        self.robot = robot
        self.mission = mission
        self.loop = GLib.MainLoop()
        self.pipeline = None
        self.webrtc = None
        self.websocket = None
        self.websocket_lock = threading.Lock()
        self.websocket_closed = threading.Event()
        self.local_peer_id = ""
        self.target_peer_id = "sfu"
        self.offer_sent = False
        self.open_labels: set[str] = set()
        self.closed_labels: set[str] = set()
        self.heartbeat_thread = None
        self.websocket_thread = None

    def run(self) -> None:
        self.setup_pipeline()
        self.connect_websocket()
        self.heartbeat_thread = threading.Thread(target=self.heartbeat_loop, daemon=True)
        self.heartbeat_thread.start()
        self.loop.run()

    def setup_pipeline(self) -> None:
        fps = max(self.robot.fps, 1)
        pipeline_description = (
            f"webrtcbin name=webrtc bundle-policy={self.robot.bundle_policy} "
            "audiotestsrc is-live=true wave=silence ! "
            "audioconvert ! audioresample ! opusenc ! rtpopuspay pt=111 ! "
            "application/x-rtp,media=audio,encoding-name=OPUS,payload=111 ! webrtc. "
            "videotestsrc is-live=true pattern=ball ! "
            f"video/x-raw,width={self.robot.rgb_width},height={self.robot.rgb_height},framerate={fps}/1 ! "
            "videoconvert ! vp8enc deadline=1 keyframe-max-dist=30 ! rtpvp8pay pt=96 ! "
            "application/x-rtp,media=video,encoding-name=VP8,payload=96 ! webrtc. "
            "videotestsrc is-live=true pattern=gradient ! "
            f"video/x-raw,width={self.robot.thermal_width},height={self.robot.thermal_height},framerate={fps}/1 ! "
            "videoconvert ! vp8enc deadline=1 keyframe-max-dist=30 ! rtpvp8pay pt=97 ! "
            "application/x-rtp,media=video,encoding-name=VP8,payload=97 ! webrtc. "
        )
        self.pipeline = Gst.parse_launch(pipeline_description)
        self.webrtc = self.pipeline.get_by_name("webrtc")
        if self.webrtc is None:
            raise RuntimeError("webrtcbin is not available")

        self.set_property_if_available("ice-transport-policy", "relay")
        turn_servers = self.mission.get("turnServers") or []
        if turn_servers:
            turn_url = turn_url_for_gstreamer(turn_servers[0])
            if turn_url:
                self.set_property_if_available("turn-server", turn_url)
                log(f"configured TURN server for GStreamer: {turn_url}")

        self.webrtc.connect("on-ice-candidate", self.on_ice_candidate)
        self.webrtc.connect("notify::ice-connection-state", self.on_ice_connection_state)
        self.webrtc.connect("notify::connection-state", self.on_connection_state)
        self.webrtc.connect("notify::signaling-state", self.on_signaling_state)

        bus = self.pipeline.get_bus()
        bus.add_signal_watch()
        bus.connect("message", self.on_bus_message)

        self.pipeline.set_state(Gst.State.PLAYING)
        log("GStreamer pipeline set to PLAYING")

    def set_property_if_available(self, name: str, value: str) -> None:
        if self.webrtc.find_property(name) is None:
            log(f"webrtcbin property missing, skipped: {name}")
            return
        self.webrtc.set_property(name, value)

    def connect_websocket(self) -> None:
        signaling_url = self.mission["sfu"]["signalingUrl"]
        log(f"signaling connecting: {signaling_url}")
        self.websocket = websocket.WebSocket()
        self.websocket.connect(
            signaling_url,
            header=[f"Authorization: Bearer {self.robot.robot_token}"],
            timeout=20,
        )
        self.websocket.settimeout(None)
        log("signaling connected")
        self.websocket_thread = threading.Thread(target=self.read_websocket_loop, daemon=True)
        self.websocket_thread.start()

    def read_websocket_loop(self) -> None:
        assert self.websocket is not None
        while not self.websocket_closed.is_set() and not self.robot.stop_event.is_set():
            try:
                raw_message = self.websocket.recv()
            except Exception as exc:
                if not self.websocket_closed.is_set() and not self.robot.stop_event.is_set():
                    log(f"websocket recv error: {exc}")
                    GLib.idle_add(self.stop_loop)
                return
            if not raw_message:
                continue
            try:
                message = json.loads(raw_message)
            except json.JSONDecodeError:
                log(f"non-json signaling message ignored: {raw_message}")
                continue
            GLib.idle_add(self.handle_signal_message, message)

    def heartbeat_loop(self) -> None:
        while not self.websocket_closed.wait(HEARTBEAT_INTERVAL_SECONDS):
            if self.robot.stop_event.is_set():
                return
            try:
                mission = self.robot.fetch_active_mission()
                if mission.get("missionCode") != self.mission.get("missionCode"):
                    log("active mission disappeared; stopping publish")
                    GLib.idle_add(self.stop_loop)
                    return
                self.robot.send_heartbeat("streaming")
            except Exception as exc:
                log(f"heartbeat failed: {exc}")

    def handle_signal_message(self, message: dict) -> bool:
        message_type = message.get("type")
        payload = message.get("payload") or {}
        target_peer_id = payload.get("targetPeerId")
        if target_peer_id and self.local_peer_id and target_peer_id != self.local_peer_id:
            return False

        if message_type == "joined":
            self.local_peer_id = payload.get("peerId") or ""
            log(f"room joined: {payload.get('room')} / {payload.get('role')}")
            return False
        if message_type == "mission-ended":
            log(f"mission ended signal received: {payload.get('room')}")
            self.stop_loop()
            return False
        if message_type in ("peer-present", "peer-joined") and payload.get("role") == "sfu":
            self.target_peer_id = payload.get("peerId") or "sfu"
            if not self.offer_sent:
                self.start_offer()
            return False
        if message_type == "answer":
            self.apply_answer(payload)
            return False
        if message_type == "publish-warning":
            log(f"publish-warning: {payload}")
            return False
        if message_type == "publish-error":
            log(f"publish-error: {payload}")
            self.stop_loop()
            return False
        return False

    def start_offer(self) -> None:
        self.offer_sent = True
        for label in DATA_CHANNEL_LABELS:
            channel = self.webrtc.emit("create-data-channel", label, None)
            channel.connect("on-open", self.on_data_channel_open, label)
            channel.connect("on-close", self.on_data_channel_close, label)
            channel.connect("on-error", self.on_data_channel_error, label)
            log(f"created DataChannel before offer: {label}")
        GLib.timeout_add(1000, self.create_offer)

    def create_offer(self) -> bool:
        log("creating SDP offer")
        promise = Gst.Promise.new_with_change_func(self.on_offer_created, None)
        self.webrtc.emit("create-offer", None, promise)
        return False

    def on_offer_created(self, promise: Gst.Promise, _user_data) -> None:
        promise.wait()
        reply = promise.get_reply()
        offer = reply.get_value("offer")
        self.webrtc.emit("set-local-description", offer, Gst.Promise.new())
        sdp = offer.sdp.as_text()
        send_sdp, normalized = normalize_media_track_msid(sdp)
        if normalized:
            log("offer media track msid normalized for canonical robot track labels")
        summarize_sdp("offer", send_sdp)
        self.send_signal(
            {
                "type": "offer",
                "payload": {
                    "targetPeerId": self.target_peer_id,
                    "type": "offer",
                    "sdp": send_sdp,
                },
            }
        )
        log("offer sent")

    def apply_answer(self, payload: dict) -> None:
        answer_sdp = payload.get("sdp")
        if not answer_sdp:
            log("answer without sdp ignored")
            return
        result, sdp_message = GstSdp.SDPMessage.new()
        if result != GstSdp.SDPResult.OK:
            raise RuntimeError(f"SDPMessage.new failed: {result}")
        result = GstSdp.sdp_message_parse_buffer(answer_sdp.encode("utf-8"), sdp_message)
        if result != GstSdp.SDPResult.OK:
            raise RuntimeError(f"parse answer SDP failed: {result}")
        answer = GstWebRTC.WebRTCSessionDescription.new(GstWebRTC.WebRTCSDPType.ANSWER, sdp_message)
        self.webrtc.emit("set-remote-description", answer, Gst.Promise.new())
        summarize_sdp("answer", answer_sdp)
        log("answer applied")

    def on_ice_candidate(self, _element, mline_index: int, candidate: str) -> None:
        self.send_signal(
            {
                "type": "candidate",
                "payload": {
                    "targetPeerId": self.target_peer_id,
                    "candidate": candidate,
                    "sdpMLineIndex": int(mline_index),
                },
            }
        )

    def on_ice_connection_state(self, element, _param) -> None:
        state = element.get_property("ice-connection-state").value_nick
        log(f"ice-connection-state={state}")
        if state in ("failed", "disconnected", "closed"):
            self.stop_loop()

    def on_connection_state(self, element, _param) -> None:
        state = element.get_property("connection-state").value_nick
        log(f"connection-state={state}")
        if state in ("failed", "closed"):
            self.stop_loop()

    def on_signaling_state(self, element, _param) -> None:
        log(f"signaling-state={element.get_property('signaling-state').value_nick}")

    def on_data_channel_open(self, channel, label: str) -> None:
        log(f"DATA_CHANNEL_OPEN label={label}")
        self.open_labels.add(label)
        if label == "channel.telemetry":
            GLib.timeout_add_seconds(DATA_INTERVAL_SECONDS, self.send_channel_payload, channel, label)
        else:
            log(f"DATA_CHANNEL_IDLE label={label} reason=payload schema not finalized")

    def on_data_channel_close(self, _channel, label: str) -> None:
        log(f"DATA_CHANNEL_CLOSE label={label}")
        self.closed_labels.add(label)

    def on_data_channel_error(self, _channel, error, label: str) -> None:
        log(f"DATA_CHANNEL_ERROR label={label}: {error}")

    def send_channel_payload(self, channel, label: str) -> bool:
        if self.websocket_closed.is_set() or self.robot.stop_event.is_set() or label in self.closed_labels:
            return False
        if label != "channel.telemetry":
            return False
        try:
            payload = self.robot.create_telemetry_payload()
            channel.emit("send-string", json.dumps(payload, separators=(",", ":")))
        except Exception as exc:
            log(f"send-string failed label={label}: {exc}")
            return False
        return True

    def send_signal(self, message: dict) -> None:
        raw = json.dumps(message)
        with self.websocket_lock:
            if self.websocket is not None:
                self.websocket.send(raw)

    def on_bus_message(self, _bus, message) -> None:
        if message.type == Gst.MessageType.ERROR:
            error, debug = message.parse_error()
            log(f"GStreamer error: {error}; debug={debug}")
            self.stop_loop()
        elif message.type == Gst.MessageType.WARNING:
            warning, debug = message.parse_warning()
            log(f"GStreamer warning: {warning}; debug={debug}")

    def stop_loop(self) -> bool:
        try:
            self.loop.quit()
        except Exception:
            pass
        return False

    def cleanup(self) -> None:
        self.websocket_closed.set()
        if self.websocket is not None:
            try:
                self.websocket.close()
            except Exception:
                pass
        if self.pipeline is not None:
            self.pipeline.set_state(Gst.State.NULL)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--server-url", required=True)
    parser.add_argument("--robot-token", default=os.environ.get("ROBOT_TOKEN", ""))
    parser.add_argument("--robot-code", required=True)
    parser.add_argument("--bundle-policy", default="max-bundle", choices=["max-bundle", "max-compat", "balanced"])
    parser.add_argument("--rgb-width", type=int, default=640)
    parser.add_argument("--rgb-height", type=int, default=360)
    parser.add_argument("--thermal-width", type=int, default=640)
    parser.add_argument("--thermal-height", type=int, default=360)
    parser.add_argument("--fps", type=int, default=15)
    args = parser.parse_args()
    if not args.robot_token:
        parser.error("--robot-token or ROBOT_TOKEN is required")
    return args


def main() -> int:
    args = parse_args()
    robot = GStreamerMockRobot(args)

    def handle_signal(_signum, _frame) -> None:
        robot.stop()

    signal.signal(signal.SIGTERM, handle_signal)
    signal.signal(signal.SIGINT, handle_signal)
    return robot.run()


if __name__ == "__main__":
    sys.exit(main())
