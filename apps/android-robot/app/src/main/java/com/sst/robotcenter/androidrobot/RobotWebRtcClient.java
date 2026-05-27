package com.sst.robotcenter.androidrobot;

import android.content.Context;

import org.json.JSONException;
import org.json.JSONObject;
import org.webrtc.AudioSource;
import org.webrtc.AudioTrack;
import org.webrtc.Camera2Enumerator;
import org.webrtc.CameraVideoCapturer;
import org.webrtc.DataChannel;
import org.webrtc.DefaultVideoDecoderFactory;
import org.webrtc.DefaultVideoEncoderFactory;
import org.webrtc.EglBase;
import org.webrtc.IceCandidate;
import org.webrtc.MediaStreamTrack;
import org.webrtc.MediaConstraints;
import org.webrtc.MediaStream;
import org.webrtc.PeerConnection;
import org.webrtc.PeerConnectionFactory;
import org.webrtc.RtpCapabilities;
import org.webrtc.RtpParameters;
import org.webrtc.RtpReceiver;
import org.webrtc.RtpTransceiver;
import org.webrtc.SurfaceTextureHelper;
import org.webrtc.VideoCapturer;
import org.webrtc.VideoSource;
import org.webrtc.VideoTrack;
import org.webrtc.SessionDescription;

import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;
import okhttp3.WebSocket;
import okhttp3.WebSocketListener;

public final class RobotWebRtcClient {
    private static boolean peerConnectionFactoryInitialized;

    private final Context applicationContext;
    private final RobotWebRtcConfig config;
    private final LogSink logSink;
    private final ExecutorService executor = Executors.newSingleThreadExecutor();

    private OkHttpClient httpClient;
    private WebSocket webSocket;
    private EglBase eglBase;
    private PeerConnectionFactory peerConnectionFactory;
    private final Map<String, PeerSession> peerSessions = new ConcurrentHashMap<>();
    private String localPeerId;
    private LocationProvider locationProvider;
    private ScheduledExecutorService sensorExecutor;
    private ScheduledExecutorService telemetryExecutor;
    private int sensorSequence;
    private int telemetrySequence;
    private long sensorStartedAtMs;
    private long telemetryStartedAtMs;

    private SurfaceTextureHelper cameraSurfaceTextureHelper;
    private SurfaceTextureHelper thermalSurfaceTextureHelper;
    private VideoCapturer cameraCapturer;
    private SyntheticThermalCapturer thermalCapturer;
    private boolean rgbCameraFrontFacing;
    private VideoSource rgbVideoSource;
    private VideoSource thermalVideoSource;
    private VideoTrack rgbVideoTrack;
    private VideoTrack thermalVideoTrack;
    private AudioSource audioSource;
    private AudioTrack audioTrack;

    private static final class PeerSession {
        private final String peerId;
        private final String role;
        private final PeerConnection peerConnection;
        private DataChannel spatialDataChannel;
        private DataChannel telemetryDataChannel;
        private DataChannel eventDataChannel;
        private DataChannel controlDataChannel;
        private boolean pendingOfferUntilIceGatheringComplete;

        private PeerSession(String peerId, String role, PeerConnection peerConnection) {
            this.peerId = peerId;
            this.role = role;
            this.peerConnection = peerConnection;
        }
    }

    public RobotWebRtcClient(Context context, RobotWebRtcConfig config, LogSink logSink) {
        this.applicationContext = context.getApplicationContext();
        this.config = config;
        this.logSink = logSink;
    }

    public void start() {
        executor.execute(() -> {
            try {
                initializeWebRtc();
                startLocationUpdates();
                startLocalMedia();
                connectSignaling();
            } catch (Exception exception) {
                log("start failed: " + exception.getMessage());
                stopInternal();
            }
        });
    }

    public void stop() {
        executor.execute(this::stopInternal);
        executor.shutdown();
    }

    public void switchRgbCamera() {
        executor.execute(() -> {
            if (!(cameraCapturer instanceof CameraVideoCapturer)) {
                log("RGB camera switch unavailable");
                return;
            }
            Camera2Enumerator enumerator = new Camera2Enumerator(applicationContext);
            boolean targetFrontFacing = !rgbCameraFrontFacing;
            String targetCameraName = findCameraName(enumerator, targetFrontFacing);
            if (targetCameraName == null) {
                log("RGB camera switch unavailable: " + cameraFacingName(targetFrontFacing) + " camera missing");
                return;
            }
            log("RGB camera switching to " + cameraFacingName(targetFrontFacing) + "...");
            ((CameraVideoCapturer) cameraCapturer).switchCamera(new CameraVideoCapturer.CameraSwitchHandler() {
                @Override
                public void onCameraSwitchDone(boolean isFrontCamera) {
                    rgbCameraFrontFacing = isFrontCamera;
                    log("RGB camera selected: " + (isFrontCamera ? "front" : "back"));
                }

                @Override
                public void onCameraSwitchError(String error) {
                    log("RGB camera switch failed: " + error);
                }
            }, targetCameraName);
        });
    }

