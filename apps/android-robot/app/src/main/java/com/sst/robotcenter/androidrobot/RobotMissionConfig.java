package com.sst.robotcenter.androidrobot;

public final class RobotMissionConfig {
    public final boolean active;
    public final String missionId;
    public final String missionCode;
    public final String missionStatus;
    public final String roomId;
    public final String signalingUrl;
    public final String turnUrl;
    public final String turnUsername;
    public final String turnPassword;

    public RobotMissionConfig(
        boolean active,
        String missionId,
        String missionCode,
        String missionStatus,
        String roomId,
        String signalingUrl,
        String turnUrl,
        String turnUsername,
        String turnPassword
    ) {
        this.active = active;
        this.missionId = missionId;
        this.missionCode = missionCode;
        this.missionStatus = missionStatus;
        this.roomId = roomId;
        this.signalingUrl = signalingUrl;
        this.turnUrl = turnUrl;
        this.turnUsername = turnUsername;
        this.turnPassword = turnPassword;
    }
}
