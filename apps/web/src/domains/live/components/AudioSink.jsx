import { useEffect, useRef } from "react";
import { stopMediaStreamTracks } from "../liveMediaCleanup.js";

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
      if (audio.srcObject === stream) {
        audio.srcObject = null;
      }
      stopMediaStreamTracks(stream);
    };
  }, [stream]);
  return <audio ref={audioRef} autoPlay />;
}