    private void initializeWebRtc() {
        if (!peerConnectionFactoryInitialized) {
            PeerConnectionFactory.InitializationOptions initializationOptions =
                PeerConnectionFactory.InitializationOptions.builder(applicationContext)
                    .setEnableInternalTracer(false)
                    .createInitializationOptions();
            PeerConnectionFactory.initialize(initializationOptions);
            peerConnectionFactoryInitialized = true;
        }

        eglBase = EglBase.create();
        DefaultVideoEncoderFactory encoderFactory = new DefaultVideoEncoderFactory(
            eglBase.getEglBaseContext(),
            true,
            true
        );
        DefaultVideoDecoderFactory decoderFactory = new DefaultVideoDecoderFactory(eglBase.getEglBaseContext());

        peerConnectionFactory = PeerConnectionFactory.builder()
            .setVideoEncoderFactory(encoderFactory)
            .setVideoDecoderFactory(decoderFactory)
            .createPeerConnectionFactory();
    }

    private void startLocationUpdates() {
        locationProvider = new LocationProvider(applicationContext, this::log);
        locationProvider.start();
    }

    private void startLocalMedia() {
        rgbVideoSource = peerConnectionFactory.createVideoSource(false);
        cameraCapturer = createCameraCapturer();
        cameraSurfaceTextureHelper = SurfaceTextureHelper.create("RgbCameraCaptureThread", eglBase.getEglBaseContext());
        cameraCapturer.initialize(cameraSurfaceTextureHelper, applicationContext, rgbVideoSource.getCapturerObserver());
        cameraCapturer.startCapture(
            RobotMediaProfile.RGB_WIDTH,
            RobotMediaProfile.RGB_HEIGHT,
            RobotMediaProfile.RGB_FPS
        );
        rgbVideoTrack = peerConnectionFactory.createVideoTrack(RobotStreamRoles.TRACK_VIDEO_1, rgbVideoSource);
        rgbVideoTrack.setEnabled(true);

        thermalVideoSource = peerConnectionFactory.createVideoSource(false);
        thermalCapturer = new SyntheticThermalCapturer();
        thermalSurfaceTextureHelper = SurfaceTextureHelper.create("ThermalCaptureThread", eglBase.getEglBaseContext());
        thermalCapturer.initialize(thermalSurfaceTextureHelper, applicationContext, thermalVideoSource.getCapturerObserver());
        thermalCapturer.startCapture(
            RobotMediaProfile.THERMAL_WIDTH,
            RobotMediaProfile.THERMAL_HEIGHT,
            RobotMediaProfile.THERMAL_FPS
        );
        thermalVideoTrack = peerConnectionFactory.createVideoTrack(RobotStreamRoles.TRACK_VIDEO_2, thermalVideoSource);
        thermalVideoTrack.setEnabled(true);

        audioSource = peerConnectionFactory.createAudioSource(new MediaConstraints());
        audioTrack = peerConnectionFactory.createAudioTrack(RobotStreamRoles.TRACK_AUDIO_1, audioSource);
        audioTrack.setEnabled(true);

        log("local media started: RGB " + RobotMediaProfile.RGB_WIDTH + "x"
            + RobotMediaProfile.RGB_HEIGHT + "@" + RobotMediaProfile.RGB_FPS
            + ", thermal " + RobotMediaProfile.THERMAL_WIDTH + "x"
            + RobotMediaProfile.THERMAL_HEIGHT + "@" + RobotMediaProfile.THERMAL_FPS
            + ", Opus audio");
    }

    private VideoCapturer createCameraCapturer() {
        Camera2Enumerator enumerator = new Camera2Enumerator(applicationContext);
        for (String deviceName : enumerator.getDeviceNames()) {
            if (enumerator.isFrontFacing(deviceName)) {
                VideoCapturer capturer = enumerator.createCapturer(deviceName, null);
                if (capturer != null) {
                    rgbCameraFrontFacing = true;
                    log("RGB camera selected: front");
                    return capturer;
                }
            }
        }
        for (String deviceName : enumerator.getDeviceNames()) {
            if (!enumerator.isBackFacing(deviceName)) {
                continue;
            }
            VideoCapturer capturer = enumerator.createCapturer(deviceName, null);
            if (capturer != null) {
                rgbCameraFrontFacing = false;
                log("RGB camera selected: back");
                return capturer;
            }
        }
        throw new IllegalStateException("no camera capturer available");
    }

