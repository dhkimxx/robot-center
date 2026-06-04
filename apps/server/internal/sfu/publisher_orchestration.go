package sfu

import (
	"fmt"
	"github.com/pion/webrtc/v4"
	"log"
	"time"
)

const publisherMediaActivityEventInterval = 10 * time.Second

func (h *Hub) handleRobotOffer(sender *peer, payload map[string]any) error {
	offerSDP := payloadString(payload, "sdp")
	if offerSDP == "" {
		return fmt.Errorf("offer sdp is empty")
	}
	robotCode, err := publisherRobotCode(sender)
	if err != nil {
		return err
	}
	if err := h.validateRobotPublisher(sender.roomID, robotCode); err != nil {
		h.sendServerSignal(sender.roomID, sender.id, "publish-error", map[string]any{
			"robotCode": robotCode,
			"reason":    "robot is not assigned to an active mission room",
		})
		return err
	}
	if normalizedOfferSDP, normalized := normalizeBundledApplicationDataChannelSDP(offerSDP); normalized {
		log.Printf("sfu robot offer sdp normalized room=%s robot=%s reason=bundled datachannel application m-line zero port", sender.roomID, robotCode)
		offerSDP = normalizedOfferSDP
	}

	peerConnection, err := h.createPeerConnection()
	if err != nil {
		return err
	}

	publisherSession := newPublisherSession(sender.id, robotCode, peerConnection)
	publisherSession.prepareConnection(sender.roomID, h)

	h.registerPublisherSession(sender.roomID, publisherSession)
	publisherAccepted := false
	defer func() {
		if !publisherAccepted {
			h.removePublisherConnection(sender.roomID, robotCode, sender.id)
		}
	}()

	localDescription, err := publisherSession.answerOffer(offerSDP)
	if err != nil {
		_ = peerConnection.Close()
		return err
	}
	publisherAccepted = true
	h.sendServerSignal(sender.roomID, sender.id, "answer", map[string]any{
		"type":      localDescription.Type.String(),
		"sdp":       localDescription.SDP,
		"robotCode": robotCode,
	})
	return nil
}

func (h *Hub) validateRobotPublisher(roomID string, robotCode string) error {
	if h.config.ValidateRobotPublisher == nil {
		return nil
	}
	return h.config.ValidateRobotPublisher(roomID, robotCode)
}

func (h *Hub) registerPublisherSession(roomID string, publisherSession *publisherSession) {
	h.mu.Lock()
	defer h.mu.Unlock()

	currentRoom := h.ensureRoomLocked(roomID)
	if existingPublisher := currentRoom.publishers[publisherSession.robotCode]; existingPublisher != nil {
		if event, ok := publisherEndedEvent(roomID, existingPublisher, "publisher_replaced", time.Now().UTC()); ok {
			h.emitPublisherEvent(event)
		}
		closePublisherSession(existingPublisher)
	}
	if publisherSession.streamBundle != nil {
		publisherSession.streamBundle.MissionCode = roomID
	}
	currentRoom.publishers[publisherSession.robotCode] = publisherSession
}

func (h *Hub) removePublisherConnection(roomID string, robotCode string, peerID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	currentRoom := h.rooms[roomID]
	if currentRoom == nil {
		return
	}
	publisher := currentRoom.publishers[robotCode]
	if publisher == nil || publisher.peerID != peerID {
		return
	}
	if event, ok := publisherEndedEvent(roomID, publisher, "peer_closed", time.Now().UTC()); ok {
		h.emitPublisherEvent(event)
	}
	closePublisherSession(publisher)
	delete(currentRoom.publishers, robotCode)
}

