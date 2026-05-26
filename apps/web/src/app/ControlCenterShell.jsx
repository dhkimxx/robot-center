import { useCallback } from "react";
import { RiListCheck3, RiRobot2Line, RiSettings3Line } from "react-icons/ri";
import { NavLink, useLocation, useNavigate } from "react-router-dom";
import NotificationStack from "../components/NotificationStack.jsx";
import Button from "../components/ui/Button.jsx";
import { navigationItems } from "../config/controlCenterConfig.js";
import { MissionModals } from "../domains/missions/MissionModals.jsx";
import { RecordingModals } from "../domains/recordings/RecordingModals.jsx";
import { RobotModals } from "../domains/robots/RobotModals.jsx";
import { useControlCenterController } from "../hooks/useControlCenterController.js";
import { makeStatusLabel, missionTypeLabel } from "../utils/formatters.js";
import { cn } from "../utils/cn.js";
import AppStatusBar from "./AppStatusBar.jsx";
import { ControlCenterRoutes } from "./ControlCenterRoutes.jsx";
import PageContextBar from "./PageContextBar.jsx";
import {
  getActiveSection,
  getRouteMissionControlCode,
  getRouteMissionReplayCode,
  getRouteSelectedMissionCode
} from "./routeUtils.js";

const navigationIcons = {
  missions: RiListCheck3,
  robots: RiRobot2Line,
  system: RiSettings3Line
};

export function ControlCenterApp() {
  const location = useLocation();
  const navigate = useNavigate();
  const navigateToAppPath = useCallback((path) => {
    navigate(path);
  }, [navigate]);
  const routeMissionControlCode = getRouteMissionControlCode(location.pathname);
  const routeMissionReplayCode = getRouteMissionReplayCode(location.pathname);
  const routeSelectedMissionCode = getRouteSelectedMissionCode(location.search);
  const activeSection = getActiveSection(location.pathname);
  const controller = useControlCenterController({
    activeSection,
    missionControlCode: routeMissionControlCode,
    missionReplayCode: routeMissionReplayCode,
    selectedMissionCode: routeSelectedMissionCode,
    navigateToPath: navigateToAppPath
  });
  const appChrome = createAppChrome({
    activeSection,
    controller,
    routeMissionControlCode,
    routeMissionReplayCode
  });

  return (
    <main className="grid h-screen min-h-0 grid-cols-[224px_minmax(0,1fr)] bg-command-950 max-[900px]:grid-cols-1 max-[900px]:overflow-auto">
      <aside className="grid h-screen min-w-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden border-r border-slate-500/10 bg-[#080a0e] px-3 py-3 text-slate-100 max-[900px]:h-auto max-[900px]:grid-rows-none max-[900px]:gap-3 max-[900px]:border-b max-[900px]:border-r-0">
        <NavLink
          aria-label="홈으로 이동"
          className="mb-3 rounded-xl border border-slate-500/10 bg-white/[0.03] px-3 py-2 no-underline transition hover:border-sapphire-400/25 hover:bg-sapphire-500/[0.08] max-[900px]:mb-0"
          reloadDocument
          to="/missions"
        >
          <p className="text-[10px] font-black uppercase tracking-normal text-sapphire-300">SST</p>
          <strong className="mt-0.5 block truncate text-sm font-black text-slate-50">Robot Control</strong>
        </NavLink>

        <nav className="grid content-start gap-1.5 overflow-auto max-[900px]:grid-cols-3">
          {navigationItems.map((item) => (
            <NavigationItem
              isSectionActive={location.pathname.startsWith(`${item.path}/`)}
              item={item}
              key={item.key}
            />
          ))}
        </nav>
      </aside>

      <section className="grid h-screen min-h-0 min-w-0 grid-rows-[52px_minmax(0,1fr)_32px] overflow-hidden max-[900px]:h-auto max-[900px]:min-h-screen max-[900px]:overflow-visible">
        <PageContextBar
          action={appChrome.contextAction}
          meta={appChrome.contextMeta}
          title={appChrome.contextTitle}
        />

        <div className="grid min-h-0 min-w-0 grid-rows-[minmax(0,1fr)] overflow-hidden p-3 max-[900px]:overflow-visible">
          <ControlCenterRoutes
            missionRouteProps={controller.missionRouteProps}
            robotRouteProps={controller.robotRouteProps}
            systemRouteProps={controller.systemRouteProps}
          />
        </div>
        <AppStatusBar
          robots={controller.missionRouteProps.robots}
          statusError={controller.statusError}
          streamingStatuses={controller.missionRouteProps.streamingStatuses}
          systemStatus={controller.systemRouteProps.systemStatus}
        />
      </section>

      <NotificationStack notifications={controller.notifications} onDismiss={controller.dismissNotification} />
      <RobotModals {...controller.robotModalProps} />
      <MissionModals {...controller.missionModalProps} />
      <RecordingModals {...controller.playbackModalProps} />
    </main>
  );
}