    private String findCameraName(Camera2Enumerator enumerator, boolean frontFacing) {
        for (String deviceName : enumerator.getDeviceNames()) {
            if (frontFacing && enumerator.isFrontFacing(deviceName)) {
                return deviceName;
            }
            if (!frontFacing && enumerator.isBackFacing(deviceName)) {
                return deviceName;
            }
        }
        return null;
    }

    private String cameraFacingName(boolean frontFacing) {
        return frontFacing ? "front" : "back";
    }

    private void connectSignaling() {
        httpClient = new OkHttpClient.Builder().build();
        Request request = new Request.Builder()
            .url(config.signalingUrl)
            .header("Authorization", "Bearer " + config.robotToken)
            .build();
        webSocket = httpClient.newWebSocket(request, new WebSocketListener() {
            @Override
            public void onOpen(WebSocket webSocket, Response response) {
                log("signaling connected");
            }

            @Override
            public void onMessage(WebSocket webSocket, String text) {
                executor.execute(() -> handleSignalingMessage(text));
            }

            @Override
            public void onClosed(WebSocket webSocket, int code, String reason) {
                log("signaling closed: " + code + " " + reason);
            }

            @Override
            public void onFailure(WebSocket webSocket, Throwable throwable, Response response) {
                log("signaling failed: " + throwable.getMessage());
            }
        });
        log("signaling connecting: " + config.signalingUrl + " / room " + config.roomId);
    }

    private void handleSignalingMessage(String text) {
        try {
            JSONObject message = new JSONObject(text);
            String type = message.optString("type");
            JSONObject payload = message.optJSONObject("payload");
            if (payload == null) {
                payload = new JSONObject();
            }
            String targetPeerId = payload.optString("targetPeerId");
            if (!targetPeerId.isEmpty() && localPeerId != null && !targetPeerId.equals(localPeerId)) {
                return;
            }

            if ("answer".equals(type)) {
                handleAnswer(payload);
                return;
            }
            if ("candidate".equals(type)) {
                handleRemoteCandidate(payload);
                return;
            }
            if ("peer-joined".equals(type) || "peer-present".equals(type)) {
                String role = payload.optString("role");
                if ("sfu".equals(role)) {
                    createFreshOffer(payload.optString("peerId"), role, type);
                }
                return;
            }
            if ("peer-left".equals(type)) {
                closePeerSession(payload.optString("peerId"));
                log("peer left: " + payload.optString("role") + " " + payload.optString("peerId"));
                return;
            }
            if ("joined".equals(type)) {
                localPeerId = payload.optString("peerId");
                log("room joined: " + payload.optString("room") + " / " + payload.optString("role"));
                return;
            }
            if ("waiting-for-peer".equals(type)) {
                log("waiting for SFU");
                return;
            }
            log("server message: " + type);
        } catch (JSONException exception) {
            log("invalid signaling message: " + exception.getMessage());
        }
    }

