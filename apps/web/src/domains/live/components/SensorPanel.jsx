import SectionHeader from "../../../components/ui/SectionHeader.jsx";
import Surface from "../../../components/ui/Surface.jsx";
import {
  formatDateTime,
  formatNumber
} from "../../../utils/formatters.js";
import { MetricTile } from "./MetricTile.jsx";

function createLegacyMetrics(sensor) {
  return [
    { label: "CO", value: sensor?.payload?.coPpm ?? sensor?.coPpm, unit: "ppm" },
    { label: "O2", value: sensor?.payload?.oxygenPercent ?? sensor?.oxygenPercent, unit: "%" },
    { label: "온도", value: sensor?.payload?.temperatureCelsius ?? sensor?.temperatureCelsius, unit: "C" },
    { label: "습도", value: sensor?.payload?.humidityPercent ?? sensor?.humidityPercent, unit: "%" }
  ];
}

function createSensorMetrics(sensor) {
  if (Array.isArray(sensor?.sensors) && sensor.sensors.length > 0) {
    return sensor.sensors;
  }
  return createLegacyMetrics(sensor);
}

function formatMetricValue(value) {
  if (typeof value === "number") {
    return formatNumber(value);
  }
  if (value === null || value === undefined || value === "") {
    return "-";
  }
  return String(value);
}

export function SensorPanel({ className = "", sensor }) {
  const metrics = createSensorMetrics(sensor);
  return (
    <Surface className={["grid min-h-0 grid-rows-[auto_minmax(0,1fr)] gap-3 overflow-hidden p-3", className].filter(Boolean).join(" ")}>
      <SectionHeader className="mb-0" title="센서" meta={sensor ? formatDateTime(sensor.sentAt ?? sensor.receivedAt) : "대기"} />
      <div className="grid min-h-0 grid-cols-2 gap-2">
        {metrics.map((metric) => (
          <MetricTile
            key={`${metric.label}-${metric.unit}`}
            label={metric.label}
            value={formatMetricValue(metric.value)}
            unit={metric.unit}
          />
        ))}
      </div>
    </Surface>
  );
}
