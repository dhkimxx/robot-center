#!/usr/bin/env python3
import argparse
import asyncio
import json
import math
import os
import signal
from datetime import datetime, timezone
from fractions import Fraction
from typing import Any
from urllib.parse import parse_qsl, urlencode, urlsplit, urlunsplit

import aiohttp
import av
import numpy as np
from aiortc import (
    AudioStreamTrack,
    RTCConfiguration,
    RTCIceServer,
    RTCPeerConnection,
    RTCSessionDescription,
    RTCRtpSender,
    VideoStreamTrack,
)
from aiortc.sdp import candidate_from_sdp

INITIAL_RECONNECT_DELAY_SECONDS = 1
MAX_RECONNECT_DELAY_SECONDS = 30
DETECTION_BBOX_PATTERNS: tuple[dict[str, float], ...] = (
    {"x": 0.50, "y": 0.50, "width": 0.04, "height": 0.04},
    {"x": 0.22, "y": 0.24, "width": 0.18, "height": 0.18},
    {"x": 0.05, "y": 0.12, "width": 0.80, "height": 0.14},
    {"x": 0.43, "y": 0.06, "width": 0.14, "height": 0.82},
    {"x": 0.12, "y": 0.12, "width": 0.70, "height": 0.64},
    {"x": 0.72, "y": 0.68, "width": 0.28, "height": 0.32},
)


def utc_now_iso() -> str:
    return datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")


def trim_trailing_slash(value: str) -> str:
    return value.rstrip("/")


def websocket_url_with_room(base_url: str, room_id: str) -> str:
    split_url = urlsplit(base_url)
    query_items = [
        (key, value)
        for key, value in parse_qsl(split_url.query, keep_blank_values=True)
        if key != "room"
    ]
    query_items.append(("room", room_id))
    return urlunsplit(
        (
            split_url.scheme,
            split_url.netloc,
            split_url.path,
            urlencode(query_items),
            split_url.fragment,
        )
    )


def robot_signaling_url_from_server_url(server_url: str) -> str:
    split_url = urlsplit(server_url)
    scheme = "wss" if split_url.scheme == "https" else "ws"
    return urlunsplit((scheme, split_url.netloc, "/api/v1/robot/sfu/ws", "", ""))


def strip_non_relay_candidates(sdp: str) -> str:
    lines = []
    for line in sdp.splitlines():
        if line.startswith("a=candidate:") and " typ relay " not in f" {line} ":
            continue
        lines.append(line)
    return "\r\n".join(lines) + "\r\n"


def is_relay_candidate(candidate_line: str) -> bool:
    return " typ relay " in f" {candidate_line} " or candidate_line.endswith(" typ relay")


def summarize_ice_candidates(sdp: str) -> str:
    counts = {"relay": 0, "host": 0, "srflx": 0, "prflx": 0, "unknown": 0}
    total = 0
    for line in sdp.splitlines():
        if not line.startswith("a=candidate:"):
            continue
        total += 1
        candidate_type = "unknown"
        parts = line.split()
        if "typ" in parts:
            type_index = parts.index("typ") + 1
            if type_index < len(parts):
                candidate_type = parts[type_index]
        counts[candidate_type if candidate_type in counts else "unknown"] += 1
    return (
        f"total={total} relay={counts['relay']} host={counts['host']} "
        f"srflx={counts['srflx']} prflx={counts['prflx']} unknown={counts['unknown']}"
    )


def codec_preferences(kind: str, preferred_mime_type: str) -> list[Any]:
    capabilities = RTCRtpSender.getCapabilities(kind)
    preferred = [
        codec
        for codec in capabilities.codecs
        if codec.mimeType.lower() == preferred_mime_type.lower()
    ]
    fallback = [
        codec
        for codec in capabilities.codecs
        if codec.mimeType.lower() != preferred_mime_type.lower()
    ]
    return preferred + fallback


ROBOT_PALETTES = [
    ((36, 123, 255), (18, 220, 155), (255, 255, 255)),
    ((255, 92, 92), (255, 190, 64), (35, 42, 60)),
    ((128, 92, 255), (78, 220, 255), (255, 255, 255)),
    ((36, 180, 105), (250, 225, 65), (18, 35, 45)),
    ((255, 110, 205), (108, 240, 170), (255, 255, 255)),
    ((245, 130, 32), (60, 210, 230), (20, 26, 38)),
]

SEVEN_SEGMENT_DIGITS = {
    "0": ("a", "b", "c", "d", "e", "f"),
    "1": ("b", "c"),
    "2": ("a", "b", "g", "e", "d"),
    "3": ("a", "b", "c", "d", "g"),
    "4": ("f", "g", "b", "c"),
    "5": ("a", "f", "g", "c", "d"),
    "6": ("a", "f", "g", "c", "d", "e"),
    "7": ("a", "b", "c"),
    "8": ("a", "b", "c", "d", "e", "f", "g"),
    "9": ("a", "b", "c", "d", "f", "g"),
}