    private void createFreshOffer(String targetPeerId, String role, String reason) {
        if (targetPeerId == null || targetPeerId.isEmpty()) {
            log("ignored peer without id: " + role);
            return;
        }
        closePeerSession(targetPeerId);
        log("creating fresh offer for " + role + " " + targetPeerId + ": " + reason);

        PeerConnection.IceServer iceServer = PeerConnection.IceServer.builder(config.turnUrl)
            .setUsername(config.turnUsername)
            .setPassword(config.turnPassword)
            .createIceServer();
        PeerConnection.RTCConfiguration rtcConfiguration =
            new PeerConnection.RTCConfiguration(Collections.singletonList(iceServer));
        rtcConfiguration.iceTransportsType = PeerConnection.IceTransportsType.RELAY;
        rtcConfiguration.sdpSemantics = PeerConnection.SdpSemantics.UNIFIED_PLAN;

        PeerConnection peerConnection = peerConnectionFactory.createPeerConnection(
            rtcConfiguration,
            createPeerConnectionObserver(targetPeerId)
        );
        if (peerConnection == null) {
            throw new IllegalStateException("failed to create PeerConnection");
        }
        PeerSession session = new PeerSession(targetPeerId, role, peerConnection);
        peerSessions.put(targetPeerId, session);

        addVideoTrackWithH264Preference(
            peerConnection,
            rgbVideoTrack,
            RobotStreamRoles.TRACK_VIDEO_1,
            "RGB",
            RobotMediaProfile.RGB_MAX_BITRATE_BPS,
            RobotMediaProfile.RGB_MIN_BITRATE_BPS,
            RobotMediaProfile.RGB_FPS,
            1.0
        );
        addVideoTrackWithH264Preference(
            peerConnection,
            thermalVideoTrack,
            RobotStreamRoles.TRACK_VIDEO_2,
            "Thermal",
            RobotMediaProfile.THERMAL_MAX_BITRATE_BPS,
            RobotMediaProfile.THERMAL_MIN_BITRATE_BPS,
            RobotMediaProfile.THERMAL_FPS,
            1.0
        );
        peerConnection.addTrack(audioTrack, Collections.singletonList(RobotStreamRoles.TRACK_AUDIO_1));
        session.telemetryDataChannel = createDataChannel(session, RobotStreamRoles.CHANNEL_TELEMETRY);
        session.spatialDataChannel = createDataChannel(session, RobotStreamRoles.CHANNEL_SPATIAL);
        session.eventDataChannel = createDataChannel(session, RobotStreamRoles.CHANNEL_EVENT);
        session.controlDataChannel = createDataChannel(session, RobotStreamRoles.CHANNEL_CONTROL);

        peerConnection.createOffer(new SimpleSdpObserver() {
            @Override
            public void onCreateSuccess(SessionDescription description) {
                if (!isCurrentSession(session)) {
                    return;
                }
                SessionDescription sanitizedDescription = new SessionDescription(
                    description.type,
                    stripNonRelaySdpCandidates(description.description)
                );
                peerConnection.setLocalDescription(new SimpleSdpObserver() {
                    @Override
                    public void onSetSuccess() {
                        session.pendingOfferUntilIceGatheringComplete = true;
                        sendPendingOfferIfIceGatheringComplete(session);
                    }

                    @Override
                    public void onSetFailure(String error) {
                        log("set local offer failed for " + session.role + ": " + error);
                    }
                }, sanitizedDescription);
            }

            @Override
            public void onCreateFailure(String error) {
                log("create offer failed for " + session.role + ": " + error);
            }
        }, new MediaConstraints());
    }

    private DataChannel createDataChannel(PeerSession session, String label) {
        DataChannel dataChannel = session.peerConnection.createDataChannel(label, new DataChannel.Init());
        dataChannel.registerObserver(createDataChannelObserver(session, label));
        return dataChannel;
    }

    private void addVideoTrackWithH264Preference(
        PeerConnection peerConnection,
        VideoTrack track,
        String streamId,
        String label,
        int maxBitrateBps,
        int minBitrateBps,
        int maxFramerate,
        double scaleResolutionDownBy
    ) {
        RtpTransceiver.RtpTransceiverInit transceiverInit = new RtpTransceiver.RtpTransceiverInit(
            RtpTransceiver.RtpTransceiverDirection.SEND_ONLY,
            Collections.singletonList(streamId)
        );
        RtpTransceiver transceiver = peerConnection.addTransceiver(track, transceiverInit);
        if (transceiver == null) {
            throw new IllegalStateException("failed to add " + label + " video transceiver");
        }
        preferH264Codec(transceiver, label);
        configureVideoEncoding(
            transceiver,
            label,
            maxBitrateBps,
            minBitrateBps,
            maxFramerate,
            scaleResolutionDownBy
        );
    }

    private void configureVideoEncoding(
        RtpTransceiver transceiver,
        String label,
        int maxBitrateBps,
        int minBitrateBps,
        int maxFramerate,
        double scaleResolutionDownBy
    ) {
        RtpParameters parameters = transceiver.getSender().getParameters();
        if (parameters == null || parameters.encodings == null || parameters.encodings.isEmpty()) {
            log(label + " video encoding preference skipped: encodings missing");
            return;
        }

        parameters.degradationPreference = RtpParameters.DegradationPreference.MAINTAIN_RESOLUTION;
        RtpParameters.Encoding encoding = parameters.encodings.get(0);
        encoding.active = true;
        encoding.maxBitrateBps = maxBitrateBps;
        encoding.minBitrateBps = minBitrateBps;
        encoding.maxFramerate = maxFramerate;
        encoding.scaleResolutionDownBy = scaleResolutionDownBy;

        boolean applied = transceiver.getSender().setParameters(parameters);
        if (applied) {
            log(label + " video encoding preference: max " + (maxBitrateBps / 1000)
                + "kbps, min " + (minBitrateBps / 1000) + "kbps, "
                + maxFramerate + "fps, scale " + scaleResolutionDownBy);
        } else {
            log(label + " video encoding preference failed");
        }
    }

