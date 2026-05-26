import { useEffect, useMemo, useState } from "react";
import Button from "../../components/ui/Button.jsx";
import EmptyState from "../../components/ui/EmptyState.jsx";
import SectionHeader from "../../components/ui/SectionHeader.jsx";
import StatusBadge from "../../components/ui/StatusBadge.jsx";
import Surface from "../../components/ui/Surface.jsx";
import { RecordingObjectList, makeRecordingStatusTone } from "../recordings/RecordingObjectList.jsx";
import {
  makeRecordingRobotGroups,
  makeRecordingSessionGroups
} from "../recordings/recordingHelpers.js";
import {
  formatDateTime,
  formatDurationSeconds,
  makeStatusLabel,
  missionTypeLabel
} from "../../utils/formatters.js";
import { cn } from "../../utils/cn.js";
import { formatMissionRobotCount, getMissionRobotDetails } from "./missionHelpers.js";

export function MissionReplayScreen({
  mission,
  missingMissionCode,
  observedStreams,
  onBackToMissionList,
  onOpenPlaybackFile,
  recordings,
  robots,
  streamingStatuses
}) {
  const missionRecordings = useMemo(
    () => recordings.filter((recording) => recording.missionCode === mission?.missionCode),
    [mission?.missionCode, recordings]
  );
  const sessionGroups = useMemo(() => makeRecordingSessionGroups(missionRecordings), [missionRecordings]);
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

  if (!mission) {
    return (
      <Surface className="grid content-start gap-4">
        <SectionHeader
          action={<Button size="sm" onClick={onBackToMissionList}>임무 목록</Button>}
          title="임무 리플레이"
          meta={missingMissionCode || "임무 없음"}
        />
        <EmptyState>리플레이할 임무를 찾을 수 없습니다.</EmptyState>
      </Surface>
    );
  }

  const robotDetails = getMissionRobotDetails(mission, robots, streamingStatuses, observedStreams);

  return (
    <Surface className="grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)] gap-3 overflow-hidden p-3">
      <div className="grid gap-3 rounded-xl border border-slate-500/20 bg-white/[0.045] p-3">
        <div className="flex min-w-0 items-center justify-between gap-3">
          <div className="min-w-0">
            <h2 className="truncate text-lg font-black text-slate-50">임무 리플레이</h2>
            <span className="mt-0.5 block truncate text-sm font-semibold text-slate-400">
              {mission.missionCode} / {makeStatusLabel(mission.status)}
            </span>
          </div>
          <Button size="sm" onClick={onBackToMissionList}>임무 목록</Button>
        </div>
        <div className="flex min-w-0 items-start justify-between gap-4 max-[900px]:grid">
          <div className="min-w-0">
            <strong className="block truncate text-base font-black text-slate-50">{mission.name}</strong>
            <span className="mt-1 block text-sm font-bold text-slate-400">
              {missionTypeLabel(mission.missionType)} / {formatMissionRobotCount(robotDetails)}
            </span>
          </div>
          <span className="shrink-0 rounded-full border border-slate-500/20 bg-slate-500/[0.14] px-3 py-1 text-xs font-bold text-slate-300">
            {sessionGroups.length}개 세션 / {missionRecordings.length}개 청크
          </span>
        </div>
        {mission.siteNote ? (
          <p className="text-sm font-semibold leading-relaxed text-slate-400">{mission.siteNote}</p>
        ) : null}
      </div>

      {robotGroups.length === 0 ? (
        <EmptyState>이 임무에 저장된 리플레이 파일이 없습니다.</EmptyState>
      ) : (
        <div className="grid min-h-0 grid-cols-[280px_minmax(0,1fr)] gap-3 overflow-hidden max-[1180px]:grid-cols-1">
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

          <section className="grid min-h-0 min-w-0 grid-rows-[auto_minmax(0,1fr)] gap-2 overflow-hidden">
            <div className="flex min-h-11 items-center justify-between gap-3 rounded-lg border border-slate-500/20 bg-white/[0.035] px-3">
              <div>
                <strong className="block text-sm font-black text-slate-50">{selectedRobotGroup?.robotCode}</strong>
                <span className="mt-0.5 block text-xs font-semibold text-slate-500">최근 저장순</span>
              </div>
              <small className="text-xs font-semibold text-slate-500">{selectedRobotGroup?.sessionCount ?? 0}개 세션</small>
            </div>
            <div className="grid min-h-0 content-start gap-3 overflow-auto pr-1">
              {(selectedRobotGroup?.sessions ?? []).map((session) => (
                <section className="grid gap-3 rounded-xl border border-slate-500/20 bg-white/[0.045] p-3" key={session.id}>
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
              ))}
            </div>
          </section>
        </div>
      )}
    </Surface>
  );
}
