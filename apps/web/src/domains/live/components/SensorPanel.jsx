import {
  formatDateTime,
  formatNumber
} from "../../../utils/formatters.js";
import { MetricTile } from "./MetricTile.jsx";

export function SensorPanel({ className = "", sensor }) {
  const panelClassName = ["surface", "sensor-surface", className].filter(Boolean).join(" ");
  return (
    <article className={panelClassName}>
      <div className="section-heading">
        <h2>센서</h2>
        <span>{sensor ? formatDateTime(sensor.sentAt ?? sensor.receivedAt) : "대기"}</span>
      </div>
      <div className="metric-grid">
        <MetricTile label="CO" value={formatNumber(sensor?.payload?.coPpm ?? sensor?.coPpm)} unit="ppm" />
        <MetricTile label="O2" value={formatNumber(sensor?.payload?.oxygenPercent ?? sensor?.oxygenPercent, 2)} unit="%" />
        <MetricTile label="온도" value={formatNumber(sensor?.payload?.temperatureCelsius ?? sensor?.temperatureCelsius)} unit="C" />
        <MetricTile label="습도" value={formatNumber(sensor?.payload?.humidityPercent ?? sensor?.humidityPercent)} unit="%" />
      </div>
    </article>
  );
}
