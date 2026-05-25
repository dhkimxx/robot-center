import SectionHeader from "../../../components/ui/SectionHeader.jsx";
import Surface from "../../../components/ui/Surface.jsx";
import {
  formatDateTime,
  formatNumber
} from "../../../utils/formatters.js";
import { MetricTile } from "./MetricTile.jsx";

export function SensorPanel({ className = "", sensor }) {
  return (
    <Surface className={["grid min-h-0 grid-rows-[auto_minmax(0,1fr)] gap-3 overflow-hidden p-3", className].filter(Boolean).join(" ")}>
      <SectionHeader className="mb-0" title="센서" meta={sensor ? formatDateTime(sensor.sentAt ?? sensor.receivedAt) : "대기"} />
      <div className="grid min-h-0 grid-cols-2 gap-2">
        <MetricTile label="CO" value={formatNumber(sensor?.payload?.coPpm ?? sensor?.coPpm)} unit="ppm" />
        <MetricTile label="O2" value={formatNumber(sensor?.payload?.oxygenPercent ?? sensor?.oxygenPercent, 2)} unit="%" />
        <MetricTile label="온도" value={formatNumber(sensor?.payload?.temperatureCelsius ?? sensor?.temperatureCelsius)} unit="C" />
        <MetricTile label="습도" value={formatNumber(sensor?.payload?.humidityPercent ?? sensor?.humidityPercent)} unit="%" />
      </div>
    </Surface>
  );
}
