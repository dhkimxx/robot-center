import { useCallback } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import NotificationStack from "../components/NotificationStack.jsx";
import { MissionModals } from "../domains/missions/MissionModals.jsx";
import { RecordingModals } from "../domains/recordings/RecordingModals.jsx";
import { RobotModals } from "../domains/robots/RobotModals.jsx";
import { useControlCenterController } from "../hooks/useControlCenterController.js";
import AppStatusBar from "./AppStatusBar.jsx";
import AppSidebar from "./AppSidebar.jsx";
import { ControlCenterRoutes } from "./ControlCenterRoutes.jsx";
import PageContextBar from "./PageContextBar.jsx";
import {
  getActiveSection,
  getRouteMissionControlCode,
  getRouteMissionReplayCode,
  getRouteSelectedMissionCode
} from "./routeUtils.js";

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

  return (
    <main className="grid h-screen min-h-0 grid-cols-[224px_minmax(0,1fr)] bg-command-950 max-[900px]:grid-cols-1 max-[900px]:overflow-auto">
      <AppSidebar />

      <section className="grid h-screen min-h-0 min-w-0 grid-rows-[52px_minmax(0,1fr)_32px] overflow-hidden max-[900px]:h-auto max-[900px]:min-h-screen max-[900px]:overflow-visible">
        <PageContextBar
          action={controller.pageChrome.action}
          meta={controller.pageChrome.meta}
          title={controller.pageChrome.title}
        />

        <div className="grid min-h-0 min-w-0 grid-rows-[minmax(0,1fr)] overflow-hidden p-3 max-[900px]:overflow-visible">
          <ControlCenterRoutes
            missionRouteProps={controller.missionRouteProps}
            robotRouteProps={controller.robotRouteProps}
            systemRouteProps={controller.systemRouteProps}
          />
        </div>
        <AppStatusBar
          liveStatuses={controller.missionRouteProps.liveStatuses}
          robots={controller.missionRouteProps.robots}
          statusError={controller.statusError}
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