    private void preferH264Codec(RtpTransceiver transceiver, String label) {
        RtpCapabilities capabilities =
            peerConnectionFactory.getRtpSenderCapabilities(MediaStreamTrack.MediaType.MEDIA_TYPE_VIDEO);
        if (capabilities == null || capabilities.codecs == null) {
            log(label + " video codec preference skipped: capabilities missing");
            return;
        }

        List<RtpCapabilities.CodecCapability> preferredCodecs = new ArrayList<>();
        List<RtpCapabilities.CodecCapability> fallbackCodecs = new ArrayList<>();
        for (RtpCapabilities.CodecCapability codec : capabilities.codecs) {
            if (isH264Codec(codec)) {
                preferredCodecs.add(codec);
            } else {
                fallbackCodecs.add(codec);
            }
        }
        if (preferredCodecs.isEmpty()) {
            log(label + " video codec preference skipped: H.264 unavailable");
            return;
        }

        preferredCodecs.addAll(fallbackCodecs);
        try {
            transceiver.setCodecPreferences(preferredCodecs);
            log(label + " video codec preference: H.264 first");
        } catch (RuntimeException exception) {
            log(label + " video codec preference failed: " + exception.getMessage());
        }
    }

    private boolean isH264Codec(RtpCapabilities.CodecCapability codec) {
        return "video/H264".equalsIgnoreCase(codec.mimeType)
            || "H264".equalsIgnoreCase(codec.name);
    }

    private PeerConnection.Observer createPeerConnectionObserver(String peerId) {
        return new PeerConnection.Observer() {
            @Override
            public void onSignalingChange(PeerConnection.SignalingState signalingState) {
                PeerSession session = peerSessions.get(peerId);
                log("signaling state " + peerLabel(session, peerId) + ": " + signalingState);
            }

            @Override
            public void onIceConnectionChange(PeerConnection.IceConnectionState iceConnectionState) {
                PeerSession session = peerSessions.get(peerId);
                log("ICE state " + peerLabel(session, peerId) + ": " + iceConnectionState);
            }

            @Override
            public void onIceConnectionReceivingChange(boolean receiving) {
            }

            @Override
            public void onIceGatheringChange(PeerConnection.IceGatheringState iceGatheringState) {
                PeerSession session = peerSessions.get(peerId);
                log("ICE gathering " + peerLabel(session, peerId) + ": " + iceGatheringState);
                if (iceGatheringState == PeerConnection.IceGatheringState.COMPLETE) {
                    if (session != null) {
                        sendPendingOfferIfIceGatheringComplete(session);
                        sendSignal("candidate", createEndOfCandidatesPayload(), session.peerId);
                    }
                }
            }

            @Override
            public void onIceCandidate(IceCandidate candidate) {
                PeerSession session = peerSessions.get(peerId);
                if (session == null) {
                    return;
                }
                String candidateLine = normalizeLocalCandidate(candidate.sdp);
                if (!isRelayCandidate(candidateLine)) {
                    log("local non-relay candidate blocked for " + session.role);
                    return;
                }
                sendSignal("candidate", createCandidatePayload(candidate, candidateLine), session.peerId);
            }

            @Override
            public void onIceCandidatesRemoved(IceCandidate[] candidates) {
            }

            @Override
            public void onAddStream(MediaStream stream) {
            }

            @Override
            public void onRemoveStream(MediaStream stream) {
            }

            @Override
            public void onDataChannel(DataChannel dataChannel) {
            }

            @Override
            public void onRenegotiationNeeded() {
            }

            @Override
            public void onAddTrack(RtpReceiver receiver, MediaStream[] mediaStreams) {
            }
        };
    }

    private DataChannel.Observer createDataChannelObserver(PeerSession session, String label) {
        return new DataChannel.Observer() {
            @Override
            public void onBufferedAmountChange(long previousAmount) {
            }

            @Override
            public void onStateChange() {
                if (!isCurrentSession(session)) {
                    return;
                }
                DataChannel channel = dataChannelForLabel(session, label);
                if (channel == null) {
                    return;
                }
                log(label + " DataChannel " + peerLabel(session, session.peerId) + ": " + channel.state());
                if (channel.state() == DataChannel.State.OPEN) {
                    if (RobotStreamRoles.CHANNEL_SPATIAL.equals(label)) {
                        startSensorStreaming();
                        return;
                    }
                    if (RobotStreamRoles.CHANNEL_TELEMETRY.equals(label)) {
                        startTelemetryStreaming();
                    }
                    return;
                }
                if (channel.state() == DataChannel.State.CLOSED) {
                    if (RobotStreamRoles.CHANNEL_SPATIAL.equals(label)
                        && !hasOpenDataChannel(RobotStreamRoles.CHANNEL_SPATIAL)) {
                        stopSensorStreaming();
                    }
                    if (RobotStreamRoles.CHANNEL_TELEMETRY.equals(label)
                        && !hasOpenDataChannel(RobotStreamRoles.CHANNEL_TELEMETRY)) {
                        stopTelemetryStreaming();
                    }
                }
            }

            @Override
            public void onMessage(DataChannel.Buffer buffer) {
            }
        };
    }

