import Button from "../../components/ui/Button.jsx";
import { countStreamingRobotsFromLiveStatuses } from "../missions/missionHelpers.js";

export function createRobotPageChrome({
  isLoading = false,
  liveStatuses,
  onOpenCreateRobotModal,
  robots
}) {
  const onlineRobotCount = robots.filter((robot) => ["online", "streaming"].includes(robot.status)).length;
  const streamingRobotCount = countStreamingRobotsFromLiveStatuses(liveStatuses);

  return {
    action: (
      <Button size="sm" variant="primary" onClick={onOpenCreateRobotModal}>
        로봇 등록
      </Button>
    ),
    meta: isLoading ? "로봇 정보를 불러오는 중" : `등록 ${robots.length}대 · online ${onlineRobotCount}대 · 송출 ${streamingRobotCount}개`,
    title: "로봇"
  };
}
