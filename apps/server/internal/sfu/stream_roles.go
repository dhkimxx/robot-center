package sfu

import (
	"strings"

	"robot-center/apps/server/internal/utils"

	"github.com/pion/webrtc/v4"
)

const (
	StreamRoleTrackVideo1 = "track.video_1"
	StreamRoleTrackVideo2 = "track.video_2"
	StreamRoleTrackAudio1 = "track.audio_1"
	StreamRoleTrackAudio2 = "track.audio_2"

	StreamRoleChannelTelemetry = "channel.telemetry"
	StreamRoleChannelSpatial   = "channel.spatial"
	StreamRoleChannelEvent     = "channel.event"
	StreamRoleChannelControl   = "channel.control"
)

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
	Role string
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

func normalizeTrackRole(track *webrtc.TrackRemote, usedRoles map[string]*publishedTrack) string {
	raw := strings.ToLower(strings.TrimSpace(track.StreamID() + " " + track.ID()))
	for _, role := range []string{StreamRoleTrackVideo1, StreamRoleTrackVideo2, StreamRoleTrackAudio1, StreamRoleTrackAudio2} {
		if strings.Contains(raw, role) {
			return role
		}
	}

	// Legacy compatibility only. New robot clients should publish canonical
	// slot IDs such as track.video_1 rather than semantic names.
	if strings.Contains(raw, "thermal") {
		return StreamRoleTrackVideo2
	}
	if strings.Contains(raw, "rgb") {
		return StreamRoleTrackVideo1
	}
	if strings.Contains(raw, "audio") {
		return firstAvailableRole([]string{StreamRoleTrackAudio1, StreamRoleTrackAudio2}, usedRoles)
	}
	if track.Kind() == webrtc.RTPCodecTypeAudio {
		return firstAvailableRole([]string{StreamRoleTrackAudio1, StreamRoleTrackAudio2}, usedRoles)
	}
	if track.Kind() == webrtc.RTPCodecTypeVideo {
		return firstAvailableRole([]string{StreamRoleTrackVideo1, StreamRoleTrackVideo2}, usedRoles)
	}
	return utils.SafeTrackToken(track.Kind().String())
}

func normalizeDataChannelRole(label string) string {
	normalized := strings.ToLower(strings.TrimSpace(label))
	switch normalized {
	case StreamRoleChannelTelemetry, "telemetry", "sensor":
		return StreamRoleChannelTelemetry
	case StreamRoleChannelSpatial, "spatial":
		return StreamRoleChannelSpatial
	case StreamRoleChannelEvent, "event":
		return StreamRoleChannelEvent
	case StreamRoleChannelControl, "control":
		return StreamRoleChannelControl
	default:
		return utils.SafeTrackToken(label)
	}
}

func firstAvailableRole(candidates []string, usedRoles map[string]*publishedTrack) string {
	for _, candidate := range candidates {
		if !trackRoleUsed(candidate, usedRoles) {
			return candidate
		}
	}
	return candidates[0]
}

func trackRoleUsed(role string, usedRoles map[string]*publishedTrack) bool {
	for trackKey, publishedTrack := range usedRoles {
		if publishedTrack != nil && publishedTrack.label == role {
			return true
		}
		if strings.HasSuffix(trackKey, ":"+utils.SafeTrackToken(role)) || trackKey == utils.SafeTrackToken(role) {
			return true
		}
	}
	return false
}
