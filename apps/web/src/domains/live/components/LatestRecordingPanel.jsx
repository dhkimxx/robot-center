import Button from "../../../components/ui/Button.jsx";
import EmptyState from "../../../components/ui/EmptyState.jsx";
import SectionHeader from "../../../components/ui/SectionHeader.jsx";
import Surface from "../../../components/ui/Surface.jsx";
import { cn } from "../../../utils/cn.js";
import {
  formatDateTime,
  makeStatusLabel
} from "../../../utils/formatters.js";
import { createRecordingPlaybackFile } from "../../recordings/recordingHelpers.js";

export function LatestRecordingPanel({ recording, compact = false, onOpenRecordings, onPlayRecording, playbackRecording }) {
  const playableFile = createRecordingPlaybackFile(playbackRecording ?? recording);
  return (
    <Surface className={cn("grid gap-3", compact && "rounded-xl p-3 shadow-none")}>
      <SectionHeader
        className="mb-0"
        title="녹화 저장"
        meta={playableFile ? "재생 가능" : recording ? makeStatusLabel(recording.status) : "대기"}
      />
      {!recording ? (
        <EmptyState>진행 중인 임무가 시작되면 저장 청크가 생성됩니다.</EmptyState>
      ) : (
        <div className="grid gap-3 rounded-lg border border-slate-500/20 bg-white/[0.045] p-3">
          <div className="min-w-0">
            <strong className="block truncate text-sm font-bold text-slate-50">{recording.missionCode} / #{recording.chunkIndex}</strong>
            <span className="mt-1 block text-xs font-semibold leading-relaxed text-slate-400">
              {formatDateTime(recording.startedAt)} - {formatDateTime(recording.endedAt)}
            </span>
          </div>
          <div className="flex flex-wrap gap-2">
            {playableFile ? (
              <Button size="sm" variant="primary" onClick={onPlayRecording}>
                재생
              </Button>
            ) : null}
            <Button size="sm" onClick={onOpenRecordings}>
              녹화 목록
            </Button>
          </div>
        </div>
      )}
    </Surface>
  );
}
