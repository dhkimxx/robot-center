# GStreamer Mock Robot

This mock robot is a Linux/Docker harness for validating robot publishers that use GStreamer `webrtcbin`.

It follows the robot-facing contract:

```text
1. POST /api/v1/robot/heartbeat
2. GET /api/v1/robot/mission
3. connect mission.sfu.signalingUrl
4. create media tracks and canonical DataChannels before create-offer
5. send WebRTC offer
6. apply SFU answer
7. publish media and DataChannel payloads after DataChannel open
```

The mock intentionally uses:

- GStreamer `webrtcbin`
- `bundle-policy=max-bundle` by default
- OPUS audio
- two VP8 video test tracks
- `channel.telemetry`, `channel.spatial`, `channel.event`, `channel.control`

Run through the project script:

```bash
APP_SERVER_URL=http://control-server.example:18080 ./scripts/mock-robot-gstreamer.sh
```

The script creates or resolves one robot, ensures an active mission, builds the Docker image, and keeps the mock running in a `screen` session.
