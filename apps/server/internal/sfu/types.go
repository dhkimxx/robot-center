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
	RoomID          string        `json:"roomId"`
	RobotCount      int           `json:"robotCount"`
	OperatorCount   int           `json:"operatorCount"`
	RecorderCount   int           `json:"recorderCount"`
	MediaMode       string        `json:"mediaMode"`
	PublishedTracks []string      `json:"publishedTracks,omitempty"`
	Peers           []PeerSummary `json:"peers"`
}

type ObservedRoomSummary struct {
	RoomID     string                     `json:"roomId"`
	MediaMode  string                     `json:"mediaMode"`
	Publishers []ObservedPublisherSummary `json:"publishers"`
}

type ObservedPublisherSummary struct {
	RobotCode        string     `json:"robotCode"`
	PublisherPeerID  string     `json:"publisherPeerId"`
	State            string     `json:"state"`
	ICEState         string     `json:"iceState,omitempty"`
	TrackCount       int        `json:"trackCount"`
	DataChannelCount int        `json:"dataChannelCount"`
	SubscriberCount  int        `json:"subscriberCount"`
	Tracks           []string   `json:"tracks"`
	DataChannels     []string   `json:"dataChannels"`
	JoinedAt         time.Time  `json:"joinedAt"`
	LastTrackAt      *time.Time `json:"lastTrackAt,omitempty"`
	LastDataAt       *time.Time `json:"lastDataAt,omitempty"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type PeerSummary struct {
	PeerID            string    `json:"peerId"`
	Role              string    `json:"role"`
	RobotCode         string    `json:"robotCode,omitempty"`
	SelectedRobotCode string    `json:"selectedRobotCode,omitempty"`
	JoinedAt          time.Time `json:"joinedAt"`
}

type room struct {
	id          string
	peers       map[string]*peer
	publishers  map[string]*publisherSession
	subscribers map[string]*subscriberSession
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
	peerID          string
	robotCode       string
	peerConnection  *webrtc.PeerConnection
	streamBundle    *RobotStreamBundle
	publishedTracks map[string]*publishedTrack
	joinedAt        time.Time
	iceState        string
	lastTrackAt     *time.Time
	lastDataAt      *time.Time
	updatedAt       time.Time
}

type publishedTrack struct {
	key        string
	robotCode  string
	label      string
	remoteSSRC uint32
	track      *webrtc.TrackLocalStaticRTP
}

type subscriberSession struct {
	peerID               string
	role                 string
	selectedRobotCode    string
	peerConnection       *webrtc.PeerConnection
	dataChannels         map[string]*webrtc.DataChannel
	attachedTracks       map[string]struct{}
	attachedTrackSenders map[string]*webrtc.RTPSender
	pendingOffer         bool
	needsOffer           bool
}

type signalMessage struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload,omitempty"`
}
