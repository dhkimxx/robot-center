import Button from "../../../components/ui/Button.jsx";
import EmptyState from "../../../components/ui/EmptyState.jsx";
import SectionHeader from "../../../components/ui/SectionHeader.jsx";
import Surface from "../../../components/ui/Surface.jsx";
import {
  makeLiveStatusLabel,
  makeStatusLabel,
  missionTypeLabel
} from "../../../utils/formatters.js";
import {
  connectedLiveConnectionStatuses,
  reconnectableLiveStatuses
} from "../liveConnectionStates.js";
import { createEmptyLiveSession } from "../liveHelpers.js";
import { AudioSink } from "./AudioSink.jsx";
import { ConnectionStatusPanel } from "./ConnectionStatusPanel.jsx";
import { EventPanel } from "./EventPanel.jsx";
import { LatestRecordingPanel } from "./LatestRecordingPanel.jsx";
import { MissionRobotSelector } from "./MissionRobotSelector.jsx";
import { RobotMap } from "./RobotMap.jsx";
import { SensorPanel } from "./SensorPanel.jsx";
import { VideoPane } from "./VideoPane.jsx";

export function MissionControlView({
  latestRecording,
  latestSensor,
  latestTelemetry,
  liveEvents,
  liveSessions,
  mission,
  missionTargets,
  onBackToMissionList,
  onEndMission,
  onOpenRecordings,
  onPlayLatestRecording,
  onReconnectSelectedMissionTarget,
  onStartMission,
  operationStatuses,
  playbackRecording,
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

  return (
    <section className="grid h-full min-h-0 grid-cols-[minmax(0,1fr)_336px] items-stretch gap-3.5 max-[1240px]:grid-cols-1">
      <Surface className="grid h-full min-h-0 grid-rows-[auto_auto_auto_minmax(0,1fr)] gap-3 overflow-hidden">
        <SectionHeader
          className="mb-0 items-start"
          title={mission.name}
          meta={`${mission.missionCode} / ${missionTypeLabel(mission.missionType)} / ${makeStatusLabel(mission.status)} / ${missionTargets.length}대`}
          action={(
            <div className="flex flex-wrap justify-end gap-2">
              <Button size="sm" onClick={onBackToMissionList}>임무 목록</Button>
              <Button size="sm" disabled={mission.status !== "ready"} onClick={() => onStartMission(mission.missionCode)}>시작</Button>
              <Button size="sm" disabled={mission.status !== "active"} onClick={() => onEndMission(mission.missionCode)}>종료</Button>
            </div>
          )}
        />

        <div className="flex items-center justify-between gap-3 rounded-lg border border-slate-500/20 bg-white/[0.045] p-3 max-[900px]:grid">
          <div className="min-w-0">
            <strong className="block truncate text-sm font-bold text-slate-50">{mission.missionCode}</strong>
            <span className="mt-1 block truncate text-xs font-bold text-slate-400">
              {selectedTarget ? `선택 ${selectedTarget.robotCode} ${makeLiveStatusLabel(selectedSession.status)} / 연결 ${connectedCount}/${missionTargets.length}대` : "임무에 배정된 로봇이 없습니다."}
            </span>
          </div>
          <div className="grid min-w-[min(100%,520px)] gap-2 max-[900px]:min-w-0">
            <label className="grid gap-1">
              <span className="text-xs font-bold text-slate-400">관제 로봇</span>
              <select
                className="min-w-[220px] max-[900px]:min-w-0"
                disabled={missionTargets.length === 0}
                value={selectedTarget?.key ?? ""}
                onChange={(event) => setSelectedMissionTargetKey(event.target.value)}
              >
                {missionTargets.length === 0 ? (
                  <option value="">선택 없음</option>
                ) : (
                  missionTargets.map((target) => (
                    <option key={target.key} value={target.key}>
                      {target.robot?.displayName ?? target.robotCode} / {target.robotCode}
                    </option>
                  ))
                )}
              </select>
            </label>
            {canReconnectSelectedRobot ? (
              <div className="flex justify-end">
                <Button size="sm" onClick={onReconnectSelectedMissionTarget}>재연결</Button>
              </div>
            ) : null}
          </div>
        </div>

        <MissionRobotSelector
          liveSessions={liveSessions}
          missionTargets={missionTargets}
          selectedTarget={selectedTarget}
          setSelectedMissionTargetKey={setSelectedMissionTargetKey}
        />

        {!selectedTarget ? (
          <EmptyState>관제할 로봇을 선택할 수 없습니다.</EmptyState>
        ) : (
          <div className="grid min-h-0 grid-cols-2 grid-rows-2 gap-3 max-[900px]:grid-cols-1 max-[900px]:grid-rows-none">
            <VideoPane className="min-h-0" label="RGB" stream={selectedSession.videoStreams.rgb} />
            <VideoPane className="min-h-0" label="Thermal" stream={selectedSession.videoStreams.thermal} thermal />
            <RobotMap className="min-h-0" telemetry={latestTelemetry} />
            <SensorPanel className="min-h-0" sensor={latestSensor} />
          </div>
        )}
        <AudioSink stream={selectedSession.videoStreams.audio} />
      </Surface>

      <aside className="grid min-h-0 content-start gap-3 overflow-auto">
        <ConnectionStatusPanel statuses={operationStatuses} />
        <LatestRecordingPanel
          onOpenRecordings={onOpenRecordings}
          onPlayRecording={onPlayLatestRecording}
          playbackRecording={playbackRecording}
          recording={latestRecording}
        />
        <EventPanel liveEvents={liveEvents} />
      </aside>
    </section>
  );
}
