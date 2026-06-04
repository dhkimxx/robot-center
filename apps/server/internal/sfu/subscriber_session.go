package sfu

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
)

const maxPendingRemoteCandidates = 64

func newSubscriberSession(peerID string, role string, selectedRobotCode string, peerConnection *webrtc.PeerConnection) *subscriberSession {
	return &subscriberSession{
		peerID:               peerID,
		role:                 role,
		selectedRobotCode:    strings.TrimSpace(selectedRobotCode),
		peerConnection:       peerConnection,
		dataChannels:         map[string]*webrtc.DataChannel{},
		attachedTracks:       map[string]struct{}{},
		attachedTrackSenders: map[string]*webrtc.RTPSender{},
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

func (s *subscriberSession) deferOffer() {
	s.needsOffer = true
}

func (s *subscriberSession) createLocalOffer() (*webrtc.SessionDescription, error) {
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

func (s *subscriberSession) queueRemoteCandidate(candidate webrtc.ICECandidateInit) int {
	if len(s.pendingRemoteCandidates) >= maxPendingRemoteCandidates {
		copy(s.pendingRemoteCandidates, s.pendingRemoteCandidates[1:])
		s.pendingRemoteCandidates[len(s.pendingRemoteCandidates)-1] = candidate
		return len(s.pendingRemoteCandidates)
	}
	s.pendingRemoteCandidates = append(s.pendingRemoteCandidates, candidate)
	return len(s.pendingRemoteCandidates)
}

func (s *subscriberSession) drainPendingRemoteCandidates() []webrtc.ICECandidateInit {
	if len(s.pendingRemoteCandidates) == 0 {
		return nil
	}
	candidates := append([]webrtc.ICECandidateInit(nil), s.pendingRemoteCandidates...)
	s.pendingRemoteCandidates = nil
	return candidates
}

func (s *subscriberSession) beginOffer(currentRoom *room, forwardRTCP func(roomID string, trackKey string, packets []rtcp.Packet), requestKeyFrames func(roomID string, trackKey string, count int, interval time.Duration)) bool {
	if !s.isReadyToSubscribe() {
		return s.detachPublishedTracks(currentRoom)
	}
	offerRequired := s.attachPublishedTracks(currentRoom, forwardRTCP, requestKeyFrames)
	if s.ensureDataChannels() {
		offerRequired = true
	}
	if offerRequired {
		s.pendingOffer = true
		s.needsOffer = false
	}
	return offerRequired
}

func (s *subscriberSession) selectRobot(robotCode string) {
	if s.role == "recorder" {
		return
	}
	s.selectedRobotCode = strings.TrimSpace(robotCode)
}

func (s *subscriberSession) isReadyToSubscribe() bool {
	return s.role == "recorder" || strings.TrimSpace(s.selectedRobotCode) != ""
}

func (s *subscriberSession) shouldReceiveRobot(robotCode string) bool {
	if s.role == "recorder" {
		return true
	}
	return strings.TrimSpace(robotCode) != "" && strings.TrimSpace(robotCode) == strings.TrimSpace(s.selectedRobotCode)
}

func (s *subscriberSession) attachPublishedTracks(currentRoom *room, forwardRTCP func(roomID string, trackKey string, packets []rtcp.Packet), requestKeyFrames func(roomID string, trackKey string, count int, interval time.Duration)) bool {
	if s.attachedTracks == nil {
		s.attachedTracks = map[string]struct{}{}
	}
	if s.attachedTrackSenders == nil {
		s.attachedTrackSenders = map[string]*webrtc.RTPSender{}
	}
	changed := false
	trackKeys := make([]string, 0)
	tracksByKey := map[string]*publishedTrack{}
	for _, publisher := range currentRoom.publishers {
		for trackKey, publishedTrack := range publisher.publishedTracks {
			if publishedTrack == nil || !s.shouldReceiveRobot(publishedTrack.robotCode) {
				continue
			}
			trackKeys = append(trackKeys, trackKey)
			tracksByKey[trackKey] = publishedTrack
		}
	}
	sort.Strings(trackKeys)
	desiredTrackKeys := map[string]struct{}{}
	for _, trackKey := range trackKeys {
		desiredTrackKeys[trackKey] = struct{}{}
	}

	for trackKey, sender := range s.attachedTrackSenders {
		if _, ok := desiredTrackKeys[trackKey]; ok {
			continue
		}
		if sender != nil {
			if err := s.peerConnection.RemoveTrack(sender); err != nil {
				log.Printf("sfu subscriber remove track failed room=%s peer=%s track=%s: %v", currentRoom.id, s.peerID, trackKey, err)
			}
		}
		delete(s.attachedTrackSenders, trackKey)
		delete(s.attachedTracks, trackKey)
		changed = true
	}

	for _, trackKey := range trackKeys {
		if _, ok := s.attachedTracks[trackKey]; ok {
			continue
		}
		publishedTrack := tracksByKey[trackKey]
		if publishedTrack == nil || publishedTrack.track == nil {
			continue
		}
		sender, err := s.peerConnection.AddTrack(publishedTrack.track)
		if err != nil {
			log.Printf("sfu subscriber add track failed room=%s peer=%s track=%s: %v", currentRoom.id, s.peerID, trackKey, err)
			continue
		}
		s.attachedTracks[trackKey] = struct{}{}
		s.attachedTrackSenders[trackKey] = sender
		changed = true
		go s.forwardRTCP(currentRoom.id, trackKey, sender, forwardRTCP)
		go requestKeyFrames(currentRoom.id, trackKey, 5, time.Second)
	}
	return changed
}

func (s *subscriberSession) detachPublishedTracks(currentRoom *room) bool {
	if s.attachedTracks == nil {
		s.attachedTracks = map[string]struct{}{}
	}
	if s.attachedTrackSenders == nil {
		s.attachedTrackSenders = map[string]*webrtc.RTPSender{}
	}
	changed := false
	for trackKey, sender := range s.attachedTrackSenders {
		if sender != nil {
			if err := s.peerConnection.RemoveTrack(sender); err != nil {
				log.Printf("sfu subscriber remove track failed room=%s peer=%s track=%s: %v", currentRoom.id, s.peerID, trackKey, err)
			}
		}
		delete(s.attachedTrackSenders, trackKey)
		delete(s.attachedTracks, trackKey)
		changed = true
	}
	return changed
}

func (s *subscriberSession) ensureDataChannels() bool {
	created := false
	for _, label := range canonicalDataChannelRoles {
		if s.dataChannels[label] != nil {
			continue
		}
		dataChannel, err := s.peerConnection.CreateDataChannel(label, nil)
		if err != nil {
			log.Printf("sfu subscriber datachannel create failed peer=%s label=%s: %v", s.peerID, label, err)
			continue
		}
		dataChannel.OnOpen(func() {
			log.Printf("sfu subscriber datachannel open peer=%s role=%s label=%s", s.peerID, s.role, label)
		})
		dataChannel.OnClose(func() {
			log.Printf("sfu subscriber datachannel closed peer=%s role=%s label=%s", s.peerID, s.role, label)
		})
		dataChannel.OnError(func(err error) {
			log.Printf("sfu subscriber datachannel error peer=%s role=%s label=%s: %v", s.peerID, s.role, label, err)
		})
		s.dataChannels[label] = dataChannel
		created = true
	}
	return created
}

func (s *subscriberSession) forwardRTCP(roomID string, trackKey string, sender *webrtc.RTPSender, forward func(roomID string, trackKey string, packets []rtcp.Packet)) {
	buffer := make([]byte, 1500)
	for {
		byteCount, _, err := sender.Read(buffer)
		if err != nil {
			return
		}
		packets, err := rtcp.Unmarshal(buffer[:byteCount])
		if err != nil {
			continue
		}
		forward(roomID, trackKey, packets)
	}
}
