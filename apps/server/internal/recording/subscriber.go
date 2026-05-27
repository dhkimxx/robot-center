package recording

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"

	"robot-center/apps/server/internal/domain"
)

type SubscriberStatus struct {
	Rooms     []SubscriberRoomStatus  `json:"rooms"`
	Queues    RecorderDataQueueStatus `json:"queues"`
	UpdatedAt time.Time               `json:"updatedAt"`
}

type SubscriberRoomStatus struct {
	RoomID               string                  `json:"roomId"`
	MissionCode          string                  `json:"missionCode"`
	RobotCode            string                  `json:"robotCode"`
	RobotCodes           []string                `json:"robotCodes,omitempty"`
	Robots               []SubscriberRobotStatus `json:"robots,omitempty"`
	SignalingState       string                  `json:"signalingState"`
	ICEState             string                  `json:"iceState"`
	TrackCount           int                     `json:"trackCount"`
	DataChannelCount     int                     `json:"dataChannelCount"`
	DataMessageCount     int                     `json:"dataMessageCount"`
	SensorStoredCount    int                     `json:"sensorStoredCount"`
	TelemetryStoredCount int                     `json:"telemetryStoredCount"`
	LastTrackLabel       string                  `json:"lastTrackLabel"`
	LastDataLabel        string                  `json:"lastDataLabel"`
	LastDataMessageAt    time.Time               `json:"lastDataMessageAt,omitempty"`
	LastPersistedLabel   string                  `json:"lastPersistedLabel,omitempty"`
	LastPersistedAt      time.Time               `json:"lastPersistedAt,omitempty"`
	LastError            string                  `json:"lastError,omitempty"`
	UpdatedAt            time.Time               `json:"updatedAt"`
}

type SubscriberRobotStatus struct {
	RobotCode        string    `json:"robotCode"`
	TrackCount       int       `json:"trackCount"`
	DataChannelCount int       `json:"dataChannelCount"`
	LastTrackAt      time.Time `json:"lastTrackAt,omitempty"`
	LastDataAt       time.Time `json:"lastDataAt,omitempty"`
	LastPersistedAt  time.Time `json:"lastPersistedAt,omitempty"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type recorderSignalMessage struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload,omitempty"`
}

type recorderSessionStatus struct {
	missionCode          string
	robotCode            string
	robotCodes           map[string]struct{}
	signalingState       string
	iceState             string
	trackLabels          map[string]struct{}
	dataChannelLabels    map[string]struct{}
	robotStatuses        map[string]recorderRobotRuntime
	dataMessageCount     int
	sensorStoredCount    int
	telemetryStoredCount int
	lastTrackLabel       string
	lastDataLabel        string
	lastDataMessageAt    time.Time
	lastPersistedLabel   string
	lastPersistedAt      time.Time
	lastError            string
	updatedAt            time.Time
}

type recorderRobotRuntime struct {
	trackLabels       map[string]struct{}
	dataChannelLabels map[string]struct{}
	lastTrackAt       time.Time
	lastDataAt        time.Time
	lastPersistedAt   time.Time
	updatedAt         time.Time
}

func (w *Worker) runSubscriberLoop(ctx context.Context) {
	if strings.TrimSpace(w.config.SFUWebSocketInternalBaseURL) == "" {
		log.Println("recorder-worker subscriber disabled: SFU_WS_INTERNAL_BASE_URL is empty")
		return
	}

	ticker := time.NewTicker(w.config.PollInterval)
	defer ticker.Stop()

	log.Printf("recorder-worker subscriber watching app-server=%s signaling=%s", w.config.AppServerURL, w.config.SFURecorderWebSocketURL())
	w.syncSubscriberTargets(ctx)
	for {
		select {
		case <-ctx.Done():
			w.stopSubscriberSessions()
			return
		case <-ticker.C:
			w.syncSubscriberTargets(ctx)
		}
	}
}

func (w *Worker) syncSubscriberTargets(ctx context.Context) {
	targets, err := w.appServerClient.FetchRecordingTargets(ctx)
	if err != nil {
		log.Printf("recorder-worker subscriber target fetch failed: %v", err)
		return
	}

	targetsByRoom := groupRecordingTargetsByMission(targets)
	activeRooms := map[string]struct{}{}
	for roomID, roomTargets := range targetsByRoom {
		activeRooms[roomID] = struct{}{}
		w.ensureSubscriberSession(ctx, roomID, roomTargets)
	}

	w.subscriberMu.Lock()
	for roomID, cancel := range w.subscriberCancels {
		if _, ok := activeRooms[roomID]; ok {
			continue
		}
		cancel()
		delete(w.subscriberCancels, roomID)
		delete(w.subscriberStatuses, roomID)
	}
	w.subscriberMu.Unlock()
}

