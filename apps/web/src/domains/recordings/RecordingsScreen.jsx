import { useEffect, useMemo, useState } from "react";
import Button from "../../components/ui/Button.jsx";
import EmptyState from "../../components/ui/EmptyState.jsx";
import SectionHeader from "../../components/ui/SectionHeader.jsx";
import StatusBadge from "../../components/ui/StatusBadge.jsx";
import Surface from "../../components/ui/Surface.jsx";
import { cn } from "../../utils/cn.js";
import { formatDateTime, formatDurationSeconds, makeStatusLabel } from "../../utils/formatters.js";
import {
  createRecordingPlaybackFile,
  getRecordingObjectEntries,
  isPlayableRecordingFile,
  makeFileStatusLabel,
  makeRecordingRobotGroups,
  makeRecordingSessionGroups
} from "./recordingHelpers.js";

function makeRecordingStatusTone(status) {
  if (["available", "uploaded", "completed"].includes(status)) {
    return "success";
  }
  if (["failed", "error"].includes(status)) {
    return "danger";
  }
  if (["recording", "pending", "uploading"].includes(status)) {
    return "warning";
  }
  return "neutral";
}

export default function RecordingsScreen({ onOpenPlaybackFile, recordings }) {
  const sessionGroups = useMemo(() => makeRecordingSessionGroups(recordings), [recordings]);
  const robotGroups = useMemo(() => makeRecordingRobotGroups(sessionGroups), [sessionGroups]);
  const [selectedRobotCode, setSelectedRobotCode] = useState("");
  const selectedRobotGroup = robotGroups.find((group) => group.robotCode === selectedRobotCode) ?? robotGroups[0] ?? null;

  useEffect(() => {
    if (robotGroups.length === 0) {
      setSelectedRobotCode("");
      return;
    }
    if (!selectedRobotCode || !robotGroups.some((group) => group.robotCode === selectedRobotCode)) {
      setSelectedRobotCode(robotGroups[0].robotCode);
    }
  }, [robotGroups, selectedRobotCode]);

  return (
    <Surface className="grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden">
      <SectionHeader
        meta={`${robotGroups.length}대 / ${sessionGroups.length}개 세션 / ${recordings.length}개 청크`}
        title="로봇별 녹화"
      />
      {robotGroups.length === 0 ? (
        <div className="grid gap-3">
          <EmptyState>아직 생성된 녹화 메타데이터가 없습니다.</EmptyState>
        </div>
      ) : (
        <div className="grid min-h-0 grid-cols-[260px_minmax(0,1fr)] gap-3 overflow-hidden max-[1180px]:grid-cols-1">
          <aside className="grid min-h-0 content-start gap-2 overflow-auto pr-1">
            {robotGroups.map((group) => (
              <button
                className={cn(
                  "grid min-h-16 w-full gap-1 rounded-lg border border-slate-500/20 bg-white/[0.045] px-3 py-2 text-left transition hover:border-sapphire-500/[0.45] hover:bg-sapphire-500/[0.12]",
                  selectedRobotGroup?.robotCode === group.robotCode && "border-sapphire-500/55 bg-sapphire-500/[0.10] shadow-[inset_3px_0_0_var(--color-sapphire)]"
                )}
                key={group.robotCode}
                type="button"
                onClick={() => setSelectedRobotCode(group.robotCode)}
              >
                <strong className="truncate text-sm font-bold text-slate-50">{group.robotCode}</strong>
                <span className="truncate text-xs font-semibold text-slate-400">{group.sessionCount}개 세션 / {group.chunkCount}개 청크</span>
                <small className="truncate text-xs font-semibold text-slate-500">최근 {formatDateTime(group.latestAt)}</small>
              </button>
            ))}
          </aside>
          <section className="grid min-h-0 min-w-0 content-start gap-3 overflow-auto pr-1">
            <div className="flex min-h-12 items-center justify-between gap-3 px-1">
              <div>
                <strong className="block text-base font-bold text-slate-50">{selectedRobotGroup?.robotCode}</strong>
                <span className="mt-1 block text-sm font-semibold text-slate-400">최근 녹화순</span>
              </div>
              <small className="text-xs font-semibold text-slate-500">{selectedRobotGroup?.sessionCount ?? 0}개 세션</small>
            </div>
            {(selectedRobotGroup?.sessions ?? []).map((session) => (
              <section className="grid gap-3 rounded-xl border border-slate-500/20 bg-white/[0.045] p-4" key={session.id}>
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
                <div className="grid gap-3">
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
                        onSelectVideo={(entry) => onOpenPlaybackFile(createRecordingPlaybackFile(recording, entry))}
                        recording={recording}
                      />
                    </div>
                  ))}
                </div>
              </section>
            ))}
          </section>
        </div>
      )}
    </Surface>
  );
}

function RecordingObjectList({ onSelectVideo, recording }) {
  const entries = getRecordingObjectEntries(recording);
  return (
    <div className="grid grid-cols-[repeat(auto-fit,minmax(180px,1fr))] gap-2">
      {entries.map((entry) => (
        <div className="flex min-h-11 min-w-0 items-center justify-between gap-3 rounded-lg border border-slate-500/20 bg-white/[0.045] p-2.5" key={`${recording.id}-${entry.type ?? entry.label}`}>
          <div className="min-w-0">
            <strong className="block truncate text-sm font-bold text-slate-50">{entry.label}</strong>
            <span className="mt-1 block break-words text-xs font-semibold leading-snug text-slate-400">{entry.contentType ?? "메타데이터"}</span>
          </div>
          {isPlayableRecordingFile(entry) ? (
            <Button size="sm" variant="primary" onClick={() => onSelectVideo?.(entry)}>
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
