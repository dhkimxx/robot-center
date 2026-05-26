package sfu

import (
	"sort"
	"time"
)

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
				PeerID:            peer.id,
				Role:              peer.role,
				RobotCode:         peer.robotCode,
				SelectedRobotCode: peer.selectedRobotCode,
				JoinedAt:          peer.joinedAt,
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

func (h *Hub) ObservedRooms() []ObservedRoomSummary {
	h.mu.RLock()
	defer h.mu.RUnlock()

	summaries := make([]ObservedRoomSummary, 0, len(h.rooms))
	for _, room := range h.rooms {
		summary := ObservedRoomSummary{
			RoomID:     room.id,
			MediaMode:  "go_sfu",
			Publishers: make([]ObservedPublisherSummary, 0, len(room.publishers)),
		}
		for _, publisher := range room.publishers {
			tracks := make([]string, 0, len(publisher.publishedTracks))
			for trackKey := range publisher.publishedTracks {
				tracks = append(tracks, trackKey)
			}
			sort.Strings(tracks)

			dataChannels := make([]string, 0)
			if publisher.streamBundle != nil {
				for label := range publisher.streamBundle.DataChannels {
					dataChannels = append(dataChannels, label)
				}
			}
			sort.Strings(dataChannels)

			summary.Publishers = append(summary.Publishers, ObservedPublisherSummary{
				RobotCode:        publisher.robotCode,
				PublisherPeerID:  publisher.peerID,
				State:            observedPublisherState(publisher),
				ICEState:         publisher.iceState,
				TrackCount:       len(tracks),
				DataChannelCount: len(dataChannels),
				SubscriberCount:  room.subscriberCountForRobot(publisher.robotCode),
				Tracks:           tracks,
				DataChannels:     dataChannels,
				JoinedAt:         publisher.joinedAt,
				LastTrackAt:      cloneTimePointer(publisher.lastTrackAt),
				LastDataAt:       cloneTimePointer(publisher.lastDataAt),
				UpdatedAt:        publisher.updatedAt,
			})
		}
		sort.Slice(summary.Publishers, func(i, j int) bool {
			if summary.Publishers[i].RobotCode != summary.Publishers[j].RobotCode {
				return summary.Publishers[i].RobotCode < summary.Publishers[j].RobotCode
			}
			return summary.Publishers[i].PublisherPeerID < summary.Publishers[j].PublisherPeerID
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

func (h *Hub) findPublisherLocked(roomID string, robotCode string, peerID string) *publisherSession {
	currentRoom := h.rooms[roomID]
	if currentRoom == nil {
		return nil
	}
	publisher := currentRoom.publishers[robotCode]
	if publisher != nil && publisher.peerID == peerID {
		return publisher
	}
	for _, publisher := range currentRoom.publishers {
		if publisher.peerID == peerID {
			return publisher
		}
	}
	return nil
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

func (r *room) subscriberCountForRobot(robotCode string) int {
	count := 0
	for _, session := range r.subscribers {
		if session != nil && session.shouldReceiveRobot(robotCode) {
			count++
		}
	}
	return count
}

func observedPublisherState(publisher *publisherSession) string {
	if publisher == nil {
		return "unknown"
	}
	if len(publisher.publishedTracks) > 0 || publisher.lastTrackAt != nil || publisher.lastDataAt != nil {
		return "publishing"
	}
	switch publisher.iceState {
	case "connected", "completed":
		return "connected"
	case "failed", "disconnected", "closed":
		return publisher.iceState
	default:
		return "joined"
	}
}

func cloneTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := value.UTC()
	return &cloned
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
