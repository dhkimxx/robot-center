import { useEffect, useRef } from "react";

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
  const videoPaneClassName = [
    "video-pane",
    thermal ? "thermal" : "",
    compact ? "compact" : "",
    className
  ].filter(Boolean).join(" ");
  return (
    <div className={videoPaneClassName}>
      <video ref={videoRef} autoPlay playsInline muted={label !== "Audio"} />
      {!stream ? <span>{label} 대기</span> : null}
      <strong>{label}</strong>
    </div>
  );
}
