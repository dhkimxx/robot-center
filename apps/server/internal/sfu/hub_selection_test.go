package sfu

import (
	"testing"
)

func TestNewOperatorSubscriberSessionUsesValidatedSelectedRobotCode(t *testing.T) {
	hub := NewHub()
	roomID := "mission-001"
	operatorPeer := testPeer("operator-peer", roomID, "operator", "")
	operatorPeer.selectedRobotCode = "robot-001"

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
	if session.selectedRobotCode != "robot-001" {
		t.Fatalf("selectedRobotCode = %q, want selected robot from validated peer state", session.selectedRobotCode)
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

func TestHubStoresValidatedSelectionBeforeRobotPublishes(t *testing.T) {
	hub := NewHub(Config{
		ValidateRobotSelection: func(roomID string, robotCode string) error {
			if roomID != "mission-001" || robotCode != "robot-001" {
				t.Fatalf("selection validator received room=%q robot=%q", roomID, robotCode)
			}
			return nil
		},
	})
	roomID := "mission-001"
	operatorPeer := testPeer("operator-peer", roomID, "operator", "")
	robotCode := "robot-001"
	trackKey := publishedTrackKey(robotCode, StreamRoleTrackVideo1)

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

	if err := hub.handleSubscriberRobotSelection(operatorPeer, map[string]any{"robotCode": robotCode}); err != nil {
		t.Fatalf("expected validated robot selection to be stored while publisher is absent: %v", err)
	}
	message := readPeerSignal(t, operatorPeer)
	if message.Type != "select-robot-ack" || message.Payload["robotCode"] != robotCode || message.Payload["streamState"] != "waiting_for_publisher" {
		t.Fatalf("expected waiting select-robot-ack, got %#v", message)
	}
	hub.mu.RLock()
	selectedRobotCode := hub.rooms[roomID].peers[operatorPeer.id].selectedRobotCode
	hub.mu.RUnlock()
	if selectedRobotCode != robotCode {
		t.Fatalf("selectedRobotCode = %q, want %q", selectedRobotCode, robotCode)
	}

	hub.mu.Lock()
	currentRoom := hub.rooms[roomID]
	currentRoom.publishers[robotCode] = &publisherSession{
		robotCode: robotCode,
		publishedTracks: map[string]*publishedTrack{
			trackKey: {
				key:       trackKey,
				robotCode: robotCode,
				label:     StreamRoleTrackVideo1,
			},
		},
	}
	hub.mu.Unlock()

	hub.ensureSubscriberOffer(roomID, operatorPeer.id)
	waitForSubscriberSelection(t, hub, roomID, operatorPeer.id, robotCode)
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
	hub.mu.RLock()
	selectedRobotCode := hub.rooms[roomID].peers[operatorPeer.id].selectedRobotCode
	operatorRobotCode := hub.rooms[roomID].peers[operatorPeer.id].robotCode
	hub.mu.RUnlock()
	if operatorRobotCode != "" || selectedRobotCode != "robot-001" {
		t.Fatalf("operator state robotCode=%q selectedRobotCode=%q, want identity empty and selection robot-001", operatorRobotCode, selectedRobotCode)
	}
	waitForSubscriberSelection(t, hub, roomID, operatorPeer.id, "robot-001")
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
