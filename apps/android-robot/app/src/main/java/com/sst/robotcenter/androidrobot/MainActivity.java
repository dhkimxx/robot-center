package com.sst.robotcenter.androidrobot;

import android.Manifest;
import android.app.Activity;
import android.content.pm.PackageManager;
import android.graphics.Color;
import android.os.Bundle;
import android.text.InputType;
import android.text.method.ScrollingMovementMethod;
import android.util.Log;
import android.view.Gravity;
import android.view.ViewGroup;
import android.view.WindowManager;
import android.widget.Button;
import android.widget.EditText;
import android.widget.LinearLayout;
import android.widget.ScrollView;
import android.widget.TextView;
import android.widget.Toast;

import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.Locale;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

public class MainActivity extends Activity {
    private static final int RUNTIME_PERMISSION_REQUEST_CODE = 1001;
    private static final String LOG_TAG = "RobotCenterMock";

    private EditText serverUrlInput;
    private EditText robotCodeInput;
    private EditText robotTokenInput;
    private TextView statusText;
    private TextView logText;
    private RobotCenterApiClient apiClient;
    private RobotWebRtcClient robotClient;
    private ScheduledExecutorService centerExecutor;
    private volatile String activeMissionCode;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        getWindow().addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON);
        ScrollView contentView = createContentView();
        setContentView(contentView);
        if (getIntent().getBooleanExtra("autoConnect", false)) {
            contentView.post(this::connectRobotClient);
        }
    }

    @Override
    protected void onDestroy() {
        stopRobotClient();
        super.onDestroy();
    }

    private ScrollView createContentView() {
        ScrollView scrollView = new ScrollView(this);
        scrollView.setFillViewport(true);
        scrollView.setBackgroundColor(Color.rgb(229, 233, 240));

        LinearLayout root = new LinearLayout(this);
        root.setOrientation(LinearLayout.VERTICAL);
        root.setPadding(dp(16), dp(14), dp(16), dp(16));
        scrollView.addView(root, new ScrollView.LayoutParams(
            ViewGroup.LayoutParams.MATCH_PARENT,
            ViewGroup.LayoutParams.WRAP_CONTENT
        ));

        TextView title = new TextView(this);
        title.setText("Android Robot Mock");
        title.setTextColor(Color.rgb(15, 23, 42));
        title.setTextSize(22);
        title.setGravity(Gravity.START);
        root.addView(title);

        TextView subtitle = new TextView(this);
        subtitle.setText("Robot registration client + RGB/Thermal/Audio WebRTC publisher");
        subtitle.setTextColor(Color.rgb(71, 85, 105));
        subtitle.setTextSize(13);
        subtitle.setPadding(0, dp(4), 0, dp(14));
        root.addView(subtitle);

        serverUrlInput = addInput(root, "Server URL", getIntentValue("serverUrl", "http://192.168.20.26:18080"), false);
        robotCodeInput = addInput(root, "Robot Code", getIntentValue("robotCode", "robot-001"), false);
        robotTokenInput = addInput(root, "Robot Token", getIntentValue("robotToken", ""), true);

        LinearLayout buttonRow = new LinearLayout(this);
        buttonRow.setOrientation(LinearLayout.HORIZONTAL);
        buttonRow.setPadding(0, dp(10), 0, dp(10));
        root.addView(buttonRow, new LinearLayout.LayoutParams(
            ViewGroup.LayoutParams.MATCH_PARENT,
            ViewGroup.LayoutParams.WRAP_CONTENT
        ));

        Button connectButton = new Button(this);
        connectButton.setText("Connect Center");
        connectButton.setAllCaps(false);
        connectButton.setOnClickListener(view -> connectRobotClient());
        buttonRow.addView(connectButton, new LinearLayout.LayoutParams(0, dp(44), 1));

        Button disconnectButton = new Button(this);
        disconnectButton.setText("Disconnect");
        disconnectButton.setAllCaps(false);
        disconnectButton.setOnClickListener(view -> stopRobotClient());
        LinearLayout.LayoutParams disconnectLayout = new LinearLayout.LayoutParams(0, dp(44), 1);
        disconnectLayout.setMargins(dp(8), 0, 0, 0);
        buttonRow.addView(disconnectButton, disconnectLayout);

        Button switchCameraButton = new Button(this);
        switchCameraButton.setText("Switch RGB Camera");
        switchCameraButton.setAllCaps(false);
        switchCameraButton.setOnClickListener(view -> switchRgbCamera());
        root.addView(switchCameraButton, new LinearLayout.LayoutParams(
            ViewGroup.LayoutParams.MATCH_PARENT,
            dp(44)
        ));

        statusText = new TextView(this);
        statusText.setText("Status: idle");
        statusText.setTextColor(Color.rgb(30, 64, 175));
        statusText.setTextSize(14);
        statusText.setPadding(0, dp(4), 0, dp(8));
        root.addView(statusText);

        logText = new TextView(this);
        logText.setTextColor(Color.rgb(30, 41, 59));
        logText.setTextSize(12);
        logText.setMinLines(16);
        logText.setGravity(Gravity.BOTTOM);
        logText.setMovementMethod(new ScrollingMovementMethod());
        logText.setPadding(dp(10), dp(10), dp(10), dp(10));
        logText.setBackgroundColor(Color.WHITE);
        root.addView(logText, new LinearLayout.LayoutParams(
            ViewGroup.LayoutParams.MATCH_PARENT,
            ViewGroup.LayoutParams.WRAP_CONTENT
        ));

        return scrollView;
    }

    private EditText addInput(LinearLayout parent, String label, String value, boolean secret) {
        TextView labelView = new TextView(this);
        labelView.setText(label);
        labelView.setTextColor(Color.rgb(100, 116, 139));
        labelView.setTextSize(12);
        parent.addView(labelView);

        EditText input = new EditText(this);
        input.setSingleLine(true);
        input.setText(value);
        input.setTextSize(14);
        input.setSelectAllOnFocus(false);
        input.setPadding(dp(10), 0, dp(10), 0);
        if (secret) {
            input.setInputType(InputType.TYPE_CLASS_TEXT | InputType.TYPE_TEXT_VARIATION_VISIBLE_PASSWORD);
        }
        parent.addView(input, new LinearLayout.LayoutParams(
            ViewGroup.LayoutParams.MATCH_PARENT,
            dp(42)
        ));
        return input;
    }

    private String getIntentValue(String key, String fallback) {
        String value = getIntent().getStringExtra(key);
        if (value == null || value.trim().isEmpty()) {
            return fallback;
        }
        return value;
    }

    private void connectRobotClient() {
        if (!hasRequiredPermissions()) {
            requestPermissions(
                new String[] {
                    Manifest.permission.CAMERA,
                    Manifest.permission.RECORD_AUDIO,
                    Manifest.permission.ACCESS_FINE_LOCATION,
                    Manifest.permission.ACCESS_COARSE_LOCATION
                },
                RUNTIME_PERMISSION_REQUEST_CODE
            );
            return;
        }

        String serverUrl = serverUrlInput.getText().toString().trim();
        String robotCode = robotCodeInput.getText().toString().trim();
        String robotToken = robotTokenInput.getText().toString().trim();
        if (serverUrl.isEmpty() || robotCode.isEmpty() || robotToken.isEmpty()) {
            Toast.makeText(this, "Server URL, Robot Code, and Robot Token are required.", Toast.LENGTH_SHORT).show();
            return;
        }

        stopRobotClient();
        apiClient = new RobotCenterApiClient(serverUrl, robotCode, robotToken);
        centerExecutor = Executors.newSingleThreadScheduledExecutor();
        centerExecutor.execute(() -> {
            sendHeartbeatOnce();
            pollMissionOnce();
        });
        centerExecutor.scheduleAtFixedRate(() -> {
            sendHeartbeatOnce();
            pollMissionOnce();
        }, 5, 5, TimeUnit.SECONDS);
        setStatus("center connected");
    }

    private void sendHeartbeatOnce() {
        try {
            apiClient.sendHeartbeat(activeMissionCode == null ? "online" : "streaming");
            appendLog("heartbeat accepted");
        } catch (Exception exception) {
            appendLog("heartbeat failed: " + exception.getMessage());
        }
    }

    private void pollMissionOnce() {
        try {
            RobotMissionConfig mission = apiClient.fetchMission();
            if (!mission.active) {
                activeMissionCode = null;
                appendLog("mission polling: " + mission.missionStatus);
                setStatus("online / waiting mission");
                return;
            }
            if (mission.missionCode != null && mission.missionCode.equals(activeMissionCode) && robotClient != null) {
                return;
            }
            appendLog("active mission received: " + mission.missionCode);
            startWebRtcForMission(mission);
        } catch (Exception exception) {
            appendLog("mission polling failed: " + exception.getMessage());
        }
    }

    private void startWebRtcForMission(RobotMissionConfig mission) {
        runOnUiThread(() -> {
            if (robotClient != null) {
                robotClient.stop();
                robotClient = null;
            }
            RobotWebRtcConfig config = new RobotWebRtcConfig(
                mission.signalingUrl,
                mission.turnUrl,
                mission.turnUsername,
                mission.turnPassword,
                robotCodeInput.getText().toString().trim(),
                mission.missionId,
                mission.missionCode,
                mission.roomId
            );
            robotClient = new RobotWebRtcClient(this, config, this::appendLog);
            robotClient.start();
            activeMissionCode = mission.missionCode;
            setStatus("streaming / " + mission.missionCode + " / " + mission.roomId);
        });
    }

    private void switchRgbCamera() {
        if (robotClient == null) {
            Toast.makeText(this, "Start streaming first.", Toast.LENGTH_SHORT).show();
            return;
        }
        robotClient.switchRgbCamera();
    }

    private void stopRobotClient() {
        if (centerExecutor != null) {
            centerExecutor.shutdownNow();
            centerExecutor = null;
        }
        if (apiClient != null) {
            apiClient.close();
            apiClient = null;
        }
        if (robotClient != null) {
            robotClient.stop();
            robotClient = null;
        }
        activeMissionCode = null;
        setStatus("idle");
    }

    private boolean hasRequiredPermissions() {
        boolean hasCamera = checkSelfPermission(Manifest.permission.CAMERA) == PackageManager.PERMISSION_GRANTED;
        boolean hasAudio = checkSelfPermission(Manifest.permission.RECORD_AUDIO) == PackageManager.PERMISSION_GRANTED;
        boolean hasLocation = checkSelfPermission(Manifest.permission.ACCESS_FINE_LOCATION) == PackageManager.PERMISSION_GRANTED
            || checkSelfPermission(Manifest.permission.ACCESS_COARSE_LOCATION) == PackageManager.PERMISSION_GRANTED;
        return hasCamera && hasAudio && hasLocation;
    }

    private void setStatus(String status) {
        runOnUiThread(() -> statusText.setText("Status: " + status));
    }

    private void appendLog(String message) {
        Log.d(LOG_TAG, message);
        runOnUiThread(() -> {
            String timestamp = new SimpleDateFormat("HH:mm:ss", Locale.KOREA).format(new Date());
            String nextLog = "[" + timestamp + "] " + message + "\n" + logText.getText();
            logText.setText(nextLog);
        });
    }

    @Override
    public void onRequestPermissionsResult(int requestCode, String[] permissions, int[] grantResults) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults);
        if (requestCode != RUNTIME_PERMISSION_REQUEST_CODE) {
            return;
        }
        if (hasRequiredPermissions()) {
            connectRobotClient();
            return;
        }
        Toast.makeText(this, "Camera, microphone, and location permissions are required.", Toast.LENGTH_SHORT).show();
    }

    private int dp(int value) {
        return Math.round(value * getResources().getDisplayMetrics().density);
    }
}
