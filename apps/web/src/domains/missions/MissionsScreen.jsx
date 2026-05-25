import { useEffect, useMemo, useRef, useState } from "react";
import L from "leaflet";
import { Circle, MapContainer, Marker, TileLayer, useMap } from "react-leaflet";
import {
  formatDateTime,
  formatElapsedTime,
  formatNumber,
  getTelemetryPositionState,
  makeLiveStatusLabel,
  makeStatusLabel,
  missionTypeLabel
} from "../../utils/formatters.js";
import {
  createEmptyLiveSession,
  formatMediaChannelCount,
  formatStreamingSubscriberCount
} from "../live/liveHelpers.js";
import {
  formatMissionRobotCount,
  getMissionRobotDetails
} from "./missionHelpers.js";
import { createRecordingPlaybackFile } from "../recordings/recordingHelpers.js";
import "leaflet/dist/leaflet.css";

const closedMissionStatuses = new Set(["completed", "ended", "cancelled"]);

export default function MissionsScreen({
  controlMission,
  latestRecording,
  latestSensor,
  latestTelemetry,
  liveEvents,
  liveSessions,
  missionTargets,
  missions,
  onBackToMissionList,
  onEndMission,
  onOpenCreateMissionModal,
  onOpenMissionControl,
  onOpenRecordings,
  onPlayLatestRecording,
  onReconnectSelectedMissionTarget,
  onSelectMission,
  onStartMission,
  operationStatuses,
  playbackRecording,
  robots,
  selectedMission,
  selectedMissionTargetKey,
  setSelectedMissionTargetKey
}) {
  const orderedMissions = useMemo(() => {
    const statusOrder = { active: 0, ready: 1, completed: 2, ended: 2, cancelled: 3 };
    return [...missions].sort((left, right) => {
      const leftOrder = statusOrder[left.status] ?? 9;
      const rightOrder = statusOrder[right.status] ?? 9;
      if (leftOrder !== rightOrder) {
        return leftOrder - rightOrder;
      }
      return (right.startedAt ?? right.createdAt ?? "").localeCompare(left.startedAt ?? left.createdAt ?? "");
    });
  }, [missions]);
  const activeMissionCount = missions.filter((mission) => mission.status === "active").length;

  if (controlMission) {
    return (
      <MissionControlView
        latestRecording={latestRecording}
        latestSensor={latestSensor}
        latestTelemetry={latestTelemetry}
        liveEvents={liveEvents}
        liveSessions={liveSessions}
        mission={controlMission}
        missionTargets={missionTargets}
        onBackToMissionList={onBackToMissionList}
        onEndMission={onEndMission}
        onOpenRecordings={onOpenRecordings}
        onPlayLatestRecording={onPlayLatestRecording}
        onReconnectSelectedMissionTarget={onReconnectSelectedMissionTarget}
        onStartMission={onStartMission}
        operationStatuses={operationStatuses}
        playbackRecording={playbackRecording}
        selectedMissionTargetKey={selectedMissionTargetKey}
        setSelectedMissionTargetKey={setSelectedMissionTargetKey}
      />
    );
  }

  return (
    <section className="mission-management-layout">
      <article className="surface mission-list-surface">
        <div className="section-heading">
          <div>
            <h2>진행 임무</h2>
            <span>진행 {activeMissionCount}건 / 전체 {missions.length}건</span>
          </div>
          <button className="primary-button compact-button" type="button" onClick={onOpenCreateMissionModal}>임무 생성</button>
        </div>
        <div className="list-block">
          {missions.length === 0 ? (
            <p className="empty-state">생성된 임무가 없습니다.</p>
          ) : (
            orderedMissions.map((mission) => {
              const isSelectedMission = selectedMission?.missionCode === mission.missionCode;
              const isClosedMission = closedMissionStatuses.has(mission.status);
              const robotDetails = getMissionRobotDetails(mission, robots);
              const missionRowClassName = [
                "row-item",
                "mission-row",
                isSelectedMission ? "active" : "",
                isClosedMission ? "closed" : ""
              ].filter(Boolean).join(" ");
              return (
                <div
                  className={missionRowClassName}
                  key={mission.missionCode}
                >
                  <button
                    aria-label={`${mission.name} ${mission.missionCode} 선택`}
                    aria-pressed={isSelectedMission}
                    className="mission-row-select"
                    type="button"
                    onClick={() => onSelectMission(mission.missionCode)}
                  >
                    <strong>{mission.name}</strong>
                    <span>
                      {mission.missionCode} / {missionTypeLabel(mission.missionType)} / {makeStatusLabel(mission.status)} / {formatMissionRobotCount(robotDetails)}
                    </span>
                    {robotDetails.length > 0 ? (
                      <span className="mission-row-robots">
                        {robotDetails.map((robot) => `${robot.robotCode} ${makeStatusLabel(robot.status)}`).join(" / ")}
                      </span>
                    ) : null}
                  </button>
                </div>
              );
            })
          )}
        </div>
      </article>

      <article className="surface">
        <div className="section-heading">
          <h2>임무 상세</h2>
          <span>{selectedMission?.missionCode ?? "선택 없음"}</span>
        </div>
        {!selectedMission ? (
          <p className="empty-state">임무를 선택하세요.</p>
        ) : (
          <MissionDetailPanel
            mission={selectedMission}
            onEndMission={onEndMission}
            onOpenMissionControl={onOpenMissionControl}
            onStartMission={onStartMission}
            robotDetails={getMissionRobotDetails(selectedMission, robots)}
          />
        )}
      </article>
    </section>
  );
}

