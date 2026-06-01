package sfu

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func newTestSFUServer(hub *Hub) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/robot/sfu/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.ServePeer(w, r, PeerJoinRequest{
			RoomID:    r.URL.Query().Get("room"),
			Role:      "robot",
			RobotCode: r.URL.Query().Get("robotCode"),
		})
	})
	mux.HandleFunc("GET /sfu/operator/ws", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("robotCode") != "" {
			http.Error(w, "robotCode query is not allowed for operator websocket", http.StatusBadRequest)
			return
		}
		hub.ServePeer(w, r, PeerJoinRequest{
			RoomID: r.URL.Query().Get("room"),
			Role:   "operator",
		})
	})
	mux.HandleFunc("GET /sfu/recorder/ws", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("robotCode") != "" {
			http.Error(w, "robotCode query is not allowed for recorder websocket", http.StatusBadRequest)
			return
		}
		hub.ServePeer(w, r, PeerJoinRequest{
			RoomID: r.URL.Query().Get("room"),
			Role:   "recorder",
		})
	})
	return httptest.NewServer(mux)
}

func dialPeer(t *testing.T, url string) *websocket.Conn {
	t.Helper()

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", url, err)
	}
	return conn
}

func dialPeerResponse(t *testing.T, url string) (*websocket.Conn, *http.Response, error) {
	t.Helper()

	return websocket.DefaultDialer.Dial(url, nil)
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

func waitForSubscriberSelection(t *testing.T, hub *Hub, roomID string, peerID string, robotCode string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		hub.mu.RLock()
		currentRoom := hub.rooms[roomID]
		var selectedRobotCode string
		if currentRoom != nil && currentRoom.subscribers[peerID] != nil {
			selectedRobotCode = currentRoom.subscribers[peerID].selectedRobotCode
		}
		hub.mu.RUnlock()
		if selectedRobotCode == robotCode {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("subscriber %s did not select %s", peerID, robotCode)
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
