import { RiRefreshLine } from "react-icons/ri";
import Button from "../../../components/ui/Button.jsx";
import SectionHeader from "../../../components/ui/SectionHeader.jsx";
import Surface from "../../../components/ui/Surface.jsx";
import {
  formatDateTime,
  formatNumber
} from "../../../utils/formatters.js";
import { createSensorMetrics } from "../sensorDisplayMetrics.js";
import { MetricTile } from "./MetricTile.jsx";

function formatMetricValue(value) {
  if (typeof value === "number") {
    return formatNumber(value);
  }
  if (value === null || value === undefined || value === "") {
    return "-";
  }
  return String(value);
}

export function SensorPanel({ className = "", isRefreshing = false, onRefresh, sensor, sourceLabel = "" }) {
  const metrics = createSensorMetrics(sensor);
  const valueMeta = sensor
    ? `${metrics.length}개 · ${formatDateTime(sensor.receivedAt)}`
    : "대기";
  const meta = sourceLabel ? `${sourceLabel} · ${valueMeta}` : valueMeta;
  const refreshAction = onRefresh ? (
    <Button
      aria-label="센서 최신 저장값 갱신"
      disabled={isRefreshing}
      onClick={onRefresh}
      size="icon"
      title="센서 최신 저장값 갱신"
      variant="ghost"
    >
      <RiRefreshLine aria-hidden className={isRefreshing ? "h-4 w-4 animate-spin" : "h-4 w-4"} />
    </Button>
  ) : null;

  return (
    <Surface className={["grid min-h-0 grid-rows-[auto_minmax(0,1fr)] gap-3 overflow-hidden p-3", className].filter(Boolean).join(" ")}>
      <SectionHeader action={refreshAction} className="mb-0" title="센서" meta={meta} />
      {metrics.length === 0 ? (
        <div className="grid min-h-0 place-items-center rounded-lg border border-slate-500/20 bg-white/[0.035] text-sm font-bold text-slate-500">
          수신된 센서값 없음
        </div>
      ) : (
        <div className="grid min-h-0 auto-rows-min grid-cols-[repeat(auto-fit,minmax(128px,1fr))] gap-2 overflow-y-auto pr-1">
          {metrics.map((metric) => (
            <MetricTile
              compact={metrics.length > 4}
              key={metric.key ?? `${metric.label}-${metric.unit}`}
              label={metric.label}
              alarmLevel={metric.alarmLevel}
              value={formatMetricValue(metric.value)}
              unit={metric.unit}
            />
          ))}
        </div>
      )}
    </Surface>
  );
}