function MissionDetailPanel({ mission, onEndMission, onOpenMissionControl, onStartMission, robotDetails }) {
  return (
    <div className="mission-detail-panel">
      <div>
        <strong>{mission.name}</strong>
        <span>{mission.missionCode}</span>
      </div>
      <div className="mission-detail-meta">
        <span>시나리오 {missionTypeLabel(mission.missionType)}</span>
        <span>상태 {makeStatusLabel(mission.status)}</span>
        <span>배정 로봇 {formatMissionRobotCount(robotDetails)}</span>
        <span>현장 메모 {mission.siteNote || "-"}</span>
      </div>
      <div className="mission-detail-robots">
        {robotDetails.length === 0 ? (
          <span className="mission-robot-chip muted">미배정</span>
        ) : (
          robotDetails.map((robot) => (
            <span className="mission-robot-chip" key={robot.robotCode}>
              {robot.robotCode} · {makeStatusLabel(robot.status)}
            </span>
          ))
        )}
      </div>
      <div className="button-row mission-detail-actions">
        <button className="primary-button" type="button" onClick={() => onOpenMissionControl(mission)}>관제 진입</button>
        <button type="button" disabled={mission.status !== "ready"} onClick={() => onStartMission(mission.missionCode)}>시작</button>
        <button type="button" disabled={mission.status !== "active"} onClick={() => onEndMission(mission.missionCode)}>종료</button>
      </div>
    </div>
  );
}

function MissionControlView({
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
    return ["connected", "completed"].includes(session.status);
  }).length;
  const canReconnectSelectedRobot = mission.status === "active"
    && Boolean(selectedTarget)
    && selectedSession.events.length > 0
    && ["closed", "disconnected", "failed", "signaling closed", "signaling error"].includes(selectedSession.status);

  return (
    <section className="mission-control-layout">
      <article className="surface mission-control-surface">
        <div className="section-heading mission-control-heading">
          <div>
            <h2>{mission.name}</h2>
            <span>{mission.missionCode} / {missionTypeLabel(mission.missionType)} / {makeStatusLabel(mission.status)} / {missionTargets.length}대</span>
          </div>
          <div className="button-row mission-actions">
            <button type="button" onClick={onBackToMissionList}>임무 목록</button>
            <button type="button" disabled={mission.status !== "ready"} onClick={() => onStartMission(mission.missionCode)}>시작</button>
            <button type="button" disabled={mission.status !== "active"} onClick={() => onEndMission(mission.missionCode)}>종료</button>
          </div>
        </div>

        <div className="mission-command-bar">
          <div>
            <strong>{mission.missionCode}</strong>
            <span>{selectedTarget ? `선택 ${selectedTarget.robotCode} ${makeLiveStatusLabel(selectedSession.status)} / 연결 ${connectedCount}/${missionTargets.length}대` : "임무에 배정된 로봇이 없습니다."}</span>
          </div>
          <div className="mission-command-controls">
            <label className="mission-robot-select-label">
              <span>관제 로봇</span>
              <select
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
              <div className="button-row control-actions">
                <button type="button" onClick={onReconnectSelectedMissionTarget}>재연결</button>
              </div>
            ) : null}
          </div>
        </div>

        <div className="mission-robot-selector">
          {missionTargets.length === 0 ? (
            <p className="empty-state">임무에 배정된 로봇이 없습니다.</p>
          ) : (
            missionTargets.map((target) => {
              const session = liveSessions[target.key] ?? createEmptyLiveSession();
              return (
                <button
                  className={selectedTarget?.key === target.key ? "mission-robot-button active" : "mission-robot-button"}
                  key={target.key}
                  type="button"
                  onClick={() => setSelectedMissionTargetKey(target.key)}
                >
                  <div>
                    <strong>{target.robot?.displayName ?? target.robotCode}</strong>
                    <span>{target.robotCode} / {makeStatusLabel(target.robot?.status ?? "offline")}</span>
                    <span>{formatMediaChannelCount(target.streamingStatus)} / {formatStreamingSubscriberCount(target.streamingStatus)}</span>
                  </div>
                  <small>{makeLiveStatusLabel(session.status)}</small>
                </button>
              );
            })
          )}
        </div>

        {!selectedTarget ? (
          <p className="empty-state">관제할 로봇을 선택할 수 없습니다.</p>
        ) : (
          <div className="control-quadrants">
            <VideoPane className="control-quadrant" label="RGB" stream={selectedSession.videoStreams.rgb} />
            <VideoPane className="control-quadrant" label="Thermal" stream={selectedSession.videoStreams.thermal} thermal />
            <RobotMap className="control-quadrant" telemetry={latestTelemetry} />
            <SensorPanel className="control-quadrant" sensor={latestSensor} />
          </div>
        )}
        <AudioSink stream={selectedSession.videoStreams.audio} />
      </article>

      <aside className="right-rail mission-control-rail">
        <ConnectionStatusPanel statuses={operationStatuses} />
        <LatestRecordingPanel
          onOpenRecordings={onOpenRecordings}
          onPlayRecording={onPlayLatestRecording}
          playbackRecording={playbackRecording}
          recording={latestRecording}
        />
        <article className="surface event-panel">
          <div className="section-heading">
            <h2>이벤트</h2>
            <span>{liveEvents.length}건</span>
          </div>
          <div className="event-list">
            {liveEvents.length === 0 ? (
              <p className="empty-state">관제 연결 이벤트가 없습니다.</p>
            ) : (
              liveEvents.map((event) => (
                <div className="event-row" key={event.id}>
                  <span>{formatDateTime(event.at)}</span>
                  <strong>{event.message}</strong>
                </div>
              ))
            )}
          </div>
        </article>
      </aside>
    </section>
  );
}

