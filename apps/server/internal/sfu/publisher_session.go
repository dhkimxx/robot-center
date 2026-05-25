package sfu

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pion/webrtc/v4"
)

func newPublisherSession(peerID string, robotCode string, peerConnection *webrtc.PeerConnection) *publisherSession {
	streamBundle := newRobotStreamBundle("", robotCode)
	return &publisherSession{
		peerID:          peerID,
		robotCode:       robotCode,
		peerConnection:  peerConnection,
		streamBundle:    streamBundle,
		publishedTracks: map[string]*publishedTrack{},
	}
}

func publisherRobotCode(sender *peer, payload map[string]any) (string, error) {
	robotCode := strings.TrimSpace(sender.robotCode)
	if robotCode == "" {
		robotCode = strings.TrimSpace(payloadString(payload, "robotCode"))
	}
	if robotCode == "" {
		return "", fmt.Errorf("robotCode is required")
	}
	return robotCode, nil
}

func (s *publisherSession) prepareConnection(roomID string, hub *Hub) {
	s.peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		// Offers and answers are sent after ICE gathering completes, so relay
		// candidates are already embedded in SDP. Avoid trickle ordering issues
		// with Android/browser clients in the PoC.
		_ = candidate
	})
	s.peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("sfu robot ICE room=%s robot=%s peer=%s state=%s", roomID, s.robotCode, s.peerID, state.String())
	})
	s.peerConnection.OnTrack(func(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		hub.publishRobotTrack(roomID, s.robotCode, track)
	})
	s.peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		label := normalizeDataChannelRole(dataChannel.Label())
		if s.streamBundle != nil {
			s.streamBundle.DataChannels[label] = &PublishedDataChannel{Role: label}
		}
		log.Printf("sfu robot datachannel room=%s robot=%s label=%s", roomID, s.robotCode, label)
		dataChannel.OnMessage(func(message webrtc.DataChannelMessage) {
			hub.forwardDataChannelMessage(roomID, s.robotCode, label, message.Data)
		})
	})
}

func (s *publisherSession) answerOffer(offerSDP string) (*webrtc.SessionDescription, error) {
	if err := s.peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offerSDP,
	}); err != nil {
		return nil, err
	}
	answer, err := s.peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}
	gatherComplete := webrtc.GatheringCompletePromise(s.peerConnection)
	if err := s.peerConnection.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	select {
	case <-gatherComplete:
	case <-time.After(5 * time.Second):
	}

	localDescription := s.peerConnection.LocalDescription()
	if localDescription == nil {
		return nil, fmt.Errorf("local answer is missing")
	}
	return localDescription, nil
}
