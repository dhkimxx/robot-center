package sfu

import (
	"testing"

	"github.com/pion/webrtc/v4"
)

func TestOperatorSubscriberDoesNotAutoSelectRobot(t *testing.T) {
	session := &subscriberSession{role: "operator"}
	currentRoom := &room{
		publishers: map[string]*publisherSession{
			"robot-002": {},
			"robot-001": {},
		},
	}

	if session.beginOffer(currentRoom, nil, nil) {
		t.Fatalf("operator should not create an offer before selecting a robot")
	}
	if session.selectedRobotCode != "" {
		t.Fatalf("selectedRobotCode = %q, want empty before explicit selection", session.selectedRobotCode)
	}
	if session.shouldReceiveRobot("robot-001") || session.shouldReceiveRobot("robot-002") {
		t.Fatalf("operator should not receive robot streams before explicit selection")
	}
}

func TestOperatorSubscriberReceivesOnlySelectedRobot(t *testing.T) {
	session := &subscriberSession{role: "operator"}

	session.selectRobot("robot-001")
	if !session.shouldReceiveRobot("robot-001") {
		t.Fatalf("operator should receive the selected robot")
	}
	if session.shouldReceiveRobot("robot-002") {
		t.Fatalf("operator should not receive an unselected robot")
	}

	session.selectRobot("robot-002")
	if !session.shouldReceiveRobot("robot-002") {
		t.Fatalf("operator should receive the newly selected robot")
	}
	if session.shouldReceiveRobot("robot-001") {
		t.Fatalf("operator should stop receiving the previous robot")
	}
}

func TestRecorderSubscriberReceivesEveryRobot(t *testing.T) {
	session := &subscriberSession{role: "recorder"}
	session.selectRobot("robot-001")

	if !session.shouldReceiveRobot("robot-001") || !session.shouldReceiveRobot("robot-002") {
		t.Fatalf("recorder should receive all robot streams")
	}
}

func TestCanonicalDataChannelRoles(t *testing.T) {
	cases := map[string]string{
		"channel.telemetry": StreamRoleChannelTelemetry,
		"channel.event":     StreamRoleChannelEvent,
		"channel.spatial":   StreamRoleChannelSpatial,
		"channel.control":   StreamRoleChannelControl,
		"telemetry":         "telemetry",
		"sensor":            "sensor",
		"event":             "event",
		"spatial":           "spatial",
		"control":           "control",
	}
	for input, want := range cases {
		if got := normalizeDataChannelRole(input); got != want {
			t.Fatalf("normalizeDataChannelRole(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestHubQueuesSubscriberCandidateBeforeRemoteDescription(t *testing.T) {
	hub := NewHub()
	roomID := "mission-001"
	operatorPeer := testPeer("operator-peer", roomID, "operator", "")
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		t.Fatal(err)
	}
	defer peerConnection.Close()

	session := newSubscriberSession(operatorPeer.id, "operator", "robot-001", peerConnection)
	hub.mu.Lock()
	hub.rooms[roomID] = &room{
		id: roomID,
		peers: map[string]*peer{
			operatorPeer.id: operatorPeer,
		},
		publishers: map[string]*publisherSession{},
		subscribers: map[string]*subscriberSession{
			operatorPeer.id: session,
		},
	}
	hub.mu.Unlock()

	err = hub.handleRemoteCandidate(operatorPeer, map[string]any{
		"candidate":     "candidate:0 1 udp 2122252543 192.0.2.1 3478 typ host",
		"sdpMid":        "0",
		"sdpMLineIndex": float64(0),
	})
	if err != nil {
		t.Fatalf("expected early subscriber candidate to be queued, got error: %v", err)
	}
	if len(session.pendingRemoteCandidates) != 1 {
		t.Fatalf("pending candidate count = %d, want 1", len(session.pendingRemoteCandidates))
	}
	if session.pendingRemoteCandidates[0].Candidate == "" {
		t.Fatalf("expected candidate to be preserved")
	}
}

func TestSubscriberSessionDrainsPendingRemoteCandidates(t *testing.T) {
	session := &subscriberSession{}
	for i := 0; i < maxPendingRemoteCandidates+1; i++ {
		session.queueRemoteCandidate(webrtc.ICECandidateInit{Candidate: "candidate"})
	}
	if len(session.pendingRemoteCandidates) != maxPendingRemoteCandidates {
		t.Fatalf("pending candidate count = %d, want capped %d", len(session.pendingRemoteCandidates), maxPendingRemoteCandidates)
	}

	candidates := session.drainPendingRemoteCandidates()
	if len(candidates) != maxPendingRemoteCandidates {
		t.Fatalf("drained candidate count = %d, want %d", len(candidates), maxPendingRemoteCandidates)
	}
	if len(session.pendingRemoteCandidates) != 0 {
		t.Fatalf("pending candidate queue should be empty after drain")
	}
}
