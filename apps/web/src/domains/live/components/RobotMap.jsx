import { useEffect, useMemo, useState } from "react";
import L from "leaflet";
import { Circle, MapContainer, Marker, TileLayer, useMap } from "react-leaflet";
import {
  formatElapsedTime,
  formatNumber,
  getTelemetryPositionState
} from "../../../utils/formatters.js";
import "leaflet/dist/leaflet.css";

function MapRecenter({ position }) {
  const map = useMap();

  useEffect(() => {
    map.setView(position, map.getZoom(), { animate: true });
  }, [map, position]);

  return null;
}

export function RobotMap({ className = "", telemetry }) {
  const [now, setNow] = useState(Date.now());

  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, []);

  const positionState = getTelemetryPositionState(telemetry, now);
  const mapPosition = useMemo(() => {
    if (!positionState.hasPosition) {
      return null;
    }
    return [Number(positionState.latitude), Number(positionState.longitude)];
  }, [positionState.hasPosition, positionState.latitude, positionState.longitude]);
  const accuracyMeter = Number(positionState.accuracyMeter);
  const accuracyRadius = Number.isFinite(accuracyMeter) ? Math.max(8, accuracyMeter) : 20;
  const markerIcon = useMemo(() => L.divIcon({
    className: positionState.isFresh ? "robot-location-marker" : "robot-location-marker stale",
    html: `<span class="robot-location-dot"></span><span class="robot-location-label">${positionState.statusLabel}</span>`,
    iconAnchor: [8, 8]
  }), [positionState.isFresh, positionState.statusLabel]);

  const panelClassName = ["surface", "map-surface", className].filter(Boolean).join(" ");

  return (
    <article className={panelClassName}>
      <div className="section-heading">
        <h2>위치</h2>
        <span>{positionState.statusLabel}</span>
      </div>
      <div className={mapPosition ? "map-canvas has-position" : "map-canvas empty"}>
        {mapPosition ? (
          <MapContainer
            center={mapPosition}
            className="robot-location-map"
            zoom={17}
            zoomControl={false}
            attributionControl={false}
            scrollWheelZoom
          >
            <TileLayer
              url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
              maxZoom={19}
            />
            <Circle
              center={mapPosition}
              pathOptions={{ color: "#2f6fdb", fillColor: "#2f6fdb", fillOpacity: 0.16, opacity: 0.42, weight: 1 }}
              radius={accuracyRadius}
            />
            <Marker icon={markerIcon} position={mapPosition} />
            <MapRecenter position={mapPosition} />
          </MapContainer>
        ) : (
          <span className="map-empty">GPS 대기</span>
        )}
      </div>
      <div className="coordinate-row">
        <span>Lat {positionState.hasPosition ? formatNumber(positionState.latitude, 6) : "-"}</span>
        <span>Lng {positionState.hasPosition ? formatNumber(positionState.longitude, 6) : "-"}</span>
      </div>
      <div className="position-meta-row">
        <span>{positionState.hasPosition ? `수신 ${formatElapsedTime(positionState.timestamp, now)}` : "위치 미수신"}</span>
        <span>{positionState.accuracyMeter ? `오차 ${formatNumber(positionState.accuracyMeter)}m` : "오차 -"}</span>
      </div>
    </article>
  );
}
