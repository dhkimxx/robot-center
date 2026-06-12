import { useCallback, useState } from "react";
import EmptyState from "../../../components/ui/EmptyState.jsx";
import Surface from "../../../components/ui/Surface.jsx";
import { formatDateTimeFull } from "../../../utils/formatters.js";
import { makeLiveRecordingTimingLabel, makeLiveStreamTimingLabel } from "../../missions/missionHelpers.js";
import {
  readDetectionOverlaySettings,
  writeDetectionOverlaySettings
} from "../detectionOverlaySettings.js";
import {
  connectedLiveConnectionStatuses,
  reconnectableLiveStatuses
} from "../liveConnectionStates.js";
import { makeLiveRobotDiagnostics } from "../liveDiagnostics.js";
import { createEmptyLiveSession } from "../liveHelpers.js";
import { AudioSink } from "./AudioSink.jsx";
import { ControlStatusSummary } from "./ControlStatusSummary.jsx";
import { DetectionOverlayControls } from "./DetectionOverlayControls.jsx";
import { EventPanel } from "./EventPanel.jsx";
import { MissionControlToolbar } from "./MissionControlToolbar.jsx";
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
  const [detectionOverlaySettings, setDetectionOverlaySettings] = useState(readDetectionOverlaySettings);
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
  const selectedDiagnostics = makeLiveRobotDiagnostics({
    session: selectedSession,
    target: selectedTarget
  });
  const updateDetectionOverlaySetting = useCallback((field, value) => {
    setDetectionOverlaySettings((current) => writeDetectionOverlaySettings({
      ...current,
      [field]: value
    }));
  }, []);
  const detectionOverlayTtlMs = detectionOverlaySettings.ttlSeconds * 1000;

  return (
    <section
      className="grid h-full min-h-0 grid-rows-[auto_minmax(0,1.35fr)_minmax(0,0.95fr)_minmax(128px,0.5fr)] gap-3 overflow-hidden max-[1100px]:h-auto max-[1100px]:grid-rows-none max-[1100px]:overflow-auto"
      data-testid="live-control-layout"
    >
      <div
        className="relative z-30 grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(300px,0.35fr)] gap-3 max-[900px]:grid-cols-1"
        data-testid="live-control-topbar"
      >
        <MissionControlToolbar
          canReconnectSelectedRobot={canReconnectSelectedRobot}
          liveSessions={liveSessions}
          mission={mission}
          missionConnectionLabel={missionConnectionLabel}
          missionStartLabel={missionStartLabel}
          missionTargets={missionTargets}
          onOpenMissionReplay={onOpenMissionReplay}
          onReconnectSelectedMissionTarget={onReconnectSelectedMissionTarget}
          selectedMissionTargetKey={selectedMissionTargetKey}
          selectedRecordingTimingLabel={selectedRecordingTimingLabel}
          selectedStreamTimingLabel={selectedStreamTimingLabel}
          setSelectedMissionTargetKey={setSelectedMissionTargetKey}
        />
        <ControlStatusSummary diagnostics={selectedDiagnostics} statuses={operationStatuses} />
      </div>

      <div className="relative z-0 grid min-h-0 grid-rows-[auto_minmax(0,1fr)] gap-2" data-testid="live-control-video-grid">
        <div className="flex min-h-8 items-center justify-end">
          <DetectionOverlayControls
            maxDetections={detectionOverlaySettings.maxDetections}
            onChange={updateDetectionOverlaySetting}
            ttlSeconds={detectionOverlaySettings.ttlSeconds}
          />
        </div>
        {!selectedTarget ? (
          <Surface className="grid min-h-0">
            <EmptyState>관제할 로봇을 선택할 수 없습니다.</EmptyState>
          </Surface>
        ) : (
          <div className="grid min-h-0 grid-cols-2 gap-3 max-[900px]:grid-cols-1" data-testid="live-control-video-panes">
            <VideoPane
              className="min-h-0"
              detectionOverlay={selectedSession.detectionOverlays?.rgb}
              detectionOverlayMaxCount={detectionOverlaySettings.maxDetections}
              detectionOverlayTtlMs={detectionOverlayTtlMs}
              label="RGB"
              stream={selectedSession.videoStreams.rgb}
            />
            <VideoPane
              className="min-h-0"
              detectionOverlay={selectedSession.detectionOverlays?.thermal}
              detectionOverlayMaxCount={detectionOverlaySettings.maxDetections}
              detectionOverlayTtlMs={detectionOverlayTtlMs}
              label="Thermal"
              stream={selectedSession.videoStreams.thermal}
              thermal
            />
          </div>
        )}
      </div>

      <div className="grid min-h-0 grid-cols-2 gap-3 max-[900px]:grid-cols-1" data-testid="live-control-map-sensor-grid">
        <RobotMap className="min-h-0" telemetry={latestTelemetry} />
        <SensorPanel
          className="min-h-0"
          isRefreshing={isSensorSnapshotRefreshing}
          onRefresh={onRefreshSensorSnapshot}
          sensor={latestSensor}
          sourceLabel={latestSensorSourceLabel}
        />
      </div>

      <EventPanel className="min-h-0" data-testid="live-control-event-panel" liveEvents={liveEvents} />
      <AudioSink stream={selectedSession.videoStreams.audio} />
    </section>
  );
}
