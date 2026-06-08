package sfu

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pion/webrtc/v4"

	"robot-center/apps/server/internal/monitorlog"
)

func newPublisherSession(peerID string, robotCode string, peerConnection *webrtc.PeerConnection) *publisherSession {
	streamBundle := newRobotStreamBundle("", robotCode)
	now := time.Now().UTC()
	return &publisherSession{
		peerID:          peerID,
		robotCode:       robotCode,
		peerConnection:  peerConnection,
		streamBundle:    streamBundle,
		publishedTracks: map[string]*publishedTrack{},
		joinedAt:        now,
		updatedAt:       now,
	}
}

func publisherRobotCode(sender *peer) (string, error) {
	robotCode := strings.TrimSpace(sender.robotCode)
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
		hub.markPublisherICEState(roomID, s.robotCode, s.peerID, state.String())
	})
	s.peerConnection.OnTrack(func(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		hub.publishRobotTrack(roomID, s.robotCode, track)
	})
	s.peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		label := normalizeDataChannelRole(dataChannel.Label())
		hub.markPublisherDataChannel(roomID, s.robotCode, s.peerID, label)
		log.Printf("sfu robot datachannel detected room=%s robot=%s label=%s", roomID, s.robotCode, label)
		dataChannel.OnOpen(func() {
			hub.markPublisherDataChannelOpen(roomID, s.robotCode, s.peerID, label)
			log.Printf("sfu robot datachannel open room=%s robot=%s label=%s", roomID, s.robotCode, label)
		})
		dataChannel.OnClose(func() {
			hub.markPublisherDataChannelClosed(roomID, s.robotCode, s.peerID, label)
			log.Printf("sfu robot datachannel closed room=%s robot=%s label=%s", roomID, s.robotCode, label)
		})
		dataChannel.OnError(func(err error) {
			hub.markPublisherDataChannelError(roomID, s.robotCode, s.peerID, label, err)
			log.Printf("sfu robot datachannel error room=%s robot=%s label=%s: %v", roomID, s.robotCode, label, err)
		})
		dataChannel.OnMessage(func(message webrtc.DataChannelMessage) {
			messageCount, firstMessage := hub.markPublisherDataActivity(roomID, s.robotCode, s.peerID, label)
			if firstMessage {
				monitorlog.Event("sfu", "robot_datachannel_first_message", "room", roomID, "robot", s.robotCode, "peer", s.peerID, "label", label, "bytes", len(message.Data), "messageCount", messageCount)
			}
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
