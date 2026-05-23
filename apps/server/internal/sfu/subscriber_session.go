package sfu

import (
	"fmt"
	"log"
	"time"

	"github.com/pion/webrtc/v4"
)

func newSubscriberSession(peerID string, peerConnection *webrtc.PeerConnection) *subscriberSession {
	return &subscriberSession{
		peerID:         peerID,
		peerConnection: peerConnection,
		dataChannels:   map[string]*webrtc.DataChannel{},
		attachedTracks: map[string]struct{}{},
	}
}

func (s *subscriberSession) configureConnection(roomID string, targetPeer *peer) {
	s.peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		_ = candidate
	})
	s.peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("sfu subscriber ICE room=%s peer=%s role=%s state=%s", roomID, targetPeer.id, targetPeer.role, state.String())
	})
}

func (s *subscriberSession) createOffer() (*webrtc.SessionDescription, error) {
	offer, err := s.peerConnection.CreateOffer(nil)
	if err != nil {
		return nil, err
	}
	gatherComplete := webrtc.GatheringCompletePromise(s.peerConnection)
	if err := s.peerConnection.SetLocalDescription(offer); err != nil {
		return nil, err
	}
	select {
	case <-gatherComplete:
	case <-time.After(5 * time.Second):
	}

	localDescription := s.peerConnection.LocalDescription()
	if localDescription == nil {
		return nil, fmt.Errorf("local offer is missing")
	}
	return localDescription, nil
}
