package com.sst.robotcenter.androidrobot;

import android.content.Context;

import org.webrtc.CapturerObserver;
import org.webrtc.JavaI420Buffer;
import org.webrtc.SurfaceTextureHelper;
import org.webrtc.VideoCapturer;
import org.webrtc.VideoFrame;

import java.nio.ByteBuffer;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

public final class SyntheticThermalCapturer implements VideoCapturer {
    private CapturerObserver capturerObserver;
    private ScheduledExecutorService frameExecutor;
    private int width = 640;
    private int height = 480;
    private int fps = 30;
    private long startedAtNs;

    @Override
    public void initialize(SurfaceTextureHelper surfaceTextureHelper, Context applicationContext, CapturerObserver capturerObserver) {
        this.capturerObserver = capturerObserver;
    }

    @Override
    public void startCapture(int width, int height, int fps) {
        this.width = Math.max(160, width);
        this.height = Math.max(120, height);
        this.fps = Math.max(1, fps);
        this.startedAtNs = System.nanoTime();
        this.frameExecutor = Executors.newSingleThreadScheduledExecutor();
        this.capturerObserver.onCapturerStarted(true);

        long frameIntervalMs = Math.max(16, 1000L / this.fps);
        this.frameExecutor.scheduleAtFixedRate(this::emitFrame, 0, frameIntervalMs, TimeUnit.MILLISECONDS);
    }

    @Override
    public void stopCapture() throws InterruptedException {
        if (frameExecutor != null) {
            frameExecutor.shutdownNow();
            frameExecutor.awaitTermination(500, TimeUnit.MILLISECONDS);
            frameExecutor = null;
        }
        if (capturerObserver != null) {
            capturerObserver.onCapturerStopped();
        }
    }

    @Override
    public void changeCaptureFormat(int width, int height, int fps) {
        this.width = Math.max(160, width);
        this.height = Math.max(120, height);
        this.fps = Math.max(1, fps);
    }

    @Override
    public void dispose() {
        try {
            stopCapture();
        } catch (InterruptedException exception) {
            Thread.currentThread().interrupt();
        }
    }

    @Override
    public boolean isScreencast() {
        return false;
    }

    private void emitFrame() {
        if (capturerObserver == null) {
            return;
        }

        long nowNs = System.nanoTime();
        double elapsedSeconds = (nowNs - startedAtNs) / 1_000_000_000.0;
        JavaI420Buffer buffer = JavaI420Buffer.allocate(width, height);
        fillLumaPlane(buffer, elapsedSeconds);
        fillChromaPlane(buffer.getDataU(), buffer.getStrideU(), (width + 1) / 2, (height + 1) / 2, 96);
        fillChromaPlane(buffer.getDataV(), buffer.getStrideV(), (width + 1) / 2, (height + 1) / 2, 168);

        VideoFrame frame = new VideoFrame(buffer, 0, nowNs);
        capturerObserver.onFrameCaptured(frame);
        frame.release();
    }

    private void fillLumaPlane(JavaI420Buffer buffer, double elapsedSeconds) {
        ByteBuffer dataY = buffer.getDataY();
        int strideY = buffer.getStrideY();
        double centerX = Math.sin(elapsedSeconds * 0.8) * 0.42;
        double centerY = Math.cos(elapsedSeconds * 0.65) * 0.28;

        for (int y = 0; y < height; y++) {
            double normalY = ((double) y / Math.max(1, height - 1)) * 2.0 - 1.0;
            for (int x = 0; x < width; x++) {
                double normalX = ((double) x / Math.max(1, width - 1)) * 2.0 - 1.0;
                double hotspot = Math.exp(-((Math.pow(normalX - centerX, 2) / 0.05) + (Math.pow(normalY - centerY, 2) / 0.08)));
                double background = 0.25
                    + 0.18 * Math.sin(normalX * 5.0 + elapsedSeconds * 1.3)
                    + 0.12 * Math.cos(normalY * 6.0 - elapsedSeconds);
                int heat = clampToByte((background + hotspot * 0.85) * 255.0);
                dataY.put(y * strideY + x, (byte) heat);
            }
        }

        int scanline = (int) ((elapsedSeconds * 90) % height);
        for (int x = 0; x < width; x++) {
            dataY.put(scanline * strideY + x, (byte) 255);
            if (scanline + 1 < height) {
                dataY.put((scanline + 1) * strideY + x, (byte) 255);
            }
        }
    }

    private void fillChromaPlane(ByteBuffer plane, int stride, int planeWidth, int planeHeight, int value) {
        byte chroma = (byte) clampToByte(value);
        for (int y = 0; y < planeHeight; y++) {
            for (int x = 0; x < planeWidth; x++) {
                plane.put(y * stride + x, chroma);
            }
        }
    }

    private int clampToByte(double value) {
        return Math.max(0, Math.min(255, (int) Math.round(value)));
    }
}
