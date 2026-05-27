package sfu

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
