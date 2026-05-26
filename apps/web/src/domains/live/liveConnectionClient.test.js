import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createLiveConnectionClient } from "./liveConnectionClient.js";
import { LiveCloseReason, LiveSessionStatus } from "./liveConnectionStates.js";

const originalWebSocket = globalThis.WebSocket;
const originalRTCPeerConnection = globalThis.RTCPeerConnection;

class FakeWebSocket {
  static CONNECTING = 0;
  static OPEN = 1;
  static CLOSING = 2;
  static CLOSED = 3;
  static instances = [];

  constructor(url) {
    this.url = url;
    this.readyState = FakeWebSocket.OPEN;
    this.sentMessages = [];
    FakeWebSocket.instances.push(this);
  }

  send(message) {
    this.sentMessages.push(message);
  }

  close(code, reason) {
    this.readyState = FakeWebSocket.CLOSED;
    this.closeCode = code;
    this.closeReason = reason;
    this.onclose?.({ code, reason, wasClean: true });
  }
}

class FakePeerConnection {
  static instances = [];

  constructor() {
    this.iceConnectionState = "new";
    this.localDescription = null;
    this.signalingState = "stable";
    this.receiverTrack = { stop: vi.fn() };
    this.senderTrack = { stop: vi.fn() };
    this.addIceCandidate = vi.fn(async () => {});
    this.createAnswer = vi.fn(async () => ({ sdp: "answer-sdp", type: "answer" }));
    this.getReceivers = vi.fn(() => [{ track: this.receiverTrack }]);
    this.getSenders = vi.fn(() => [{ track: this.senderTrack }]);
    this.setLocalDescription = vi.fn(async (description) => {
      this.localDescription = description;
    });
    this.setRemoteDescription = vi.fn(async () => {});
    FakePeerConnection.instances.push(this);
  }

  close() {
    this.closed = true;
    this.signalingState = "closed";
  }
}

function createTestClient(overrides = {}) {
  return createLiveConnectionClient({
    missionRoomId: "mission-006",
    onDataChannelMessage: vi.fn(),
    onEvent: vi.fn(),
    onStatusChange: vi.fn(),
    onTrack: vi.fn(),
    robotCode: "robot-001",
    rtcConfig: {
      iceServers: [],
      signalingUrl: "ws://127.0.0.1/sfu/ws"
    },
    ...overrides
  });
}

describe("createLiveConnectionClient", () => {
  beforeEach(() => {
    FakeWebSocket.instances = [];
    FakePeerConnection.instances = [];
    vi.stubGlobal("WebSocket", FakeWebSocket);
    vi.stubGlobal("RTCPeerConnection", FakePeerConnection);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
    globalThis.WebSocket = originalWebSocket;
    globalThis.RTCPeerConnection = originalRTCPeerConnection;
  });

  it("records malformed signaling messages as signaling errors", async () => {
    const onEvent = vi.fn();
    const onStatusChange = vi.fn();
    createTestClient({ onEvent, onStatusChange });
    const websocket = FakeWebSocket.instances[0];

    await websocket.onmessage({ data: "{" });

    expect(onStatusChange).toHaveBeenCalledWith(LiveSessionStatus.SIGNALING_ERROR);
    expect(onEvent).toHaveBeenCalledWith(expect.stringContaining("관제 연결 오류"));
    expect(FakePeerConnection.instances[0].closed).toBeUndefined();
  });

  it("closes transports when offer handling fails", async () => {
    const onEvent = vi.fn();
    const onStatusChange = vi.fn();
    const onClose = vi.fn();
    const client = createTestClient({ onEvent, onStatusChange });
    client.onClose(onClose);
    const websocket = FakeWebSocket.instances[0];
    const peerConnection = FakePeerConnection.instances[0];
    peerConnection.setRemoteDescription.mockRejectedValueOnce(new Error("bad sdp"));

    await websocket.onmessage({
      data: JSON.stringify({
        type: "offer",
        payload: {
          fromPeerId: "sfu",
          sdp: "bad-offer"
        }
      })
    });

    expect(onStatusChange).toHaveBeenCalledWith(LiveSessionStatus.SIGNALING_ERROR);
    expect(onEvent).toHaveBeenCalledWith("관제 연결 오류: bad sdp");
    expect(websocket.closeReason).toBe(LiveCloseReason.CONNECTION_FAILED);
    expect(peerConnection.closed).toBe(true);
    expect(onClose).toHaveBeenCalledWith(expect.objectContaining({
      reason: LiveCloseReason.CONNECTION_FAILED
    }));
  });

  it("stops peer tracks and removes handlers when closed", () => {
    const client = createTestClient();
    const websocket = FakeWebSocket.instances[0];
    const peerConnection = FakePeerConnection.instances[0];

    client.close(LiveCloseReason.NAVIGATION);

    expect(peerConnection.receiverTrack.stop).toHaveBeenCalledTimes(1);
    expect(peerConnection.senderTrack.stop).toHaveBeenCalledTimes(1);
    expect(peerConnection.closed).toBe(true);
    expect(peerConnection.ontrack).toBeNull();
    expect(peerConnection.onicecandidate).toBeNull();
    expect(peerConnection.oniceconnectionstatechange).toBeNull();
    expect(peerConnection.ondatachannel).toBeNull();
    expect(websocket.onopen).toBeNull();
    expect(websocket.onerror).toBeNull();
    expect(websocket.onmessage).toBeNull();
    expect(websocket.onclose).toBeNull();
  });

  it("clears data channel handlers when closed", () => {
    const client = createTestClient();
    const peerConnection = FakePeerConnection.instances[0];
    const channel = {
      label: "channel.telemetry",
      readyState: "open",
      close: vi.fn()
    };

    peerConnection.ondatachannel({ channel });
    expect(channel.onopen).toEqual(expect.any(Function));
    expect(channel.onclose).toEqual(expect.any(Function));
    expect(channel.onmessage).toEqual(expect.any(Function));

    client.close(LiveCloseReason.NAVIGATION);

    expect(channel.close).toHaveBeenCalledTimes(1);
    expect(channel.onopen).toBeNull();
    expect(channel.onclose).toBeNull();
    expect(channel.onmessage).toBeNull();
    expect(channel.onerror).toBeNull();
  });
});
