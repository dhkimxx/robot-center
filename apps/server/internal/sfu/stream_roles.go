package sfu

import (
	"strings"
	"time"

	"robot-center/apps/server/internal/utils"

	"github.com/pion/webrtc/v4"
)

const (
	StreamRoleTrackVideo1 = "track.video_1"
	StreamRoleTrackVideo2 = "track.video_2"
	StreamRoleTrackAudio1 = "track.audio_1"
	StreamRoleTrackAudio2 = "track.audio_2"

	StreamRoleTrackUnmappedPrefix = "unmapped."

	StreamRoleChannelTelemetry = "channel.telemetry"
	StreamRoleChannelSpatial   = "channel.spatial"
	StreamRoleChannelEvent     = "channel.event"
	StreamRoleChannelControl   = "channel.control"
)

var canonicalTrackRoles = []string{
	StreamRoleTrackVideo1,
	StreamRoleTrackVideo2,
	StreamRoleTrackAudio1,
	StreamRoleTrackAudio2,
}

var canonicalDataChannelRoles = []string{
	StreamRoleChannelTelemetry,
	StreamRoleChannelSpatial,
	StreamRoleChannelEvent,
	StreamRoleChannelControl,
}

type RobotStreamMetadata struct {
	DisplayLabels map[string]string
}

type PublishedDataChannel struct {
	Role          string
	State         string
	DetectedAt    *time.Time
	OpenedAt      *time.Time
	LastMessageAt *time.Time
	MessageCount  int
	ClosedAt      *time.Time
	LastError     string
}

type RobotStreamBundle struct {
	MissionCode  string
	RobotCode    string
	Tracks       map[string]*publishedTrack
	DataChannels map[string]*PublishedDataChannel
	Metadata     RobotStreamMetadata
}

func newRobotStreamBundle(missionCode string, robotCode string) *RobotStreamBundle {
	return &RobotStreamBundle{
		MissionCode:  strings.TrimSpace(missionCode),
		RobotCode:    strings.TrimSpace(robotCode),
		Tracks:       map[string]*publishedTrack{},
		DataChannels: map[string]*PublishedDataChannel{},
		Metadata: RobotStreamMetadata{
			DisplayLabels: map[string]string{},
		},
	}
}

func normalizeTrackRole(track *webrtc.TrackRemote, _ map[string]*publishedTrack) string {
	raw := strings.ToLower(strings.TrimSpace(track.StreamID() + " " + track.ID()))
	for _, role := range canonicalTrackRoles {
		if strings.Contains(raw, role) {
			return role
		}
	}
	return unmappedTrackRole(track)
}

func isCanonicalTrackRole(role string) bool {
	for _, canonicalRole := range canonicalTrackRoles {
		if role == canonicalRole {
			return true
		}
	}
	return false
}

func unmappedTrackRole(track *webrtc.TrackRemote) string {
	for _, candidate := range []string{track.ID(), track.StreamID(), track.Kind().String()} {
		token := utils.SafeTrackToken(candidate)
		if token != "unknown" {
			return StreamRoleTrackUnmappedPrefix + token
		}
	}
	return StreamRoleTrackUnmappedPrefix + "unknown"
}

func normalizeDataChannelRole(label string) string {
	normalized := strings.ToLower(strings.TrimSpace(label))
	switch normalized {
	case StreamRoleChannelTelemetry:
		return StreamRoleChannelTelemetry
	case StreamRoleChannelSpatial:
		return StreamRoleChannelSpatial
	case StreamRoleChannelEvent:
		return StreamRoleChannelEvent
	case StreamRoleChannelControl:
		return StreamRoleChannelControl
	default:
		return utils.SafeTrackToken(label)
	}
}
