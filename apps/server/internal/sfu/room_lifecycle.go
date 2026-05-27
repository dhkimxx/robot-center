package sfu

func (h *Hub) CloseRoom(roomID string) {
	h.mu.Lock()
	currentRoom := h.rooms[roomID]
	if currentRoom == nil {
		h.mu.Unlock()
		return
	}
	peers := make([]*peer, 0, len(currentRoom.peers))
	for _, roomPeer := range currentRoom.peers {
		roomPeer.enqueue(signalMessage{
			Type: "mission-ended",
			Payload: map[string]any{
				"room": roomID,
			},
		})
		peers = append(peers, roomPeer)
	}
	h.closePublisherConnectionsLocked(currentRoom)
	h.closeSubscriberConnectionsLocked(currentRoom)
	h.mu.Unlock()

	for _, roomPeer := range peers {
		_ = roomPeer.conn.Close()
	}
}
