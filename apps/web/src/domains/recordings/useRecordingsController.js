import { useState } from "react";

export function useRecordingsController() {
  const [recordingPlaybackFile, setRecordingPlaybackFile] = useState(null);

  return {
    recordingPlaybackFile,
    setRecordingPlaybackFile
  };
}
