# Python Mock Robot

This app is a local robot publisher used when Android devices are unavailable.

It follows the same control-plane and media-plane path as the robot sample:

```text
Python Mock Robot
-> robot-gateway REST
-> WebRTC publish once
-> Go SFU
-> Browser subscribe
-> Recorder subscribe
```

Run through the project script:

```bash
./scripts/python-mock-robots-up.sh
```

The script starts two mock robots by default and keeps them running in `screen`
sessions:

- `robot-center-pyrobot-001`
- `robot-center-pyrobot-002`

The mock publishes:

- RGB video: synthetic H.264-preferred track
- Thermal video: synthetic H.264-preferred track
- Audio: silent Opus-preferred track
- DataChannels: `sensor`, `telemetry`
