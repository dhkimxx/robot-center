package com.sst.robotcenter.androidrobot;

import android.location.Location;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.time.Instant;

public final class SensorPayloadFactory {
    private SensorPayloadFactory() {
    }

    public static String createSensorPayload(int sequence, long startedAtMs, String robotCode) {
        long nowMs = System.currentTimeMillis();
        double elapsedSeconds = (nowMs - startedAtMs) / 1000.0;
        try {
            JSONObject pose = new JSONObject()
                .put("x", round(1.2 + elapsedSeconds * 0.08, 2))
                .put("y", round(0.7 * Math.sin(elapsedSeconds / 5.0), 2))
                .put("yawDeg", round((elapsedSeconds * 7.5) % 360.0, 1));

            return new JSONObject()
                .put("messageId", robotCode + "-sensor-" + sequence)
                .put("schemaVersion", "1.0")
                .put("messageType", "spatial")
                .put("sequence", sequence)
                .put("sentAt", Instant.ofEpochMilli(nowMs).toString())
                .put("descriptors", new JSONArray()
                    .put(new JSONObject()
                        .put("sensorId", "spatial.odometry_1")
                        .put("displayName", "Odometry")
                        .put("sensorType", "odometry")
                        .put("valueType", "object")
                        .put("unit", "m")
                        .put("sampleRateHz", 1)
                        .put("enabled", true)
                        .put("metadata", new JSONObject().put("frameId", "odom"))))
                .put("samples", new JSONArray()
                    .put(new JSONObject()
                        .put("sensorId", "spatial.odometry_1")
                        .put("timestamp", Instant.ofEpochMilli(nowMs).toString())
                        .put("sequence", sequence)
                        .put("quality", "mock")
                        .put("values", pose)))
                .toString();
        } catch (JSONException exception) {
            throw new IllegalStateException("failed to create sensor payload", exception);
        }
    }

    public static String createTelemetryPayload(
        int sequence,
        long startedAtMs,
        String robotCode,
        Location location
    ) {
        long nowMs = System.currentTimeMillis();
        double elapsedSeconds = (nowMs - startedAtMs) / 1000.0;
        try {
            JSONObject position = new JSONObject().put("positionAvailable", location != null);
            if (location != null) {
                copyJsonObject(createGpsPosition(location), position);
            }

            return new JSONObject()
                .put("messageId", robotCode + "-telemetry-" + sequence)
                .put("schemaVersion", "1.0")
                .put("messageType", "telemetry")
                .put("sequence", sequence)
                .put("sentAt", Instant.ofEpochMilli(nowMs).toString())
                .put("descriptors", new JSONArray()
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.position_1")
                        .put("displayName", "GPS")
                        .put("sensorType", "position")
                        .put("valueType", "object")
                        .put("sampleRateHz", 1)
                        .put("enabled", true))
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.battery_1")
                        .put("displayName", "Battery")
                        .put("sensorType", "battery")
                        .put("valueType", "object")
                        .put("unit", "percent")
                        .put("sampleRateHz", 1)
                        .put("enabled", true))
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.environment_1")
                        .put("displayName", "Environment")
                        .put("sensorType", "environment")
                        .put("valueType", "object")
                        .put("sampleRateHz", 1)
                        .put("enabled", true)))
                .put("samples", new JSONArray()
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.position_1")
                        .put("timestamp", Instant.ofEpochMilli(nowMs).toString())
                        .put("sequence", sequence)
                        .put("quality", location == null ? "unknown" : "good")
                        .put("values", position))
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.battery_1")
                        .put("timestamp", Instant.ofEpochMilli(nowMs).toString())
                        .put("sequence", sequence)
                        .put("quality", "mock")
                        .put("values", new JSONObject()
                            .put("batteryPercent", round(82.0 - ((elapsedSeconds * 0.01) % 8.0), 1))))
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.environment_1")
                        .put("timestamp", Instant.ofEpochMilli(nowMs).toString())
                        .put("sequence", sequence)
                        .put("quality", "mock")
                        .put("values", new JSONObject()
                            .put("networkState", "android")
                            .put("coPpm", round(18.0 + 8.0 * Math.sin(elapsedSeconds / 8.0) + 3.0 * Math.sin(elapsedSeconds / 2.1), 1))
                            .put("oxygenPercent", round(20.8 + 0.2 * Math.cos(elapsedSeconds / 11.0), 2))
                            .put("temperatureCelsius", round(29.0 + 2.8 * Math.sin(elapsedSeconds / 13.0), 1))
                            .put("humidityPercent", round(61.0 + 5.0 * Math.cos(elapsedSeconds / 17.0), 1)))))
                .toString();
        } catch (JSONException exception) {
            throw new IllegalStateException("failed to create telemetry payload", exception);
        }
    }

    private static JSONObject createGpsPosition(Location location) throws JSONException {
        JSONObject position = new JSONObject()
            .put("coordinateType", "gps")
            .put("provider", location.getProvider())
            .put("latitude", round(location.getLatitude(), 6))
            .put("longitude", round(location.getLongitude(), 6));

        if (location.hasAltitude()) {
            position.put("altitudeMeter", round(location.getAltitude(), 1));
        }
        if (location.hasAccuracy()) {
            position.put("accuracyMeter", round(location.getAccuracy(), 1));
        }
        if (location.hasBearing()) {
            position.put("headingDegree", round(location.getBearing(), 1));
        }
        if (location.hasSpeed()) {
            position.put("speedMeterPerSecond", round(location.getSpeed(), 2));
        }
        if (location.getTime() > 0) {
            position.put("fixTime", Instant.ofEpochMilli(location.getTime()).toString());
        }
        return position;
    }

    private static void copyJsonObject(JSONObject source, JSONObject destination) throws JSONException {
        JSONArray names = source.names();
        if (names == null) {
            return;
        }
        for (int index = 0; index < names.length(); index++) {
            String name = names.getString(index);
            destination.put(name, source.get(name));
        }
    }

    private static double round(double value, int digits) {
        double scale = Math.pow(10, digits);
        return Math.round(value * scale) / scale;
    }
}
