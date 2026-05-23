package com.sst.robotcenter.androidrobot;

import android.Manifest;
import android.content.Context;
import android.content.pm.PackageManager;
import android.location.Location;
import android.location.LocationListener;
import android.location.LocationManager;
import android.os.Bundle;
import android.os.Looper;

public final class LocationProvider {
    private final Context applicationContext;
    private final LogSink logSink;
    private final LocationManager locationManager;
    private final LocationListener locationListener;

    private volatile Location latestLocation;
    private boolean listening;

    public LocationProvider(Context context, LogSink logSink) {
        this.applicationContext = context.getApplicationContext();
        this.logSink = logSink;
        this.locationManager = (LocationManager) applicationContext.getSystemService(Context.LOCATION_SERVICE);
        this.locationListener = new LocationListener() {
            @Override
            public void onLocationChanged(Location location) {
                updateLatestLocation(location);
            }

            @Override
            public void onProviderEnabled(String provider) {
                log("location provider enabled: " + provider);
            }

            @Override
            public void onProviderDisabled(String provider) {
                log("location provider disabled: " + provider);
            }

            @Override
            public void onStatusChanged(String provider, int status, Bundle extras) {
            }
        };
    }

    public void start() {
        if (listening || locationManager == null) {
            return;
        }
        if (!hasLocationPermission()) {
            log("location permission missing; telemetry GPS will be unavailable");
            return;
        }

        updateFromLastKnown(LocationManager.GPS_PROVIDER);
        updateFromLastKnown(LocationManager.NETWORK_PROVIDER);

        boolean requested = false;
        requested = requestProvider(LocationManager.GPS_PROVIDER) || requested;
        requested = requestProvider(LocationManager.NETWORK_PROVIDER) || requested;
        listening = requested;
        if (!requested) {
            log("no enabled location provider; telemetry GPS will be unavailable");
        }
    }

    public void stop() {
        if (locationManager == null || !listening) {
            return;
        }
        locationManager.removeUpdates(locationListener);
        listening = false;
    }

    public Location getLatestLocation() {
        Location location = latestLocation;
        return location == null ? null : new Location(location);
    }

    private boolean requestProvider(String provider) {
        if (!hasLocationPermission() || !locationManager.isProviderEnabled(provider)) {
            return false;
        }
        try {
            locationManager.requestLocationUpdates(provider, 1000L, 0.5f, locationListener, Looper.getMainLooper());
            log("location updates requested: " + provider);
            return true;
        } catch (RuntimeException exception) {
            log("location provider request failed: " + provider + " / " + exception.getMessage());
            return false;
        }
    }

    private void updateFromLastKnown(String provider) {
        if (!hasLocationPermission() || !locationManager.isProviderEnabled(provider)) {
            return;
        }
        try {
            updateLatestLocation(locationManager.getLastKnownLocation(provider));
        } catch (RuntimeException exception) {
            log("last known location failed: " + provider + " / " + exception.getMessage());
        }
    }

    private void updateLatestLocation(Location location) {
        if (location == null) {
            return;
        }
        Location current = latestLocation;
        if (current == null || location.getTime() >= current.getTime()) {
            latestLocation = new Location(location);
            log("location updated: " + location.getProvider());
        }
    }

    private boolean hasLocationPermission() {
        return applicationContext.checkSelfPermission(Manifest.permission.ACCESS_FINE_LOCATION) == PackageManager.PERMISSION_GRANTED
            || applicationContext.checkSelfPermission(Manifest.permission.ACCESS_COARSE_LOCATION) == PackageManager.PERMISSION_GRANTED;
    }

    private void log(String message) {
        logSink.log(message);
    }
}
