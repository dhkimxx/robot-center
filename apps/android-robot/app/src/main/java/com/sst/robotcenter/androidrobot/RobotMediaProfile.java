package com.sst.robotcenter.androidrobot;

public final class RobotMediaProfile {
    public static final int RGB_WIDTH = 1920;
    public static final int RGB_HEIGHT = 1080;
    public static final int RGB_FPS = 30;
    public static final int RGB_MAX_BITRATE_BPS = 5_000_000;
    public static final int RGB_MIN_BITRATE_BPS = 2_500_000;
    public static final int RGB_BITRATE_KBPS = RGB_MAX_BITRATE_BPS / 1000;

    public static final int THERMAL_WIDTH = 640;
    public static final int THERMAL_HEIGHT = 480;
    public static final int THERMAL_FPS = 30;
    public static final int THERMAL_MAX_BITRATE_BPS = 800_000;
    public static final int THERMAL_MIN_BITRATE_BPS = 400_000;
    public static final int THERMAL_BITRATE_KBPS = THERMAL_MAX_BITRATE_BPS / 1000;

    private RobotMediaProfile() {
    }
}
