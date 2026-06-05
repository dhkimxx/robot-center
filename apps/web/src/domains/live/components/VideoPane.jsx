import { useEffect, useRef, useState } from "react";
import { cn } from "../../../utils/cn.js";
import { stopMediaStreamTracks } from "../liveMediaCleanup.js";

const emptyVideoMetrics = {
  width: 0,
  height: 0,
  fps: 0
};

export function VideoPane({ className = "", compact = false, label, stream, thermal = false }) {
  const videoRef = useRef(null);
  const [isVideoReady, setIsVideoReady] = useState(false);
  const [videoMetrics, setVideoMetrics] = useState(emptyVideoMetrics);

  useEffect(() => {
    const video = videoRef.current;
    setIsVideoReady(false);
    setVideoMetrics(emptyVideoMetrics);
    if (!video) {
      return undefined;
    }
    const markVideoReady = () => setIsVideoReady(true);
    const markVideoWaiting = () => setIsVideoReady(false);
    video.addEventListener("loadeddata", markVideoReady);
    video.addEventListener("canplay", markVideoReady);
    video.addEventListener("playing", markVideoReady);
    video.addEventListener("waiting", markVideoWaiting);
    video.addEventListener("stalled", markVideoWaiting);
    video.srcObject = stream;
    if (stream) {
      void video.play().catch(() => {});
    }
    return () => {
      video.removeEventListener("loadeddata", markVideoReady);
      video.removeEventListener("canplay", markVideoReady);
      video.removeEventListener("playing", markVideoReady);
      video.removeEventListener("waiting", markVideoWaiting);
      video.removeEventListener("stalled", markVideoWaiting);
      if (video.srcObject === stream) {
        video.srcObject = null;
      }
      stopMediaStreamTracks(stream);
    };
  }, [stream]);

  useEffect(() => {
    const video = videoRef.current;
    setVideoMetrics(emptyVideoMetrics);
    if (!video || !stream) {
      return undefined;
    }

    let cancelled = false;
    let callbackId = 0;
    let intervalId = 0;
    let frameCount = 0;
    let sampleStartedAt = 0;
    let previousTotalFrames = 0;
    let previousMeasuredAt = 0;

    const publishMetrics = (fps = 0) => {
      if (cancelled) {
        return;
      }
      const width = video.videoWidth || 0;
      const height = video.videoHeight || 0;
      setVideoMetrics((current) => {
        const roundedFPS = fps > 0 ? Math.round(fps) : current.fps;
        if (current.width === width && current.height === height && current.fps === roundedFPS) {
          return current;
        }
        return { width, height, fps: roundedFPS };
      });
    };

    const onVideoFrame = (now) => {
      if (cancelled) {
        return;
      }
      frameCount += 1;
      if (!sampleStartedAt) {
        sampleStartedAt = now;
        publishMetrics();
      }
      const elapsedMs = now - sampleStartedAt;
      if (elapsedMs >= 1000) {
        publishMetrics((frameCount * 1000) / elapsedMs);
        frameCount = 0;
        sampleStartedAt = now;
      }
      callbackId = video.requestVideoFrameCallback(onVideoFrame);
    };

    const updateFromPlaybackQuality = () => {
      const quality = typeof video.getVideoPlaybackQuality === "function"
        ? video.getVideoPlaybackQuality()
        : null;
      const now = Date.now();
      if (!quality || !previousMeasuredAt) {
        previousMeasuredAt = now;
        previousTotalFrames = quality?.totalVideoFrames ?? 0;
        publishMetrics();
        return;
      }
      const elapsedMs = now - previousMeasuredAt;
      const frameDelta = (quality.totalVideoFrames ?? 0) - previousTotalFrames;
      previousMeasuredAt = now;
      previousTotalFrames = quality.totalVideoFrames ?? previousTotalFrames;
      publishMetrics(elapsedMs > 0 ? (frameDelta * 1000) / elapsedMs : 0);
    };

    const publishCurrentMetrics = () => publishMetrics();
    video.addEventListener("loadedmetadata", publishCurrentMetrics);
    video.addEventListener("resize", publishCurrentMetrics);
    if (typeof video.requestVideoFrameCallback === "function") {
      callbackId = video.requestVideoFrameCallback(onVideoFrame);
    } else {
      updateFromPlaybackQuality();
      intervalId = window.setInterval(updateFromPlaybackQuality, 1000);
    }

    return () => {
      cancelled = true;
      if (callbackId && typeof video.cancelVideoFrameCallback === "function") {
        video.cancelVideoFrameCallback(callbackId);
      }
      if (intervalId) {
        window.clearInterval(intervalId);
      }
      video.removeEventListener("loadedmetadata", publishCurrentMetrics);
      video.removeEventListener("resize", publishCurrentMetrics);
    };
  }, [stream]);

  const shouldShowLoading = !stream || !isVideoReady;
  const resolutionLabel = videoMetrics.width && videoMetrics.height ? `${videoMetrics.width}x${videoMetrics.height}` : "-";
  const fpsLabel = videoMetrics.fps > 0 ? `${videoMetrics.fps} fps` : "- fps";

  return (
    <div
      className={cn(
        "relative min-h-[260px] overflow-hidden rounded-xl border border-slate-500/20 bg-command-950",
        thermal && "bg-[#111015]",
        compact && "min-h-[180px]",
        className
      )}
    >
      <video className="absolute inset-0 h-full w-full object-contain" ref={videoRef} autoPlay playsInline muted={label !== "Audio"} />
      {shouldShowLoading ? (
        <div className="absolute inset-0 grid place-items-center bg-command-950/60">
          <div className="grid justify-items-center gap-3">
            <span
              aria-hidden
              className="h-8 w-8 rounded-full border-2 border-slate-700 border-t-sapphire-400 motion-safe:animate-spin"
            />
            <span className="text-sm font-bold text-slate-500">{label} 영상 로드 중</span>
          </div>
        </div>
      ) : null}
      <strong className="absolute left-3 top-3 rounded-lg border border-sapphire-500/25 bg-command-950/80 px-3 py-1.5 text-sm font-bold text-slate-100">
        {label}
      </strong>
      <div className="absolute right-3 top-3 flex items-center gap-2 rounded-lg border border-slate-500/20 bg-command-950/80 px-2.5 py-1.5 text-[11px] font-bold text-slate-200 shadow-lg shadow-black/20">
        <span className="tabular-nums">{resolutionLabel}</span>
        <span className="h-3 w-px bg-slate-500/40" />
        <span className="tabular-nums text-slate-300">{fpsLabel}</span>
      </div>
    </div>
  );
}
