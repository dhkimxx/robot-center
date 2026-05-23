import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { requestJson, websocketUrlWithQuery } from "../api/controlCenterApi.js";
import {
  formatDateTime,
  formatElapsedTime,
  getTelemetryPositionState,
  makeLiveChannelLabel,
  makeLiveStatusLabel,
  makePeerRoleLabel,
  makeStatusLabel,
  makeStatusTone
} from "../utils/formatters.js";
import {
  createEmptyLiveSession,
  findLatestSampleForRobot,
  findRobotCodeForRemoteTrack,
  findRobotCodeFromDataMessage,
  findTrackSlot,
  formatMediaChannelCount,
  formatStreamingSubscriberCount,
  makeLiveTargetKey
} from "../domains/live/liveHelpers.js";
import {
  createInitialMissionForm,
  createMissionRobotTargets,
  getMissionCodeFromRobotKey,
  makeMissionConnectionKey,
  makeMissionRoomId,
  makeMissionRobotKey
} from "../domains/missions/missionHelpers.js";
import {
  createRecordingPlaybackFile,
  findLatestRecordingForTarget
} from "../domains/recordings/recordingHelpers.js";
import { createInitialRobotForm, createRobotEditForm } from "../domains/robots/robotHelpers.js";
import {
  findNavigationItemByPath,
  getNavigationKeyFromPath,
  getNavigationPath,
  normalizePathname
} from "../utils/navigation.js";
import { waitForIceGatheringComplete } from "../utils/webrtc.js";

