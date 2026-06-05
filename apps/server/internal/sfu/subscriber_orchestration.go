package sfu

import (
	"fmt"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	"log"
	"strings"
	"time"
)

func (h *Hub) ensureRoomSubscriberOffers(roomID string) {
	h.mu.RLock()
	currentRoom := h.rooms[roomID]
	if currentRoom == nil {
		h.mu.RUnlock()
		return
	}
	peerIDs := make([]string, 0, len(currentRoom.peers))
	for _, peer := range currentRoom.peers {
		if isSubscriberRole(peer.role) {
			peerIDs = append(peerIDs, peer.id)
		}
	}
	h.mu.RUnlock()

	for _, peerID := range peerIDs {
		h.ensureSubscriberOffer(roomID, peerID)
	}
}

func (h *Hub) ensureSubscriberOffer(roomID string, peerID string) {
	h.mu.Lock()
	currentRoom := h.rooms[roomID]
	if currentRoom == nil || !currentRoom.hasPublishedTracks() {
		h.mu.Unlock()
		return
	}
	targetPeer := currentRoom.peers[peerID]
	if targetPeer == nil || !isSubscriberRole(targetPeer.role) {
		h.mu.Unlock()
		return
	}
	session := currentRoom.subscribers[peerID]
	if session == nil {
		peerConnection, err := h.createPeerConnection()
		if err != nil {
			h.mu.Unlock()
			log.Printf("sfu subscriber peer connection failed room=%s peer=%s: %v", roomID, peerID, err)
			return
		}
		session = newSubscriberSession(peerID, targetPeer.role, selectedRobotCodeForNewSubscriber(targetPeer), peerConnection)
		currentRoom.subscribers[peerID] = session
		session.configureConnection(roomID, targetPeer)
	}
	if session.pendingOffer {
		session.deferOffer()
		h.mu.Unlock()
		return
	}
	if session.peerConnection.SignalingState() != webrtc.SignalingStateStable {
		session.deferOffer()
		h.mu.Unlock()
		return
	}

	offerRequired := session.beginOffer(currentRoom, h.forwardRTCPToRobot, h.requestKeyFrames)
	if !offerRequired {
		h.mu.Unlock()
		return
	}
	h.mu.Unlock()

	localDescription, err := session.createLocalOffer()
	if err != nil {
		h.markSubscriberOfferFailed(roomID, peerID, err)
		return
	}
	h.sendServerSignal(roomID, peerID, "offer", map[string]any{
		"type": localDescription.Type.String(),
		"sdp":  localDescription.SDP,
	})
}

func (h *Hub) requestKeyFrame(roomID string, trackKey string) {
	h.forwardRTCPToRobot(roomID, trackKey, []rtcp.Packet{
		&rtcp.PictureLossIndication{},
	})
}

func (h *Hub) requestKeyFrames(roomID string, trackKey string, count int, interval time.Duration) {
	for index := 0; index < count; index++ {
		h.requestKeyFrame(roomID, trackKey)
		time.Sleep(interval)
	}
}

func (h *Hub) forwardRTCPToRobot(roomID string, trackKey string, packets []rtcp.Packet) {
	h.mu.RLock()
	currentRoom := h.rooms[roomID]
	if currentRoom == nil {
		h.mu.RUnlock()
		return
	}
	publisher, publishedTrack := currentRoom.findPublishedTrack(trackKey)
	if publisher == nil || publisher.peerConnection == nil || publishedTrack == nil || publishedTrack.remoteSSRC == 0 {
		h.mu.RUnlock()
		return
	}
	publisherConnection := publisher.peerConnection
	remoteSSRC := publishedTrack.remoteSSRC
	h.mu.RUnlock()

	forwardPackets := make([]rtcp.Packet, 0, len(packets))
	for _, packet := range packets {
		switch packet.(type) {
		case *rtcp.PictureLossIndication:
			forwardPackets = append(forwardPackets, &rtcp.PictureLossIndication{MediaSSRC: remoteSSRC})
		case *rtcp.FullIntraRequest:
			forwardPackets = append(forwardPackets, &rtcp.FullIntraRequest{MediaSSRC: remoteSSRC})
		}
	}
	if len(forwardPackets) == 0 {
		return
	}
	if err := publisherConnection.WriteRTCP(forwardPackets); err != nil {
		log.Printf("sfu rtcp forward failed room=%s track=%s: %v", roomID, trackKey, err)
	}
}