def robot_number(robot_code: str) -> int:
    suffix = "".join(char for char in robot_code if char.isdigit())
    if not suffix:
        return sum(ord(char) for char in robot_code) % 1000
    return int(suffix[-3:])


class SyntheticVideoTrack(VideoStreamTrack):
    def __init__(self, robot_code: str, label: str, slot: str, width: int, height: int, fps: int):
        super().__init__()
        self._id = slot
        self.robot_code = robot_code
        self.label = label
        self.slot = slot
        self.width = width
        self.height = height
        self.fps = fps
        self.sequence = 0
        self.timestamp = 0
        self.frame_interval_seconds = 1 / max(fps, 1)
        self.robot_offset = sum(ord(char) for char in robot_code) % 255
        self.robot_number = robot_number(robot_code)
        self.palette = ROBOT_PALETTES[(self.robot_number - 1) % len(ROBOT_PALETTES)]

    async def recv(self) -> av.VideoFrame:
        await asyncio.sleep(self.frame_interval_seconds)
        self.sequence += 1
        self.timestamp += int(90000 / max(self.fps, 1))

        if self.label == "thermal":
            image = self.create_thermal_frame()
        else:
            image = self.create_rgb_frame()

        frame = av.VideoFrame.from_ndarray(image, format="rgb24")
        frame.pts = self.timestamp
        frame.time_base = Fraction(1, 90000)
        return frame

    def create_rgb_frame(self) -> np.ndarray:
        y_indices, x_indices = np.indices((self.height, self.width))
        base = np.zeros((self.height, self.width, 3), dtype=np.uint8)
        motion = (self.sequence * 5) % max(self.width, 1)
        primary = np.array(self.palette[0], dtype=np.uint8)
        secondary = np.array(self.palette[1], dtype=np.uint8)
        text_color = np.array(self.palette[2], dtype=np.uint8)

        robot_shift = self.robot_number * 27
        base[:, :, 0] = (x_indices * 255 // max(self.width, 1) + int(primary[0]) + robot_shift) % 255
        base[:, :, 1] = (y_indices * 255 // max(self.height, 1) + int(primary[1])) % 255
        base[:, :, 2] = (self.sequence * 2 + int(primary[2]) + robot_shift) % 255

        stripe_width = max(self.width // (10 + self.robot_number % 5), 28)
        stripe_mask = ((x_indices + y_indices + self.sequence * 6) // stripe_width) % 2 == 0
        base[stripe_mask] = ((base[stripe_mask].astype(np.uint16) * 2 + secondary.astype(np.uint16)) // 3).astype(np.uint8)

        bar_width = max(self.width // 28, 12)
        base[:, motion : min(motion + bar_width, self.width), :] = np.array([255, 255, 255], dtype=np.uint8)
        self.draw_status_blocks(base, secondary)
        self.draw_robot_identity(base, primary, secondary, text_color)
        return base

    def create_thermal_frame(self) -> np.ndarray:
        y_indices, x_indices = np.indices((self.height, self.width))
        primary = np.array(self.palette[0], dtype=np.uint8)
        secondary = np.array(self.palette[1], dtype=np.uint8)
        text_color = np.array(self.palette[2], dtype=np.uint8)
        phase = self.sequence / 12 + self.robot_number
        heat = (
            128
            + 70 * np.sin((x_indices / max(self.width, 1)) * math.pi * (3 + self.robot_number % 4) + phase)
            + 55 * np.cos((y_indices / max(self.height, 1)) * math.pi * (2 + self.robot_number % 3) - phase)
        )
        heat = np.clip(heat, 0, 255).astype(np.uint8)
        frame = np.zeros((self.height, self.width, 3), dtype=np.uint8)
        frame[:, :, 0] = np.clip(heat.astype(np.int16) + primary[0] // 5, 0, 255).astype(np.uint8)
        frame[:, :, 1] = np.clip(heat.astype(np.int16) // 2 + secondary[1] // 3, 0, 255).astype(np.uint8)
        frame[:, :, 2] = np.clip(255 - heat.astype(np.int16) + primary[2] // 6, 0, 255).astype(np.uint8)

        hotspot_x = (self.sequence * (3 + self.robot_number % 5) + self.robot_offset) % max(self.width, 1)
        hotspot_y = int(self.height * (0.28 + 0.14 * (self.robot_number % 4)))
        radius = max(min(self.width, self.height) // 10, 18)
        mask = (x_indices - hotspot_x) ** 2 + (y_indices - hotspot_y) ** 2 < radius**2
        frame[mask] = secondary
        self.draw_status_blocks(frame, primary)
        self.draw_robot_identity(frame, primary, secondary, text_color)
        return frame

    def draw_status_blocks(self, image: np.ndarray, color: np.ndarray) -> None:
        pad = max(self.width // 80, 8)
        block_height = max(self.height // 12, 18)
        block_width = max(self.width // 5, 80)
        image[pad : pad + block_height, pad : pad + block_width, :] = color

        pulse_width = int((self.sequence % 40) / 40 * block_width)
        pulse_top = pad + block_height + pad
        image[
            pulse_top : pulse_top + max(block_height // 2, 8),
            pad : pad + pulse_width,
            :,
        ] = np.array([250, 250, 250], dtype=np.uint8)

    def draw_robot_identity(
        self,
        image: np.ndarray,
        primary: np.ndarray,
        secondary: np.ndarray,
        text_color: np.ndarray,
    ) -> None:
        pad = max(min(self.width, self.height) // 32, 8)
        band_width = max(self.width // 28, 18)
        image[:, :band_width, :] = primary

        marker_count = 2 + (self.robot_number % 5)
        marker_height = max(self.height // 24, 8)
        marker_gap = max(marker_height // 2, 4)
        for index in range(marker_count):
            top = pad + index * (marker_height + marker_gap)
            bottom = min(top + marker_height, self.height - pad)
            if top < bottom:
                image[top:bottom, : band_width + pad, :] = secondary

        digits = f"{self.robot_number % 1000:03d}"
        digit_height = max(min(self.height // 4, self.width // 8), 34)
        digit_width = max(int(digit_height * 0.58), 18)
        digit_gap = max(digit_width // 5, 4)
        badge_width = digit_width * 3 + digit_gap * 4
        badge_height = digit_height + digit_gap * 2
        x = max(self.width - badge_width - pad, pad)
        y = max(self.height - badge_height - pad, pad)

        self.fill_rect(image, x, y, badge_width, badge_height, np.array([13, 19, 32], dtype=np.uint8))
        self.fill_rect(image, x, y, badge_width, max(digit_gap, 4), primary)
        for digit_index, digit in enumerate(digits):
            digit_x = x + digit_gap + digit_index * (digit_width + digit_gap)
            digit_y = y + digit_gap
            self.draw_seven_segment_digit(image, digit, digit_x, digit_y, digit_width, digit_height, text_color)

    def draw_seven_segment_digit(
        self,
        image: np.ndarray,
        digit: str,
        x: int,
        y: int,
        width: int,
        height: int,
        color: np.ndarray,
    ) -> None:
        active_segments = SEVEN_SEGMENT_DIGITS.get(digit, ())
        thickness = max(width // 6, 3)
        half_height = height // 2
        segments = {
            "a": (x + thickness, y, width - thickness * 2, thickness),
            "b": (x + width - thickness, y + thickness, thickness, half_height - thickness),
            "c": (x + width - thickness, y + half_height, thickness, half_height - thickness),
            "d": (x + thickness, y + height - thickness, width - thickness * 2, thickness),
            "e": (x, y + half_height, thickness, half_height - thickness),
            "f": (x, y + thickness, thickness, half_height - thickness),
            "g": (x + thickness, y + half_height - thickness // 2, width - thickness * 2, thickness),
        }
        for segment in active_segments:
            self.fill_rect(image, *segments[segment], color)

    def fill_rect(
        self,
        image: np.ndarray,
        x: int,
        y: int,
        width: int,
        height: int,
        color: np.ndarray,
    ) -> None:
        x1 = max(0, min(self.width, x))
        y1 = max(0, min(self.height, y))
        x2 = max(0, min(self.width, x + width))
        y2 = max(0, min(self.height, y + height))
        if x1 < x2 and y1 < y2:
            image[y1:y2, x1:x2, :] = color


class SilenceAudioTrack(AudioStreamTrack):
    def __init__(self, slot: str = "track.audio_1"):
        super().__init__()
        self._id = slot
        self.sample_rate = 48000
        self.samples_per_frame = 960
        self.timestamp = 0

    async def recv(self) -> av.AudioFrame:
        await asyncio.sleep(self.samples_per_frame / self.sample_rate)
        frame = av.AudioFrame(format="s16", layout="stereo", samples=self.samples_per_frame)
        for plane in frame.planes:
            plane.update(bytes(plane.buffer_size))
        frame.pts = self.timestamp
        frame.sample_rate = self.sample_rate
        frame.time_base = Fraction(1, self.sample_rate)
        self.timestamp += self.samples_per_frame
        return frame


class MockRobot:
    def __init__(self, args: argparse.Namespace):
        self.server_url = trim_trailing_slash(args.server_url)
        self.robot_code = args.robot_code
        self.robot_token = args.robot_token
        self.rgb_width = args.rgb_width
        self.rgb_height = args.rgb_height
        self.thermal_width = args.thermal_width
        self.thermal_height = args.thermal_height
        self.fps = args.fps
        self.override_mission_id = args.mission_id
        self.override_mission_code = args.mission_code
        self.override_room_id = args.room_id
        self.override_signaling_url = args.signaling_url
        self.override_turn_url = args.turn_url
        self.override_turn_username = args.turn_username
        self.override_turn_password = args.turn_password
        self.session: aiohttp.ClientSession | None = None
        self.peer_connection: RTCPeerConnection | None = None
        self.websocket: aiohttp.ClientWebSocketResponse | None = None
        self.local_peer_id = ""
        self.mission: dict[str, Any] = {}
        self.telemetry_channel = None
        self.event_channel = None
        self.spatial_channel = None
        self.control_channel = None
        self.stop_event = asyncio.Event()
        self.publish_stop_event = asyncio.Event()
        self.publish_tasks: list[asyncio.Task[Any]] = []
        self.pending_remote_candidates: list[dict[str, Any]] = []
        self.sequence = 0
        self.event_sequence = 0
        self.offer_sent = False

    async def run(self) -> None:
        timeout = aiohttp.ClientTimeout(total=10)
        self.session = aiohttp.ClientSession(
            timeout=timeout,
            headers={"Authorization": f"Bearer {self.robot_token}"},
        )
        try:
            await self.run_reconnect_loop()
        finally:
            await self.close()

    async def run_reconnect_loop(self) -> None:
        retry_delay_seconds = INITIAL_RECONNECT_DELAY_SECONDS
        while not self.stop_event.is_set():
            self.reset_publish_state()
            try:
                await self.send_heartbeat("online")
                if self.has_mission_override():
                    self.mission = await self.create_override_mission()
                else:
                    self.mission = await self.wait_for_active_mission()
                await self.send_heartbeat("streaming")
                await self.connect_signaling()
                retry_delay_seconds = INITIAL_RECONNECT_DELAY_SECONDS
                if not self.stop_event.is_set():
                    self.log("publish connection ended; retrying from mission lookup")
            except asyncio.CancelledError:
                raise
            except Exception as error:
                if not self.stop_event.is_set():
                    self.log(f"publish connection failed: {error}")
            finally:
                await self.report_publish_stopped()
                await self.close_publish_resources()

            if self.stop_event.is_set():
                return
            await self.sleep_until_retry(retry_delay_seconds)
            retry_delay_seconds = min(retry_delay_seconds * 2, MAX_RECONNECT_DELAY_SECONDS)

    async def sleep_until_retry(self, delay_seconds: int) -> None:
        self.log(f"reconnect retry in {delay_seconds}s")
        try:
            await asyncio.wait_for(self.stop_event.wait(), timeout=delay_seconds)
        except asyncio.TimeoutError:
            return

    def reset_publish_state(self) -> None:
        self.publish_stop_event = asyncio.Event()
        self.publish_tasks = []
        self.peer_connection = None
        self.websocket = None
        self.local_peer_id = ""
        self.mission = {}
        self.telemetry_channel = None
        self.event_channel = None
        self.spatial_channel = None
        self.control_channel = None
        self.offer_sent = False

    def has_mission_override(self) -> bool:
        return any(
            (
                self.override_mission_id,
                self.override_mission_code,
                self.override_room_id,
                self.override_signaling_url,
                self.override_turn_url,
            )
        )

    async def create_override_mission(self) -> dict[str, Any]:
        mission_id = self.override_mission_id
        mission_code = self.override_mission_code or self.override_room_id
        if not mission_id or not mission_code:
            raise ValueError("--mission-id and --mission-code are required for mission override")

        room_id = self.override_room_id or mission_code
        signaling_url = self.override_signaling_url or robot_signaling_url_from_server_url(self.server_url)
        if not signaling_url:
            raise ValueError("signalingUrl is required for mission override")

        rtc_config = await self.get_json("/api/rtc-config")
        turn_servers = self.create_turn_servers(rtc_config)
        mission = {
            "missionId": mission_id,
            "missionCode": mission_code,
            "missionStatus": "active",
            "sfu": {
                "signalingUrl": websocket_url_with_room(signaling_url, room_id),
                "iceTransportPolicy": rtc_config.get("iceTransportPolicy", "relay"),
            },
            "turnServers": turn_servers,
        }
        self.log(f"mission override active: {mission_code} room={room_id}")
        return mission

    def create_turn_servers(self, rtc_config: dict[str, Any]) -> list[dict[str, Any]]:
        if self.override_turn_url:
            return [
                {
                    "urls": [self.override_turn_url],
                    "username": self.override_turn_username,
                    "credential": self.override_turn_password,
                }
            ]
        return rtc_config.get("iceServers", [])

    def mission_room_id(self, mission: dict[str, Any]) -> str:
        return str(mission.get("roomId") or mission.get("missionCode") or "")

    async def wait_for_active_mission(self) -> dict[str, Any]:
        while not self.stop_event.is_set():
            mission = await self.fetch_active_mission()
            if mission:
                self.log(
                    "mission active: "
                    f"{mission.get('missionCode')} room={self.mission_room_id(mission)}"
                )
                return mission
            self.log("waiting for active mission")
            await asyncio.sleep(2)
        raise asyncio.CancelledError

    async def fetch_active_mission(self) -> dict[str, Any]:
        mission = await self.get_json("/api/v1/robot/mission")
        if mission.get("missionStatus") == "active":
            return mission
        return {}

    async def send_heartbeat(self, state: str) -> None:
        payload = {
            "state": state,
            "batteryPercent": self.current_battery_percent(),
            "networkQuality": "mock-local",
            "sentAt": utc_now_iso(),
        }
        await self.post_json("/api/v1/robot/heartbeat", payload)

    async def connect_signaling(self) -> None:
        signaling_url = self.mission["sfu"]["signalingUrl"]
        self.log(f"signaling connecting: {signaling_url}")
        assert self.session is not None
        headers = {"Authorization": f"Bearer {self.robot_token}"}
        async with self.session.ws_connect(signaling_url, heartbeat=20, headers=headers) as websocket:
            self.websocket = websocket
            self.log("signaling connected")
            heartbeat_task = asyncio.create_task(self.heartbeat_loop())
            try:
                async for message in websocket:
                    if message.type == aiohttp.WSMsgType.TEXT:
                        await self.handle_signaling_message(json.loads(message.data))
                    elif message.type in (
                        aiohttp.WSMsgType.CLOSED,
                        aiohttp.WSMsgType.ERROR,
                    ):
                        break
                    if self.stop_event.is_set() or self.publish_stop_event.is_set():
                        break
            finally:
                heartbeat_task.cancel()
                await asyncio.gather(heartbeat_task, return_exceptions=True)
        self.log("signaling disconnected")

    async def handle_signaling_message(self, message: dict[str, Any]) -> None:
        message_type = message.get("type")
        payload = message.get("payload") or {}
        target_peer_id = payload.get("targetPeerId")
        if target_peer_id and self.local_peer_id and target_peer_id != self.local_peer_id:
            return

        if message_type == "joined":
            self.local_peer_id = payload.get("peerId", "")
            self.log(f"room joined: {payload.get('room')} / {payload.get('role')}")
            return

        if message_type == "mission-ended":
            self.log(f"mission ended signal received: {payload.get('room')}")
            await self.stop_publish("mission ended signal")
            return

        if message_type in ("peer-present", "peer-joined") and payload.get("role") == "sfu":
            if not self.offer_sent:
                await self.create_offer(payload.get("peerId") or "sfu")
            return

        if message_type == "answer":
            await self.handle_answer(payload)
            return

        if message_type == "candidate":
            await self.handle_remote_candidate(payload)
            return

        if message_type == "peer-left":
            self.log(f"peer left: {payload.get('role')} {payload.get('peerId')}")
            return

    async def create_offer(self, target_peer_id: str) -> None:
        self.offer_sent = True
        ice_servers = []
        for server in self.mission.get("turnServers", []):
            urls = server.get("urls") or []
            if isinstance(urls, str):
                urls = [urls]
            ice_servers.append(
                RTCIceServer(
                    urls=urls,
                    username=server.get("username"),
                    credential=server.get("credential"),
                )
            )

        self.peer_connection = RTCPeerConnection(RTCConfiguration(iceServers=ice_servers))
        self.configure_peer_connection()
        self.add_media_tracks()
        self.telemetry_channel = self.peer_connection.createDataChannel("channel.telemetry")
        self.event_channel = self.peer_connection.createDataChannel("channel.event")
        self.spatial_channel = self.peer_connection.createDataChannel("channel.spatial")
        self.control_channel = self.peer_connection.createDataChannel("channel.control")
        self.attach_data_channel_logging("channel.telemetry", self.telemetry_channel)
        self.attach_data_channel_logging("channel.event", self.event_channel)
        self.attach_data_channel_logging("channel.spatial", self.spatial_channel)
        self.attach_data_channel_logging("channel.control", self.control_channel)

        self.publish_tasks.extend(
            asyncio.create_task(self.data_channel_loop(label, channel))
            for label, channel in (
                ("channel.telemetry", self.telemetry_channel),
                ("channel.event", self.event_channel),
            )
        )

        offer = await self.peer_connection.createOffer()
        await self.peer_connection.setLocalDescription(offer)
        await self.wait_for_ice_gathering_complete()

        local_description = self.peer_connection.localDescription
        if local_description is None:
            raise RuntimeError("local offer is missing")

        offer_sdp = strip_non_relay_candidates(local_description.sdp)
        self.log(f"offer candidates: {summarize_ice_candidates(offer_sdp)}")
        await self.send_signal(
            "offer",
            {
                "targetPeerId": target_peer_id,
                "type": local_description.type,
                "sdp": offer_sdp,
            },
        )
        self.log("offer sent")

    def configure_peer_connection(self) -> None:
        assert self.peer_connection is not None

        @self.peer_connection.on("connectionstatechange")
        async def on_connection_state_change() -> None:
            assert self.peer_connection is not None
            self.log(f"pc state: {self.peer_connection.connectionState}")

        @self.peer_connection.on("iceconnectionstatechange")
        async def on_ice_connection_state_change() -> None:
            assert self.peer_connection is not None
            self.log(f"ICE state: {self.peer_connection.iceConnectionState}")

        @self.peer_connection.on("icegatheringstatechange")
        async def on_ice_gathering_state_change() -> None:
            assert self.peer_connection is not None
            self.log(f"ICE gathering: {self.peer_connection.iceGatheringState}")

    def attach_data_channel_logging(self, label: str, channel: Any) -> None:
        @channel.on("open")
        def on_open() -> None:
            self.log(f"{label} DataChannel open")
            if label not in ("channel.telemetry", "channel.event"):
                self.log(f"{label} DataChannel idle: payload schema not finalized")

        @channel.on("close")
        def on_close() -> None:
            self.log(f"{label} DataChannel closed")

    def add_media_tracks(self) -> None:
        assert self.peer_connection is not None
        rgb_track = SyntheticVideoTrack(
            self.robot_code,
            "rgb",
            "track.video_1",
            self.rgb_width,
            self.rgb_height,
            self.fps,
        )
        thermal_track = SyntheticVideoTrack(
            self.robot_code,
            "thermal",
            "track.video_2",
            self.thermal_width,
            self.thermal_height,
            self.fps,
        )
        audio_track = SilenceAudioTrack("track.audio_1")

        rgb_transceiver = self.peer_connection.addTransceiver(rgb_track, direction="sendonly")
        thermal_transceiver = self.peer_connection.addTransceiver(
            thermal_track,
            direction="sendonly",
        )
        audio_transceiver = self.peer_connection.addTransceiver(audio_track, direction="sendonly")

        video_preferences = codec_preferences("video", "video/H264")
        audio_preferences = codec_preferences("audio", "audio/opus")
        if video_preferences:
            rgb_transceiver.setCodecPreferences(video_preferences)
            thermal_transceiver.setCodecPreferences(video_preferences)
        if audio_preferences:
            audio_transceiver.setCodecPreferences(audio_preferences)

    async def handle_answer(self, payload: dict[str, Any]) -> None:
        if self.peer_connection is None:
            return
        answer_sdp = payload.get("sdp", "")
        if not answer_sdp:
            return
        self.log(f"answer candidates: {summarize_ice_candidates(answer_sdp)}")
        await self.peer_connection.setRemoteDescription(
            RTCSessionDescription(sdp=answer_sdp, type=payload.get("type", "answer"))
        )
        self.log("answer applied")
        await self.apply_pending_remote_candidates()

    async def handle_remote_candidate(self, payload: dict[str, Any]) -> None:
        candidate_line = payload.get("candidate", "")
        if not candidate_line:
            return
        if not is_relay_candidate(candidate_line):
            self.log("remote non-relay candidate ignored")
            return
        if self.peer_connection is None:
            return
        if self.peer_connection.remoteDescription is None:
            self.pending_remote_candidates.append(payload)
            return
        await self.apply_remote_candidate(payload)

    async def apply_pending_remote_candidates(self) -> None:
        pending_candidates = self.pending_remote_candidates
        self.pending_remote_candidates = []
        for payload in pending_candidates:
            await self.apply_remote_candidate(payload)

    async def apply_remote_candidate(self, payload: dict[str, Any]) -> None:
        if self.peer_connection is None:
            return
        candidate_line = payload.get("candidate", "")
        if not candidate_line:
            return
        candidate = candidate_from_sdp(candidate_line)
        candidate.sdpMid = payload.get("sdpMid")
        candidate.sdpMLineIndex = payload.get("sdpMLineIndex")
        await self.peer_connection.addIceCandidate(candidate)
        self.log("remote relay candidate added")

    async def wait_for_ice_gathering_complete(self) -> None:
        assert self.peer_connection is not None
        for _ in range(100):
            if self.peer_connection.iceGatheringState == "complete":
                return
            await asyncio.sleep(0.1)

    async def data_channel_loop(self, label: str, channel: Any) -> None:
        while not self.stop_event.is_set() and not self.publish_stop_event.is_set():
            await asyncio.sleep(1)
            if channel.readyState != "open":
                continue
            if label == "channel.telemetry":
                self.sequence += 1
                payload = self.create_telemetry_payload()
            elif label == "channel.event":
                self.event_sequence += 1
                payload = self.create_event_payload(self.event_sequence)
            else:
                continue
            channel.send(json.dumps(payload, separators=(",", ":")))

    async def heartbeat_loop(self) -> None:
        while not self.stop_event.is_set() and not self.publish_stop_event.is_set():
            await asyncio.sleep(10)
            active_mission = await self.fetch_active_mission()
            if not active_mission or active_mission.get("missionCode") != self.mission.get("missionCode"):
                self.log("active mission disappeared; stopping publish")
                await self.stop_publish("active mission ended")
                return
            try:
                await self.send_heartbeat("streaming")
            except Exception as error:
                self.log(f"heartbeat failed: {error}")

    def create_telemetry_payload(self) -> dict[str, Any]:
        latitude, longitude = self.current_position()
        return {
            "messageId": f"{self.robot_code}-telemetry-{self.sequence}",
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
                    "label": "H2S",
                    "unit": "ppm",
                    "enabled": True,
                },
                {
                    "sensorId": "telemetry.gas.channel_3",
                    "sensorType": "gas",
                    "label": "O2",
                    "unit": "%Vol",
                    "enabled": True,
                },
                {
                    "sensorId": "telemetry.gas.channel_4",
                    "sensorType": "gas",
                    "label": "CH4",
                    "unit": "%LEL",
                    "enabled": True,
                },
                {
                    "sensorId": "telemetry.gas.channel_5",
                    "sensorType": "gas",
                    "label": "TEMP",
                    "unit": "degC",
                    "enabled": True,
                },
                {
                    "sensorId": "telemetry.gas.channel_6",
                    "sensorType": "gas",
                    "label": "HUM",
                    "unit": "RH",
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
                        "altitudeMeter": 45.0 + (self.sequence % 8),
                        "accuracyMeter": 4.5,
                        "headingDegree": (self.sequence * 12) % 360,
                        "speedMeterPerSecond": 0.4 + (self.sequence % 4) * 0.1,
                    },
                },
                {
                    "sensorId": "telemetry.gas.channel_1",
                    "timestamp": utc_now_iso(),
                    "values": {
                        "concentration": 13.0 + self.sequence % 3,
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
                        "concentration": 2.1 + (self.sequence % 3) * 0.1,
                        "scale_code": 1,
                        "alarm_code": 0,
                        "alarm": "normal",
                        "low_alarm": 5.0,
                        "high_alarm": 10.0,
                        "valid": True,
                    },
                },
                {
                    "sensorId": "telemetry.gas.channel_3",
                    "timestamp": utc_now_iso(),
                    "values": {
                        "concentration": 21.1 + math.sin(self.sequence / 10) * 0.2,
                        "scale_code": 1,
                        "alarm_code": 0,
                        "alarm": "normal",
                        "low_alarm": 19.5,
                        "high_alarm": 23.5,
                        "valid": True,
                    },
                },
                {
                    "sensorId": "telemetry.gas.channel_4",
                    "timestamp": utc_now_iso(),
                    "values": {
                        "concentration": 6.8 + self.sequence % 4,
                        "scale_code": 1,
                        "alarm_code": 0,
                        "alarm": "normal",
                        "low_alarm": 20.0,
                        "high_alarm": 40.0,
                        "valid": True,
                    },
                },
                {
                    "sensorId": "telemetry.gas.channel_5",
                    "timestamp": utc_now_iso(),
                    "values": {
                        "concentration": 26.6 + math.sin(self.sequence / 5) * 2,
                        "scale_code": 1,
                        "alarm_code": 0,
                        "alarm": "normal",
                        "low_alarm": 5.0,
                        "high_alarm": 50.0,
                        "valid": True,
                    },
                },
                {
                    "sensorId": "telemetry.gas.channel_6",
                    "timestamp": utc_now_iso(),
                    "values": {
                        "concentration": 48.9 + math.cos(self.sequence / 8) * 5,
                        "scale_code": 1,
                        "alarm_code": 0,
                        "alarm": "normal",
                        "low_alarm": 20.0,
                        "high_alarm": 80.0,
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

    def create_event_payload(self, sequence: int) -> dict[str, Any]:
        occurred_at = utc_now_iso()

        def detection_snapshot_event(track_id: str, offset: int) -> dict[str, Any]:
            tick = sequence + offset
            class_name = ["person", "smoke", "vehicle"][tick % 3]
            bbox = DETECTION_BBOX_PATTERNS[tick % len(DETECTION_BBOX_PATTERNS)]
            should_clear_snapshot = (track_id == "track.video_1" and sequence % 15 == 0) or (
                track_id == "track.video_2" and sequence % 20 == 0
            )
            detections = [] if should_clear_snapshot else [
                {
                    "className": class_name,
                    "confidence": round(0.72 + (tick % 20) / 100, 2),
                    "bbox": {
                        "x": bbox["x"],
                        "y": bbox["y"],
                        "width": bbox["width"],
                        "height": bbox["height"],
                    },
                }
            ]
            return {
                "eventId": f"{self.robot_code}-detect-{track_id.replace('.', '_')}-{sequence}",
                "eventType": "detection.object",
                "timestamp": occurred_at,
                "values": {
                    "trackId": track_id,
                    "detections": detections,
                },
            }

        events: list[dict[str, Any]] = [
            detection_snapshot_event("track.video_1", 0),
            detection_snapshot_event("track.video_2", 1),
        ]
        if sequence % 10 == 0:
            events.append(
                {
                    "eventId": f"{self.robot_code}-mission-{sequence}",
                    "eventType": "mission.event",
                    "timestamp": occurred_at,
                    "values": {
                        "severity": "info",
                        "title": "Mock mission event",
                        "description": f"{self.robot_code} event sequence {sequence}",
                    },
                }
            )
        return {
            "messageId": f"{self.robot_code}-event-{sequence}",
            "messageType": "event",
            "events": events,
        }

    def current_position(self) -> tuple[float, float]:
        robot_delta = (sum(ord(char) for char in self.robot_code) % 20) / 10000
        movement = math.sin(self.sequence / 10) / 10000
        return 37.5665 + robot_delta + movement, 126.9780 + robot_delta + movement

    def current_battery_percent(self) -> int:
        return max(30, 92 - (self.sequence // 15) % 35)

    async def send_signal(self, message_type: str, payload: dict[str, Any]) -> None:
        if self.websocket is None:
            return
        await self.websocket.send_json({"type": message_type, "payload": payload})

    async def get_json(self, path: str) -> dict[str, Any]:
        assert self.session is not None
        async with self.session.get(self.server_url + path) as response:
            response.raise_for_status()
            return await response.json()

    async def post_json(self, path: str, payload: dict[str, Any]) -> dict[str, Any]:
        assert self.session is not None
        async with self.session.post(self.server_url + path, json=payload) as response:
            response.raise_for_status()
            return await response.json()

    async def report_publish_stopped(self) -> None:
        if not self.mission or self.session is None or self.session.closed:
            return
        try:
            await self.send_heartbeat("online")
            self.log("publish stopped")
        except Exception as error:
            self.log(f"publish stopped report failed: {error}")

    async def stop_publish(self, reason: str) -> None:
        self.log(f"stopping publish: {reason}")
        self.publish_stop_event.set()
        if self.websocket is not None and not self.websocket.closed:
            await self.websocket.close()
        if self.peer_connection is not None:
            await self.peer_connection.close()
            self.peer_connection = None

    async def close_publish_resources(self) -> None:
        self.publish_stop_event.set()
        for task in self.publish_tasks:
            task.cancel()
        if self.publish_tasks:
            await asyncio.gather(*self.publish_tasks, return_exceptions=True)
        self.publish_tasks = []
        if self.websocket is not None and not self.websocket.closed:
            await self.websocket.close()
        self.websocket = None
        if self.peer_connection is not None:
            await self.peer_connection.close()
            self.peer_connection = None
        self.telemetry_channel = None
        self.event_channel = None
        self.spatial_channel = None
        self.control_channel = None

    async def close(self) -> None:
        self.stop_event.set()
        self.publish_stop_event.set()
        await self.close_publish_resources()
        if self.session is not None:
            await self.session.close()

    def request_stop(self) -> None:
        self.stop_event.set()
        self.publish_stop_event.set()

    def log(self, message: str) -> None:
        print(f"[{utc_now_iso()}] [{self.robot_code}] {message}", flush=True)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Robot Center Python Mock Robot")
    parser.add_argument("--server-url", required=True)
    parser.add_argument("--robot-code", required=True)
    parser.add_argument("--robot-token", default=os.environ.get("ROBOT_TOKEN"))
    parser.add_argument("--rgb-width", type=int, default=1280)
    parser.add_argument("--rgb-height", type=int, default=720)
    parser.add_argument("--thermal-width", type=int, default=640)
    parser.add_argument("--thermal-height", type=int, default=480)
    parser.add_argument("--fps", type=int, default=15)
    parser.add_argument("--mission-id", default="")
    parser.add_argument("--mission-code", default="")
    parser.add_argument("--room-id", default="")
    parser.add_argument("--signaling-url", default="")
    parser.add_argument("--turn-url", default="")
    parser.add_argument("--turn-username", default="")
    parser.add_argument("--turn-password", default="")
    args = parser.parse_args()
    if not args.robot_token:
        parser.error("--robot-token or ROBOT_TOKEN is required")
    return args


async def main() -> None:
    robot = MockRobot(parse_args())
    loop = asyncio.get_running_loop()
    for signal_name in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(signal_name, robot.request_stop)
    await robot.run()


if __name__ == "__main__":
    asyncio.run(main())
