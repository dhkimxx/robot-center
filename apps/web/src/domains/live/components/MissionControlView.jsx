import {
  makeLiveStatusLabel,
  makeStatusLabel,
  missionTypeLabel
} from "../../../utils/formatters.js";
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

        <MissionRobotSelector
          liveSessions={liveSessions}
          missionTargets={missionTargets}
          selectedTarget={selectedTarget}
          setSelectedMissionTargetKey={setSelectedMissionTargetKey}
        />

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
        <EventPanel liveEvents={liveEvents} />
      </aside>
    </section>
  );
}
