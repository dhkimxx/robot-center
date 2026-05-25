package sfu

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestHubAnnouncesServerPeerAndRoomPeers(t *testing.T) {
	server := httptest.NewServer(NewHub())
	defer server.Close()

	websocketURL := "ws" + strings.TrimPrefix(server.URL, "http")
	robot := dialPeer(t, websocketURL+"?room=mission-001&role=robot&robotCode=robot-001")
	defer robot.Close()
	operator := dialPeer(t, websocketURL+"?room=mission-001&role=operator")
	defer operator.Close()

	if message := readMessage(t, robot); message.Type != "joined" {
		t.Fatalf("expected robot joined message, got %#v", message)
	}
	if message := readMessage(t, robot); message.Type != "peer-present" || message.Payload["role"] != "sfu" || message.Payload["peerId"] != serverPeerID {
		t.Fatalf("expected robot peer-present sfu, got %#v", message)
	}
	if message := readMessage(t, operator); message.Type != "joined" {
		t.Fatalf("expected operator joined message, got %#v", message)
	}
	if message := readMessage(t, operator); message.Type != "peer-present" || message.Payload["role"] != "sfu" || message.Payload["peerId"] != serverPeerID {
		t.Fatalf("expected operator peer-present sfu, got %#v", message)
	}
	if message := readMessage(t, operator); message.Type != "peer-present" || message.Payload["role"] != "robot" || message.Payload["robotCode"] != "robot-001" {
		t.Fatalf("expected operator peer-present robot, got %#v", message)
	}
	if message := readMessage(t, robot); message.Type != "peer-joined" || message.Payload["role"] != "operator" {
		t.Fatalf("expected robot peer-joined operator, got %#v", message)
	}
}

func TestOperatorQueryRobotCodeDoesNotPreselectRobot(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(hub)
	defer server.Close()

	websocketURL := "ws" + strings.TrimPrefix(server.URL, "http")
	operator := dialPeer(t, websocketURL+"?room=mission-001&role=operator&robotCode=robot-001")
	defer operator.Close()

	if message := readMessage(t, operator); message.Type != "joined" {
		t.Fatalf("expected operator joined message, got %#v", message)
	}
	if message := readMessage(t, operator); message.Type != "peer-present" || message.Payload["role"] != "sfu" {
		t.Fatalf("expected operator peer-present sfu, got %#v", message)
	}

	summary := waitForRoomSummary(t, hub, "mission-001")
	if len(summary.Peers) != 1 || summary.Peers[0].RobotCode != "" {
		t.Fatalf("expected operator query robotCode to be ignored, got %#v", summary.Peers)
	}

	hub.mu.RLock()
	currentRoom := hub.rooms["mission-001"]
	if currentRoom == nil || currentRoom.peers[summary.Peers[0].PeerID].robotCode != "" {
		t.Fatalf("expected operator peer robotCode to stay empty")
	}
	hub.mu.RUnlock()
}

func TestNewOperatorSubscriberSessionIgnoresPeerRobotCode(t *testing.T) {
	hub := NewHub()
	roomID := "mission-001"
	operatorPeer := testPeer("operator-peer", roomID, "operator", "robot-001")

	hub.mu.Lock()
	hub.rooms[roomID] = &room{
		id: roomID,
		peers: map[string]*peer{
			operatorPeer.id: operatorPeer,
		},
		publishers: map[string]*publisherSession{
			"robot-001": {
				robotCode: "robot-001",
				publishedTracks: map[string]*publishedTrack{
					publishedTrackKey("robot-001", StreamRoleTrackVideo1): {
						key:       publishedTrackKey("robot-001", StreamRoleTrackVideo1),
						robotCode: "robot-001",
						label:     StreamRoleTrackVideo1,
					},
				},
			},
		},
		subscribers: map[string]*subscriberSession{},
	}
	hub.mu.Unlock()

	hub.ensureSubscriberOffer(roomID, operatorPeer.id)

	hub.mu.RLock()
	session := hub.rooms[roomID].subscribers[operatorPeer.id]
	hub.mu.RUnlock()
	if session == nil {
		t.Fatal("expected subscriber session to be created")
	}
	if session.selectedRobotCode != "" {
		t.Fatalf("selectedRobotCode = %q, want empty before select-robot ack", session.selectedRobotCode)
	}
}