func (h *Hub) markSubscriberOfferFailed(roomID string, peerID string, err error) {
	h.mu.Lock()
	if currentRoom := h.rooms[roomID]; currentRoom != nil {
		if session := currentRoom.subscribers[peerID]; session != nil {
			session.pendingOffer = false
			session.needsOffer = true
		}
	}
	h.mu.Unlock()
	log.Printf("sfu subscriber offer failed room=%s peer=%s: %v", roomID, peerID, err)
}

func (h *Hub) handleSubscriberAnswer(sender *peer, payload map[string]any) error {
	answerSDP := payloadString(payload, "sdp")
	if answerSDP == "" {
		return fmt.Errorf("answer sdp is empty")
	}

	h.mu.RLock()
	currentRoom := h.rooms[sender.roomID]
	var session *subscriberSession
	if currentRoom != nil {
		session = currentRoom.subscribers[sender.id]
	}
	h.mu.RUnlock()
	if session == nil || session.peerConnection == nil {
		return fmt.Errorf("subscriber session is missing")
	}

	if err := session.peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  answerSDP,
	}); err != nil {
		return err
	}

	h.mu.Lock()
	needsOffer := false
	var pendingCandidates []webrtc.ICECandidateInit
	if currentRoom := h.rooms[sender.roomID]; currentRoom != nil {
		if currentSession := currentRoom.subscribers[sender.id]; currentSession != nil {
			currentSession.pendingOffer = false
			needsOffer = currentSession.needsOffer
			currentSession.needsOffer = false
			pendingCandidates = currentSession.drainPendingRemoteCandidates()
		}
	}
	h.mu.Unlock()
	h.addSubscriberPendingRemoteCandidates(sender.roomID, sender.id, session.peerConnection, pendingCandidates)
	if needsOffer {
		go h.ensureSubscriberOffer(sender.roomID, sender.id)
	}
	return nil
}

func (h *Hub) addSubscriberPendingRemoteCandidates(roomID string, peerID string, peerConnection *webrtc.PeerConnection, candidates []webrtc.ICECandidateInit) {
	if peerConnection == nil || len(candidates) == 0 {
		return
	}
	for _, candidate := range candidates {
		if err := peerConnection.AddICECandidate(candidate); err != nil {
			log.Printf("sfu queued subscriber candidate ignored room=%s peer=%s: %v", roomID, peerID, err)
		}
	}
	log.Printf("sfu subscriber candidate queue flushed room=%s peer=%s count=%d", roomID, peerID, len(candidates))
}

