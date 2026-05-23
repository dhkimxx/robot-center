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
	rgbA := publishedTrackKey("robot-001", "rgb")
	rgbB := publishedTrackKey("robot-002", "rgb")
	if rgbA == rgbB {
		t.Fatalf("expected robot-scoped track keys, got %q and %q", rgbA, rgbB)
	}
	if localTrackID("robot-001", "thermal") != "robot-001-thermal" {
		t.Fatalf("expected robot-scoped local track id")
	}
	if localStreamID("robot-001") != "robot-robot-001" {
		t.Fatalf("expected robot-scoped local stream id")
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