function ConnectionStatusPanel({ compact = false, statuses }) {
  return (
    <article className={compact ? "status-surface embedded" : "surface status-surface"}>
      <div className="section-heading">
        <h2>연결 상태</h2>
        <span>{statuses.filter((status) => status.tone === "ok").length}/{statuses.length}</span>
      </div>
      <div className="status-grid">
        {statuses.map((status) => (
          <div className={`status-cell ${status.tone}`} key={status.label}>
            <span>{status.label}</span>
            <strong>{status.value}</strong>
            <small>{status.detail}</small>
          </div>
        ))}
      </div>
    </article>
  );
}

function LatestRecordingPanel({ recording, compact = false, onOpenRecordings, onPlayRecording, playbackRecording }) {
  const playableFile = createRecordingPlaybackFile(playbackRecording ?? recording);
  return (
    <article className={compact ? "latest-recording compact" : "surface latest-recording"}>
      <div className="section-heading">
        <h2>녹화 저장</h2>
        <span>{playableFile ? "재생 가능" : recording ? makeStatusLabel(recording.status) : "대기"}</span>
      </div>
      {!recording ? (
        <p className="empty-state">진행 중인 임무가 시작되면 저장 청크가 생성됩니다.</p>
      ) : (
        <div className="latest-recording-body">
          <div>
            <strong>{recording.missionCode} / #{recording.chunkIndex}</strong>
            <span>{formatDateTime(recording.startedAt)} - {formatDateTime(recording.endedAt)}</span>
          </div>
          <div className="latest-recording-actions">
            {playableFile ? (
              <button className="object-link" type="button" onClick={onPlayRecording}>
                재생
              </button>
            ) : null}
            <button className="object-link secondary" type="button" onClick={onOpenRecordings}>
              녹화 목록
            </button>
          </div>
        </div>
      )}
    </article>
  );
}

function SensorPanel({ className = "", sensor }) {
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

function VideoPane({ className = "", compact = false, label, stream, thermal = false }) {
  const videoRef = useRef(null);
  useEffect(() => {
    const video = videoRef.current;
    if (!video) {
      return undefined;
    }
    video.srcObject = stream;
    if (stream) {
      void video.play().catch(() => {});
    }
    return () => {
      video.srcObject = null;
    };
  }, [stream]);
  const videoPaneClassName = [
    "video-pane",
    thermal ? "thermal" : "",
    compact ? "compact" : "",
    className
  ].filter(Boolean).join(" ");
  return (
    <div className={videoPaneClassName}>
      <video ref={videoRef} autoPlay playsInline muted={label !== "Audio"} />
      {!stream ? <span>{label} 대기</span> : null}
      <strong>{label}</strong>
    </div>
  );
}

function AudioSink({ stream }) {
  const audioRef = useRef(null);
  useEffect(() => {
    const audio = audioRef.current;
    if (!audio) {
      return undefined;
    }
    audio.srcObject = stream;
    if (stream) {
      void audio.play().catch(() => {});
    }
    return () => {
      audio.srcObject = null;
    };
  }, [stream]);
  return <audio ref={audioRef} autoPlay />;
}

function MapRecenter({ position }) {
  const map = useMap();

  useEffect(() => {
    map.setView(position, map.getZoom(), { animate: true });
  }, [map, position]);

  return null;
}

function RobotMap({ className = "", telemetry }) {
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

function MetricTile({ label, value, unit }) {
  return (
    <div className="metric-tile">
      <span>{label}</span>
      <strong>{value}</strong>
      <small>{unit}</small>
    </div>
  );
}
