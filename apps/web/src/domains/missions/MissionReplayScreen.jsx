import { useMemo } from "react";
import Button from "../../components/ui/Button.jsx";
import EmptyState from "../../components/ui/EmptyState.jsx";
import SectionHeader from "../../components/ui/SectionHeader.jsx";
import Surface from "../../components/ui/Surface.jsx";
import { ListSkeleton, PanelSkeleton } from "../../components/ui/Skeleton.jsx";
import {
  formatDateTimeFull,
  makeStatusLabel,
  missionTypeLabel
} from "../../utils/formatters.js";
import { formatMissionRobotCount, getMissionRobotDetails } from "./missionHelpers.js";
import { MissionReplayChunkPanel } from "./MissionReplayChunkPanel.jsx";
import { MissionReplayRobotList } from "./MissionReplayRobotList.jsx";
import {
  createRobotDisplayNamesByCode,
  getRobotDisplayName
} from "./missionReplayHelpers.js";
import { useMissionReplayRecordings } from "./useMissionReplayRecordings.js";

export function MissionReplayScreen({
  isLoading = false,
  mission,
  missingMissionCode,
  onBackToMissionList,
  onOpenPlaybackFile,
  robots = []
}) {
  const missionCode = mission?.missionCode ?? "";
  const robotDisplayNamesByCode = useMemo(() => createRobotDisplayNamesByCode(robots), [robots]);
  const robotDetails = useMemo(() => getMissionRobotDetails(mission, robots), [mission, robots]);
  const {
    chunkState,
    isChunkLoading,
    isLoadingMore,
    isSummaryLoading,
    loadMoreChunks,
    reloadChunks,
    reloadSummary,
    robotSummaries,
    selectedRobotCode,
    selectedRobotSummary,
    sessionGroups,
    setSelectedRobotCode,
    summaryState
  } = useMissionReplayRecordings(missionCode);
  const selectedRobotDisplayName = getRobotDisplayName(robotDisplayNamesByCode, selectedRobotSummary?.robotCode);

  if (!mission && isLoading) {
    return (
      <Surface className="grid h-full min-h-0 content-start gap-4">
        <SectionHeader
          action={<Button size="sm" onClick={() => onBackToMissionList()}>임무 목록</Button>}
          title="임무 리플레이"
          meta={missingMissionCode || "확인 중"}
        />
        <PanelSkeleton rows={5} />
      </Surface>
    );
  }

  if (!mission) {
    return (
      <Surface className="grid content-start gap-4">
        <SectionHeader
          action={<Button size="sm" onClick={() => onBackToMissionList()}>임무 목록</Button>}
          title="임무 리플레이"
          meta={missingMissionCode || "임무 없음"}
        />
        <EmptyState>리플레이할 임무를 찾을 수 없습니다.</EmptyState>
      </Surface>
    );
  }

  return (
    <Surface className="grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)] gap-3 overflow-hidden p-3">
      <MissionReplayHeader
        mission={mission}
        onBackToMissionList={onBackToMissionList}
        recordedRobotCount={robotSummaries.length}
        robotDetails={robotDetails}
        totalChunks={summaryState.summary?.totalChunks ?? 0}
      />

      {isLoading || isSummaryLoading ? (
        <div className="grid min-h-0 grid-cols-[300px_minmax(0,1fr)] gap-3 overflow-hidden max-[1180px]:grid-cols-1">
          <ListSkeleton count={4} />
          <PanelSkeleton rows={8} />
        </div>
      ) : summaryState.status === "error" ? (
        <div className="grid content-start gap-3">
          <EmptyState>녹화 요약을 불러오지 못했습니다. {summaryState.error}</EmptyState>
          <Button className="justify-self-start" size="sm" onClick={reloadSummary}>
            다시 조회
          </Button>
        </div>
      ) : robotSummaries.length === 0 ? (
        <EmptyState>이 임무에 저장된 리플레이 파일이 없습니다.</EmptyState>
      ) : (
        <div className="grid min-h-0 grid-cols-[300px_minmax(0,1fr)] gap-3 overflow-hidden max-[1180px]:grid-cols-1">
          <MissionReplayRobotList
            onSelectRobot={setSelectedRobotCode}
            robotDisplayNamesByCode={robotDisplayNamesByCode}
            robotSummaries={robotSummaries}
            selectedRobotCode={selectedRobotSummary?.robotCode ?? selectedRobotCode}
          />
          <MissionReplayChunkPanel
            chunkState={chunkState}
            isChunkLoading={isChunkLoading}
            isLoadingMore={isLoadingMore}
            onLoadMoreChunks={loadMoreChunks}
            onOpenPlaybackFile={onOpenPlaybackFile}
            onReloadChunks={reloadChunks}
            selectedRobotDisplayName={selectedRobotDisplayName}
            selectedRobotSummary={selectedRobotSummary}
            sessionGroups={sessionGroups}
          />
        </div>
      )}
    </Surface>
  );
}

function MissionReplayHeader({
  mission,
  onBackToMissionList,
  recordedRobotCount,
  robotDetails,
  totalChunks
}) {
  return (
    <div className="grid gap-3 rounded-xl border border-slate-500/20 bg-white/[0.045] p-3">
      <div className="flex min-w-0 items-center justify-between gap-3">
        <div className="min-w-0">
          <h2 className="truncate text-lg font-black text-slate-50">임무 리플레이</h2>
          <span className="mt-0.5 block truncate text-sm font-semibold text-slate-400">
            {mission.missionCode} / {makeStatusLabel(mission.status)}
          </span>
        </div>
        <Button size="sm" onClick={() => onBackToMissionList()}>임무 목록</Button>
      </div>
      <div className="flex min-w-0 items-start justify-between gap-4 max-[900px]:grid">
        <div className="min-w-0">
          <strong className="block truncate text-base font-black text-slate-50">{mission.name}</strong>
          <span className="mt-1 block text-sm font-bold text-slate-400">
            {missionTypeLabel(mission.missionType)} / {formatMissionRobotCount(robotDetails)}
          </span>
          <span className="mt-1 block text-xs font-semibold text-slate-500">
            시작 {formatDateTimeFull(mission.startedAt)} / 생성 {formatDateTimeFull(mission.createdAt)}
          </span>
        </div>
        <div className="flex shrink-0 flex-wrap justify-end gap-2 max-[900px]:justify-start">
          <span className="rounded-full border border-slate-500/20 bg-slate-500/[0.14] px-3 py-1 text-xs font-bold text-slate-300">
            녹화 로봇 {recordedRobotCount}대
          </span>
          <span className="rounded-full border border-slate-500/20 bg-slate-500/[0.14] px-3 py-1 text-xs font-bold text-slate-300">
            전체 {totalChunks}개 청크
          </span>
        </div>
      </div>
      {mission.siteNote ? (
        <p className="text-sm font-semibold leading-relaxed text-slate-400">{mission.siteNote}</p>
      ) : null}
    </div>
  );
}