    private void startSensorStreaming() {
        if (sensorExecutor != null) {
            return;
        }
        sensorSequence = 0;
        sensorStartedAtMs = System.currentTimeMillis();
        sensorExecutor = Executors.newSingleThreadScheduledExecutor();
        sensorExecutor.scheduleAtFixedRate(() -> {
            String payload = SensorPayloadFactory.createSensorPayload(
                sensorSequence++,
                sensorStartedAtMs,
                config.robotCode
            );
            byte[] bytes = payload.getBytes(StandardCharsets.UTF_8);
            for (PeerSession session : peerSessions.values()) {
                DataChannel channel = session.spatialDataChannel;
                if (channel == null || channel.state() != DataChannel.State.OPEN) {
                    continue;
                }
                ByteBuffer buffer = ByteBuffer.wrap(bytes);
                channel.send(new DataChannel.Buffer(buffer, false));
            }
        }, 0, 1, TimeUnit.SECONDS);
    }

    private void stopSensorStreaming() {
        if (sensorExecutor != null) {
            sensorExecutor.shutdownNow();
            sensorExecutor = null;
        }
    }

    private void startTelemetryStreaming() {
        if (telemetryExecutor != null) {
            return;
        }
        telemetrySequence = 0;
        telemetryStartedAtMs = System.currentTimeMillis();
        telemetryExecutor = Executors.newSingleThreadScheduledExecutor();
        telemetryExecutor.scheduleAtFixedRate(() -> {
            String payload = SensorPayloadFactory.createTelemetryPayload(
                telemetrySequence++,
                telemetryStartedAtMs,
                config.robotCode,
                locationProvider == null ? null : locationProvider.getLatestLocation()
            );
            byte[] bytes = payload.getBytes(StandardCharsets.UTF_8);
            for (PeerSession session : peerSessions.values()) {
                DataChannel channel = session.telemetryDataChannel;
                if (channel == null || channel.state() != DataChannel.State.OPEN) {
                    continue;
                }
                ByteBuffer buffer = ByteBuffer.wrap(bytes);
                channel.send(new DataChannel.Buffer(buffer, false));
            }
        }, 0, 1, TimeUnit.SECONDS);
    }

    private void stopTelemetryStreaming() {
        if (telemetryExecutor != null) {
            telemetryExecutor.shutdownNow();
            telemetryExecutor = null;
        }
    }

    private boolean hasOpenDataChannel(String label) {
        for (PeerSession session : peerSessions.values()) {
            DataChannel channel = dataChannelForLabel(session, label);
            if (channel != null && channel.state() == DataChannel.State.OPEN) {
                return true;
            }
        }
        return false;
    }

    private DataChannel dataChannelForLabel(PeerSession session, String label) {
        if (session == null) {
            return null;
        }
        if (RobotStreamRoles.CHANNEL_TELEMETRY.equals(label)) {
            return session.telemetryDataChannel;
        }
        if (RobotStreamRoles.CHANNEL_SPATIAL.equals(label)) {
            return session.spatialDataChannel;
        }
        if (RobotStreamRoles.CHANNEL_EVENT.equals(label)) {
            return session.eventDataChannel;
        }
        if (RobotStreamRoles.CHANNEL_CONTROL.equals(label)) {
            return session.controlDataChannel;
        }
        return null;
    }

    private void handleAnswer(JSONObject payload) {
        String fromPeerId = payload.optString("fromPeerId");
        PeerSession session = peerSessionForSignal(fromPeerId);
        if (session == null) {
            log("ignored answer because PeerConnection is missing: " + fromPeerId);
            return;
        }
        if (session.peerConnection.signalingState() != PeerConnection.SignalingState.HAVE_LOCAL_OFFER) {
            log("ignored duplicate answer for " + peerLabel(session, session.peerId)
                + " in state: " + session.peerConnection.signalingState());
            return;
        }
        SessionDescription description = new SessionDescription(
            SessionDescription.Type.ANSWER,
            payload.optString("sdp")
        );
        log("answer candidates " + peerLabel(session, session.peerId)
            + ": relay " + countRelayCandidates(description.description)
            + " / total " + countCandidates(description.description));
        session.peerConnection.setRemoteDescription(new SimpleSdpObserver() {
            @Override
            public void onSetSuccess() {
                log("remote answer applied for " + peerLabel(session, session.peerId));
            }

            @Override
            public void onSetFailure(String error) {
                log("set remote answer failed for " + peerLabel(session, session.peerId) + ": " + error);
            }
        }, description);
    }

