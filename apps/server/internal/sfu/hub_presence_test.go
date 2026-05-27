package sfu

import (
	"net/http/httptest"
	"strings"
	"testing"
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
	if len(summary.Peers) != 1 || summary.Peers[0].RobotCode != "" || summary.Peers[0].SelectedRobotCode != "" {
		t.Fatalf("expected operator query robotCode to be ignored, got %#v", summary.Peers)
	}

	hub.mu.RLock()
	currentRoom := hub.rooms["mission-001"]
	if currentRoom == nil ||
		currentRoom.peers[summary.Peers[0].PeerID].robotCode != "" ||
		currentRoom.peers[summary.Peers[0].PeerID].selectedRobotCode != "" {
		t.Fatalf("expected operator peer robotCode and selectedRobotCode to stay empty")
	}
	hub.mu.RUnlock()
}
