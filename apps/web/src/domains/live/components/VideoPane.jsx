import { useEffect, useRef, useState } from "react";
import { cn } from "../../../utils/cn.js";
import { stopMediaStreamTracks } from "../liveMediaCleanup.js";

export function VideoPane({ className = "", compact = false, label, stream, thermal = false }) {
  const videoRef = useRef(null);
  const [isVideoReady, setIsVideoReady] = useState(false);

  useEffect(() => {
    const video = videoRef.current;
    setIsVideoReady(false);
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

  const shouldShowLoading = !stream || !isVideoReady;

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
    </div>
  );
}
