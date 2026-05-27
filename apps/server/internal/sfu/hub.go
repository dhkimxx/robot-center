package sfu

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
	"time"
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
	http.NotFound(w, r)
}

func (h *Hub) ServePeer(w http.ResponseWriter, r *http.Request, request PeerJoinRequest) {
	roomID := strings.TrimSpace(request.RoomID)
	role := strings.TrimSpace(request.Role)
	robotCode := strings.TrimSpace(request.RobotCode)
	if roomID == "" || role == "" {
		http.Error(w, "room and role are required", http.StatusBadRequest)
		return
	}
	if role == "robot" && robotCode == "" {
		http.Error(w, "robotCode is required for robot peers", http.StatusInternalServerError)
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
