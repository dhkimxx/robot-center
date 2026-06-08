package sfu

import (
	"testing"

	"github.com/pion/webrtc/v4"
)

func TestSendServerCandidateSignals(t *testing.T) {
	hub := NewHub()
	targetPeer := testPeer("peer_operator", "mission-001", "operator", "")
	hub.rooms["mission-001"] = &room{
		id:          "mission-001",
		peers:       map[string]*peer{targetPeer.id: targetPeer},
		publishers:  map[string]*publisherSession{},
		subscribers: map[string]*subscriberSession{},
	}

	sdpMid := "0"
	sdpMLineIndex := uint16(0)
	hub.sendServerCandidateSignals("mission-001", targetPeer.id, []webrtc.ICECandidateInit{
		{},
		{
			Candidate:     "candidate:1 1 udp 2130706431 192.0.2.10 49160 typ relay",
			SDPMid:        &sdpMid,
			SDPMLineIndex: &sdpMLineIndex,
		},
	})

	message := readPeerSignal(t, targetPeer)
	if message.Type != "candidate" {
		t.Fatalf("message type = %s, want candidate", message.Type)
	}
	if message.Payload["fromRole"] != "sfu" {
		t.Fatalf("fromRole = %v, want sfu", message.Payload["fromRole"])
	}
	if message.Payload["targetPeerId"] != targetPeer.id {
		t.Fatalf("targetPeerId = %v, want %s", message.Payload["targetPeerId"], targetPeer.id)
	}
	if message.Payload["candidate"] == "" {
		t.Fatal("candidate payload should not be empty")
	}
	if message.Payload["sdpMid"] != sdpMid {
		t.Fatalf("sdpMid = %v, want %s", message.Payload["sdpMid"], sdpMid)
	}
	if message.Payload["sdpMLineIndex"] != sdpMLineIndex {
		t.Fatalf("sdpMLineIndex = %v, want %d", message.Payload["sdpMLineIndex"], sdpMLineIndex)
	}
}