function createAppChrome({
  activeSection,
  controller,
  routeMissionControlCode,
  routeMissionReplayCode
}) {
  const missions = controller.missionRouteProps.missions;
  const robots = controller.missionRouteProps.robots;
  const streamingStatuses = controller.missionRouteProps.streamingStatuses;
  const activeMissionCount = missions.filter((mission) => mission.status === "active").length;
  const readyMissionCount = missions.filter((mission) => mission.status === "ready").length;
  const closedMissionCount = missions.filter((mission) => ["ended", "completed", "cancelled"].includes(mission.status)).length;
  const streamingRobotCount = streamingStatuses.filter((status) => status.status === "streaming").length;
  const controlMission = controller.missionRouteProps.controlMission;
  const replayMission = controller.missionRouteProps.replayMission;

  if (routeMissionControlCode) {
    const controlStatus = controlMission ? makeStatusLabel(controlMission.status) : "확인 중";
    const controlType = controlMission ? missionTypeLabel(controlMission.missionType) : "임무";
    return {
      contextAction: (
        <>
          <Button
            as={NavLink}
            reloadDocument
            size="sm"
            to={`/missions?selected=${encodeURIComponent(routeMissionControlCode)}`}
            onClick={() => controller.missionRouteProps.onBackToMissionList({ navigate: false })}
          >
            임무 목록
          </Button>
          <Button
            size="sm"
            disabled={controlMission?.status !== "ready"}
            onClick={() => controller.missionRouteProps.onStartMission(routeMissionControlCode)}
          >
            시작
          </Button>
          <Button
            size="sm"
            disabled={controlMission?.status !== "active"}
            onClick={() => controller.missionRouteProps.onEndMission(routeMissionControlCode)}
          >
            종료
          </Button>
        </>
      ),
      contextMeta: `${controlMission?.missionCode ?? routeMissionControlCode} · ${controlType} · ${controlStatus} · 로봇 ${controller.missionRouteProps.missionTargets.length}대`,
      contextTitle: controlMission?.name ?? "실시간 관제"
    };
  }

  if (routeMissionReplayCode) {
    return {
      contextMeta: `${replayMission?.missionCode ?? routeMissionReplayCode} · 녹화 리플레이`,
      contextTitle: replayMission?.name ?? "종료 임무 리플레이"
    };
  }

  if (activeSection === "robots") {
    const onlineRobotCount = robots.filter((robot) => ["online", "streaming"].includes(robot.status)).length;
    return {
      contextAction: (
        <Button size="sm" variant="primary" onClick={controller.robotRouteProps.onOpenCreateRobotModal}>
          로봇 등록
        </Button>
      ),
      contextMeta: `등록 ${robots.length}대 · online ${onlineRobotCount}대 · 송출 ${streamingRobotCount}개`,
      contextTitle: "로봇"
    };
  }

  if (activeSection === "system") {
    const roomCount = controller.systemRouteProps.systemStatus?.summary?.sfuRooms
      ?? controller.systemRouteProps.systemStatus?.sfuRooms?.length
      ?? 0;
    const componentCount = controller.systemRouteProps.systemStatus?.components?.length ?? 0;
    return {
      contextMeta: `서비스 ${componentCount}개 · 실시간 연결 ${roomCount}개 · 송출 ${streamingRobotCount}개`,
      contextTitle: "시스템"
    };
  }

  return {
    contextAction: (
      <Button size="sm" variant="primary" onClick={controller.missionRouteProps.onOpenCreateMissionModal}>
        임무 생성
      </Button>
    ),
    contextMeta: `진행 ${activeMissionCount}건 · 대기 ${readyMissionCount}건 · 종료 ${closedMissionCount}건 · 송출 로봇 ${streamingRobotCount}개`,
    contextTitle: "임무"
  };
}

function NavigationItem({ isSectionActive, item }) {
  const Icon = navigationIcons[item.key] ?? RiListCheck3;
  return (
    <NavLink
      className={({ isActive }) => cn(
        "flex h-10 min-w-0 items-center gap-2.5 rounded-lg border border-transparent px-3 text-sm font-extrabold text-slate-400 no-underline transition hover:bg-white/[0.04] hover:text-slate-50",
        (isActive || isSectionActive) && "border-sapphire-300/25 bg-sapphire-500/[0.13] text-white shadow-[inset_3px_0_0_var(--color-sapphire)]"
      )}
      end={item.path !== "/missions"}
      reloadDocument
      to={item.path}
    >
      <Icon aria-hidden className="h-4 w-4 shrink-0 text-current" />
      <span className="min-w-0 truncate">{item.label}</span>
    </NavLink>
  );
}
