import Button from "../../components/ui/Button.jsx";
import EmptyState from "../../components/ui/EmptyState.jsx";
import StatusBadge from "../../components/ui/StatusBadge.jsx";
import { PanelSkeleton } from "../../components/ui/Skeleton.jsx";
import { RecordingObjectList, makeRecordingStatusTone } from "../recordings/RecordingObjectList.jsx";
import {
  formatDateTime,
  formatDurationSeconds,
  makeStatusLabel
} from "../../utils/formatters.js";
import {
  makeFileAvailabilityLabel,
  makeLoadedChunkLabel,
  replayChunkPageSize
} from "./missionReplayHelpers.js";

export function MissionReplayChunkPanel({
  chunkState,
  isChunkLoading,
  isLoadingMore,
  onLoadMoreChunks,
  onOpenPlaybackFile,
  onReloadChunks,
  selectedRobotDisplayName,
  selectedRobotSummary,
  sessionGroups
}) {
  return (
    <section className="grid min-h-0 min-w-0 grid-rows-[auto_minmax(0,1fr)] gap-2 overflow-hidden">
      <MissionReplayChunkHeader
        chunkState={chunkState}
        selectedRobotDisplayName={selectedRobotDisplayName}
        selectedRobotSummary={selectedRobotSummary}
      />
      <div className="grid min-h-0 content-start gap-3 overflow-auto pr-1">
        {isChunkLoading ? (
          <PanelSkeleton rows={8} />
        ) : chunkState.status === "error" && chunkState.chunks.length === 0 ? (
          <div className="grid content-start gap-3">
            <EmptyState>녹화 청크를 불러오지 못했습니다. {chunkState.error}</EmptyState>
            <Button className="justify-self-start" size="sm" onClick={onReloadChunks}>
              다시 조회
            </Button>
          </div>
        ) : sessionGroups.length === 0 ? (
          <EmptyState>선택한 로봇의 리플레이 청크가 없습니다.</EmptyState>
        ) : (
          <>
            {sessionGroups.map((session) => (
              <MissionReplaySessionGroup
                key={session.id}
                onOpenPlaybackFile={onOpenPlaybackFile}
                session={session}
              />
            ))}
            <div className="flex min-h-10 items-center justify-center">
              {chunkState.page?.hasMore ? (
                <Button disabled={isLoadingMore} size="sm" onClick={onLoadMoreChunks}>
                  {isLoadingMore ? "불러오는 중" : `이전 청크 ${replayChunkPageSize}개 더 보기`}
                </Button>
              ) : (
                <span className="text-xs font-semibold text-slate-500">
                  {makeLoadedChunkLabel(chunkState.chunks.length, chunkState.page?.total ?? chunkState.chunks.length)}
                </span>
              )}
            </div>
            {chunkState.status === "error" && chunkState.chunks.length > 0 ? (
              <EmptyState>추가 청크를 불러오지 못했습니다. {chunkState.error}</EmptyState>
            ) : null}
          </>
        )}
      </div>
    </section>
  );
}

function MissionReplayChunkHeader({
  chunkState,
  selectedRobotDisplayName,
  selectedRobotSummary
}) {
  return (
    <div className="grid gap-2 rounded-lg border border-slate-500/20 bg-white/[0.035] p-3">
      <div className="flex min-h-9 items-center justify-between gap-3">
        <div className="min-w-0">
          <strong className="block truncate text-sm font-black text-slate-50">{selectedRobotDisplayName}</strong>
          <span className="mt-0.5 block truncate text-xs font-semibold text-slate-500">
            {selectedRobotSummary?.robotCode ? `${selectedRobotSummary.robotCode} · 최근 저장순` : "최근 저장순"}
          </span>
        </div>
        <small className="shrink-0 text-xs font-semibold text-slate-500">
          {makeLoadedChunkLabel(chunkState.chunks.length, chunkState.page?.total ?? selectedRobotSummary?.chunkCount ?? 0)}
        </small>
      </div>
      {selectedRobotSummary ? (
        <div className="flex flex-wrap gap-2">
          <StatusBadge size="xs" tone="success">완료 {selectedRobotSummary.uploadedChunkCount}개</StatusBadge>
          <StatusBadge size="xs" tone="info">진행 {selectedRobotSummary.recordingChunkCount}개</StatusBadge>
          <StatusBadge size="xs" tone={selectedRobotSummary.partialChunkCount > 0 ? "warning" : "neutral"}>
            부분 {selectedRobotSummary.partialChunkCount}개
          </StatusBadge>
          <StatusBadge size="xs" tone="neutral">{makeFileAvailabilityLabel(selectedRobotSummary, "rgb_audio_mp4", "RGB")}</StatusBadge>
          <StatusBadge size="xs" tone="neutral">{makeFileAvailabilityLabel(selectedRobotSummary, "thermal_mp4", "Thermal")}</StatusBadge>
        </div>
      ) : null}
    </div>
  );
}

function MissionReplaySessionGroup({ onOpenPlaybackFile, session }) {
  return (
    <section className="grid gap-3 rounded-xl border border-slate-500/20 bg-white/[0.045] p-3">
      <div className="flex items-start justify-between gap-4 max-[900px]:flex-col">
        <div className="min-w-0">
          <strong className="block truncate text-base font-bold text-slate-50">{session.missionCode}</strong>
          <span className="mt-1 block text-sm font-semibold text-slate-400">
            {formatDateTime(session.startedAt)} - {formatDateTime(session.endedAt)}
          </span>
        </div>
        <div className="grid justify-items-end gap-1 max-[900px]:justify-items-start">
          <StatusBadge tone={makeRecordingStatusTone(session.status)}>{makeStatusLabel(session.status)}</StatusBadge>
          <small className="text-xs font-semibold text-slate-500">{session.availableFileCount}/{session.fileCount} 파일 저장</small>
        </div>
      </div>
      <div className="grid gap-2">
        {session.chunks.map((recording) => (
          <div className="grid gap-3 rounded-lg border border-slate-500/20 bg-white/[0.045] p-3" key={recording.id}>
            <div className="grid gap-1">
              <strong className="text-sm font-bold text-slate-50">청크 #{recording.chunkIndex}</strong>
              <span className="text-xs font-semibold leading-relaxed text-slate-400">
                {formatDurationSeconds(recording.durationSeconds)} / {formatDateTime(recording.startedAt)} - {formatDateTime(recording.endedAt)}
              </span>
              <span className="text-xs font-semibold leading-relaxed text-slate-400">
                상태 {makeStatusLabel(recording.status)} / 갱신 {formatDateTime(recording.updatedAt)}
              </span>
            </div>
            <RecordingObjectList
              onOpenPlaybackFile={onOpenPlaybackFile}
              recording={recording}
            />
          </div>
        ))}
      </div>
    </section>
  );
}
