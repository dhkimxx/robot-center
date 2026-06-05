package sfu

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

type Config struct {
	TURNURL                string
	TURNUsername           string
	TURNPassword           string
	ValidateRobotPublisher func(roomID string, robotCode string) error
	ValidateRobotSelection func(roomID string, robotCode string) error
	OnPublisherEvent       PublisherEventHandler
}

type PublisherEventHandler func(PublisherEvent)

type PublisherEventType string

const (
	PublisherEventMediaStarted PublisherEventType = "media_started"
	PublisherEventMediaActive  PublisherEventType = "media_active"
	PublisherEventEnded        PublisherEventType = "ended"
)

type PublisherEvent struct {
	Type            PublisherEventType
	RoomID          string
	RobotCode       string
	PublisherPeerID string
	Reason          string
	ObservedAt      time.Time
}

type PeerJoinRequest struct {
	RoomID    string
	Role      string
	RobotCode string
}

type Hub struct {
	mu       sync.RWMutex
	config   Config
	rooms    map[string]*room
	upgrader websocket.Upgrader
}

type RoomSummary struct {
	RoomID          string                     `json:"roomId"`
	RobotCount      int                        `json:"robotCount"`
	OperatorCount   int                        `json:"operatorCount"`
	RecorderCount   int                        `json:"recorderCount"`
	MediaMode       string                     `json:"mediaMode"`
	PublishedTracks []string                   `json:"publishedTracks,omitempty"`
	Publishers      []ObservedPublisherSummary `json:"publishers,omitempty"`
	Peers           []PeerSummary              `json:"peers"`
}

type ObservedRoomSummary struct {
	RoomID     string                     `json:"roomId"`
	MediaMode  string                     `json:"mediaMode"`
	Publishers []ObservedPublisherSummary `json:"publishers"`
}

type ObservedPublisherSummary struct {
	RobotCode         string                       `json:"robotCode"`
	PublisherPeerID   string                       `json:"publisherPeerId"`
	State             string                       `json:"state"`
	ICEState          string                       `json:"iceState,omitempty"`
	TrackCount        int                          `json:"trackCount"`
	DataChannelCount  int                          `json:"dataChannelCount"`
	SubscriberCount   int                          `json:"subscriberCount"`
	Tracks            []string                     `json:"tracks"`
	DataChannels      []string                     `json:"dataChannels"`
	DataChannelStates []ObservedDataChannelSummary `json:"dataChannelStates,omitempty"`
	JoinedAt          time.Time                    `json:"joinedAt"`
	FirstTrackAt      *time.Time                   `json:"firstTrackAt,omitempty"`
	LastTrackAt       *time.Time                   `json:"lastTrackAt,omitempty"`
	LastDataAt        *time.Time                   `json:"lastDataAt,omitempty"`
	UpdatedAt         time.Time                    `json:"updatedAt"`
}

type ObservedDataChannelSummary struct {
	Label         string     `json:"label"`
	State         string     `json:"state"`
	DetectedAt    *time.Time `json:"detectedAt,omitempty"`
	OpenedAt      *time.Time `json:"openedAt,omitempty"`
	LastMessageAt *time.Time `json:"lastMessageAt,omitempty"`
	MessageCount  int        `json:"messageCount"`
	ClosedAt      *time.Time `json:"closedAt,omitempty"`
	LastError     string     `json:"lastError,omitempty"`
}

type PeerSummary struct {
	PeerID            string    `json:"peerId"`
	Role              string    `json:"role"`
	RobotCode         string    `json:"robotCode,omitempty"`
	SelectedRobotCode string    `json:"selectedRobotCode,omitempty"`
	JoinedAt          time.Time `json:"joinedAt"`
}

type room struct {
	id                    string
	peers                 map[string]*peer
	publishers            map[string]*publisherSession
	subscribers           map[string]*subscriberSession
	subscriberOfferTimers map[string]*time.Timer
}

type peer struct {
	id                string
	roomID            string
	role              string
	robotCode         string
	selectedRobotCode string
	joinedAt          time.Time
	conn              *websocket.Conn
	send              chan signalMessage
}

type publisherSession struct {
	peerID           string
	robotCode        string
	peerConnection   *webrtc.PeerConnection
	streamBundle     *RobotStreamBundle
	publishedTracks  map[string]*publishedTrack
	joinedAt         time.Time
	iceState         string
	firstTrackAt     *time.Time
	lastTrackAt      *time.Time
	lastMediaEventAt time.Time
	lastDataAt       *time.Time
	updatedAt        time.Time
}

type publishedTrack struct {
	key             string
	robotCode       string
	label           string
	publisherPeerID string
	remoteSSRC      uint32
	track           *webrtc.TrackLocalStaticRTP
}

type subscriberTrackAttachment struct {
	sourceTrack     *webrtc.TrackLocalStaticRTP
	sender          *webrtc.RTPSender
	publisherPeerID string
	robotCode       string
}

type subscriberSession struct {
	peerID                  string
	role                    string
	selectedRobotCode       string
	peerConnection          *webrtc.PeerConnection
	dataChannels            map[string]*webrtc.DataChannel
	attachedTracks          map[string]subscriberTrackAttachment
	pendingRemoteCandidates []webrtc.ICECandidateInit
	pendingOffer            bool
	needsOffer              bool
}

type signalMessage struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload,omitempty"`
}
