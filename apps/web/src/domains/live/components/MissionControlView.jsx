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
import {
  createPresetLiveDashboardLayout,
  moveLiveDashboardWidget,
  readLiveDashboardLayout,
  resizeLiveDashboardWidget,
  writeLiveDashboardLayout
} from "../liveDashboardLayout.js";
import { makeLiveRobotDiagnostics } from "../liveDiagnostics.js";
import { createEmptyLiveSession } from "../liveHelpers.js";
import { AudioSink } from "./AudioSink.jsx";
import { ControlStatusSummary } from "./ControlStatusSummary.jsx";
import { DetectionOverlayControls } from "./DetectionOverlayControls.jsx";
import { LiveDashboardControls } from "./LiveDashboardControls.jsx";
import { LiveDashboardGrid } from "./LiveDashboardGrid.jsx";
import { LiveDashboardWidgetContent } from "./LiveDashboardWidgetContent.jsx";
import { MissionControlToolbar } from "./MissionControlToolbar.jsx";

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
  selectedMissionTargetKey,
  setSelectedMissionTargetKey
}) {
  const [detectionOverlaySettings, setDetectionOverlaySettings] = useState(readDetectionOverlaySettings);
  const [dashboardLayout, setDashboardLayout] = useState(readLiveDashboardLayout);
  const [draftDashboardLayout, setDraftDashboardLayout] = useState(() => dashboardLayout);
  const [isDashboardEditing, setIsDashboardEditing] = useState(false);
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
  const startDashboardEditing = useCallback(() => {
    setDraftDashboardLayout(dashboardLayout);
    setIsDashboardEditing(true);
  }, [dashboardLayout]);
  const cancelDashboardEditing = useCallback(() => {
    setDraftDashboardLayout(dashboardLayout);
    setIsDashboardEditing(false);
  }, [dashboardLayout]);
  const saveDashboardLayout = useCallback(() => {
    setDashboardLayout(writeLiveDashboardLayout(draftDashboardLayout));
    setIsDashboardEditing(false);
  }, [draftDashboardLayout]);
  const resetDashboardLayout = useCallback(() => {
    const nextLayout = createPresetLiveDashboardLayout();
    setDraftDashboardLayout(nextLayout);
    if (!isDashboardEditing) {
      setDashboardLayout(writeLiveDashboardLayout(nextLayout));
    }
  }, [isDashboardEditing]);
  const changeDashboardPreset = useCallback((presetId) => {
    const nextLayout = createPresetLiveDashboardLayout(presetId);
    if (isDashboardEditing) {
      setDraftDashboardLayout(nextLayout);
      return;
    }
    setDashboardLayout(writeLiveDashboardLayout(nextLayout));
  }, [isDashboardEditing]);
  const resizeDashboardWidget = useCallback((widgetId, nextSize) => {
    setDraftDashboardLayout((current) => resizeLiveDashboardWidget(current, widgetId, nextSize));
  }, []);
  const moveDashboardWidget = useCallback((widgetId, nextPosition) => {
    setDraftDashboardLayout((current) => moveLiveDashboardWidget(current, widgetId, nextPosition));
  }, []);
  const detectionOverlayTtlMs = detectionOverlaySettings.ttlSeconds * 1000;
  const activeDashboardLayout = isDashboardEditing ? draftDashboardLayout : dashboardLayout;
  const renderDashboardWidget = useCallback((widgetId) => {
    return (
      <LiveDashboardWidgetContent
        detectionOverlaySettings={detectionOverlaySettings}
        detectionOverlayTtlMs={detectionOverlayTtlMs}
        isLayoutEditing={isDashboardEditing}
        isSensorSnapshotRefreshing={isSensorSnapshotRefreshing}
        latestSensor={latestSensor}
        latestSensorSourceLabel={latestSensorSourceLabel}
        latestTelemetry={latestTelemetry}
        liveEvents={liveEvents}
        onRefreshSensorSnapshot={onRefreshSensorSnapshot}
        selectedSession={selectedSession}
        widgetId={widgetId}
      />
    );
  }, [
    detectionOverlaySettings,
    detectionOverlayTtlMs,
    isDashboardEditing,
    isSensorSnapshotRefreshing,
    latestSensor,
    latestSensorSourceLabel,
    latestTelemetry,
    liveEvents,
    onRefreshSensorSnapshot,
    selectedSession
  ]);

  return (
    <section
      className="grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)] gap-3 overflow-hidden max-[1100px]:h-auto max-[1100px]:grid-rows-none max-[1100px]:overflow-auto"
      data-testid="live-control-layout"
    >
      <div
        className="relative z-30 grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(520px,0.46fr)] gap-3 max-[1320px]:grid-cols-1"
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
        <ControlStatusSummary diagnostics={selectedDiagnostics} />
      </div>

      <div
        className="relative z-0 grid min-h-0 grid-rows-[auto_minmax(0,1fr)] gap-2 overflow-hidden"
        data-testid="live-dashboard-shell"
      >
        <div className="flex min-h-8 flex-wrap items-center justify-between gap-2">
          <DetectionOverlayControls
            maxDetections={detectionOverlaySettings.maxDetections}
            onChange={updateDetectionOverlaySetting}
            ttlSeconds={detectionOverlaySettings.ttlSeconds}
          />
          <LiveDashboardControls
            isEditing={isDashboardEditing}
            layout={activeDashboardLayout}
            onCancel={cancelDashboardEditing}
            onChangePreset={changeDashboardPreset}
            onReset={resetDashboardLayout}
            onSave={saveDashboardLayout}
            onStartEdit={startDashboardEditing}
          />
        </div>
        {!selectedTarget ? (
          <Surface className="grid min-h-0">
            <EmptyState>관제할 로봇을 선택할 수 없습니다.</EmptyState>
          </Surface>
        ) : (
          <LiveDashboardGrid
            isEditing={isDashboardEditing}
            layout={activeDashboardLayout}
            onMoveWidget={moveDashboardWidget}
            onResizeWidget={resizeDashboardWidget}
            renderWidget={renderDashboardWidget}
          />
        )}
      </div>
      <AudioSink stream={selectedSession.videoStreams.audio} />
    </section>
  );
}
