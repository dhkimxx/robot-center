package sfu

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestHubRejectsRobotOfferWhenPublisherGuardFails(t *testing.T) {
	guardError := errors.New("inactive mission assignment")
	hub := NewHub(Config{
		ValidateRobotPublisher: func(roomID string, robotCode string) error {
			if roomID != "mission-ended" || robotCode != "robot-001" {
				t.Fatalf("guard received room=%q robot=%q", roomID, robotCode)
			}
			return guardError
		},
	})
	robotPeer := testPeer("robot-peer", "mission-ended", "robot", "robot-001")

	hub.mu.Lock()
	hub.rooms[robotPeer.roomID] = &room{
		id: robotPeer.roomID,
		peers: map[string]*peer{
			robotPeer.id: robotPeer,
		},
		publishers:  map[string]*publisherSession{},
		subscribers: map[string]*subscriberSession{},
	}
	hub.mu.Unlock()

	err := hub.handleRobotOffer(robotPeer, map[string]any{"sdp": "fake-offer"})
	if !errors.Is(err, guardError) {
		t.Fatalf("handleRobotOffer error = %v, want %v", err, guardError)
	}
	message := readPeerSignal(t, robotPeer)
	if message.Type != "publish-error" || message.Payload["robotCode"] != "robot-001" {
		t.Fatalf("expected publish-error for guarded robot offer, got %#v", message)
	}

	hub.mu.RLock()
	publisherCount := len(hub.rooms[robotPeer.roomID].publishers)
	hub.mu.RUnlock()
	if publisherCount != 0 {
		t.Fatalf("expected guarded robot offer not to register publisher, got %d", publisherCount)
	}
}

func TestHubSummarizesMultiPublisherMissionRoom(t *testing.T) {
	hub := NewHub()
	server := newTestSFUServer(hub)
	defer server.Close()

	websocketURL := "ws" + strings.TrimPrefix(server.URL, "http")
	robotA := dialPeer(t, websocketURL+"/api/v1/robot/sfu/ws?room=mission-001&robotCode=robot-001")
	defer robotA.Close()
	robotB := dialPeer(t, websocketURL+"/api/v1/robot/sfu/ws?room=mission-001&robotCode=robot-002")
	defer robotB.Close()
	operator := dialPeer(t, websocketURL+"/sfu/operator/ws?room=mission-001")
	defer operator.Close()
	recorder := dialPeer(t, websocketURL+"/sfu/recorder/ws?room=mission-001")
	defer recorder.Close()

	summary := waitForRoomSummary(t, hub, "mission-001")
	if summary.RobotCount != 2 || summary.OperatorCount != 1 || summary.RecorderCount != 1 {
		t.Fatalf("expected 2 robots, 1 operator, 1 recorder, got %#v", summary)
	}
	if !summaryHasRobot(summary, "robot-001") || !summaryHasRobot(summary, "robot-002") {
		t.Fatalf("expected robot codes in peers, got %#v", summary.Peers)
	}
}

func TestHubObservedRoomsSummarizePublisherActivity(t *testing.T) {
	hub := NewHub()
	roomID := "mission-001"
	observedAt := time.Now().UTC()

	hub.mu.Lock()
	currentRoom := hub.ensureRoomLocked(roomID)
	publisher := newPublisherSession("robot-peer", "robot-001", nil)
	publisher.iceState = "connected"
	publisher.lastTrackAt = &observedAt
	publisher.lastDataAt = &observedAt
	publisher.updatedAt = observedAt
	publisher.publishedTracks[publishedTrackKey("robot-001", StreamRoleTrackVideo1)] = &publishedTrack{
		key:       publishedTrackKey("robot-001", StreamRoleTrackVideo1),
		robotCode: "robot-001",
		label:     StreamRoleTrackVideo1,
	}
	publisher.streamBundle.DataChannels[StreamRoleChannelTelemetry] = &PublishedDataChannel{Role: StreamRoleChannelTelemetry}
	currentRoom.publishers[publisher.robotCode] = publisher
	currentRoom.subscribers["operator-peer"] = newSubscriberSession("operator-peer", "operator", "robot-001", nil)
	currentRoom.subscribers["recorder-peer"] = newSubscriberSession("recorder-peer", "recorder", "", nil)
	hub.mu.Unlock()

	rooms := hub.ObservedRooms()
	if len(rooms) != 1 {
		t.Fatalf("expected one observed room, got %#v", rooms)
	}
	publishers := rooms[0].Publishers
	if len(publishers) != 1 {
		t.Fatalf("expected one observed publisher, got %#v", rooms[0])
	}
	publisherSummary := publishers[0]
	if publisherSummary.State != "publishing" || publisherSummary.TrackCount != 1 || publisherSummary.DataChannelCount != 1 {
		t.Fatalf("unexpected observed publisher summary: %#v", publisherSummary)
	}
	if len(publisherSummary.DataChannelStates) != 1 || publisherSummary.DataChannelStates[0].Label != StreamRoleChannelTelemetry {
		t.Fatalf("expected telemetry data channel lifecycle summary, got %#v", publisherSummary.DataChannelStates)
	}
	if publisherSummary.DataChannelStates[0].State != "detected" {
		t.Fatalf("expected default data channel state to be detected, got %#v", publisherSummary.DataChannelStates[0])
	}
	if publisherSummary.SubscriberCount != 2 {
		t.Fatalf("expected operator and recorder subscriber count, got %#v", publisherSummary)
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