func (w *Worker) ensureSubscriberSession(ctx context.Context, roomID string, targets []domain.Mission) {
	if len(targets) == 0 {
		return
	}
	target := targets[0]
	robotCodes := robotCodesFromTargets(targets)
	w.subscriberMu.Lock()
	if _, ok := w.subscriberCancels[roomID]; ok {
		status := w.subscriberStatuses[roomID]
		status.robotCode = firstRobotCode(robotCodes)
		status.robotCodes = robotCodeSet(robotCodes)
		status.updatedAt = time.Now().UTC()
		w.subscriberStatuses[roomID] = status
		w.subscriberMu.Unlock()
		return
	}
	sessionCtx, cancel := context.WithCancel(ctx)
	w.subscriberCancels[roomID] = cancel
	w.subscriberStatuses[roomID] = recorderSessionStatus{
		missionCode:       target.MissionCode,
		robotCode:         firstRobotCode(robotCodes),
		robotCodes:        robotCodeSet(robotCodes),
		signalingState:    "starting",
		iceState:          "new",
		trackLabels:       map[string]struct{}{},
		dataChannelLabels: map[string]struct{}{},
		robotStatuses:     map[string]recorderRobotRuntime{},
		updatedAt:         time.Now().UTC(),
	}
	w.subscriberMu.Unlock()

	go func() {
		defer func() {
			w.subscriberMu.Lock()
			delete(w.subscriberCancels, roomID)
			w.subscriberMu.Unlock()
		}()
		w.runRecorderSession(sessionCtx, target)
	}()
}

func (w *Worker) stopSubscriberSessions() {
	w.subscriberMu.Lock()
	defer w.subscriberMu.Unlock()
	for _, cancel := range w.subscriberCancels {
		cancel()
	}
	w.subscriberCancels = map[string]context.CancelFunc{}
}

func (w *Worker) SubscriberStatus() SubscriberStatus {
	w.subscriberMu.RLock()
	defer w.subscriberMu.RUnlock()

	status := SubscriberStatus{
		Rooms:     make([]SubscriberRoomStatus, 0, len(w.subscriberStatuses)),
		Queues:    w.RecorderDataQueueStatus(),
		UpdatedAt: time.Now().UTC(),
	}
	for roomID, roomStatus := range w.subscriberStatuses {
		status.Rooms = append(status.Rooms, SubscriberRoomStatus{
			RoomID:               roomID,
			MissionCode:          roomStatus.missionCode,
			RobotCode:            roomStatus.robotCode,
			RobotCodes:           sortedRobotCodes(roomStatus.robotCodes),
			Robots:               subscriberRobotStatuses(roomStatus),
			SignalingState:       roomStatus.signalingState,
			ICEState:             roomStatus.iceState,
			TrackCount:           len(roomStatus.trackLabels),
			DataChannelCount:     len(roomStatus.dataChannelLabels),
			DataMessageCount:     roomStatus.dataMessageCount,
			SensorStoredCount:    roomStatus.sensorStoredCount,
			TelemetryStoredCount: roomStatus.telemetryStoredCount,
			LastTrackLabel:       roomStatus.lastTrackLabel,
			LastDataLabel:        roomStatus.lastDataLabel,
			LastDataMessageAt:    roomStatus.lastDataMessageAt,
			LastPersistedLabel:   roomStatus.lastPersistedLabel,
			LastPersistedAt:      roomStatus.lastPersistedAt,
			LastError:            roomStatus.lastError,
			UpdatedAt:            roomStatus.updatedAt,
		})
	}
	return status
}