    private void handleRemoteCandidate(JSONObject payload) {
        String fromPeerId = payload.optString("fromPeerId");
        PeerSession session = peerSessionForSignal(fromPeerId);
        if (session == null) {
            log("ignored candidate because PeerConnection is missing: " + fromPeerId);
            return;
        }
        String candidateLine = payload.optString("candidate");
        if (candidateLine == null || candidateLine.isEmpty()) {
            return;
        }
        if (!isRelayCandidate(candidateLine)) {
            log("remote non-relay candidate ignored for " + peerLabel(session, session.peerId));
            return;
        }
        log("remote relay candidate added for " + peerLabel(session, session.peerId));
        IceCandidate candidate = new IceCandidate(
            payload.optString("sdpMid"),
            payload.optInt("sdpMLineIndex"),
            candidateLine
        );
        session.peerConnection.addIceCandidate(candidate);
    }

    private void sendSessionDescription(PeerSession session, String type, SessionDescription description) {
        log(type + " candidates " + peerLabel(session, session.peerId)
            + ": relay " + countRelayCandidates(description.description)
            + " / total " + countCandidates(description.description));
        sendSignal(type, createSessionDescriptionPayload(description), session.peerId);
        log(type + " sent to " + peerLabel(session, session.peerId));
    }

    private void sendPendingOfferIfIceGatheringComplete(PeerSession session) {
        if (!isCurrentSession(session) || !session.pendingOfferUntilIceGatheringComplete) {
            return;
        }
        if (session.peerConnection.iceGatheringState() != PeerConnection.IceGatheringState.COMPLETE) {
            return;
        }
        SessionDescription localDescription = session.peerConnection.getLocalDescription();
        if (localDescription == null) {
            return;
        }
        session.pendingOfferUntilIceGatheringComplete = false;
        sendSessionDescription(session, "offer", localDescription);
    }

    private void sendSignal(String type, JSONObject payload, String targetPeerId) {
        WebSocket socket = webSocket;
        if (socket == null) {
            log("signal send failed, socket missing: " + type);
            return;
        }
        try {
            JSONObject message = new JSONObject()
                .put("type", type)
                .put("payload", payload);
            if (targetPeerId != null && !targetPeerId.isEmpty()) {
                payload.put("targetPeerId", targetPeerId);
            }
            socket.send(message.toString());
        } catch (JSONException exception) {
            log("signal serialization failed: " + exception.getMessage());
        }
    }

    private JSONObject createEndOfCandidatesPayload() {
        try {
            return new JSONObject().put("candidate", "");
        } catch (JSONException exception) {
            throw new IllegalStateException("failed to create end-of-candidates payload", exception);
        }
    }

    private JSONObject createCandidatePayload(IceCandidate candidate, String candidateLine) {
        try {
            return new JSONObject()
                .put("candidate", candidateLine)
                .put("sdpMid", candidate.sdpMid)
                .put("sdpMLineIndex", candidate.sdpMLineIndex);
        } catch (JSONException exception) {
            throw new IllegalStateException("failed to create candidate payload", exception);
        }
    }

    private JSONObject createSessionDescriptionPayload(SessionDescription description) {
        try {
            return new JSONObject()
                .put("type", description.type.canonicalForm())
                .put("sdp", stripNonRelaySdpCandidates(description.description));
        } catch (JSONException exception) {
            throw new IllegalStateException("failed to create SDP payload", exception);
        }
    }

    private String stripNonRelaySdpCandidates(String sdp) {
        String[] lines = sdp.replace("\r\n", "\n").split("\n");
        StringBuilder builder = new StringBuilder();
        int removed = 0;
        for (String line : lines) {
            if (line.startsWith("a=candidate:") && !isRelayCandidate(line)) {
                removed++;
                continue;
            }
            builder.append(line).append("\r\n");
        }
        if (removed > 0) {
            log("removed " + removed + " non-relay SDP candidates");
        }
        return builder.toString();
    }

    private String normalizeLocalCandidate(String candidateLine) {
        if (candidateLine == null || candidateLine.startsWith("candidate:")) {
            return candidateLine;
        }
        return "candidate:" + candidateLine;
    }

    private boolean isRelayCandidate(String candidateLine) {
        return candidateLine != null
            && (candidateLine.contains(" typ relay ") || candidateLine.endsWith(" typ relay"));
    }

    private int countCandidates(String sdp) {
        if (sdp == null || sdp.isEmpty()) {
            return 0;
        }
        int count = 0;
        for (String line : sdp.replace("\r\n", "\n").split("\n")) {
            if (line.startsWith("a=candidate:")) {
                count++;
            }
        }
        return count;
    }