func TestHubSummarizesMultiPublisherMissionRoom(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(hub)
	defer server.Close()

	websocketURL := "ws" + strings.TrimPrefix(server.URL, "http")
	robotA := dialPeer(t, websocketURL+"?room=mission-001&role=robot&robotCode=robot-001")
	defer robotA.Close()
	robotB := dialPeer(t, websocketURL+"?room=mission-001&role=robot&robotCode=robot-002")
	defer robotB.Close()
	operator := dialPeer(t, websocketURL+"?room=mission-001&role=operator")
	defer operator.Close()
	recorder := dialPeer(t, websocketURL+"?room=mission-001&role=recorder")
	defer recorder.Close()

	summary := waitForRoomSummary(t, hub, "mission-001")
	if summary.RobotCount != 2 || summary.OperatorCount != 1 || summary.RecorderCount != 1 {
		t.Fatalf("expected 2 robots, 1 operator, 1 recorder, got %#v", summary)
	}
	if !summaryHasRobot(summary, "robot-001") || !summaryHasRobot(summary, "robot-002") {
		t.Fatalf("expected robot codes in peers, got %#v", summary.Peers)
	}
}

func TestHubPublishedTrackKeysAreRobotScoped(t *testing.T) {
	rgbA := publishedTrackKey("robot-001", StreamRoleTrackVideo1)
	rgbB := publishedTrackKey("robot-002", StreamRoleTrackVideo1)
	if rgbA == rgbB {
		t.Fatalf("expected robot-scoped track keys, got %q and %q", rgbA, rgbB)
	}
	if localTrackID("robot-001", StreamRoleTrackVideo2) != "robot-001-track.video_2" {
		t.Fatalf("expected robot-scoped local track id")
	}
	if localStreamID("robot-001") != "robot-robot-001" {
		t.Fatalf("expected robot-scoped local stream id")
	}
}

func TestHubRejectsSelectingRobotWithoutPublishedBundle(t *testing.T) {
	hub := NewHub()
	roomID := "mission-001"
	operatorPeer := testPeer("operator-peer", roomID, "operator", "")

	hub.mu.Lock()
	hub.rooms[roomID] = &room{
		id: roomID,
		peers: map[string]*peer{
			operatorPeer.id: operatorPeer,
		},
		publishers:  map[string]*publisherSession{},
		subscribers: map[string]*subscriberSession{},
	}
	hub.mu.Unlock()

	err := hub.handleSubscriberRobotSelection(operatorPeer, map[string]any{"robotCode": "robot-missing"})
	if err == nil {
		t.Fatal("expected missing robot selection to fail")
	}

	message := readPeerSignal(t, operatorPeer)
	if message.Type != "select-robot-error" || message.Payload["robotCode"] != "robot-missing" {
		t.Fatalf("expected select-robot-error for missing robot, got %#v", message)
	}
}

func TestHubAcknowledgesSelectingPublishedRobotBundle(t *testing.T) {
	hub := NewHub()
	roomID := "mission-001"
	operatorPeer := testPeer("operator-peer", roomID, "operator", "")
	robotPeer := testPeer("robot-peer", roomID, "robot", "robot-001")
	bundle := newRobotStreamBundle(roomID, "robot-001")
	trackKey := publishedTrackKey("robot-001", StreamRoleTrackVideo1)
	track := &publishedTrack{key: trackKey, robotCode: "robot-001", label: StreamRoleTrackVideo1}
	bundle.Tracks[StreamRoleTrackVideo1] = track

	hub.mu.Lock()
	hub.rooms[roomID] = &room{
		id: roomID,
		peers: map[string]*peer{
			operatorPeer.id: operatorPeer,
			robotPeer.id:    robotPeer,
		},
		publishers: map[string]*publisherSession{
			"robot-001": {
				peerID:          robotPeer.id,
				robotCode:       "robot-001",
				streamBundle:    bundle,
				publishedTracks: map[string]*publishedTrack{trackKey: track},
			},
		},
		subscribers: map[string]*subscriberSession{},
	}
	hub.mu.Unlock()

	if err := hub.handleSubscriberRobotSelection(operatorPeer, map[string]any{"robotCode": "robot-001"}); err != nil {
		t.Fatalf("expected published robot selection to succeed: %v", err)
	}

	message := readPeerSignal(t, operatorPeer)
	if message.Type != "select-robot-ack" || message.Payload["robotCode"] != "robot-001" {
		t.Fatalf("expected select-robot-ack, got %#v", message)
	}
}

