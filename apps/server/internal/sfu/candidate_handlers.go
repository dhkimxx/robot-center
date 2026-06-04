package sfu

import (
	"log"

	"github.com/pion/webrtc/v4"
)

func (h *Hub) handleRemoteCandidate(sender *peer, payload map[string]any) error {
	candidate := payloadString(payload, "candidate")
	if candidate == "" {
		return nil
	}
	remoteCandidate := webrtc.ICECandidateInit{
		Candidate:     candidate,
		SDPMid:        payloadStringPointer(payload, "sdpMid"),
		SDPMLineIndex: payloadUint16Pointer(payload, "sdpMLineIndex"),
	}

	h.mu.Lock()
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
				if peerConnection != nil && peerConnection.RemoteDescription() == nil {
					pendingCount := session.queueRemoteCandidate(remoteCandidate)
					h.mu.Unlock()
					log.Printf("sfu subscriber candidate queued room=%s peer=%s role=%s pending=%d", sender.roomID, sender.id, sender.role, pendingCount)
					return nil
				}
			}
		}
	}
	h.mu.Unlock()
	if peerConnection == nil {
		return nil
	}
	return peerConnection.AddICECandidate(remoteCandidate)
}
