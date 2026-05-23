package sfu

import (
	"strings"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

func classifyTrack(track *webrtc.TrackRemote) string {
	raw := strings.ToLower(track.StreamID() + " " + track.ID() + " " + track.Codec().MimeType)
	if strings.Contains(raw, "thermal") {
		return "thermal"
	}
	if strings.Contains(raw, "rgb") {
		return "rgb"
	}
	if track.Kind() == webrtc.RTPCodecTypeAudio {
		return "audio"
	}
	if track.Kind() == webrtc.RTPCodecTypeVideo {
		return "video"
	}
	return track.Kind().String()
}

func publishedTrackKey(robotCode string, label string) string {
	return safeTrackToken(robotCode) + ":" + safeTrackToken(label)
}

func localTrackID(robotCode string, label string) string {
	return safeTrackToken(robotCode) + "-" + safeTrackToken(label)
}

func localStreamID(robotCode string) string {
	return "robot-" + safeTrackToken(robotCode)
}

func safeTrackToken(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", " ", "_")
	return replacer.Replace(value)
}

func cloneRTPPacket(packet *rtp.Packet) *rtp.Packet {
	if packet == nil {
		return nil
	}
	clone := *packet
	clone.Payload = append([]byte(nil), packet.Payload...)
	clone.Header.Extension = false
	clone.Header.ExtensionProfile = 0
	clone.Header.Extensions = nil
	return &clone
}