func TestHubSelectionErrorsIncludeReason(t *testing.T) {
	hub := NewHub()
	operatorPeer := testPeer("operator-peer", "mission-001", "operator", "")

	err := hub.handleSubscriberRobotSelection(operatorPeer, map[string]any{})
	if err == nil {
		t.Fatal("expected blank robot selection to fail")
	}
	message := readPeerSignal(t, operatorPeer)
	if message.Type != "select-robot-error" || message.Payload["reason"] == "" {
		t.Fatalf("expected select-robot-error reason, got %#v", message)
	}
}

func TestHubRejectsRecorderRobotSelection(t *testing.T) {
	hub := NewHub()
	roomID := "mission-001"
	recorderPeer := testPeer("recorder-peer", roomID, "recorder", "")

	hub.mu.Lock()
	hub.rooms[roomID] = &room{
		id: roomID,
		peers: map[string]*peer{
			recorderPeer.id: recorderPeer,
		},
		publishers:  map[string]*publisherSession{},
		subscribers: map[string]*subscriberSession{},
	}
	hub.mu.Unlock()

	err := hub.handleSubscriberRobotSelection(recorderPeer, map[string]any{"robotCode": "robot-001"})
	if err == nil {
		t.Fatal("expected recorder robot selection to fail")
	}
	message := readPeerSignal(t, recorderPeer)
	if message.Type != "select-robot-error" || message.Payload["robotCode"] != "robot-001" {
		t.Fatalf("expected recorder select-robot-error, got %#v", message)
	}
}

func TestHubSubscriberLeaveDoesNotCloseOtherRoomSessions(t *testing.T) {
	hub := NewHub()
	roomID := "mission-001"
	robotPeer := testPeer("robot-peer", roomID, "robot", "robot-001")
	operatorPeer := testPeer("operator-peer", roomID, "operator", "")
	recorderPeer := testPeer("recorder-peer", roomID, "recorder", "")

	hub.mu.Lock()
	hub.rooms[roomID] = &room{
		id: roomID,
		peers: map[string]*peer{
			robotPeer.id:    robotPeer,
			operatorPeer.id: operatorPeer,
			recorderPeer.id: recorderPeer,
		},
		publishers: map[string]*publisherSession{
			"robot-001": {
				peerID:          robotPeer.id,
				robotCode:       "robot-001",
				publishedTracks: map[string]*publishedTrack{},
			},
		},
		subscribers: map[string]*subscriberSession{
			operatorPeer.id: {peerID: operatorPeer.id},
			recorderPeer.id: {peerID: recorderPeer.id},
		},
	}
	hub.mu.Unlock()

	hub.unregisterPeer(operatorPeer)

	hub.mu.RLock()
	currentRoom := hub.rooms[roomID]
	if currentRoom == nil {
		t.Fatal("expected room to remain")
	}
	if currentRoom.publishers["robot-001"] == nil {
		t.Fatalf("expected publisher to remain after one subscriber leaves")
	}
	if currentRoom.subscribers[recorderPeer.id] == nil {
		t.Fatalf("expected other subscriber to remain")
	}
	hub.mu.RUnlock()
}

func dialPeer(t *testing.T, url string) *websocket.Conn {
	t.Helper()

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", url, err)
	}
	return conn
}

func readMessage(t *testing.T, conn *websocket.Conn) signalMessage {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	var message signalMessage
	if err := conn.ReadJSON(&message); err != nil {
		t.Fatalf("read message: %v", err)
	}
	return message
}

func readPeerSignal(t *testing.T, peer *peer) signalMessage {
	t.Helper()

	select {
	case message := <-peer.send:
		return message
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for peer signal")
		return signalMessage{}
	}
}

func waitForRoomSummary(t *testing.T, hub *Hub, roomID string) RoomSummary {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		for _, summary := range hub.Summaries() {
			if summary.RoomID == roomID {
				return summary
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("room %s summary not found", roomID)
	return RoomSummary{}
}

func summaryHasRobot(summary RoomSummary, robotCode string) bool {
	for _, peer := range summary.Peers {
		if peer.Role == "robot" && peer.RobotCode == robotCode {
			return true
		}
	}
	return false
}

func testPeer(peerID string, roomID string, role string, robotCode string) *peer {
	return &peer{
		id:        peerID,
		roomID:    roomID,
		role:      role,
		robotCode: robotCode,
		joinedAt:  time.Now().UTC(),
		send:      make(chan signalMessage, 8),
	}
}
