import { EventPanel } from "./EventPanel.jsx";
import { RobotMap } from "./RobotMap.jsx";
import { SensorPanel } from "./SensorPanel.jsx";
import { VideoPane } from "./VideoPane.jsx";

export function LiveDashboardWidgetContent({
  detectionOverlaySettings,
  detectionOverlayTtlMs,
  isLayoutEditing,
  isSensorSnapshotRefreshing,
  latestSensor,
  latestSensorSourceLabel,
  latestTelemetry,
  liveEvents,
  onRefreshSensorSnapshot,
  selectedSession,
  widgetId
}) {
  switch (widgetId) {
    case "event":
      return (
        <EventPanel
          className="h-full min-h-0"
          data-testid="live-control-event-panel"
          liveEvents={liveEvents}
        />
      );
    case "map":
      return <RobotMap className="h-full min-h-0" isPreviewDisabled={isLayoutEditing} telemetry={latestTelemetry} />;
    case "rgb":
      return (
        <VideoPane
          className="h-full min-h-0"
          detectionOverlay={selectedSession.detectionOverlays?.rgb}
          detectionOverlayMaxCount={detectionOverlaySettings.maxDetections}
          detectionOverlayTtlMs={detectionOverlayTtlMs}
          label="RGB"
          stream={selectedSession.videoStreams.rgb}
        />
      );
    case "sensor":
      return (
        <SensorPanel
          className="h-full min-h-0"
          isRefreshing={isSensorSnapshotRefreshing}
          onRefresh={onRefreshSensorSnapshot}
          sensor={latestSensor}
          sourceLabel={latestSensorSourceLabel}
        />
      );
    case "thermal":
      return (
        <VideoPane
          className="h-full min-h-0"
          detectionOverlay={selectedSession.detectionOverlays?.thermal}
          detectionOverlayMaxCount={detectionOverlaySettings.maxDetections}
          detectionOverlayTtlMs={detectionOverlayTtlMs}
          label="Thermal"
          stream={selectedSession.videoStreams.thermal}
          thermal
        />
      );
    default:
      return null;
  }
}
