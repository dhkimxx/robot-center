import { useEffect, useRef } from "react";

export function AudioSink({ stream }) {
  const audioRef = useRef(null);
  useEffect(() => {
    const audio = audioRef.current;
    if (!audio) {
      return undefined;
    }
    audio.srcObject = stream;
    if (stream) {
      void audio.play().catch(() => {});
    }
    return () => {
      audio.srcObject = null;
    };
  }, [stream]);
  return <audio ref={audioRef} autoPlay />;
}
