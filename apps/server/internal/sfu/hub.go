package sfu

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/interceptor"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
)

const serverPeerID = "sfu"

func NewHub(configs ...Config) *Hub {
	cfg := Config{}
	if len(configs) > 0 {
		cfg = configs[0]
	}
	return &Hub{
		config: cfg,
		rooms:  map[string]*room{},
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		},
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	roomID := strings.TrimSpace(r.URL.Query().Get("room"))
	role := strings.TrimSpace(r.URL.Query().Get("role"))
	robotCode := strings.TrimSpace(r.URL.Query().Get("robotCode"))
	if roomID == "" || role == "" {
		http.Error(w, "room and role query parameters are required", http.StatusBadRequest)
		return
	}
	if role == "robot" && robotCode == "" {
		http.Error(w, "robotCode query parameter is required for robot peers", http.StatusBadRequest)
		return
	}
	if role != "robot" {
		robotCode = ""
	}

	connection, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	peer := &peer{
		id:        "peer_" + randomHex(8),
		roomID:    roomID,
		role:      role,
		robotCode: robotCode,
		joinedAt:  time.Now().UTC(),
		conn:      connection,
		send:      make(chan signalMessage, 32),
	}

	existingPeers := h.registerPeer(peer)
	go peer.writeLoop()

	peer.send <- signalMessage{
		Type:    "joined",
		Payload: peerPresencePayload(peer),
	}
	peer.send <- serverPeerPresentMessage(peer.roomID)

	for _, existingPeer := range existingPeers {
		peer.send <- signalMessage{
			Type:    "peer-present",
			Payload: peerPresencePayload(existingPeer),
		}
	}

	h.broadcast(peer, signalMessage{
		Type:    "peer-joined",
		Payload: peerPresencePayload(peer),
	})

	if isSubscriberRole(peer.role) {
		go h.ensureSubscriberOffer(peer.roomID, peer.id)
	}

	peer.readLoop(h)
}

func (h *Hub) handleSignal(sender *peer, message signalMessage) {
	if message.Payload == nil {
		message.Payload = map[string]any{}
	}
	message.Payload["room"] = sender.roomID
	message.Payload["fromRole"] = sender.role
	message.Payload["fromPeerId"] = sender.id

	switch message.Type {
	case "offer":
		if sender.role == "robot" && isTargetingServer(message.Payload) {
			if err := h.handleRobotOffer(sender, message.Payload); err != nil {
				log.Printf("sfu robot offer failed room=%s peer=%s: %v", sender.roomID, sender.id, err)
			}
			return
		}
	case "answer":
		if isSubscriberRole(sender.role) && isTargetingServer(message.Payload) {
			if err := h.handleSubscriberAnswer(sender, message.Payload); err != nil {
				log.Printf("sfu subscriber answer failed room=%s peer=%s: %v", sender.roomID, sender.id, err)
			}
			return
		}
	case "candidate":
		if isTargetingServer(message.Payload) {
			if err := h.handleRemoteCandidate(sender, message.Payload); err != nil {
				log.Printf("sfu candidate ignored room=%s peer=%s: %v", sender.roomID, sender.id, err)
			}
			return
		}
	case "select-robot":
		if isSubscriberRole(sender.role) {
			if err := h.handleSubscriberRobotSelection(sender, message.Payload); err != nil {
				log.Printf("sfu subscriber selection failed room=%s peer=%s: %v", sender.roomID, sender.id, err)
			}
			return
		}
	}
}

