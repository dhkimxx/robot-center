import {
  makeLiveStatusLabel,
  makeStatusLabel
} from "../../../utils/formatters.js";
import {
  createEmptyLiveSession,
  formatMediaChannelCount,
  formatStreamingSubscriberCount
} from "../liveHelpers.js";

export function MissionRobotSelector({
  liveSessions,
  missionTargets,
  selectedTarget,
  setSelectedMissionTargetKey
}) {
  return (
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
  );
}
