import { RecordingPlaybackModal } from "./RecordingPlaybackModal.jsx";

export function RecordingModals({
  recordingPlaybackFile,
  setRecordingPlaybackFile
}) {
  return recordingPlaybackFile ? (
    <RecordingPlaybackModal
      file={recordingPlaybackFile}
      onClose={() => setRecordingPlaybackFile(null)}
    />
  ) : null;
}
