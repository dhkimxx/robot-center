import { useEffect, useRef } from "react";
import { cn } from "../../../utils/cn.js";

export function VideoPane({ className = "", compact = false, label, stream, thermal = false }) {
  const videoRef = useRef(null);
  useEffect(() => {
    const video = videoRef.current;
    if (!video) {
      return undefined;
    }
    video.srcObject = stream;
    if (stream) {
      void video.play().catch(() => {});
    }
    return () => {
      video.srcObject = null;
    };
  }, [stream]);
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
      {!stream ? (
        <span className="absolute inset-0 grid place-items-center text-sm font-bold text-slate-500">{label} 대기</span>
      ) : null}
      <strong className="absolute left-3 top-3 rounded-lg border border-sapphire-500/25 bg-command-950/80 px-3 py-1.5 text-sm font-bold text-slate-100">
        {label}
      </strong>
    </div>
  );
}
