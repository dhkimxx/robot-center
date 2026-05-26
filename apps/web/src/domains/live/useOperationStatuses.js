import {
  formatDateTime,
  formatElapsedTime,
  makeLiveStatusLabel,
  makeStatusLabel,
  makeStatusTone
} from "../../utils/formatters.js";
import {
  formatMediaChannelCount,
  formatStreamingSubscriberCount
} from "./liveHelpers.js";

export function useOperationStatuses({
  activeLiveStream,
  latestPositionState,
  primaryRobot,
  selectedLiveSession,
  selectedLiveTarget,
  statusError
}) {
  return [
    {
      label: "관제 서비스",
      value: statusError ? "대기" : "정상",
      detail: statusError || "정상 응답",
      tone: statusError ? "danger" : "ok"
    },
    {
      label: "로봇",
      value: selectedLiveTarget?.robot ? makeStatusLabel(selectedLiveTarget.robot.status) : primaryRobot ? makeStatusLabel(primaryRobot.status) : "미등록",
      detail: selectedLiveTarget?.robot ? `${selectedLiveTarget.robot.robotCode} / 최근 ${formatDateTime(selectedLiveTarget.robot.lastSeenAt)}` : primaryRobot ? `${primaryRobot.robotCode} / 최근 ${formatDateTime(primaryRobot.lastSeenAt)}` : "등록 필요",
      tone: makeStatusTone(selectedLiveTarget?.robot?.status ?? primaryRobot?.status)
    },
    {
      label: "실시간 링크",
      value: makeLiveStatusLabel(selectedLiveSession.status),
      detail: activeLiveStream ? `${formatMediaChannelCount(activeLiveStream)} / ${formatStreamingSubscriberCount(activeLiveStream)}` : "송출 대기",
      tone: makeStatusTone(selectedLiveSession.status)
    },
    {
      label: "위치",
      value: latestPositionState.statusLabel,
      detail: latestPositionState.hasPosition ? `수신 ${formatElapsedTime(latestPositionState.timestamp)}` : "GPS 대기",
      tone: makeStatusTone(latestPositionState.statusLabel)
    }
  ];
}
