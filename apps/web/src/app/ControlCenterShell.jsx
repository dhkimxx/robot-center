import { NavLink, useLocation, useNavigate } from "react-router-dom";
import NotificationStack from "../components/NotificationStack.jsx";
import { navigationItems } from "../config/controlCenterConfig.js";
import { MissionModals } from "../domains/missions/MissionModals.jsx";
import { RecordingModals } from "../domains/recordings/RecordingModals.jsx";
import { RobotModals } from "../domains/robots/RobotModals.jsx";
import { useControlCenterController } from "../hooks/useControlCenterController.js";
import { cn } from "../utils/cn.js";
import { ControlCenterRoutes } from "./ControlCenterRoutes.jsx";
import { getActiveSection, getRouteMissionControlCode } from "./routeUtils.js";

export function ControlCenterApp() {
  const location = useLocation();
  const navigate = useNavigate();
  const routeMissionControlCode = getRouteMissionControlCode(location.pathname);
  const activeSection = getActiveSection(location.pathname);
  const controller = useControlCenterController({
    activeSection,
    missionControlCode: routeMissionControlCode,
    navigateToPath: navigate
  });

  return (
    <main className="grid h-screen min-h-0 grid-cols-[184px_minmax(0,1fr)] bg-command-950 max-[900px]:grid-cols-1 max-[900px]:overflow-auto">
      <aside className="h-screen min-w-0 overflow-auto border-r border-slate-500/10 bg-[#080a0e] px-3 py-4 text-slate-100 max-[900px]:h-auto max-[900px]:border-b max-[900px]:border-r-0">
        <nav className="grid gap-2 max-[900px]:grid-cols-4">
          {navigationItems.map((item) => (
            <NavLink
              className={({ isActive }) => cn(
                "flex h-10 items-center rounded-lg border border-transparent px-3 text-sm font-extrabold text-slate-400 no-underline transition hover:bg-white/[0.04] hover:text-slate-50",
                (isActive || location.pathname.startsWith(`${item.path}/`)) && "border-sapphire-300/25 bg-sapphire-500/[0.13] text-white shadow-[inset_3px_0_0_var(--color-sapphire)]"
              )}
              end={item.path !== "/missions"}
              key={item.key}
              to={item.path}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
      </aside>

      <section className="grid h-screen min-h-0 min-w-0 grid-rows-[auto_minmax(0,1fr)] gap-3 overflow-hidden px-4 pb-4 pt-3 max-[900px]:h-auto max-[900px]:min-h-screen max-[900px]:overflow-visible">
        <div className="grid min-w-0 gap-2 border-b border-slate-500/10 pb-3">
          <header className="flex min-h-11 items-center justify-between gap-4 px-0.5 max-[640px]:items-start">
            <div>
              <p className="mb-1 text-[11px] font-extrabold uppercase tracking-normal text-sapphire-300">SST Robot Control PoC</p>
              <h1 className="text-2xl font-black leading-tight tracking-normal text-slate-50">Sapphire Command Center</h1>
            </div>
          </header>
          {controller.statusError ? (
            <div className="rounded-xl border border-red-300/25 bg-red-400/[0.12] px-3 py-2 text-sm font-bold text-red-100" role="status">
              서버 응답 대기: {controller.statusError}
            </div>
          ) : null}
        </div>

        <div className="grid min-h-0 min-w-0 grid-rows-[minmax(0,1fr)] overflow-hidden max-[900px]:overflow-visible">
          <ControlCenterRoutes
            missionRouteProps={controller.missionRouteProps}
            recordingRouteProps={controller.recordingRouteProps}
            robotRouteProps={controller.robotRouteProps}
            systemRouteProps={controller.systemRouteProps}
          />
        </div>
      </section>

      <NotificationStack notifications={controller.notifications} onDismiss={controller.dismissNotification} />
      <RobotModals {...controller.robotModalProps} />
      <MissionModals {...controller.missionModalProps} />
      <RecordingModals {...controller.playbackModalProps} />
    </main>
  );
}
