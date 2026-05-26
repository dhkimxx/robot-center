import { useCallback, useEffect, useMemo, useState } from "react";
import {
  createMissionRequest,
  endMissionRequest,
  startMissionRequest
} from "../../api/missionsApi.js";
import {
  createInitialMissionForm,
  createMissionRobotTargets,
  getBusyRobotReasonForMissionCreate,
  isClosedMission
} from "./missionHelpers.js";

export function useMissionManagementController({
  disconnectMissionByCode,
  loadAll,
  missions,
  navigateToPath,
  observedStreams,
  resolveStoredLiveTargetKey,
  robots,
  routeSelectedMissionCode = "",
  setMissionControlCode,
  setSelectedLiveTargetKey,
  showNotification
}) {
  const [missionForm, setMissionForm] = useState(createInitialMissionForm);
  const [missionModal, setMissionModal] = useState(null);
  const [selectedMissionManagementCode, setSelectedMissionManagementCode] = useState("");

  const selectedMission = useMemo(
    () => missions.find((mission) => mission.missionCode === selectedMissionManagementCode)
      ?? missions.find((mission) => !isClosedMission(mission))
      ?? missions[0]
      ?? null,
    [missions, selectedMissionManagementCode]
  );

  useEffect(() => {
    const selectedRobotCodes = missionForm.robotCodes ?? [];
    if (selectedRobotCodes.length === 0 && robots.length > 0) {
      const firstAssignableRobot = robots.find((robot) => !getBusyRobotReasonForMissionCreate(robot.robotCode, missions, Date.now(), observedStreams));
      setMissionForm((current) => ({
        ...current,
        robotCode: firstAssignableRobot?.robotCode ?? "",
        robotCodes: firstAssignableRobot ? [firstAssignableRobot.robotCode] : []
      }));
    }
  }, [missionForm.robotCodes, missions, observedStreams, robots]);

  useEffect(() => {
    if (missions.length === 0) {
      setSelectedMissionManagementCode("");
      return;
    }
    if (!selectedMissionManagementCode || !missions.some((mission) => mission.missionCode === selectedMissionManagementCode)) {
      setSelectedMissionManagementCode((missions.find((mission) => !isClosedMission(mission)) ?? missions[0]).missionCode);
    }
  }, [missions, selectedMissionManagementCode]);

  useEffect(() => {
    if (!routeSelectedMissionCode || !missions.some((mission) => mission.missionCode === routeSelectedMissionCode)) {
      return;
    }
    setSelectedMissionManagementCode(routeSelectedMissionCode);
  }, [missions, routeSelectedMissionCode]);

  const closeMissionModal = useCallback(() => {
    setMissionModal(null);
  }, []);

  const openMissionControl = useCallback((mission, options = {}) => {
    const targets = createMissionRobotTargets(mission, robots, observedStreams);
    if (options.navigate !== false && navigateToPath) {
      navigateToPath(`/missions/${encodeURIComponent(mission.missionCode)}/control`);
    }
    setSelectedMissionManagementCode(mission.missionCode);
    setMissionControlCode(mission.missionCode);
    setSelectedLiveTargetKey(resolveStoredLiveTargetKey(targets));
  }, [navigateToPath, observedStreams, resolveStoredLiveTargetKey, robots, setMissionControlCode, setSelectedLiveTargetKey]);

  const closeMissionControl = useCallback((missionControlCode, options = {}) => {
    if (missionControlCode) {
      disconnectMissionByCode(missionControlCode);
    }
    setMissionControlCode("");
    if (options.navigate !== false && navigateToPath) {
      const selectedQuery = missionControlCode ? `?selected=${encodeURIComponent(missionControlCode)}` : "";
      navigateToPath(`/missions${selectedQuery}`);
    }
  }, [disconnectMissionByCode, navigateToPath, setMissionControlCode]);

  const openMissionReplay = useCallback((mission, options = {}) => {
    if (!mission?.missionCode) {
      return;
    }
    setSelectedMissionManagementCode(mission.missionCode);
    if (options.navigate !== false && navigateToPath) {
      navigateToPath(`/missions/${encodeURIComponent(mission.missionCode)}/replay`);
    }
  }, [navigateToPath]);

  function openMissionCreateModal() {
    setMissionForm((current) => {
      const robotCodes = current.robotCodes?.length > 0
        ? current.robotCodes
        : current.robotCode
          ? [current.robotCode]
          : [];
      const assignableRobotCodes = robotCodes.filter((robotCode) => !getBusyRobotReasonForMissionCreate(robotCode, missions, Date.now(), observedStreams));
      return {
        ...createInitialMissionForm(),
        robotCode: assignableRobotCodes[0] ?? "",
        robotCodes: assignableRobotCodes
      };
    });
    setMissionModal("create");
  }

  async function createMission(event) {
    event.preventDefault();
    try {
      const robotCodes = (missionForm.robotCodes ?? []).filter((robotCode) => !getBusyRobotReasonForMissionCreate(robotCode, missions, Date.now(), observedStreams));
      if (robotCodes.length === 0) {
        showNotification("배정 가능한 로봇이 없습니다.", "warning");
        return;
      }
      const legacyRobotCode = robotCodes[0] ?? "";
      const payload = await createMissionRequest({
        ...missionForm,
        robotCode: legacyRobotCode,
        robotCodes
      });
      showNotification(`${payload.mission.missionCode} 생성 완료`, "success");
      setMissionForm((current) => {
        const currentRobotCodes = current.robotCodes ?? [];
        return {
          ...createInitialMissionForm(),
          robotCode: currentRobotCodes[0] ?? "",
          robotCodes: currentRobotCodes
        };
      });
      setSelectedMissionManagementCode(payload.mission.missionCode);
      closeMissionModal();
      await loadAll();
    } catch (error) {
      showNotification(error instanceof Error ? error.message : "임무 생성 실패", "danger");
    }
  }

  async function startMission(missionCode) {
    try {
      const payload = await startMissionRequest(missionCode);
      showNotification(`${payload.mission.missionCode} 시작`, "success");
      openMissionControl(payload.mission);
      await loadAll();
    } catch (error) {
      showNotification(error instanceof Error ? error.message : "임무 시작 실패", "danger");
    }
  }

  async function endMission(missionCode) {
    try {
      const payload = await endMissionRequest(missionCode);
      showNotification(`${payload.mission.missionCode} 종료`, "success");
      disconnectMissionByCode(payload.mission.missionCode);
      await loadAll();
    } catch (error) {
      showNotification(error instanceof Error ? error.message : "임무 종료 실패", "danger");
    }
  }

  return {
    closeMissionControl,
    closeMissionModal,
    createMission,
    endMission,
    missionForm,
    missionModal,
    openMissionControl,
    openMissionCreateModal,
    openMissionReplay,
    selectedMission,
    selectedMissionManagementCode,
    setMissionForm,
    setSelectedMissionManagementCode,
    startMission
  };
}
