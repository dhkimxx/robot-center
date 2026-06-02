package sfu

import (
	"robot-center/apps/server/internal/utils"

	"github.com/pion/rtp"
)

func publishedTrackKey(robotCode string, label string) string {
	return utils.SafeTrackToken(robotCode) + ":" + utils.SafeTrackToken(label)
}

func localTrackID(robotCode string, label string) string {
	return utils.SafeTrackToken(robotCode) + "-" + utils.SafeTrackToken(label)
}

func localStreamID(robotCode string) string {
	return "robot-" + utils.SafeTrackToken(robotCode)
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