export function useControlCenterController() {
  const [activeTab, setActiveTab] = useState(() => getNavigationKeyFromPath(window.location.pathname));
  const [systemStatus, setSystemStatus] = useState(null);
  const [robots, setRobots] = useState([]);
  const [missions, setMissions] = useState([]);
  const [streamingStatuses, setStreamingStatuses] = useState([]);
  const [recordings, setRecordings] = useState([]);
  const [serverTelemetry, setServerTelemetry] = useState([]);
  const [serverSensors, setServerSensors] = useState([]);
  const [connectionInfo, setConnectionInfo] = useState(null);
  const [statusError, setStatusError] = useState("");
  const [notifications, setNotifications] = useState([]);
  const [robotForm, setRobotForm] = useState(createInitialRobotForm);
  const [selectedRobotManagementCode, setSelectedRobotManagementCode] = useState("");
  const [robotEditForm, setRobotEditForm] = useState(() => createRobotEditForm(null));
  const [missionForm, setMissionForm] = useState(createInitialMissionForm);
  const [recordingPlaybackFile, setRecordingPlaybackFile] = useState(null);
  const [missionControlCode, setMissionControlCode] = useState("");
  const [robotModal, setRobotModal] = useState(null);
  const [missionModal, setMissionModal] = useState(null);
  const [pendingArchiveRobotCode, setPendingArchiveRobotCode] = useState("");
  const [selectedMissionManagementCode, setSelectedMissionManagementCode] = useState("");

  const [selectedLiveTargetKey, setSelectedLiveTargetKey] = useState("");
  const [liveSessions, setLiveSessions] = useState({});
  const liveConnectionsRef = useRef(new Map());
  const notificationSequenceRef = useRef(0);

  const navigateToTab = useCallback((tabKey, options = {}) => {
    const nextPath = getNavigationPath(tabKey);
    setActiveTab(tabKey);
    if (normalizePathname(window.location.pathname) === nextPath) {
      return;
    }
    const historyMethod = options.replace ? "replaceState" : "pushState";
    window.history[historyMethod]({ tabKey }, "", nextPath);
  }, []);

  useEffect(() => {
    const currentNavigationItem = findNavigationItemByPath(window.location.pathname);
    if (!currentNavigationItem || normalizePathname(window.location.pathname) === "/") {
      navigateToTab(getNavigationKeyFromPath(window.location.pathname), { replace: true });
    }

    const handlePopState = () => {
      setActiveTab(getNavigationKeyFromPath(window.location.pathname));
    };
    window.addEventListener("popstate", handlePopState);
    return () => window.removeEventListener("popstate", handlePopState);
  }, [navigateToTab]);

  const activeMissions = useMemo(
    () => missions.filter((mission) => mission.status === "active"),
    [missions]
  );
  const primaryRobot = useMemo(() => robots[0] ?? null, [robots]);
  const selectedRobot = useMemo(
    () => robots.find((robot) => robot.robotCode === selectedRobotManagementCode) ?? robots[0] ?? null,
    [robots, selectedRobotManagementCode]
  );
  const selectedMission = useMemo(
    () => missions.find((mission) => mission.missionCode === selectedMissionManagementCode) ?? missions[0] ?? null,
    [missions, selectedMissionManagementCode]
  );
  const pendingArchiveRobot = useMemo(
    () => robots.find((robot) => robot.robotCode === pendingArchiveRobotCode) ?? null,
    [pendingArchiveRobotCode, robots]
  );
  const missionControlMission = useMemo(
    () => missions.find((mission) => mission.missionCode === missionControlCode) ?? null,
    [missionControlCode, missions]
  );
  const activeLiveTargets = useMemo(
    () => activeMissions
      .flatMap((mission) => createMissionRobotTargets(mission, robots, streamingStatuses)),
    [activeMissions, robots, streamingStatuses]
  );
  const missionControlTargets = useMemo(() => {
    if (!missionControlMission) {
      return [];
    }
    return createMissionRobotTargets(missionControlMission, robots, streamingStatuses);
  }, [missionControlMission, robots, streamingStatuses]);
  const liveTargets = useMemo(
    () => (missionControlMission ? missionControlTargets : activeLiveTargets),
    [activeLiveTargets, missionControlMission, missionControlTargets]
  );
  const selectedLiveTarget = useMemo(
    () => liveTargets.find((target) => target.key === selectedLiveTargetKey) ?? liveTargets[0] ?? null,
    [liveTargets, selectedLiveTargetKey]
  );
  const selectedLiveSession = useMemo(
    () => liveSessions[makeLiveTargetKey(selectedLiveTarget)] ?? createEmptyLiveSession(),
    [liveSessions, selectedLiveTarget]
  );
  const activeStreamingStatus = useMemo(() => {
    if (!selectedLiveTarget) {
      return null;
    }
    return selectedLiveTarget.streamingStatus;
  }, [selectedLiveTarget]);
  const selectedMissionCode = selectedLiveTarget?.mission?.missionCode ?? "";
  const selectedRobotCode = selectedLiveTarget?.robotCode ?? "";
  const latestServerTelemetry = useMemo(
    () => findLatestSampleForRobot(serverTelemetry, selectedRobotCode),
    [selectedRobotCode, serverTelemetry]
  );
  const latestServerSensor = useMemo(
    () => findLatestSampleForRobot(serverSensors, selectedRobotCode),
    [selectedRobotCode, serverSensors]
  );
  const latestTelemetry = selectedLiveSession.telemetry ?? latestServerTelemetry;
  const latestSensor = selectedLiveSession.sensor ?? latestServerSensor;
  const latestPositionState = getTelemetryPositionState(latestTelemetry);
  const latestRecording = useMemo(
    () => findLatestRecordingForTarget(recordings, selectedMissionCode, selectedRobotCode),
    [recordings, selectedMissionCode, selectedRobotCode]
  );
  const latestPlayableRecording = useMemo(
    () => findLatestRecordingForTarget(
      recordings,
      selectedMissionCode,
      selectedRobotCode,
      (recording) => Boolean(createRecordingPlaybackFile(recording))
    ),
    [recordings, selectedMissionCode, selectedRobotCode]
  );

  const showNotification = useCallback((message, tone = "info") => {
    notificationSequenceRef.current += 1;
    const notification = {
      id: `notification-${Date.now()}-${notificationSequenceRef.current}`,
      message,
      tone
    };
    setNotifications((current) => [...current, notification].slice(-5));
  }, []);

  const dismissNotification = useCallback((notificationId) => {
    setNotifications((current) => current.filter((notification) => notification.id !== notificationId));
  }, []);

  const updateLiveSession = useCallback((targetKey, updater) => {
    setLiveSessions((current) => {
      const previous = current[targetKey] ?? createEmptyLiveSession();
      return {
        ...current,
        [targetKey]: updater(previous)
      };
    });
  }, []);

  const appendLiveEvent = useCallback((targetKey, message) => {
    if (!targetKey) {
      return;
    }
    updateLiveSession(targetKey, (session) => ({
      ...session,
      events: [
        { id: `${Date.now()}-${Math.random()}`, message, at: new Date().toISOString() },
        ...session.events
      ].slice(0, 40)
    }));
  }, [updateLiveSession]);

  const disconnectLiveTarget = useCallback((targetKey) => {
    const missionCode = getMissionCodeFromRobotKey(targetKey);
    const connectionKey = makeMissionConnectionKey(missionCode);
    const connection = liveConnectionsRef.current.get(connectionKey) ?? liveConnectionsRef.current.get(targetKey);
    if (connection?.websocket) {
      connection.websocket.close(1000, "operator disconnected");
    }
    if (connection?.peerConnection) {
      connection.peerConnection.close();
    }
    liveConnectionsRef.current.delete(connectionKey);
    liveConnectionsRef.current.delete(targetKey);
    const targetKeys = liveTargets
      .filter((candidate) => candidate.mission.missionCode === missionCode)
      .map((candidate) => candidate.key);
    (targetKeys.length > 0 ? targetKeys : [targetKey]).forEach((candidateKey) => {
      updateLiveSession(candidateKey, (session) => ({
        ...session,
        status: "disconnected",
        videoStreams: { rgb: null, thermal: null, audio: null }
      }));
    });
  }, [liveTargets, updateLiveSession]);

  const disconnectAllLiveTargets = useCallback(() => {
    const missionCodes = new Set(liveTargets.map((target) => target.mission.missionCode));
    if (missionCodes.size === 0) {
      Array.from(liveConnectionsRef.current.keys()).forEach((connectionKey) => {
        const connection = liveConnectionsRef.current.get(connectionKey);
        if (connection?.websocket) {
          connection.websocket.close(1000, "operator disconnected");
        }
        if (connection?.peerConnection) {
          connection.peerConnection.close();
        }
        liveConnectionsRef.current.delete(connectionKey);
      });
      return;
    }
    missionCodes.forEach((missionCode) => {
      disconnectLiveTarget(makeMissionRobotKey(missionCode, ""));
    });
  }, [disconnectLiveTarget, liveTargets]);

  useEffect(() => () => {
    liveConnectionsRef.current.forEach((connection) => {
      if (connection?.websocket) {
        connection.websocket.close(1000, "operator disconnected");
      }
      if (connection?.peerConnection) {
        connection.peerConnection.close();
      }
    });
    liveConnectionsRef.current.clear();
  }, []);

  const disconnectMissionLiveTargets = useCallback((targets) => {
    const missionCodes = new Set(targets.map((target) => target.mission.missionCode));
    missionCodes.forEach((missionCode) => {
      disconnectLiveTarget(makeMissionRobotKey(missionCode, ""));
    });
  }, [disconnectLiveTarget]);

  const disconnectMissionByCode = useCallback((missionCode) => {
    disconnectLiveTarget(makeMissionRobotKey(missionCode, ""));
  }, [disconnectLiveTarget]);

  const persistDataChannelMessage = useCallback((targetKey, label, message) => {
    let parsed;
    try {
      parsed = JSON.parse(message);
    } catch {
      appendLiveEvent(targetKey, `${makeLiveChannelLabel(label)} 데이터 해석 실패`);
      return;
    }

    if (label === "telemetry") {
      updateLiveSession(targetKey, (session) => ({ ...session, telemetry: parsed }));
      return;
    }

    if (label === "sensor") {
      updateLiveSession(targetKey, (session) => ({ ...session, sensor: parsed }));
      return;
    }

    appendLiveEvent(targetKey, `${makeLiveChannelLabel(label)} 데이터 수신`);
  }, [appendLiveEvent, updateLiveSession]);

  const connectLiveTarget = useCallback(async (target) => {
    if (!target) {
      showNotification("선택한 임무에 연결할 로봇이 없습니다.", "warning");
      return;
    }

    const missionCode = target.mission.missionCode;
    const missionRoomId = target.missionRoomId || makeMissionRoomId(target.mission);
    const missionTargetsForRoom = liveTargets.filter((candidate) => candidate.mission.missionCode === missionCode);
    const missionTargets = missionTargetsForRoom.length > 0 ? missionTargetsForRoom : [target];
    const targetKey = makeLiveTargetKey(target);
    const connectionKey = makeMissionConnectionKey(missionCode);
    disconnectLiveTarget(targetKey);
    setMissionControlCode(missionCode);
    setSelectedLiveTargetKey(targetKey);
    navigateToTab("missions");
    missionTargets.forEach((missionTarget) => {
      updateLiveSession(missionTarget.key, (session) => ({
        ...session,
        status: "connecting",
        videoStreams: { rgb: null, thermal: null, audio: null }
      }));
    });

    try {
      const rtcConfig = await requestJson("/api/rtc-config");
      const websocket = new WebSocket(websocketUrlWithQuery(rtcConfig.signalingUrl, {
        room: missionRoomId,
        role: "operator"
      }));
      const peerConnection = new RTCPeerConnection({
        iceServers: rtcConfig.iceServers ?? [],
        iceTransportPolicy: rtcConfig.iceTransportPolicy ?? "relay"
      });
      let selfPeerId = "";
      let remoteServerPeerId = "";
      let videoTrackOrder = 0;

      liveConnectionsRef.current.set(connectionKey, { websocket, peerConnection });

      peerConnection.onicecandidate = (event) => {
        if (websocket.readyState !== WebSocket.OPEN) {
          return;
        }
        const payload = event.candidate
          ? {
              candidate: event.candidate.candidate,
              sdpMid: event.candidate.sdpMid,
              sdpMLineIndex: event.candidate.sdpMLineIndex
            }
          : { candidate: "" };
        if (remoteServerPeerId) {
          payload.targetPeerId = remoteServerPeerId;
        }
        websocket.send(JSON.stringify({
          type: "candidate",
          payload
        }));
      };

      peerConnection.oniceconnectionstatechange = () => {
        missionTargets.forEach((missionTarget) => {
          updateLiveSession(missionTarget.key, (session) => ({ ...session, status: peerConnection.iceConnectionState }));
        });
        appendLiveEvent(targetKey, `실시간 연결 ${makeLiveStatusLabel(peerConnection.iceConnectionState)}`);
      };

      peerConnection.ontrack = (event) => {
        const robotCode = findRobotCodeForRemoteTrack(event, missionTargets) || target.robotCode;
        const routedTargetKey = makeMissionRobotKey(missionCode, robotCode);
        const stream = new MediaStream([event.track]);
        const slot = findTrackSlot(event, videoTrackOrder);
        if (slot !== "audio") {
          videoTrackOrder += 1;
        }
        if (slot === "audio") {
          updateLiveSession(routedTargetKey, (session) => ({
            ...session,
            videoStreams: { ...session.videoStreams, audio: stream }
          }));
          appendLiveEvent(routedTargetKey, "오디오 수신");
          return;
        }

        updateLiveSession(routedTargetKey, (session) => ({
          ...session,
          videoStreams: { ...session.videoStreams, [slot]: stream }
        }));
        appendLiveEvent(routedTargetKey, `${makeLiveChannelLabel(slot)} 영상 수신`);
      };

      peerConnection.ondatachannel = (event) => {
        const channel = event.channel;
        appendLiveEvent(targetKey, `${makeLiveChannelLabel(channel.label)} 데이터 연결 생성`);
        channel.onopen = () => appendLiveEvent(targetKey, `${makeLiveChannelLabel(channel.label)} 데이터 연결됨`);
        channel.onclose = () => appendLiveEvent(targetKey, `${makeLiveChannelLabel(channel.label)} 데이터 종료`);
        channel.onmessage = (messageEvent) => {
          const robotCode = findRobotCodeFromDataMessage(messageEvent.data) || target.robotCode;
          persistDataChannelMessage(makeMissionRobotKey(missionCode, robotCode), channel.label, messageEvent.data);
        };
      };

      websocket.onopen = () => {
        missionTargets.forEach((missionTarget) => {
          updateLiveSession(missionTarget.key, (session) => ({ ...session, status: "signaling connected" }));
        });
        appendLiveEvent(targetKey, "관제 연결 준비");
      };
      websocket.onclose = () => {
        missionTargets.forEach((missionTarget) => {
          updateLiveSession(missionTarget.key, (session) => ({ ...session, status: "signaling closed" }));
        });
        appendLiveEvent(targetKey, "관제 연결 종료");
        liveConnectionsRef.current.delete(connectionKey);
      };
      websocket.onerror = () => {
        missionTargets.forEach((missionTarget) => {
          updateLiveSession(missionTarget.key, (session) => ({ ...session, status: "signaling error" }));
        });
        appendLiveEvent(targetKey, "관제 연결 오류");
      };
      websocket.onmessage = async (event) => {
        const message = JSON.parse(event.data);
        const payload = message.payload ?? {};
        if (payload.targetPeerId && selfPeerId && payload.targetPeerId !== selfPeerId) {
          return;
        }
        if (message.type === "joined") {
          selfPeerId = payload.peerId ?? "";
          appendLiveEvent(targetKey, `${makePeerRoleLabel(payload.role)} 연결 확인`);
          return;
        }
        if (message.type === "peer-present" || message.type === "peer-joined") {
          appendLiveEvent(targetKey, `${makePeerRoleLabel(payload.role)} 참여`);
          return;
        }
        if (message.type === "offer") {
          remoteServerPeerId = payload.fromPeerId ?? remoteServerPeerId;
          appendLiveEvent(targetKey, "영상 연결 요청 수신");
          await peerConnection.setRemoteDescription({ type: "offer", sdp: payload.sdp });
          const answer = await peerConnection.createAnswer();
          await peerConnection.setLocalDescription(answer);
          await waitForIceGatheringComplete(peerConnection);
          const localDescription = peerConnection.localDescription ?? answer;
          const answerPayload = {
            type: localDescription.type,
            sdp: localDescription.sdp
          };
          if (remoteServerPeerId) {
            answerPayload.targetPeerId = remoteServerPeerId;
          }
          websocket.send(JSON.stringify({
            type: "answer",
            payload: answerPayload
          }));
          appendLiveEvent(targetKey, "영상 연결 응답 전송");
          return;
        }
        if (message.type === "candidate" && payload.candidate) {
          await peerConnection.addIceCandidate(payload);
        }
      };
    } catch (error) {
      missionTargets.forEach((missionTarget) => {
        updateLiveSession(missionTarget.key, (session) => ({ ...session, status: "failed" }));
      });
      showNotification(error instanceof Error ? error.message : "관제 연결 실패", "danger");
      appendLiveEvent(targetKey, `관제 연결 실패: ${error instanceof Error ? error.message : "알 수 없음"}`);
      disconnectLiveTarget(targetKey);
    }
  }, [appendLiveEvent, disconnectLiveTarget, liveTargets, navigateToTab, persistDataChannelMessage, showNotification, updateLiveSession]);

  const connectLive = useCallback(() => {
    void connectLiveTarget(selectedLiveTarget);
  }, [connectLiveTarget, selectedLiveTarget]);

  const connectAllLiveTargets = useCallback(() => {
    const connectedMissionCodes = new Set();
    liveTargets.forEach((target) => {
      if (connectedMissionCodes.has(target.mission.missionCode)) {
        return;
      }
      connectedMissionCodes.add(target.mission.missionCode);
      void connectLiveTarget(target);
    });
  }, [connectLiveTarget, liveTargets]);

  const disconnectLive = useCallback(() => {
    if (selectedLiveTarget) {
      disconnectLiveTarget(makeLiveTargetKey(selectedLiveTarget));
    } else {
      disconnectAllLiveTargets();
    }
  }, [disconnectAllLiveTargets, disconnectLiveTarget, selectedLiveTarget]);

  const loadAll = useCallback(async () => {
    const [statusPayload, robotPayload, missionPayload, streamingPayload, recordingPayload] = await Promise.all([
      requestJson("/api/system/status"),
      requestJson("/api/robots"),
      requestJson("/api/missions"),
      requestJson("/api/streaming-statuses"),
      requestJson("/api/recordings")
    ]);
    setSystemStatus(statusPayload);
    setRobots(robotPayload.robots ?? []);
    setMissions(missionPayload.missions ?? []);
    setStreamingStatuses(streamingPayload.streamingStatuses ?? []);
    setRecordings(recordingPayload.recordings ?? []);
    setStatusError("");
  }, []);

  useEffect(() => {
    let cancelled = false;
    async function loadInitial() {
      try {
        await loadAll();
      } catch (error) {
        if (!cancelled) {
          setStatusError(error instanceof Error ? error.message : "status load failed");
        }
      }
    }
    loadInitial();
    const timer = window.setInterval(loadInitial, 5000);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [loadAll]);

  useEffect(() => {
    const selectedRobotCodes = missionForm.robotCodes ?? [];
    if (selectedRobotCodes.length === 0 && robots.length > 0) {
      setMissionForm((current) => ({
        ...current,
        robotCode: robots[0].robotCode,
        robotCodes: [robots[0].robotCode]
      }));
    }
  }, [missionForm.robotCodes, robots]);

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
    setRobotEditForm(createRobotEditForm(selectedRobot));
  }, [selectedRobot]);

  useEffect(() => {
    if (missions.length === 0) {
      setSelectedMissionManagementCode("");
      return;
    }
    if (!selectedMissionManagementCode || !missions.some((mission) => mission.missionCode === selectedMissionManagementCode)) {
      setSelectedMissionManagementCode(missions[0].missionCode);
    }
  }, [missions, selectedMissionManagementCode]);

  useEffect(() => {
    if (missionControlCode && !missions.some((mission) => mission.missionCode === missionControlCode)) {
      setMissionControlCode("");
    }
  }, [missionControlCode, missions]);

  useEffect(() => {
    if (liveTargets.length === 0) {
      setSelectedLiveTargetKey("");
      return;
    }
    if (!selectedLiveTargetKey || !liveTargets.some((target) => target.key === selectedLiveTargetKey)) {
      setSelectedLiveTargetKey(liveTargets[0].key);
    }
  }, [liveTargets, selectedLiveTargetKey]);

  useEffect(() => {
    let cancelled = false;
    async function loadMissionSamples() {
      if (!selectedLiveTarget) {
        setServerTelemetry([]);
        setServerSensors([]);
        return;
      }
      try {
        const [telemetryPayload, sensorPayload] = await Promise.all([
          requestJson(`/api/telemetry?missionId=${encodeURIComponent(selectedLiveTarget.mission.id)}`),
          requestJson(`/api/sensor-readings?missionId=${encodeURIComponent(selectedLiveTarget.mission.id)}`)
        ]);
        if (!cancelled) {
          setServerTelemetry(telemetryPayload.telemetry ?? []);
          setServerSensors(sensorPayload.sensorReadings ?? []);
        }
      } catch (error) {
        if (!cancelled) {
          appendLiveEvent(makeLiveTargetKey(selectedLiveTarget), `sample polling failed: ${error instanceof Error ? error.message : "unknown"}`);
        }
      }
    }
    loadMissionSamples();
    const timer = window.setInterval(loadMissionSamples, 3000);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [selectedLiveTarget, appendLiveEvent]);

  const playLatestRecording = useCallback(() => {
    const playbackFile = createRecordingPlaybackFile(latestPlayableRecording);
    if (!playbackFile) {
      showNotification("재생 가능한 MP4가 아직 없습니다.", "warning");
      return;
    }
    setRecordingPlaybackFile(playbackFile);
  }, [latestPlayableRecording, showNotification]);

  const openMissionControl = useCallback((mission) => {
    const targets = createMissionRobotTargets(mission, robots, streamingStatuses);
    setMissionControlCode(mission.missionCode);
    setSelectedLiveTargetKey(targets[0]?.key ?? "");
    navigateToTab("missions");
  }, [navigateToTab, robots, streamingStatuses]);

  const closeRobotModal = useCallback(() => {
    setRobotModal(null);
  }, []);

  const closeMissionModal = useCallback(() => {
    setMissionModal(null);
  }, []);

  function openRobotCreateModal() {
    setRobotForm(createInitialRobotForm());
    setRobotModal("create");
  }

  function openMissionCreateModal() {
    setMissionForm((current) => {
      const robotCodes = current.robotCodes?.length > 0
        ? current.robotCodes
        : current.robotCode
          ? [current.robotCode]
          : [];
      return {
        ...createInitialMissionForm(),
        robotCode: robotCodes[0] ?? "",
        robotCodes
      };
    });
    setMissionModal("create");
  }

  function openRobotEditModal() {
    if (!selectedRobot) {
      showNotification("수정할 로봇을 선택하세요.", "warning");
      return;
    }
    setRobotEditForm(createRobotEditForm(selectedRobot));
    setRobotModal("edit");
  }

  async function createRobot(event) {
    event.preventDefault();
    try {
      const payload = await requestJson("/api/robots", {
        method: "POST",
        body: JSON.stringify(robotForm)
      });
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
      const payload = await requestJson(`/api/robots/${robotCode}/connection-info`);
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
      const payload = await requestJson(`/api/robots/${selectedRobot.robotCode}`, {
        method: "PATCH",
        body: JSON.stringify(robotEditForm)
      });
      showNotification(`${payload.robot.robotCode} 수정 완료`, "success");
      closeRobotModal();
      await loadAll();
    } catch (error) {
      showNotification(error instanceof Error ? error.message : "로봇 수정 실패", "danger");
    }
  }

  async function rotateRobotToken(robotCode) {
    try {
      const payload = await requestJson(`/api/robots/${robotCode}/connection-token`, { method: "POST" });
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
      await requestJson(`/api/robots/${robotCode}`, { method: "DELETE" });
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

  async function createMission(event) {
    event.preventDefault();
    try {
      const robotCodes = missionForm.robotCodes ?? [];
      const legacyRobotCode = robotCodes[0] ?? "";
      const payload = await requestJson("/api/missions", {
        method: "POST",
        body: JSON.stringify({
          ...missionForm,
          robotCode: legacyRobotCode,
          robotCodes
        })
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
      const payload = await requestJson(`/api/missions/${missionCode}/start`, { method: "POST" });
      showNotification(`${payload.mission.missionCode} 시작`, "success");
      openMissionControl(payload.mission);
      await loadAll();
    } catch (error) {
      showNotification(error instanceof Error ? error.message : "임무 시작 실패", "danger");
    }
  }

  async function endMission(missionCode) {
    try {
      const payload = await requestJson(`/api/missions/${missionCode}/end`, { method: "POST" });
      showNotification(`${payload.mission.missionCode} 종료`, "success");
      disconnectMissionByCode(payload.mission.missionCode);
      await loadAll();
    } catch (error) {
      showNotification(error instanceof Error ? error.message : "임무 종료 실패", "danger");
    }
  }

  const operationStatuses = [
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
      detail: activeStreamingStatus ? `${formatMediaChannelCount(activeStreamingStatus)} / ${formatStreamingSubscriberCount(activeStreamingStatus)}` : "송출 대기",
      tone: makeStatusTone(selectedLiveSession.status)
    },
    {
      label: "위치",
      value: latestPositionState.statusLabel,
      detail: latestPositionState.hasPosition ? `수신 ${formatElapsedTime(latestPositionState.timestamp)}` : "GPS 대기",
      tone: makeStatusTone(latestPositionState.statusLabel)
    }
  ];

  return {
    activeTab,
    navigateToTab,
    systemStatus,
    robots,
    missions,
    recordings,
    connectionInfo,
    statusError,
    notifications,
    robotForm,
    setRobotForm,
    selectedRobot,
    robotEditForm,
    setRobotEditForm,
    missionForm,
    setMissionForm,
    recordingPlaybackFile,
    setRecordingPlaybackFile,
    missionControlMission,
    missionControlTargets,
    selectedLiveSession,
    liveSessions,
    latestRecording,
    latestPlayableRecording,
    latestTelemetry,
    latestSensor,
    operationStatuses,
    selectedMission,
    selectedLiveTargetKey,
    setSelectedLiveTargetKey,
    robotModal,
    missionModal,
    pendingArchiveRobotCode,
    pendingArchiveRobot,
    setPendingArchiveRobotCode,
    archiveRobot,
    closeMissionModal,
    closeRobotModal,
    confirmArchiveRobot,
    connectAllLiveTargets,
    connectLive,
    createMission,
    createRobot,
    disconnectLive,
    disconnectMissionLiveTargets,
    dismissNotification,
    endMission,
    loadConnectionInfo,
    openMissionControl,
    openMissionCreateModal,
    openRobotCreateModal,
    openRobotEditModal,
    playLatestRecording,
    rotateRobotToken,
    setMissionControlCode,
    setSelectedMissionManagementCode,
    setSelectedRobotManagementCode,
    startMission,
    updateRobot
  };
}
