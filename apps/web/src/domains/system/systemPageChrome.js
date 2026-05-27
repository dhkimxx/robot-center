import { countStreamingRobotsFromLiveStatuses } from "../missions/missionHelpers.js";

export function createSystemPageChrome({ liveStatuses, systemStatus }) {
  const roomCount = systemStatus?.summary?.sfuRooms
    ?? systemStatus?.sfuRooms?.length
    ?? 0;
  const componentCount = systemStatus?.components?.length ?? 0;
  const streamingRobotCount = countStreamingRobotsFromLiveStatuses(liveStatuses);

  return {
    meta: `서비스 ${componentCount}개 · 실시간 연결 ${roomCount}개 · 송출 ${streamingRobotCount}개`,
    title: "시스템"
  };
}