func (h *Hub) publishRobotTrack(roomID string, robotCode string, remoteTrack *webrtc.TrackRemote) {
	var label string
	var publisherPeerID string
	h.mu.RLock()
	if currentRoom := h.rooms[roomID]; currentRoom != nil {
		if publisher := currentRoom.publishers[robotCode]; publisher != nil {
			label = normalizeTrackRole(remoteTrack, publisher.publishedTracks)
			publisherPeerID = publisher.peerID
		}
	}
	h.mu.RUnlock()
	if label == "" {
		label = normalizeTrackRole(remoteTrack, nil)
	}
	trackKey := publishedTrackKey(robotCode, label)
	localTrack, err := webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, localTrackID(robotCode, label), localStreamID(robotCode))
	if err != nil {
		log.Printf("sfu local track create failed room=%s robot=%s label=%s: %v", roomID, robotCode, label, err)
		return
	}

	h.mu.Lock()
	currentRoom := h.ensureRoomLocked(roomID)
	publisher := currentRoom.publishers[robotCode]
	if publisher == nil {
		h.mu.Unlock()
		log.Printf("sfu publisher missing room=%s robot=%s label=%s", roomID, robotCode, label)
		return
	}
	if publisherPeerID == "" {
		publisherPeerID = publisher.peerID
	}
	publishedTrack := &publishedTrack{
		key:        trackKey,
		robotCode:  robotCode,
		label:      label,
		remoteSSRC: uint32(remoteTrack.SSRC()),
		track:      localTrack,
	}
	publisher.publishedTracks[trackKey] = publishedTrack
	if publisher.streamBundle != nil {
		publisher.streamBundle.Tracks[label] = publishedTrack
	}
	now := time.Now().UTC()
	publisher.lastTrackAt = &now
	publisher.updatedAt = now
	h.mu.Unlock()

	log.Printf("sfu robot track published room=%s robot=%s label=%s key=%s kind=%s codec=%s", roomID, robotCode, label, trackKey, remoteTrack.Kind().String(), remoteTrack.Codec().MimeType)
	if !isCanonicalTrackRole(label) {
		log.Printf("sfu robot track unmapped room=%s robot=%s label=%s kind=%s stream=%s id=%s allowed=%v", roomID, robotCode, label, remoteTrack.Kind().String(), remoteTrack.StreamID(), remoteTrack.ID(), canonicalTrackRoles)
		if publisherPeerID != "" {
			h.sendServerSignal(roomID, publisherPeerID, "publish-warning", map[string]any{
				"robotCode":       robotCode,
				"warningCode":     "non_canonical_track",
				"trackLabel":      label,
				"trackId":         remoteTrack.ID(),
				"streamId":        remoteTrack.StreamID(),
				"kind":            remoteTrack.Kind().String(),
				"allowedTrackIds": canonicalTrackRoles,
				"message":         "media track msid track id must be one of the canonical robot track ids",
			})
		}
	}
	go h.forwardRTP(roomID, trackKey, remoteTrack, localTrack)
	go h.ensureRoomSubscriberOffers(roomID)
}

func (h *Hub) forwardRTP(roomID string, trackKey string, remoteTrack *webrtc.TrackRemote, localTrack *webrtc.TrackLocalStaticRTP) {
	lastObservedAt := time.Time{}
	for {
		packet, _, err := remoteTrack.ReadRTP()
		if err != nil {
			log.Printf("sfu robot track ended room=%s track=%s: %v", roomID, trackKey, err)
			return
		}
		now := time.Now().UTC()
		if now.Sub(lastObservedAt) >= time.Second {
			lastObservedAt = now
			h.markPublisherTrackActivity(roomID, trackKey, now)
		}
		if err := localTrack.WriteRTP(cloneRTPPacket(packet)); err != nil {
			log.Printf("sfu rtp forward failed room=%s track=%s: %v", roomID, trackKey, err)
		}
	}
}

func (h *Hub) forwardDataChannelMessage(roomID string, robotCode string, label string, payload []byte) {
	h.mu.RLock()
	currentRoom := h.rooms[roomID]
	if currentRoom == nil {
		h.mu.RUnlock()
		return
	}
	channels := make([]*webrtc.DataChannel, 0, len(currentRoom.subscribers))
	for _, session := range currentRoom.subscribers {
		if !session.shouldReceiveRobot(robotCode) {
			continue
		}
		channel := session.dataChannels[label]
		if channel == nil || channel.ReadyState() != webrtc.DataChannelStateOpen {
			continue
		}
		channels = append(channels, channel)
	}
	h.mu.RUnlock()

	message := string(dataChannelPayloadWithContext(roomID, robotCode, label, payload))
	for _, channel := range channels {
		if err := channel.SendText(message); err != nil {
			log.Printf("sfu datachannel send failed room=%s robot=%s label=%s: %v", roomID, robotCode, label, err)
		}
	}
}

func (h *Hub) markPublisherICEState(roomID string, robotCode string, peerID string, iceState string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	publisher := h.findPublisherLocked(roomID, robotCode, peerID)
	if publisher == nil {
		return
	}
	now := time.Now().UTC()
	publisher.iceState = iceState
	publisher.updatedAt = now
}

