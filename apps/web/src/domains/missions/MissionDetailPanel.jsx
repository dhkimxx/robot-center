import {
  makeStatusLabel,
  missionTypeLabel
} from "../../utils/formatters.js";
import { formatMissionRobotCount } from "./missionHelpers.js";

export function MissionDetailPanel({ mission, onEndMission, onOpenMissionControl, onStartMission, robotDetails }) {
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
