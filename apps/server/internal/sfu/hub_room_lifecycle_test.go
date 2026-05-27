package sfu

import (
	"testing"
)

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
