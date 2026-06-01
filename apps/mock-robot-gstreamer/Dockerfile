FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        gir1.2-gst-plugins-bad-1.0 \
        gstreamer1.0-libav \
        gstreamer1.0-nice \
        gstreamer1.0-plugins-bad \
        gstreamer1.0-plugins-base \
        gstreamer1.0-plugins-good \
        gstreamer1.0-tools \
        python3 \
        python3-gi \
        python3-gst-1.0 \
        python3-requests \
        python3-websocket \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY mock_robot.py /app/mock_robot.py

ENTRYPOINT ["python3", "/app/mock_robot.py"]
