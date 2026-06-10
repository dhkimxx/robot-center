import Button from "../../components/ui/Button.jsx";
import StatusBadge from "../../components/ui/StatusBadge.jsx";
import {
  createRecordingPlaybackFile,
  getRecordingObjectEntries,
  isPlayableRecordingFile,
  makeFileStatusLabel
} from "./recordingHelpers.js";

export function makeRecordingStatusTone(status) {
  if (["available", "uploaded", "completed"].includes(status)) {
    return "success";
  }
  if (["failed", "error"].includes(status)) {
    return "danger";
  }
  if (["recording", "pending", "uploading", "finalizing"].includes(status)) {
    return "warning";
  }
  return "neutral";
}

export function RecordingObjectList({ onOpenPlaybackFile, recording }) {
  const entries = getRecordingObjectEntries(recording);
  return (
    <div className="grid grid-cols-[repeat(auto-fit,minmax(180px,1fr))] gap-2">
      {entries.map((entry) => (
        <div
          className="flex min-h-11 min-w-0 items-center justify-between gap-3 rounded-lg border border-slate-500/20 bg-white/[0.045] p-2.5"
          data-file-type={entry.type ?? ""}
          data-testid="recording-object-entry"
          key={`${recording.id}-${entry.type ?? entry.label}`}
        >
          <div className="min-w-0">
            <strong className="block truncate text-sm font-bold text-slate-50">{entry.label}</strong>
            <span className="mt-1 block break-words text-xs font-semibold leading-snug text-slate-400">{entry.contentType ?? "메타데이터"}</span>
          </div>
          {isPlayableRecordingFile(entry) ? (
            <Button
              data-file-type={entry.type ?? ""}
              data-testid="recording-playback-button"
              size="sm"
              variant="primary"
              onClick={() => onOpenPlaybackFile?.(createRecordingPlaybackFile(recording, entry))}
            >
              재생
            </Button>
          ) : entry.status === "available" && entry.url ? (
            <a
              className="inline-flex h-8 shrink-0 items-center justify-center rounded-lg border border-sapphire-500/35 bg-sapphire-600 px-3 text-xs font-semibold text-white transition hover:bg-sapphire-500"
              href={entry.url}
              target="_blank"
              rel="noreferrer"
            >
              {entry.type === "manifest" ? "보기" : "열기"}
            </a>
          ) : (
            <StatusBadge tone={makeRecordingStatusTone(entry.status)}>{makeFileStatusLabel(entry.status)}</StatusBadge>
          )}
        </div>
      ))}
    </div>
  );
}