func (w *Worker) updateSubscriberStatus(roomID string, update func(*recorderSessionStatus)) {
	w.subscriberMu.Lock()
	defer w.subscriberMu.Unlock()

	statusKey := w.resolveSubscriberStatusKeyLocked(roomID)
	status := w.subscriberStatuses[statusKey]
	if status.missionCode == "" {
		status.missionCode = statusKey
	}
	if status.robotCodes == nil {
		status.robotCodes = map[string]struct{}{}
	}
	if status.trackLabels == nil {
		status.trackLabels = map[string]struct{}{}
	}
	if status.dataChannelLabels == nil {
		status.dataChannelLabels = map[string]struct{}{}
	}
	if status.robotStatuses == nil {
		status.robotStatuses = map[string]recorderRobotRuntime{}
	}
	update(&status)
	status.updatedAt = time.Now().UTC()
	w.subscriberStatuses[statusKey] = status
}

func (w *Worker) resolveSubscriberStatusKeyLocked(roomID string) string {
	if _, ok := w.subscriberStatuses[roomID]; ok {
		return roomID
	}
	missionCode, _ := splitRecorderMediaKey(roomID)
	if missionCode != "" {
		if _, ok := w.subscriberStatuses[missionCode]; ok {
			return missionCode
		}
	}
	return roomID
}

func (w *Worker) runRecorderSession(ctx context.Context, target domain.Mission) {
	roomID := recorderSignalingRoomID(target.MissionCode)
	signalingURL := buildSignalingURL(w.config.SFURecorderWebSocketURL(), roomID)

	connection, _, err := websocket.DefaultDialer.DialContext(ctx, signalingURL, nil)
	if err != nil {
		w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
			status.signalingState = "failed"
			status.lastError = err.Error()
		})
		log.Printf("recorder-worker subscriber signaling failed room=%s: %v", roomID, err)
		return
	}
	defer connection.Close()

	peerConnection, err := w.createRecorderPeerConnection(ctx, roomID)
	if err != nil {
		w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
			status.signalingState = "failed"
			status.lastError = err.Error()
		})
		log.Printf("recorder-worker peer connection failed room=%s: %v", roomID, err)
		return
	}
	defer peerConnection.Close()
	defer w.closeRecorderSessionAudioWriters(roomID)

	w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
		status.signalingState = "connected"
		status.lastError = ""
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		<-ctx.Done()
		_ = connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "recorder stopped"))
	}()

	var selfPeerID string
	var targetPeerID string
	for {
		select {
		case <-done:
			return
		default:
		}

		_, rawMessage, err := connection.ReadMessage()
		if err != nil {
			w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
				status.signalingState = "closed"
				status.lastError = err.Error()
			})
			return
		}

		var message recorderSignalMessage
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			continue
		}
		if shouldIgnoreTargetedMessage(message.Payload, selfPeerID) {
			continue
		}

		switch message.Type {
		case "joined":
			selfPeerID = payloadString(message.Payload, "peerId")
			log.Printf("recorder-worker joined room=%s peer=%s", roomID, selfPeerID)
		case "peer-present", "peer-joined":
			if payloadString(message.Payload, "role") == "robot" {
				robotCode := payloadString(message.Payload, "robotCode")
				if robotCode != "" {
					w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
						status.robotCodes[robotCode] = struct{}{}
						if status.robotCode == "" {
							status.robotCode = robotCode
						}
					})
				}
				log.Printf("recorder-worker sees robot room=%s robot=%s peer=%s", roomID, robotCode, payloadString(message.Payload, "peerId"))
			}
		case "offer":
			fromPeerID := payloadString(message.Payload, "fromPeerId")
			if fromPeerID != "" {
				targetPeerID = fromPeerID
			}
			if err := w.answerRecorderOffer(ctx, peerConnection, connection, message.Payload, targetPeerID); err != nil {
				w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
					status.lastError = err.Error()
				})
				log.Printf("recorder-worker answer failed room=%s: %v", roomID, err)
			}
		case "candidate":
			candidate := payloadString(message.Payload, "candidate")
			if candidate == "" {
				continue
			}
			err := peerConnection.AddICECandidate(webrtc.ICECandidateInit{
				Candidate:     candidate,
				SDPMid:        payloadStringPointer(message.Payload, "sdpMid"),
				SDPMLineIndex: payloadUint16Pointer(message.Payload, "sdpMLineIndex"),
			})
			if err != nil {
				log.Printf("recorder-worker remote candidate ignored room=%s: %v", roomID, err)
			}
		}
	}
}