func (h *Hub) markPublisherTrackActivity(roomID string, trackKey string, observedAt time.Time) {
	h.mu.Lock()
	currentRoom := h.rooms[roomID]
	if currentRoom == nil {
		h.mu.Unlock()
		return
	}
	publisher, _ := currentRoom.findPublishedTrack(trackKey)
	if publisher == nil {
		h.mu.Unlock()
		return
	}
	now := observedAt.UTC()
	var event PublisherEvent
	shouldEmitEvent := false
	if publisher.firstTrackAt == nil {
		publisher.firstTrackAt = &now
		event = PublisherEvent{
			Type:            PublisherEventMediaStarted,
			RoomID:          roomID,
			RobotCode:       publisher.robotCode,
			PublisherPeerID: publisher.peerID,
			ObservedAt:      now,
		}
		shouldEmitEvent = true
		publisher.lastMediaEventAt = now
	} else if publisher.lastMediaEventAt.IsZero() || now.Sub(publisher.lastMediaEventAt) >= publisherMediaActivityEventInterval {
		event = PublisherEvent{
			Type:            PublisherEventMediaActive,
			RoomID:          roomID,
			RobotCode:       publisher.robotCode,
			PublisherPeerID: publisher.peerID,
			ObservedAt:      now,
		}
		shouldEmitEvent = true
		publisher.lastMediaEventAt = now
	}
	publisher.lastTrackAt = &now
	publisher.updatedAt = now
	h.mu.Unlock()

	if shouldEmitEvent {
		h.emitPublisherEvent(event)
	}
}

func (h *Hub) markPublisherDataChannel(roomID string, robotCode string, peerID string, label string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	publisher := h.findPublisherLocked(roomID, robotCode, peerID)
	if publisher == nil {
		return
	}
	now := time.Now().UTC()
	ensurePublishedDataChannel(publisher, label, now)
	publisher.updatedAt = now
}

func (h *Hub) markPublisherDataChannelOpen(roomID string, robotCode string, peerID string, label string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	publisher := h.findPublisherLocked(roomID, robotCode, peerID)
	if publisher == nil {
		return
	}
	now := time.Now().UTC()
	channel := ensurePublishedDataChannel(publisher, label, now)
	if channel != nil {
		channel.State = "open"
		channel.OpenedAt = cloneTimePointer(&now)
		channel.LastError = ""
	}
	publisher.updatedAt = now
}

func (h *Hub) markPublisherDataActivity(roomID string, robotCode string, peerID string, label string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	publisher := h.findPublisherLocked(roomID, robotCode, peerID)
	if publisher == nil {
		return
	}
	now := time.Now().UTC()
	channel := ensurePublishedDataChannel(publisher, label, now)
	if channel != nil {
		if channel.State != "closed" && channel.State != "error" {
			channel.State = "open"
		}
		channel.LastMessageAt = cloneTimePointer(&now)
		channel.MessageCount++
	}
	publisher.lastDataAt = &now
	publisher.updatedAt = now
}

func (h *Hub) markPublisherDataChannelClosed(roomID string, robotCode string, peerID string, label string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	publisher := h.findPublisherLocked(roomID, robotCode, peerID)
	if publisher == nil {
		return
	}
	now := time.Now().UTC()
	channel := ensurePublishedDataChannel(publisher, label, now)
	if channel != nil {
		channel.State = "closed"
		channel.ClosedAt = cloneTimePointer(&now)
	}
	publisher.updatedAt = now
}

func (h *Hub) markPublisherDataChannelError(roomID string, robotCode string, peerID string, label string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	publisher := h.findPublisherLocked(roomID, robotCode, peerID)
	if publisher == nil {
		return
	}
	now := time.Now().UTC()
	channel := ensurePublishedDataChannel(publisher, label, now)
	if channel != nil {
		if channel.State != "closed" {
			channel.State = "error"
		}
		if err != nil {
			channel.LastError = err.Error()
		}
	}
	publisher.updatedAt = now
}

func (h *Hub) closePublisherForPeerLocked(currentRoom *room, peerID string, reason string) {
	for robotCode, publisher := range currentRoom.publishers {
		if publisher.peerID != peerID {
			continue
		}
		if event, ok := publisherEndedEvent(currentRoom.id, publisher, reason, time.Now().UTC()); ok {
			h.emitPublisherEvent(event)
		}
		closePublisherSession(publisher)
		delete(currentRoom.publishers, robotCode)
	}
}

func (h *Hub) closePublisherConnectionsLocked(currentRoom *room, reason string) {
	for robotCode, publisher := range currentRoom.publishers {
		if event, ok := publisherEndedEvent(currentRoom.id, publisher, reason, time.Now().UTC()); ok {
			h.emitPublisherEvent(event)
		}
		closePublisherSession(publisher)
		delete(currentRoom.publishers, robotCode)
	}
}

func closePublisherSession(session *publisherSession) {
	if session == nil {
		return
	}
	if session.peerConnection != nil {
		_ = session.peerConnection.Close()
	}
	session.peerConnection = nil
	session.publishedTracks = map[string]*publishedTrack{}
}
