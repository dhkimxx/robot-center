import Button from "../../../components/ui/Button.jsx";
import EmptyState from "../../../components/ui/EmptyState.jsx";
import Surface from "../../../components/ui/Surface.jsx";
import { formatDateTimeFull } from "../../../utils/formatters.js";
import { makeLiveRecordingTimingLabel, makeLiveStreamTimingLabel } from "../../missions/missionHelpers.js";
import {
  connectedLiveConnectionStatuses,
  reconnectableLiveStatuses
} from "../liveConnectionStates.js";
import { createEmptyLiveSession } from "../liveHelpers.js";
import { AudioSink } from "./AudioSink.jsx";
import { ConnectionStatusPanel } from "./ConnectionStatusPanel.jsx";
import { EventPanel } from "./EventPanel.jsx";
import { MissionRobotDropdown } from "./MissionRobotDropdown.jsx";
import { RobotMap } from "./RobotMap.jsx";
import { SensorPanel } from "./SensorPanel.jsx";
import { VideoPane } from "./VideoPane.jsx";

export function MissionControlView({
  isSensorSnapshotRefreshing = false,
  latestSensor,
  latestSensorSourceLabel,
  latestTelemetry,
  liveEvents,
  liveSessions,
  mission,
  missionTargets,
  onOpenMissionReplay,
  onReconnectSelectedMissionTarget,
  onRefreshSensorSnapshot,
  operationStatuses,
  selectedMissionTargetKey,
  setSelectedMissionTargetKey
}) {
  const selectedTarget = missionTargets.find((target) => target.key === selectedMissionTargetKey) ?? missionTargets[0] ?? null;
  const selectedSession = selectedTarget ? liveSessions[selectedTarget.key] ?? createEmptyLiveSession() : createEmptyLiveSession();
  const connectedCount = missionTargets.filter((target) => {
    const session = liveSessions[target.key] ?? createEmptyLiveSession();
    return connectedLiveConnectionStatuses.has(session.status);
  }).length;
  const canReconnectSelectedRobot = mission.status === "active"
    && Boolean(selectedTarget)
    && selectedSession.events.length > 0
    && reconnectableLiveStatuses.has(selectedSession.status);
  const missionConnectionLabel = missionTargets.length > 0
    ? `연결 ${connectedCount}/${missionTargets.length}대`
    : "임무에 배정된 로봇이 없습니다.";
  const missionStartLabel = mission.startedAt ? `시작 ${formatDateTimeFull(mission.startedAt)}` : "시작 전";
  const selectedStreamTimingLabel = selectedTarget
    ? makeLiveStreamTimingLabel(selectedTarget.liveStatus?.stream)
    : "";
  const selectedRecordingTimingLabel = selectedTarget
    ? makeLiveRecordingTimingLabel(selectedTarget.liveStatus?.recording)
    : "";

  return (
    <section className="grid h-full min-h-0 grid-cols-[minmax(0,1fr)_336px] items-stretch gap-3.5 max-[1240px]:grid-cols-1">
      <div className="grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)] gap-3 overflow-hidden">
        <div className="grid min-w-0 grid-cols-2 gap-3 max-[960px]:grid-cols-1">
          <Surface className="flex min-h-[58px] items-center justify-between gap-3 px-3 py-2.5">
            <div className="min-w-0">
              <strong className="block truncate text-sm font-extrabold text-slate-50">{mission.missionCode}</strong>
              <span className="mt-0.5 block truncate text-xs font-bold text-slate-500">
                {missionTargets.length > 0 ? `${missionStartLabel} · ${missionConnectionLabel}` : missionConnectionLabel}
              </span>
            </div>
            <Button size="sm" onClick={() => onOpenMissionReplay?.(mission)}>
              리플레이 보기
            </Button>
          </Surface>
          <Surface className="grid min-h-[68px] min-w-0 gap-1 px-3 py-2.5">
            <div className="flex min-w-0 items-center gap-3">
              <span className="shrink-0 text-xs font-bold text-slate-500">관제 로봇</span>
              <div className="min-w-0 flex-1">
                <div className="w-full min-w-0">
                  <MissionRobotDropdown
                    liveSessions={liveSessions}
                    missionTargets={missionTargets}
                    selectedMissionTargetKey={selectedMissionTargetKey}
                    setSelectedMissionTargetKey={setSelectedMissionTargetKey}
                  />
                </div>
              </div>
              {canReconnectSelectedRobot ? (
                <Button size="sm" onClick={onReconnectSelectedMissionTarget}>재연결</Button>
              ) : null}
            </div>
            {selectedStreamTimingLabel ? (
              <span className="truncate pl-[70px] text-xs font-bold text-slate-500 max-[560px]:pl-0">
                {selectedStreamTimingLabel} · {selectedRecordingTimingLabel}
              </span>
            ) : null}
          </Surface>
        </div>

        {!selectedTarget ? (
          <Surface className="grid min-h-0">
            <EmptyState>관제할 로봇을 선택할 수 없습니다.</EmptyState>
          </Surface>
        ) : (
          <Surface className="grid min-h-0 overflow-hidden p-0">
            <div className="grid min-h-0 grid-cols-2 grid-rows-2 gap-3 p-0 max-[900px]:grid-cols-1 max-[900px]:grid-rows-none">
            <VideoPane
              className="min-h-0"
              detectionOverlay={selectedSession.detectionOverlays?.rgb}
              label="RGB"
              stream={selectedSession.videoStreams.rgb}
            />
            <VideoPane
              className="min-h-0"
              detectionOverlay={selectedSession.detectionOverlays?.thermal}
              label="Thermal"
              stream={selectedSession.videoStreams.thermal}
              thermal
            />
            <RobotMap className="min-h-0" telemetry={latestTelemetry} />
            <SensorPanel
              className="min-h-0"
              isRefreshing={isSensorSnapshotRefreshing}
              onRefresh={onRefreshSensorSnapshot}
              sensor={latestSensor}
              sourceLabel={latestSensorSourceLabel}
            />
            </div>
          </Surface>
        )}
        <AudioSink stream={selectedSession.videoStreams.audio} />
      </div>

      <aside className="grid min-h-0 grid-rows-[auto_minmax(0,1fr)] gap-3 overflow-hidden">
        <ConnectionStatusPanel statuses={operationStatuses} />
        <EventPanel liveEvents={liveEvents} />
      </aside>
    </section>
  );
}