func (w *Worker) createRecorderPeerConnection(ctx context.Context, roomID string) (*webrtc.PeerConnection, error) {
	mediaEngine := &webrtc.MediaEngine{}
	if err := mediaEngine.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}
	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs:       []string{w.config.TURNInternalURL},
				Username:   w.config.TURNUsername,
				Credential: w.config.TURNPassword,
			},
		},
		ICETransportPolicy: webrtc.ICETransportPolicyRelay,
	})
	if err != nil {
		return nil, err
	}

	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		// The recorder waits for ICE gathering before sending its answer, so the
		// SDP already contains relay candidates. Avoid broadcasting untargeted
		// trickle candidates in the P0 signaling hub.
		_ = candidate
	})
	peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
			status.iceState = state.String()
		})
		log.Printf("recorder-worker ICE room=%s state=%s", roomID, state.String())
	})
	peerConnection.OnTrack(func(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		robotCode, label := classifyRecorderTrack(track)
		if robotCode == "" {
			robotCode = w.singleSubscriberRobotCode(roomID)
		}
		trackLabel := recorderTrackLabel(robotCode, label)
		observedAt := time.Now().UTC()
		w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
			if robotCode != "" {
				status.robotCodes[robotCode] = struct{}{}
				if status.robotCode == "" {
					status.robotCode = robotCode
				}
				robotStatus := ensureRecorderRobotRuntime(status, robotCode)
				robotStatus.trackLabels[label] = struct{}{}
				robotStatus.lastTrackAt = observedAt
				robotStatus.updatedAt = observedAt
				status.robotStatuses[robotCode] = robotStatus
			}
			status.trackLabels[trackLabel] = struct{}{}
			status.lastTrackLabel = trackLabel
		})
		mediaKey := recorderMediaKey(roomID, robotCode)
		log.Printf("recorder-worker track room=%s robot=%s label=%s kind=%s codec=%s stream=%s id=%s", roomID, robotCode, label, track.Kind().String(), track.Codec().MimeType, track.StreamID(), track.ID())
		if track.Kind() == webrtc.RTPCodecTypeVideo && strings.EqualFold(track.Codec().MimeType, webrtc.MimeTypeH264) {
			w.recordH264Track(ctx, mediaKey, label, track)
			return
		}
		if track.Kind() == webrtc.RTPCodecTypeAudio && strings.EqualFold(track.Codec().MimeType, webrtc.MimeTypeOpus) {
			w.recordOpusTrack(ctx, mediaKey, label, track)
			return
		}
		for {
			if _, _, err := track.ReadRTP(); err != nil {
				return
			}
		}
	})
	peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		label := dataChannel.Label()
		w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
			status.dataChannelLabels[label] = struct{}{}
			status.lastDataLabel = label
		})
		log.Printf("recorder-worker datachannel room=%s label=%s", roomID, label)
		dataChannel.OnMessage(func(message webrtc.DataChannelMessage) {
			payload := append([]byte(nil), message.Data...)
			w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
				status.dataChannelLabels[label] = struct{}{}
				status.dataMessageCount++
				status.lastDataLabel = label
				status.lastDataMessageAt = time.Now().UTC()
			})
			w.enqueueRecorderDataChannelMessage(ctx, roomID, label, payload)
		})
	})

	peerConnection.OnSignalingStateChange(func(state webrtc.SignalingState) {
		w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
			status.signalingState = state.String()
		})
	})

	return peerConnection, nil
}

func (w *Worker) answerRecorderOffer(ctx context.Context, peerConnection *webrtc.PeerConnection, connection *websocket.Conn, payload map[string]any, targetPeerID string) error {
	offerSDP := payloadString(payload, "sdp")
	if offerSDP == "" {
		return fmt.Errorf("offer sdp is empty")
	}
	if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offerSDP,
	}); err != nil {
		return err
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return err
	}
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	if err := peerConnection.SetLocalDescription(answer); err != nil {
		return err
	}

	select {
	case <-gatherComplete:
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
	}

	localDescription := peerConnection.LocalDescription()
	if localDescription == nil {
		return fmt.Errorf("local answer is missing")
	}
	answerPayload := map[string]any{
		"type": localDescription.Type.String(),
		"sdp":  localDescription.SDP,
	}
	if targetPeerID != "" {
		answerPayload["targetPeerId"] = targetPeerID
	}
	return connection.WriteJSON(recorderSignalMessage{Type: "answer", Payload: answerPayload})
}
