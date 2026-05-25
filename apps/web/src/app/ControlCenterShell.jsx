import { NavLink, useLocation, useNavigate } from "react-router-dom";
import NotificationStack from "../components/NotificationStack.jsx";
import { navigationItems } from "../config/controlCenterConfig.js";
import { useControlCenterController } from "../hooks/useControlCenterController.js";
import { ControlCenterModals } from "./ControlCenterModals.jsx";
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
    <main className="app-shell">
      <aside className="sidebar">
        <nav className="nav-list">
          {navigationItems.map((item) => (
            <NavLink
              className={({ isActive }) => isActive || location.pathname.startsWith(`${item.path}/`) ? "nav-item active" : "nav-item"}
              end={item.path !== "/missions"}
              key={item.key}
              to={item.path}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
      </aside>

      <section className="workspace">
        <div className="workspace-header">
          <header className="topbar">
            <div>
              <p className="eyebrow">SST Robot Control PoC</p>
              <h1>Sapphire Command Center</h1>
            </div>
          </header>
          {controller.statusError ? (
            <div className="server-alert" role="status">
              서버 응답 대기: {controller.statusError}
            </div>
          ) : null}
        </div>

        <div className="workspace-content">
          <ControlCenterRoutes controller={controller} navigateToPath={navigate} />
        </div>
      </section>

      <NotificationStack notifications={controller.notifications} onDismiss={controller.dismissNotification} />
      <ControlCenterModals controller={controller} />
    </main>
  );
}
