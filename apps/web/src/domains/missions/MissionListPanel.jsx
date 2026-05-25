import { useMemo } from "react";
import {
  makeStatusLabel,
  missionTypeLabel
} from "../../utils/formatters.js";
import {
  formatMissionRobotCount,
  getMissionRobotDetails
} from "./missionHelpers.js";

const closedMissionStatuses = new Set(["completed", "ended", "cancelled"]);

export function MissionListPanel({
  missions,
  onOpenCreateMissionModal,
  onSelectMission,
  robots,
  selectedMission
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

  return (
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
              <div className={missionRowClassName} key={mission.missionCode}>
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
  );
}
