package com.sst.robotcenter.androidrobot;

public final class RobotWebRtcConfig {
    public final String signalingUrl;
    public final String turnUrl;
    public final String turnUsername;
    public final String turnPassword;
    public final String robotToken;
    public final String robotCode;
    public final String missionId;
    public final String missionCode;
    public final String roomId;

    public RobotWebRtcConfig(
        String signalingUrl,
        String turnUrl,
        String turnUsername,
        String turnPassword,
        String robotToken,
        String robotCode,
        String missionId,
        String missionCode,
        String roomId
    ) {
        this.signalingUrl = signalingUrl;
        this.turnUrl = turnUrl;
        this.turnUsername = turnUsername;
        this.turnPassword = turnPassword;
        this.robotToken = robotToken;
        this.robotCode = robotCode;
        this.missionId = missionId;
        this.missionCode = missionCode;
        this.roomId = roomId;
    }
}
