import { getMissionRobotCodes } from "../missions/missionHelpers.js";
import { formatDateTime, makeStatusLabel } from "../../utils/formatters.js";

export default function RobotsScreen({
  missions,
  onArchiveRobot,
  onLoadConnectionInfo,
  onOpenCreateRobotModal,
  onOpenEditRobotModal,
  onSelectRobot,
  robots,
  selectedRobot
}) {
  const selectedRobotHasOpenMission = selectedRobot
    ? missions.some((mission) => getMissionRobotCodes(mission).includes(selectedRobot.robotCode) && ["ready", "active"].includes(mission.status))
    : false;

  return (
    <section className="robot-management-layout">
      <article className="surface robot-list-surface">
        <div className="section-heading">
          <div>
            <h2>등록 로봇</h2>
            <span>{robots.length}대</span>
          </div>
          <button className="primary-button compact-button" type="button" onClick={onOpenCreateRobotModal}>로봇 등록</button>
        </div>
        <div className="list-block">
          {robots.length === 0 ? (
            <p className="empty-state">등록된 로봇이 없습니다.</p>
          ) : (
            robots.map((robot) => {
              const isSelectedRobot = selectedRobot?.robotCode === robot.robotCode;
              return (
                <div
                  className={isSelectedRobot ? "row-item robot-row active" : "row-item robot-row"}
                  key={robot.robotCode}
                >
                  <button
                    aria-label={`${robot.displayName} ${robot.robotCode} 선택`}
                    aria-pressed={isSelectedRobot}
                    className="robot-row-select"
                    type="button"
                    onClick={() => onSelectRobot(robot.robotCode)}
                  >
                    <strong>{robot.displayName}</strong>
                    <span>{robot.robotCode} / {makeStatusLabel(robot.status)} / 최근 {formatDateTime(robot.lastSeenAt)}</span>
                  </button>
                </div>
              );
            })
          )}
        </div>
      </article>

      <section className="robot-detail-stack">
        <article className="surface">
          <div className="section-heading">
            <h2>로봇 상세</h2>
            <span>{selectedRobot?.robotCode ?? "선택 없음"}</span>
          </div>
          {!selectedRobot ? (
            <p className="empty-state">로봇을 선택하세요.</p>
          ) : (
            <div className="robot-detail-panel">
              <div>
                <strong>{selectedRobot.displayName}</strong>
                <span>{selectedRobot.modelName || "모델 미지정"}</span>
              </div>
              <div className="robot-detail-meta">
                <span>상태 {makeStatusLabel(selectedRobot.status)}</span>
                <span>최근 연결 {formatDateTime(selectedRobot.lastSeenAt)}</span>
                <span>최근 송출 {formatDateTime(selectedRobot.lastStreamingAt)}</span>
                {selectedRobotHasOpenMission ? <span>삭제 불가: 진행/대기 임무 배정</span> : null}
              </div>
              <div className="button-row robot-detail-actions">
                <button className="primary-button" type="button" onClick={onOpenEditRobotModal}>수정</button>
                <button type="button" onClick={() => onLoadConnectionInfo(selectedRobot.robotCode)}>연결 정보</button>
                <button
                  className="danger-button"
                  disabled={selectedRobotHasOpenMission}
                  type="button"
                  onClick={() => onArchiveRobot(selectedRobot.robotCode)}
                >
                  삭제
                </button>
              </div>
            </div>
          )}
        </article>
      </section>
    </section>
  );
}
