import { useCallback, useState } from "react";
import { createRecordingPlaybackFile } from "./recordingHelpers.js";

export function useRecordingsController({ latestPlayableRecording, showNotification }) {
  const [recordingPlaybackFile, setRecordingPlaybackFile] = useState(null);

  const playLatestRecording = useCallback(() => {
    const playbackFile = createRecordingPlaybackFile(latestPlayableRecording);
    if (!playbackFile) {
      showNotification("재생 가능한 MP4가 아직 없습니다.", "warning");
      return;
    }
    setRecordingPlaybackFile(playbackFile);
  }, [latestPlayableRecording, showNotification]);

  return {
    playLatestRecording,
    recordingPlaybackFile,
    setRecordingPlaybackFile
  };
}
