import { Navigate, Route, Routes } from "react-router-dom";
import MissionsScreen from "../domains/missions/MissionsScreen.jsx";
import RobotsScreen from "../domains/robots/RobotsScreen.jsx";
import SystemScreen from "../domains/system/SystemScreen.jsx";

export function ControlCenterRoutes({
  missionRouteProps,
  robotRouteProps,
  systemRouteProps
}) {
  return (
    <Routes>
      <Route path="/" element={<Navigate replace to="/missions" />} />
      <Route path="/missions" element={<MissionsScreen {...missionRouteProps} controlMission={null} />} />
      <Route path="/missions/:missionCode/control" element={<MissionsScreen {...missionRouteProps} />} />
      <Route path="/missions/:missionCode/replay" element={<MissionsScreen {...missionRouteProps} controlMission={null} />} />
      <Route path="/robots" element={<RobotsScreen {...robotRouteProps} />} />
      <Route path="/system" element={<SystemScreen {...systemRouteProps} />} />
      <Route path="*" element={<Navigate replace to="/missions" />} />
    </Routes>
  );
}
