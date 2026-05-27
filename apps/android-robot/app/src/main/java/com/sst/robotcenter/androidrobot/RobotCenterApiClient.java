package com.sst.robotcenter.androidrobot;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.io.IOException;
import java.time.Instant;

import okhttp3.MediaType;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.RequestBody;
import okhttp3.Response;

public final class RobotCenterApiClient {
    private static final MediaType JSON = MediaType.get("application/json; charset=utf-8");

    private final OkHttpClient httpClient = new OkHttpClient.Builder().build();
    private final String serverUrl;
    private final String robotCode;
    private final String robotToken;

    public RobotCenterApiClient(String serverUrl, String robotCode, String robotToken) {
        this.serverUrl = trimTrailingSlash(serverUrl);
        this.robotCode = robotCode;
        this.robotToken = robotToken;
    }

    public void sendHeartbeat(String state) throws IOException, JSONException {
        JSONObject body = new JSONObject()
            .put("state", state)
            .put("sentAt", Instant.now().toString());
        post("/api/robot-gateway/heartbeat", body);
    }

    public RobotMissionConfig fetchMission() throws IOException, JSONException {
        JSONObject response = get("/api/robot-gateway/mission");
        String missionStatus = response.optString("missionStatus", "none");
        if (!"active".equals(missionStatus)) {
            return new RobotMissionConfig(
                false,
                null,
                null,
                missionStatus,
                null,
                null,
                null,
                null,
                null
            );
        }

        JSONObject sfu = response.getJSONObject("sfu");
        JSONArray turnServers = response.getJSONArray("turnServers");
        JSONObject firstTurnServer = turnServers.getJSONObject(0);
        JSONArray urls = firstTurnServer.getJSONArray("urls");
        String missionCode = response.optString("missionCode", response.optString("roomId"));
        String roomId = response.optString("roomId", missionCode);
        if (roomId.isEmpty()) {
            roomId = missionCode;
        }
        return new RobotMissionConfig(
            true,
            response.optString("missionId"),
            missionCode,
            missionStatus,
            roomId,
            sfu.getString("signalingUrl"),
            urls.getString(0),
            firstTurnServer.getString("username"),
            firstTurnServer.getString("credential")
        );
    }

    public void close() {
        httpClient.dispatcher().executorService().shutdown();
        httpClient.connectionPool().evictAll();
    }

    private JSONObject get(String path) throws IOException, JSONException {
        Request request = new Request.Builder()
            .url(serverUrl + path)
            .header("Authorization", "Bearer " + robotToken)
            .get()
            .build();
        try (Response response = httpClient.newCall(request).execute()) {
            return parseResponse(response);
        }
    }

    private JSONObject post(String path, JSONObject body) throws IOException, JSONException {
        Request request = new Request.Builder()
            .url(serverUrl + path)
            .header("Authorization", "Bearer " + robotToken)
            .post(RequestBody.create(body.toString(), JSON))
            .build();
        try (Response response = httpClient.newCall(request).execute()) {
            return parseResponse(response);
        }
    }

    private JSONObject parseResponse(Response response) throws IOException, JSONException {
        String bodyText = response.body() == null ? "" : response.body().string();
        if (!response.isSuccessful()) {
            throw new IOException("HTTP " + response.code() + ": " + bodyText);
        }
        if (bodyText.isEmpty()) {
            return new JSONObject();
        }
        return new JSONObject(bodyText);
    }

    private static String trimTrailingSlash(String value) {
        String trimmed = value.trim();
        while (trimmed.endsWith("/")) {
            trimmed = trimmed.substring(0, trimmed.length() - 1);
        }
        return trimmed;
    }

}