func (h *Hub) handleRobotOffer(sender *peer, payload map[string]any) error {
	offerSDP := payloadString(payload, "sdp")
	if offerSDP == "" {
		return fmt.Errorf("offer sdp is empty")
	}
	robotCode, err := publisherRobotCode(sender, payload)
	if err != nil {
		return err
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

func (h *Hub) registerPublisherSession(roomID string, publisherSession *publisherSession) {
	h.mu.Lock()
	defer h.mu.Unlock()

	currentRoom := h.ensureRoomLocked(roomID)
	if existingPublisher := currentRoom.publishers[publisherSession.robotCode]; existingPublisher != nil {
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
	closePublisherSession(publisher)
	delete(currentRoom.publishers, robotCode)
}

func (h *Hub) publishRobotTrack(roomID string, robotCode string, remoteTrack *webrtc.TrackRemote) {
	var label string
	h.mu.RLock()
	if currentRoom := h.rooms[roomID]; currentRoom != nil {
		if publisher := currentRoom.publishers[robotCode]; publisher != nil {
			label = normalizeTrackRole(remoteTrack, publisher.publishedTracks)
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
	h.mu.Unlock()

	log.Printf("sfu robot track published room=%s robot=%s label=%s key=%s kind=%s codec=%s", roomID, robotCode, label, trackKey, remoteTrack.Kind().String(), remoteTrack.Codec().MimeType)
	go h.forwardRTP(roomID, trackKey, remoteTrack, localTrack)
	go h.ensureRoomSubscriberOffers(roomID)
}

func (h *Hub) forwardRTP(roomID string, trackKey string, remoteTrack *webrtc.TrackRemote, localTrack *webrtc.TrackLocalStaticRTP) {
	for {
		packet, _, err := remoteTrack.ReadRTP()
		if err != nil {
			log.Printf("sfu robot track ended room=%s track=%s: %v", roomID, trackKey, err)
			return
		}
		if err := localTrack.WriteRTP(cloneRTPPacket(packet)); err != nil {
			log.Printf("sfu rtp forward failed room=%s track=%s: %v", roomID, trackKey, err)
		}
	}
}

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
		session = newSubscriberSession(peerID, targetPeer.role, "", peerConnection)
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
	if currentRoom := h.rooms[sender.roomID]; currentRoom != nil {
		if currentSession := currentRoom.subscribers[sender.id]; currentSession != nil {
			currentSession.pendingOffer = false
			needsOffer = currentSession.needsOffer
			currentSession.needsOffer = false
		}
	}
	h.mu.Unlock()
	if needsOffer {
		go h.ensureSubscriberOffer(sender.roomID, sender.id)
	}
	return nil
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
		h.sendSelectRobotError(sender, robotCode, "robot is not publishing in this mission room")
		return fmt.Errorf("robot %s is not publishing in room %s", robotCode, sender.roomID)
	}
	if publisher.streamBundle == nil || len(publisher.streamBundle.Tracks) == 0 {
		h.mu.Unlock()
		h.sendSelectRobotError(sender, robotCode, "robot stream bundle is not usable")
		return fmt.Errorf("robot %s has no usable stream bundle in room %s", robotCode, sender.roomID)
	}
	targetPeer.robotCode = robotCode
	if session := currentRoom.subscribers[sender.id]; session != nil {
		session.selectRobot(robotCode)
		session.deferOffer()
	}
	h.mu.Unlock()

	h.sendServerSignal(sender.roomID, sender.id, "select-robot-ack", map[string]any{
		"robotCode": robotCode,
	})
	go h.ensureSubscriberOffer(sender.roomID, sender.id)
	return nil
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

func (h *Hub) handleRemoteCandidate(sender *peer, payload map[string]any) error {
	candidate := payloadString(payload, "candidate")
	if candidate == "" {
		return nil
	}
	h.mu.RLock()
	currentRoom := h.rooms[sender.roomID]
	var peerConnection *webrtc.PeerConnection
	if currentRoom != nil {
		if sender.role == "robot" {
			if publisher := currentRoom.publishers[sender.robotCode]; publisher != nil && publisher.peerID == sender.id {
				peerConnection = publisher.peerConnection
			}
			if peerConnection == nil {
				for _, publisher := range currentRoom.publishers {
					if publisher.peerID == sender.id {
						peerConnection = publisher.peerConnection
						break
					}
				}
			}
		}
		if isSubscriberRole(sender.role) {
			if session := currentRoom.subscribers[sender.id]; session != nil {
				peerConnection = session.peerConnection
			}
		}
	}
	h.mu.RUnlock()
	if peerConnection == nil {
		return nil
	}
	return peerConnection.AddICECandidate(webrtc.ICECandidateInit{
		Candidate:     candidate,
		SDPMid:        payloadStringPointer(payload, "sdpMid"),
		SDPMLineIndex: payloadUint16Pointer(payload, "sdpMLineIndex"),
	})
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

	message := string(dataChannelPayloadWithRobotCode(robotCode, payload))
	for _, channel := range channels {
		if err := channel.SendText(message); err != nil {
			log.Printf("sfu datachannel send failed room=%s robot=%s label=%s: %v", roomID, robotCode, label, err)
		}
	}
}

func (h *Hub) createPeerConnection() (*webrtc.PeerConnection, error) {
	mediaEngine := &webrtc.MediaEngine{}
	if err := mediaEngine.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}
	interceptorRegistry := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(mediaEngine, interceptorRegistry); err != nil {
		return nil, err
	}
	api := webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(interceptorRegistry),
	)

	configuration := webrtc.Configuration{}
	if strings.TrimSpace(h.config.TURNURL) != "" {
		configuration.ICEServers = []webrtc.ICEServer{
			{
				URLs:       []string{h.config.TURNURL},
				Username:   h.config.TURNUsername,
				Credential: h.config.TURNPassword,
			},
		}
		configuration.ICETransportPolicy = webrtc.ICETransportPolicyRelay
	}
	return api.NewPeerConnection(configuration)
}

func (h *Hub) closePublisherForPeerLocked(currentRoom *room, peerID string) {
	for robotCode, publisher := range currentRoom.publishers {
		if publisher.peerID != peerID {
			continue
		}
		closePublisherSession(publisher)
		delete(currentRoom.publishers, robotCode)
	}
}

func (h *Hub) closePublisherConnectionsLocked(currentRoom *room) {
	for robotCode, publisher := range currentRoom.publishers {
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
	session.attachedTracks = map[string]struct{}{}
	session.attachedTrackSenders = map[string]*webrtc.RTPSender{}
	session.pendingOffer = false
	session.needsOffer = false
}

func (h *Hub) sendServerSignal(roomID string, targetPeerID string, messageType string, payload map[string]any) {
	if payload == nil {
		payload = map[string]any{}
	}
	payload["room"] = roomID
	payload["fromRole"] = "sfu"
	payload["fromPeerId"] = serverPeerID
	payload["targetPeerId"] = targetPeerID
	h.sendToPeer(roomID, targetPeerID, signalMessage{
		Type:    messageType,
		Payload: payload,
	})
}

func (h *Hub) broadcast(sender *peer, message signalMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	currentRoom := h.rooms[sender.roomID]
	if currentRoom == nil {
		return
	}
	for _, candidate := range currentRoom.peers {
		if candidate.id == sender.id {
			continue
		}
		candidate.enqueue(message)
	}
}

func (h *Hub) sendToPeer(roomID string, peerID string, message signalMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	currentRoom := h.rooms[roomID]
	if currentRoom == nil {
		return
	}
	targetPeer := currentRoom.peers[peerID]
	if targetPeer == nil {
		return
	}
	targetPeer.enqueue(message)
}
