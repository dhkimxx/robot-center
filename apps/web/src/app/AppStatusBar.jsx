import { useEffect, useState } from "react";
import { componentLabels } from "../config/controlCenterConfig.js";
import { countStreamingRobotsFromLiveStatuses } from "../domains/missions/missionHelpers.js";

function formatClock(date) {
  return new Intl.DateTimeFormat("ko-KR", {
    hour: "2-digit",
    hour12: false,
    minute: "2-digit",
    second: "2-digit"
  }).format(date);
}

function makeComponentLabel(component) {
  const label = componentLabels[component.name] ?? component.name;
  if (component.status === "ok" || component.status === "ready" || component.status === "configured") {
    return `${label} 정상`;
  }
  return `${label} ${component.status}`;
}

export default function AppStatusBar({ liveStatuses, robots, statusError, systemStatus }) {
  const [clockText, setClockText] = useState(() => formatClock(new Date()));
  const components = systemStatus?.components ?? [];
  const rooms = systemStatus?.sfuRooms ?? [];
  const onlineRobotCount = robots.filter((robot) => ["online", "streaming"].includes(robot.status)).length;
  const streamingRobotCount = countStreamingRobotsFromLiveStatuses(liveStatuses);
  const componentSummary = components.slice(0, 2).map(makeComponentLabel).join(" · ");
  const leftSummary = statusError ? `서버 응답 대기 · ${statusError}` : componentSummary || "서비스 상태 확인 중";
  const rightSummary = `실시간 연결 ${rooms.length}개 · 로봇 ${onlineRobotCount}대 online · 송출 ${streamingRobotCount}개`;

  useEffect(() => {
    const timer = window.setInterval(() => {
      setClockText(formatClock(new Date()));
    }, 1000);
    return () => window.clearInterval(timer);
  }, []);

  return (
    <footer className="flex min-h-8 min-w-0 items-center justify-between gap-4 border-t border-slate-800/80 bg-command-950/95 px-4 text-[11px] font-semibold text-slate-500">
      <span className="min-w-0 truncate">{leftSummary}</span>
      <span className="hidden shrink-0 items-center gap-3 text-right md:flex">
        <time className="font-bold tabular-nums text-slate-300" dateTime={clockText}>{clockText}</time>
        <span className="truncate text-slate-400">{rightSummary}</span>
      </span>
    </footer>
  );
}