    private int countRelayCandidates(String sdp) {
        if (sdp == null || sdp.isEmpty()) {
            return 0;
        }
        int count = 0;
        for (String line : sdp.replace("\r\n", "\n").split("\n")) {
            if (line.startsWith("a=candidate:") && isRelayCandidate(line)) {
                count++;
            }
        }
        return count;
    }

    private PeerSession peerSessionForSignal(String fromPeerId) {
        if (fromPeerId != null && !fromPeerId.isEmpty()) {
            return peerSessions.get(fromPeerId);
        }
        if (peerSessions.size() != 1) {
            return null;
        }
        return peerSessions.values().iterator().next();
    }

    private boolean isCurrentSession(PeerSession session) {
        return session != null && peerSessions.get(session.peerId) == session;
    }

    private String peerLabel(PeerSession session, String fallbackPeerId) {
        if (session == null) {
            return fallbackPeerId == null || fallbackPeerId.isEmpty() ? "peer" : fallbackPeerId;
        }
        return session.role + " " + session.peerId;
    }

    private void closePeerConnection() {
        for (String peerId : new ArrayList<>(peerSessions.keySet())) {
            closePeerSession(peerId);
        }
        stopSensorStreaming();
        stopTelemetryStreaming();
    }

    private void closePeerSession(String peerId) {
        if (peerId == null || peerId.isEmpty()) {
            return;
        }
        PeerSession session = peerSessions.remove(peerId);
        if (session == null) {
            return;
        }
        session.pendingOfferUntilIceGatheringComplete = false;
        if (session.spatialDataChannel != null) {
            session.spatialDataChannel.close();
            session.spatialDataChannel.dispose();
            session.spatialDataChannel = null;
        }
        if (session.telemetryDataChannel != null) {
            session.telemetryDataChannel.close();
            session.telemetryDataChannel.dispose();
            session.telemetryDataChannel = null;
        }
        if (session.eventDataChannel != null) {
            session.eventDataChannel.close();
            session.eventDataChannel.dispose();
            session.eventDataChannel = null;
        }
        if (session.controlDataChannel != null) {
            session.controlDataChannel.close();
            session.controlDataChannel.dispose();
            session.controlDataChannel = null;
        }
        session.peerConnection.close();
        session.peerConnection.dispose();
        if (!hasOpenDataChannel(RobotStreamRoles.CHANNEL_SPATIAL)) {
            stopSensorStreaming();
        }
        if (!hasOpenDataChannel(RobotStreamRoles.CHANNEL_TELEMETRY)) {
            stopTelemetryStreaming();
        }
    }

    private void stopInternal() {
        closePeerConnection();

        if (locationProvider != null) {
            locationProvider.stop();
            locationProvider = null;
        }

        if (webSocket != null) {
            webSocket.close(1000, "client stopped");
            webSocket = null;
        }
        if (httpClient != null) {
            httpClient.dispatcher().executorService().shutdown();
            httpClient.connectionPool().evictAll();
            httpClient = null;
        }

        stopVideoCapturer(cameraCapturer);
        stopVideoCapturer(thermalCapturer);
        cameraCapturer = null;
        thermalCapturer = null;

        if (rgbVideoTrack != null) {
            rgbVideoTrack.dispose();
            rgbVideoTrack = null;
        }
        if (thermalVideoTrack != null) {
            thermalVideoTrack.dispose();
            thermalVideoTrack = null;
        }
        if (audioTrack != null) {
            audioTrack.dispose();
            audioTrack = null;
        }
        if (rgbVideoSource != null) {
            rgbVideoSource.dispose();
            rgbVideoSource = null;
        }
        if (thermalVideoSource != null) {
            thermalVideoSource.dispose();
            thermalVideoSource = null;
        }
        if (audioSource != null) {
            audioSource.dispose();
            audioSource = null;
        }
        if (cameraSurfaceTextureHelper != null) {
            cameraSurfaceTextureHelper.dispose();
            cameraSurfaceTextureHelper = null;
        }
        if (thermalSurfaceTextureHelper != null) {
            thermalSurfaceTextureHelper.dispose();
            thermalSurfaceTextureHelper = null;
        }
        if (peerConnectionFactory != null) {
            peerConnectionFactory.dispose();
            peerConnectionFactory = null;
        }
        if (eglBase != null) {
            eglBase.release();
            eglBase = null;
        }
        log("robot client stopped");
    }

    private void stopVideoCapturer(VideoCapturer capturer) {
        if (capturer == null) {
            return;
        }
        try {
            capturer.stopCapture();
        } catch (InterruptedException exception) {
            Thread.currentThread().interrupt();
            log("video capturer stop interrupted");
        } catch (RuntimeException exception) {
            log("video capturer stop failed: " + exception.getMessage());
        }
        capturer.dispose();
    }

    private void log(String message) {
        logSink.log(message);
    }
}
