import Button from "../../components/ui/Button.jsx";
import EmptyState from "../../components/ui/EmptyState.jsx";
import SectionHeader from "../../components/ui/SectionHeader.jsx";
import StatusBadge from "../../components/ui/StatusBadge.jsx";
import Surface from "../../components/ui/Surface.jsx";
import { formatDateTime, formatDurationSeconds, makeStatusLabel } from "../../utils/formatters.js";
import { makeRecordingStatusTone } from "./RecordingObjectList.jsx";
import {
  createRecordingPlaybackFile,
  filterRecordingsByMissionCode,
  getPlayableRecordingVideoEntries,
  makeFileStatusLabel,
  sortRecordingChunksLatestFirst
} from "./recordingHelpers.js";

const visibleRecordingLimit = 6;

export function ActiveMissionRecordingsPanel({
  missionCode,
  onOpenPlaybackFile,
  recordings = []
}) {
  const missionRecordings = sortRecordingChunksLatestFirst(
    filterRecordingsByMissionCode(recordings, missionCode)
  );
  const visibleRecordings = missionRecordings.slice(0, visibleRecordingLimit);
  const playableChunkCount = missionRecordings.filter((recording) => getPlayableRecordingVideoEntries(recording).length > 0).length;

  return (
    <Surface className="grid min-h-[220px] grid-rows-[auto_minmax(0,1fr)] gap-3 overflow-hidden">
      <SectionHeader
        className="mb-0"
        title="녹화 파일"
        meta={`${playableChunkCount}/${missionRecordings.length} 청크`}
      />
      <div className="grid min-h-0 content-start gap-2 overflow-auto pr-1">
        {visibleRecordings.length === 0 ? (
          <EmptyState>저장된 녹화 chunk가 없습니다.</EmptyState>
        ) : (
          visibleRecordings.map((recording) => (
            <ActiveMissionRecordingItem
              key={recording.id}
              onOpenPlaybackFile={onOpenPlaybackFile}
              recording={recording}
            />
          ))
        )}
      </div>
    </Surface>
  );
}

function ActiveMissionRecordingItem({ onOpenPlaybackFile, recording }) {
  const playableFiles = getPlayableRecordingVideoEntries(recording);
  const status = normalizeDisplayRecordingStatus(recording);

  return (
    <div className="grid gap-2 rounded-lg border border-slate-500/20 bg-white/[0.045] p-3">
      <div className="flex min-w-0 items-start justify-between gap-3">
        <div className="min-w-0">
          <strong className="block truncate text-sm font-bold text-slate-50">
            {recording.robotCode} / #{recording.chunkIndex}
          </strong>
          <span className="mt-1 block truncate text-xs font-semibold text-slate-500">
            {formatDurationSeconds(recording.durationSeconds)} · {formatDateTime(recording.startedAt)}
          </span>
        </div>
        <StatusBadge size="xs" tone={makeRecordingStatusTone(status)}>
          {makeStatusLabel(status)}
        </StatusBadge>
      </div>

      {playableFiles.length > 0 ? (
        <div className="flex flex-wrap gap-2">
          {playableFiles.map((file) => (
            <Button
              key={`${recording.id}-${file.type}`}
              size="sm"
              variant="primary"
              onClick={() => onOpenPlaybackFile?.(createRecordingPlaybackFile(recording, file))}
            >
              {file.type === "thermal_mp4" ? "Thermal" : "RGB"}
            </Button>
          ))}
        </div>
      ) : (
        <span className="text-xs font-semibold text-slate-500">
          {makePendingRecordingLabel(recording)}
        </span>
      )}
    </div>
  );
}

function normalizeDisplayRecordingStatus(recording) {
  if (getPlayableRecordingVideoEntries(recording).length > 0) {
    return "uploaded";
  }
  return recording.status;
}

function makePendingRecordingLabel(recording) {
  const files = recording.files ?? [];
  const activeFile = files.find((file) => ["recording", "finalizing", "partial", "failed"].includes(file.status));
  return activeFile ? makeFileStatusLabel(activeFile.status) : makeStatusLabel(recording.status);
}
