import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  archiveRobotRequest,
  createRobotRequest,
  fetchRobotConnectionInfo,
  rotateRobotConnectionToken,
  updateRobotRequest
} from "../../api/robotsApi.js";
import { createInitialRobotForm, createRobotEditForm, shouldRefreshRobotEditForm } from "./robotHelpers.js";

export function useRobotManagementController({
  connectionInfo,
  loadAll,
  robots,
  setConnectionInfo,
  showNotification
}) {
  const [robotForm, setRobotForm] = useState(createInitialRobotForm);
  const [selectedRobotManagementCode, setSelectedRobotManagementCode] = useState("");
  const [robotEditForm, setRobotEditForm] = useState(() => createRobotEditForm(null));
  const [robotModal, setRobotModal] = useState(null);
  const [pendingArchiveRobotCode, setPendingArchiveRobotCode] = useState("");
  const previousRobotEditCodeRef = useRef("");

  const selectedRobot = useMemo(
    () => robots.find((robot) => robot.robotCode === selectedRobotManagementCode) ?? robots[0] ?? null,
    [robots, selectedRobotManagementCode]
  );
  const selectedRobotCode = selectedRobot?.robotCode ?? "";
  const pendingArchiveRobot = useMemo(
    () => robots.find((robot) => robot.robotCode === pendingArchiveRobotCode) ?? null,
    [pendingArchiveRobotCode, robots]
  );

  useEffect(() => {
    if (robots.length === 0) {
      setSelectedRobotManagementCode("");
      return;
    }
    if (!selectedRobotManagementCode || !robots.some((robot) => robot.robotCode === selectedRobotManagementCode)) {
      setSelectedRobotManagementCode(robots[0].robotCode);
    }
  }, [robots, selectedRobotManagementCode]);

  useEffect(() => {
    if (!shouldRefreshRobotEditForm({
      nextRobotCode: selectedRobotCode,
      previousRobotCode: previousRobotEditCodeRef.current,
      robotModal
    })) {
      return;
    }
    previousRobotEditCodeRef.current = selectedRobotCode;
    setRobotEditForm(createRobotEditForm(selectedRobot));
  }, [robotModal, selectedRobot, selectedRobotCode]);

  const closeRobotModal = useCallback(() => {
    setRobotModal(null);
  }, []);

  function openRobotCreateModal() {
    setRobotForm(createInitialRobotForm());
    setRobotModal("create");
  }

  function openRobotEditModal() {
    if (!selectedRobot) {
      showNotification("수정할 로봇을 선택하세요.", "warning");
      return;
    }
    previousRobotEditCodeRef.current = selectedRobot.robotCode;
    setRobotEditForm(createRobotEditForm(selectedRobot));
    setRobotModal("edit");
  }

  async function createRobot(event) {
    event.preventDefault();
    try {
      const payload = await createRobotRequest(robotForm);
      setConnectionInfo(payload.connectionInfo);
      showNotification(`${payload.robot.robotCode} 등록 완료`, "success");
      setRobotForm(createInitialRobotForm());
      setSelectedRobotManagementCode(payload.robot.robotCode);
      setRobotModal("connection");
      await loadAll();
    } catch (error) {
      showNotification(error instanceof Error ? error.message : "로봇 생성 실패", "danger");
    }
  }

  async function loadConnectionInfo(robotCode) {
    setSelectedRobotManagementCode(robotCode);
    try {
      const payload = await fetchRobotConnectionInfo(robotCode);
      setConnectionInfo(payload.connectionInfo);
      setRobotModal("connection");
    } catch (error) {
      showNotification(error instanceof Error ? error.message : "연결 정보 조회 실패", "danger");
    }
  }

  async function updateRobot(event) {
    event.preventDefault();
    if (!selectedRobot) {
      showNotification("수정할 로봇을 선택하세요.", "warning");
      return;
    }
    try {
      const payload = await updateRobotRequest(selectedRobot.robotCode, robotEditForm);
      showNotification(`${payload.robot.robotCode} 수정 완료`, "success");
      closeRobotModal();
      await loadAll();
    } catch (error) {
      showNotification(error instanceof Error ? error.message : "로봇 수정 실패", "danger");
    }
  }

  async function rotateRobotToken(robotCode) {
    try {
      const payload = await rotateRobotConnectionToken(robotCode);
      setConnectionInfo(payload.connectionInfo);
      setSelectedRobotManagementCode(robotCode);
      showNotification(`${robotCode} 연결 토큰 재발급 완료`, "success");
      setRobotModal("connection");
    } catch (error) {
      showNotification(error instanceof Error ? error.message : "토큰 재발급 실패", "danger");
    }
  }

  async function archiveRobot(robotCode) {
    setPendingArchiveRobotCode(robotCode);
  }

  async function confirmArchiveRobot() {
    const robotCode = pendingArchiveRobotCode;
    if (!robotCode) {
      return;
    }
    setPendingArchiveRobotCode("");
    try {
      await archiveRobotRequest(robotCode);
      setConnectionInfo((current) => current?.robotCode === robotCode ? null : current);
      if (connectionInfo?.robotCode === robotCode) {
        closeRobotModal();
      }
      showNotification(`${robotCode} 목록 제거 완료`, "success");
      await loadAll();
    } catch (error) {
      showNotification(error instanceof Error ? error.message : "로봇 삭제 실패", "danger");
    }
  }

  return {
    archiveRobot,
    closeRobotModal,
    confirmArchiveRobot,
    createRobot,
    loadConnectionInfo,
    openRobotCreateModal,
    openRobotEditModal,
    pendingArchiveRobot,
    pendingArchiveRobotCode,
    robotEditForm,
    robotForm,
    robotModal,
    rotateRobotToken,
    selectedRobot,
    selectedRobotManagementCode,
    setPendingArchiveRobotCode,
    setRobotEditForm,
    setRobotForm,
    setSelectedRobotManagementCode,
    updateRobot
  };
}
