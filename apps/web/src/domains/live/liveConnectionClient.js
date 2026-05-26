import { websocketUrlWithQuery } from "../../api/controlCenterApi.js";
import { makeLiveChannelLabel, makeLiveStatusLabel, makePeerRoleLabel } from "../../utils/formatters.js";
import { waitForIceGatheringComplete } from "../../utils/webrtc.js";
import { LiveCloseReason, LiveSessionStatus } from "./liveConnectionStates.js";
import { stopMediaTrack } from "./liveMediaCleanup.js";

export function createLiveConnectionClient({
  missionRoomId,
  onDataChannelMessage,
  onEvent,
  onStatusChange,
  onTrack,
  robotCode,
  rtcConfig
}) {
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
  let requestedCloseReason = "";
  let closeHandler = null;
  const dataChannels = new Set();

  const cleanupDataChannel = (channel) => {
    if (!channel) {
      return;
    }
    channel.onopen = null;
    channel.onclose = null;
    channel.onmessage = null;
    channel.onerror = null;
    if (typeof channel.close === "function" && ["connecting", "open"].includes(channel.readyState)) {
      try {
        channel.close();
      } catch {
        // DataChannel cleanup is best-effort during route transitions.
      }
    }
  };

  const cleanupPeerConnectionHandlers = () => {
    peerConnection.ontrack = null;
    peerConnection.onicecandidate = null;
    peerConnection.oniceconnectionstatechange = null;
    peerConnection.ondatachannel = null;
    peerConnection.onconnectionstatechange = null;
    peerConnection.onsignalingstatechange = null;
    peerConnection.onnegotiationneeded = null;
  };

  const stopPeerConnectionTracks = () => {
    if (typeof peerConnection.getReceivers === "function") {
      peerConnection.getReceivers().forEach((receiver) => stopMediaTrack(receiver?.track));
    }
    if (typeof peerConnection.getSenders === "function") {
      peerConnection.getSenders().forEach((sender) => stopMediaTrack(sender?.track));
    }
  };

  const cleanupPeerConnection = () => {
    stopPeerConnectionTracks();
    dataChannels.forEach(cleanupDataChannel);
    dataChannels.clear();
    cleanupPeerConnectionHandlers();
  };

  const cleanupWebSocketHandlers = ({ keepCloseHandler = false } = {}) => {
    websocket.onopen = null;
    websocket.onerror = null;
    websocket.onmessage = null;
    if (!keepCloseHandler) {
      websocket.onclose = null;
    }
  };

  const handleWebSocketClose = (event) => {
    const payload = {
      code: event.code,
      reason: requestedCloseReason || event.reason,
      wasClean: event.wasClean
    };
    cleanupWebSocketHandlers();
    closeHandler?.(payload);
  };

  const closeTransports = (reason = LiveCloseReason.DISCONNECTED) => {
    requestedCloseReason = reason;
    cleanupPeerConnection();
    cleanupWebSocketHandlers({ keepCloseHandler: true });
    if (websocket.readyState === WebSocket.CONNECTING || websocket.readyState === WebSocket.OPEN) {
      websocket.close(1000, reason);
    } else {
      cleanupWebSocketHandlers();
    }
    if (peerConnection.signalingState !== "closed") {
      peerConnection.close();
    }
  };

  const reportSignalingError = (error, { closeConnection = false } = {}) => {
    const message = error instanceof Error ? error.message : "알 수 없음";
    onStatusChange(LiveSessionStatus.SIGNALING_ERROR);
    onEvent(`관제 연결 오류: ${message}`);
    if (closeConnection) {
      closeTransports(LiveCloseReason.CONNECTION_FAILED);
    }
  };

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
    onStatusChange(peerConnection.iceConnectionState);
    onEvent(`실시간 연결 ${makeLiveStatusLabel(peerConnection.iceConnectionState)}`);
  };

  peerConnection.ontrack = onTrack;

  peerConnection.ondatachannel = (event) => {
    const channel = event.channel;
    dataChannels.add(channel);
    onEvent(`${makeLiveChannelLabel(channel.label)} 데이터 연결 생성`);
    channel.onopen = () => onEvent(`${makeLiveChannelLabel(channel.label)} 데이터 연결됨`);
    channel.onclose = () => {
      onEvent(`${makeLiveChannelLabel(channel.label)} 데이터 종료`);
      cleanupDataChannel(channel);
      dataChannels.delete(channel);
    };
    channel.onmessage = (messageEvent) => onDataChannelMessage(channel.label, messageEvent.data);
  };

  websocket.onopen = () => {
    websocket.send(JSON.stringify({
      type: "select-robot",
      payload: {
        targetPeerId: "sfu",
        robotCode
      }
    }));
    onStatusChange(LiveSessionStatus.SIGNALING_CONNECTED);
    onEvent("관제 연결 준비");
  };

  websocket.onerror = () => {
    onStatusChange(LiveSessionStatus.SIGNALING_ERROR);
    onEvent("관제 연결 오류");
  };

  websocket.onclose = handleWebSocketClose;

  websocket.onmessage = async (event) => {
    let messageType = "";
    try {
      const message = JSON.parse(event.data);
      messageType = message.type ?? "";
      const payload = message.payload ?? {};
      if (payload.targetPeerId && selfPeerId && payload.targetPeerId !== selfPeerId) {
        return;
      }
      if (message.type === "joined") {
        selfPeerId = payload.peerId ?? "";
        onEvent(`${makePeerRoleLabel(payload.role)} 연결 확인`);
        return;
      }
      if (message.type === "peer-present" || message.type === "peer-joined") {
        onEvent(`${makePeerRoleLabel(payload.role)} 참여`);
        return;
      }
      if (message.type === "select-robot-ack") {
        onEvent("관제 로봇 선택 반영", payload.robotCode ?? robotCode);
        return;
      }
      if (message.type === "offer") {
        remoteServerPeerId = payload.fromPeerId ?? remoteServerPeerId;
        onEvent("영상 연결 요청 수신");
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
        onEvent("영상 연결 응답 전송");
        return;
      }
      if (message.type === "candidate" && payload.candidate) {
        await peerConnection.addIceCandidate(payload);
      }
    } catch (error) {
      reportSignalingError(error, { closeConnection: messageType === "offer" });
    }
  };

  return {
    peerConnection,
    websocket,
    close(reason = LiveCloseReason.DISCONNECTED) {
      closeTransports(reason);
    },
    onClose(handler) {
      closeHandler = handler;
      websocket.onclose = handleWebSocketClose;
    }
  };
}
