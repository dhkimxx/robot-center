# Python Mock Robot

This app is a Python-based robot publisher used when Android devices are unavailable.

It follows the same control-plane and media-plane path as the robot sample:

```text
Python Mock Robot
1. heartbeat -> robot-gateway REST
2. active mission lookup
3. SFU signaling WebSocket connect
4. WebRTC offer send
5. WebRTC answer receive
6. media/DataChannel publish once to the mission room
```

Run through the project script:

```bash
APP_SERVER_URL=http://control-server.example:18080 ./scripts/mock-robots-python.sh
```

`APP_SERVER_URL` is required and must be reachable from the machine running the mock robot.

The script starts two mock robots by default and keeps them running in `screen`
sessions:

- `robot-center-pyrobot-001`
- `robot-center-pyrobot-002`

The mock publishes:

- `track.video_1`: RGB demo video
- `track.video_2`: Thermal demo video
- `track.audio_1`: silent Opus-preferred audio
- `track.audio_2`: reserved secondary audio slot
- `channel.telemetry`: `SensorDescriptor` / `SensorSample` payloads for GPS, environment, battery
- `channel.spatial`: `SensorDescriptor` / `SensorSample` payloads for IMU and odometry
- `channel.event`: robot heartbeat events
- `channel.control`: reserved stub; the mock creates the channel but does not emit control payloads

Runtime options keep the robot-team integration path explicit:

```bash
python3 apps/mock-robot-python/mock_robot.py \
  --server-url http://control-server.example:18080 \
  --robot-code robot-001 \
  --robot-token <token> \
  --rgb-width 1280 \
  --rgb-height 720 \
  --thermal-width 640 \
  --thermal-height 480 \
  --fps 15
```
