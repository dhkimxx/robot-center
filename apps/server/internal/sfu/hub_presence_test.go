package sfu

import (
	"net/http"
	"strings"
	"testing"
)

func TestGenericSFUWebSocketEndpointIsDisabled(t *testing.T) {
	server := newTestSFUServer(NewHub())
	defer server.Close()

	response, err := http.Get(server.URL + "/sfu/ws?room=mission-001&role=operator")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("expected generic /sfu/ws endpoint to be disabled, got %d", response.StatusCode)
	}
}

func TestHubAnnouncesServerPeerAndRoomPeers(t *testing.T) {
	server := newTestSFUServer(NewHub())
	defer server.Close()

	websocketURL := "ws" + strings.TrimPrefix(server.URL, "http")
	robot := dialPeer(t, websocketURL+"/api/v1/robot/sfu/ws?room=mission-001&robotCode=robot-001")
	defer robot.Close()
	operator := dialPeer(t, websocketURL+"/api/v1/operator/sfu/ws?room=mission-001")
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

func TestHubReplacesDuplicateRobotPeer(t *testing.T) {
	hub := NewHub()
	oldRobotPeer := testPeer("robot-peer-old", "mission-001", "robot", "robot-001")
	newRobotPeer := testPeer("robot-peer-new", "mission-001", "robot", "robot-001")

	if existingPeers, replacedPeers := hub.registerPeer(oldRobotPeer); len(existingPeers) != 0 || len(replacedPeers) != 0 {
		t.Fatalf("expected first robot peer to have no existing/replaced peers, got existing=%#v replaced=%#v", existingPeers, replacedPeers)
	}
	existingPeers, replacedPeers := hub.registerPeer(newRobotPeer)

	if len(existingPeers) != 0 {
		t.Fatalf("expected replaced peer to be hidden from new peer presence, got %#v", existingPeers)
	}
	if len(replacedPeers) != 1 || replacedPeers[0].id != oldRobotPeer.id {
		t.Fatalf("expected old robot peer to be replaced, got %#v", replacedPeers)
	}
	summary := hub.Summaries()[0]
	if summary.RobotCount != 1 || len(summary.Peers) != 1 || summary.Peers[0].PeerID != newRobotPeer.id {
		t.Fatalf("expected only replacement robot peer in summary, got %#v", summary)
	}
}

func TestOperatorEndpointRejectsRobotCodeQuery(t *testing.T) {
	hub := NewHub()
	server := newTestSFUServer(hub)
	defer server.Close()

	websocketURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, response, err := dialPeerResponse(t, websocketURL+"/api/v1/operator/sfu/ws?room=mission-001&robotCode=robot-001")
	if conn != nil {
		conn.Close()
	}
	if err == nil {
		t.Fatal("expected operator websocket with robotCode query to be rejected")
	}
	if response == nil || response.StatusCode != 400 {
		t.Fatalf("expected 400 response, got response=%v err=%v", response, err)
	}
}
