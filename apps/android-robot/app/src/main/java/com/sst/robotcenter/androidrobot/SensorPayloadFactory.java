package com.sst.robotcenter.androidrobot;

import android.location.Location;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.time.Instant;

public final class SensorPayloadFactory {
    private SensorPayloadFactory() {
    }

    public static String createTelemetryPayload(
        int sequence,
        long startedAtMs,
        Location location
    ) {
        long nowMs = System.currentTimeMillis();
        double elapsedSeconds = (nowMs - startedAtMs) / 1000.0;
        try {
            JSONObject position = new JSONObject().put("positionAvailable", location != null);
            if (location != null) {
                copyJsonObject(createGpsPosition(location), position);
            }
            double coPpm = round(18.0 + 8.0 * Math.sin(elapsedSeconds / 8.0) + 3.0 * Math.sin(elapsedSeconds / 2.1), 1);
            double h2sPpm = round(2.1 + 0.2 * Math.sin(elapsedSeconds / 7.0), 1);
            double oxygenPercent = round(20.8 + 0.2 * Math.cos(elapsedSeconds / 11.0), 2);
            double ch4Lel = round(6.8 + 1.5 * Math.cos(elapsedSeconds / 9.0), 1);
            double tempChannelValue = round(29.0 + 2.8 * Math.sin(elapsedSeconds / 13.0), 1);
            double humChannelValue = round(61.0 + 5.0 * Math.cos(elapsedSeconds / 17.0), 1);

            return new JSONObject()
                .put("messageId", "telemetry-" + startedAtMs + "-" + sequence)
                .put("schemaVersion", "1.0")
                .put("messageType", "telemetry")
                .put("descriptors", new JSONArray()
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.position_1")
                        .put("label", "GPS")
                        .put("sensorType", "position")
                        .put("enabled", true))
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.battery_1")
                        .put("label", "Battery")
                        .put("sensorType", "battery")
                        .put("unit", "percent")
                        .put("enabled", true))
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.gas.channel_1")
                        .put("label", "CO")
                        .put("sensorType", "gas")
                        .put("unit", "ppm")
                        .put("enabled", true))
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.gas.channel_2")
                        .put("label", "H2S")
                        .put("sensorType", "gas")
                        .put("unit", "ppm")
                        .put("enabled", true))
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.gas.channel_3")
                        .put("label", "O2")
                        .put("sensorType", "gas")
                        .put("unit", "%Vol")
                        .put("enabled", true))
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.gas.channel_4")
                        .put("label", "CH4")
                        .put("sensorType", "gas")
                        .put("unit", "%LEL")
                        .put("enabled", true))
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.gas.channel_5")
                        .put("label", "TEMP")
                        .put("sensorType", "gas")
                        .put("unit", "degC")
                        .put("enabled", true))
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.gas.channel_6")
                        .put("label", "HUM")
                        .put("sensorType", "gas")
                        .put("unit", "RH")
                        .put("enabled", true)))
                .put("samples", new JSONArray()
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.position_1")
                        .put("timestamp", Instant.ofEpochMilli(nowMs).toString())
                        .put("values", position))
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.battery_1")
                        .put("timestamp", Instant.ofEpochMilli(nowMs).toString())
                        .put("values", new JSONObject()
                            .put("batteryPercent", round(82.0 - ((elapsedSeconds * 0.01) % 8.0), 1))))
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.gas.channel_5")
                        .put("timestamp", Instant.ofEpochMilli(nowMs).toString())
                        .put("values", createGasChannelValue(tempChannelValue, 5.0, 50.0)))
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.gas.channel_6")
                        .put("timestamp", Instant.ofEpochMilli(nowMs).toString())
                        .put("values", createGasChannelValue(humChannelValue, 20.0, 80.0)))
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.gas.channel_3")
                        .put("timestamp", Instant.ofEpochMilli(nowMs).toString())
                        .put("values", createGasChannelValue(oxygenPercent, 19.5, 23.5)))
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.gas.channel_1")
                        .put("timestamp", Instant.ofEpochMilli(nowMs).toString())
                        .put("values", createGasChannelValue(coPpm, 10.0, 15.0)))
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.gas.channel_2")
                        .put("timestamp", Instant.ofEpochMilli(nowMs).toString())
                        .put("values", createGasChannelValue(h2sPpm, 5.0, 10.0)))
                    .put(new JSONObject()
                        .put("sensorId", "telemetry.gas.channel_4")
                        .put("timestamp", Instant.ofEpochMilli(nowMs).toString())
                        .put("values", createGasChannelValue(ch4Lel, 20.0, 40.0))))
                .toString();
        } catch (JSONException exception) {
            throw new IllegalStateException("failed to create telemetry payload", exception);
        }
    }

    private static JSONObject createGasChannelValue(
        double concentration,
        double lowAlarm,
        double highAlarm
    ) throws JSONException {
        return new JSONObject()
            .put("concentration", concentration)
            .put("scale_code", 1)
            .put("alarm_code", 0)
            .put("alarm", "normal")
            .put("low_alarm", lowAlarm)
            .put("high_alarm", highAlarm)
            .put("valid", true);
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
