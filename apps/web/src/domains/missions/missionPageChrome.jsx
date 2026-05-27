import { NavLink } from "react-router-dom";
import Button from "../../components/ui/Button.jsx";
import { makeStatusLabel, missionTypeLabel } from "../../utils/formatters.js";
import { countStreamingRobotsFromLiveStatuses } from "./missionHelpers.js";

export function createMissionListPageChrome({
  liveStatuses,
  missions,
  onOpenCreateMissionModal
}) {
  const activeMissionCount = missions.filter((mission) => mission.status === "active").length;
  const readyMissionCount = missions.filter((mission) => mission.status === "ready").length;
  const closedMissionCount = missions.filter((mission) => ["ended", "completed", "cancelled"].includes(mission.status)).length;
  const streamingRobotCount = countStreamingRobotsFromLiveStatuses(liveStatuses);

  return {
    action: (
      <Button size="sm" variant="primary" onClick={onOpenCreateMissionModal}>
        임무 생성
      </Button>
    ),
    meta: `진행 ${activeMissionCount}건 · 대기 ${readyMissionCount}건 · 종료 ${closedMissionCount}건 · 송출 로봇 ${streamingRobotCount}개`,
    title: "임무"
  };
}

export function createMissionControlPageChrome({
  controlMission,
  missionTargets,
  onBackToMissionList,
  onEndMission,
  onStartMission,
  routeMissionControlCode
}) {
  const controlStatus = controlMission ? makeStatusLabel(controlMission.status) : "확인 중";
  const controlType = controlMission ? missionTypeLabel(controlMission.missionType) : "임무";

  return {
    action: (
      <>
        <Button
          as={NavLink}
          reloadDocument
          size="sm"
          to={`/missions?selected=${encodeURIComponent(routeMissionControlCode)}`}
          onClick={() => onBackToMissionList({ navigate: false })}
        >
          임무 목록
        </Button>
        <Button
          size="sm"
          disabled={controlMission?.status !== "ready"}
          onClick={() => onStartMission(routeMissionControlCode)}
        >
          시작
        </Button>
        <Button
          size="sm"
          disabled={controlMission?.status !== "active"}
          onClick={() => onEndMission(routeMissionControlCode)}
        >
          종료
        </Button>
      </>
    ),
    meta: `${controlMission?.missionCode ?? routeMissionControlCode} · ${controlType} · ${controlStatus} · 로봇 ${missionTargets.length}대`,
    title: controlMission?.name ?? "실시간 관제"
  };
}

export function createMissionReplayPageChrome({
  replayMission,
  routeMissionReplayCode
}) {
  return {
    meta: `${replayMission?.missionCode ?? routeMissionReplayCode} · 녹화 리플레이`,
    title: replayMission?.name ?? "종료 임무 리플레이"
  };
}