func (h *Hub) handleSubscriberRobotSelection(sender *peer, payload map[string]any) error {
	robotCode := strings.TrimSpace(payloadString(payload, "robotCode"))
	if robotCode == "" {
		h.sendSelectRobotError(sender, robotCode, "robotCode is required")
		return fmt.Errorf("robotCode is required")
	}

	h.mu.Lock()
	currentRoom := h.rooms[sender.roomID]
	if currentRoom == nil {
		h.mu.Unlock()
		h.sendSelectRobotError(sender, robotCode, "room is missing")
		return fmt.Errorf("room is missing")
	}
	targetPeer := currentRoom.peers[sender.id]
	if targetPeer == nil || !isSubscriberRole(targetPeer.role) {
		h.mu.Unlock()
		h.sendSelectRobotError(sender, robotCode, "subscriber peer is missing")
		return fmt.Errorf("subscriber peer is missing")
	}
	if targetPeer.role == "recorder" {
		h.mu.Unlock()
		h.sendSelectRobotError(sender, robotCode, "recorder cannot select a single robot")
		return fmt.Errorf("recorder cannot select a single robot")
	}
	publisher := currentRoom.publishers[robotCode]
	if publisher == nil {
		h.mu.Unlock()
		if err := h.validateRobotSelection(sender.roomID, robotCode); err != nil {
			h.sendSelectRobotError(sender, robotCode, "robot is not assigned to this active mission room")
			return err
		}
		h.mu.Lock()
		currentRoom = h.rooms[sender.roomID]
		if currentRoom == nil {
			h.mu.Unlock()
			h.sendSelectRobotError(sender, robotCode, "room is missing")
			return fmt.Errorf("room is missing")
		}
		targetPeer = currentRoom.peers[sender.id]
		if targetPeer == nil || !isSubscriberRole(targetPeer.role) {
			h.mu.Unlock()
			h.sendSelectRobotError(sender, robotCode, "subscriber peer is missing")
			return fmt.Errorf("subscriber peer is missing")
		}
		if targetPeer.role == "recorder" {
			h.mu.Unlock()
			h.sendSelectRobotError(sender, robotCode, "recorder cannot select a single robot")
			return fmt.Errorf("recorder cannot select a single robot")
		}
		publisher = currentRoom.publishers[robotCode]
	}
	streamState := "waiting_for_publisher"
	if publisher != nil && publisher.streamBundle != nil && len(publisher.streamBundle.Tracks) > 0 {
		streamState = "publishing"
	}
	targetPeer.selectedRobotCode = robotCode
	if session := currentRoom.subscribers[sender.id]; session != nil {
		session.selectRobot(robotCode)
		session.deferOffer()
	}
	h.mu.Unlock()

	h.sendServerSignal(sender.roomID, sender.id, "select-robot-ack", map[string]any{
		"robotCode":   robotCode,
		"streamState": streamState,
	})
	go h.ensureSubscriberOffer(sender.roomID, sender.id)
	return nil
}

func (h *Hub) validateRobotSelection(roomID string, robotCode string) error {
	if h.config.ValidateRobotSelection != nil {
		return h.config.ValidateRobotSelection(roomID, robotCode)
	}
	if h.config.ValidateRobotPublisher != nil {
		return h.config.ValidateRobotPublisher(roomID, robotCode)
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	if currentRoom := h.rooms[roomID]; currentRoom != nil {
		if publisher := currentRoom.publishers[robotCode]; publisher != nil {
			return nil
		}
	}
	return fmt.Errorf("robot %s is not publishing in room %s", robotCode, roomID)
}

func (h *Hub) sendSelectRobotError(targetPeer *peer, robotCode string, reason string) {
	if targetPeer == nil {
		return
	}
	targetPeer.enqueue(signalMessage{
		Type: "select-robot-error",
		Payload: map[string]any{
			"room":         targetPeer.roomID,
			"fromRole":     "sfu",
			"fromPeerId":   serverPeerID,
			"targetPeerId": targetPeer.id,
			"robotCode":    robotCode,
			"reason":       reason,
		},
	})
}

func (h *Hub) closeSubscriberConnectionsLocked(currentRoom *room) {
	for _, session := range currentRoom.subscribers {
		closeSubscriberSession(session)
	}
	currentRoom.subscribers = map[string]*subscriberSession{}
}

func closeSubscriberSession(session *subscriberSession) {
	if session == nil {
		return
	}
	if session.peerConnection != nil {
		_ = session.peerConnection.Close()
	}
	session.peerConnection = nil
	session.dataChannels = map[string]*webrtc.DataChannel{}
	session.attachedTrackSources = map[string]*webrtc.TrackLocalStaticRTP{}
	session.attachedTrackSenders = map[string]*webrtc.RTPSender{}
	session.pendingRemoteCandidates = nil
	session.pendingOffer = false
	session.needsOffer = false
}
