package sfu

import "sort"

func (h *Hub) Summaries() []RoomSummary {
	h.mu.RLock()
	defer h.mu.RUnlock()

	summaries := make([]RoomSummary, 0, len(h.rooms))
	for _, room := range h.rooms {
		summary := RoomSummary{
			RoomID:          room.id,
			MediaMode:       "go_sfu",
			PublishedTracks: make([]string, 0),
			Peers:           make([]PeerSummary, 0, len(room.peers)),
		}
		for _, publisher := range room.publishers {
			for trackKey := range publisher.publishedTracks {
				summary.PublishedTracks = append(summary.PublishedTracks, trackKey)
			}
		}
		sort.Strings(summary.PublishedTracks)

		for _, peer := range room.peers {
			switch peer.role {
			case "robot":
				summary.RobotCount++
			case "operator":
				summary.OperatorCount++
			case "recorder":
				summary.RecorderCount++
			}
			summary.Peers = append(summary.Peers, PeerSummary{
				PeerID:    peer.id,
				Role:      peer.role,
				RobotCode: peer.robotCode,
				JoinedAt:  peer.joinedAt,
			})
		}
		sort.Slice(summary.Peers, func(i, j int) bool {
			return summary.Peers[i].JoinedAt.Before(summary.Peers[j].JoinedAt)
		})
		summaries = append(summaries, summary)
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].RoomID < summaries[j].RoomID
	})
	return summaries
}

func (h *Hub) registerPeer(joinedPeer *peer) []*peer {
	h.mu.Lock()
	defer h.mu.Unlock()

	currentRoom := h.ensureRoomLocked(joinedPeer.roomID)
	existingPeers := make([]*peer, 0, len(currentRoom.peers))
	for _, existingPeer := range currentRoom.peers {
		existingPeers = append(existingPeers, existingPeer)
	}
	currentRoom.peers[joinedPeer.id] = joinedPeer
	return existingPeers
}

func (h *Hub) ensureRoomLocked(roomID string) *room {
	currentRoom := h.rooms[roomID]
	if currentRoom == nil {
		currentRoom = &room{
			id:          roomID,
			peers:       map[string]*peer{},
			publishers:  map[string]*publisherSession{},
			subscribers: map[string]*subscriberSession{},
		}
		h.rooms[roomID] = currentRoom
	}
	return currentRoom
}

func (r *room) hasPublishedTracks() bool {
	for _, publisher := range r.publishers {
		if len(publisher.publishedTracks) > 0 {
			return true
		}
	}
	return false
}

func (r *room) findPublishedTrack(trackKey string) (*publisherSession, *publishedTrack) {
	for _, publisher := range r.publishers {
		if publishedTrack := publisher.publishedTracks[trackKey]; publishedTrack != nil {
			return publisher, publishedTrack
		}
	}
	return nil, nil
}

func (h *Hub) unregisterPeer(leavingPeer *peer) {
	h.mu.Lock()
	currentRoom := h.rooms[leavingPeer.roomID]
	if currentRoom != nil {
		delete(currentRoom.peers, leavingPeer.id)
		if leavingPeer.role == "robot" {
			h.closePublisherForPeerLocked(currentRoom, leavingPeer.id)
		}
		if session := currentRoom.subscribers[leavingPeer.id]; session != nil {
			closeSubscriberSession(session)
			delete(currentRoom.subscribers, leavingPeer.id)
		}
		if len(currentRoom.peers) == 0 {
			h.closePublisherConnectionsLocked(currentRoom)
			h.closeSubscriberConnectionsLocked(currentRoom)
			delete(h.rooms, leavingPeer.roomID)
		}
	}
	close(leavingPeer.send)
	h.mu.Unlock()

	h.broadcast(leavingPeer, signalMessage{
		Type: "peer-left",
		Payload: map[string]any{
			"room":   leavingPeer.roomID,
			"role":   leavingPeer.role,
			"peerId": leavingPeer.id,
		},
	})
}
