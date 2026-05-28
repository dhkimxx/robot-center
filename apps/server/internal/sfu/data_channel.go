package sfu

import (
	"encoding/json"
	"strings"
	"time"
)

func dataChannelPayloadWithContext(roomID string, robotCode string, channelRole string, payload []byte) []byte {
	if !json.Valid(payload) {
		return payload
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil || object == nil {
		return payload
	}
	if strings.TrimSpace(robotCode) != "" {
		object["robotCode"] = strings.TrimSpace(robotCode)
	}
	if strings.TrimSpace(roomID) != "" {
		object["missionId"] = strings.TrimSpace(roomID)
		object["missionCode"] = strings.TrimSpace(roomID)
	}
	if strings.TrimSpace(channelRole) != "" {
		object["channelRole"] = strings.TrimSpace(channelRole)
	}
	encoded, err := json.Marshal(object)
	if err != nil {
		return payload
	}
	return encoded
}

func ensurePublishedDataChannel(publisher *publisherSession, label string, observedAt time.Time) *PublishedDataChannel {
	if publisher == nil {
		return nil
	}
	if observedAt.IsZero() {
		observedAt = time.Now().UTC()
	}
	if publisher.streamBundle == nil {
		publisher.streamBundle = newRobotStreamBundle("", publisher.robotCode)
	}
	channel := publisher.streamBundle.DataChannels[label]
	if channel == nil {
		channel = &PublishedDataChannel{
			Role:       label,
			State:      "detected",
			DetectedAt: cloneTimePointer(&observedAt),
		}
		publisher.streamBundle.DataChannels[label] = channel
		return channel
	}
	if strings.TrimSpace(channel.Role) == "" {
		channel.Role = label
	}
	if strings.TrimSpace(channel.State) == "" {
		channel.State = "detected"
	}
	if channel.DetectedAt == nil {
		channel.DetectedAt = cloneTimePointer(&observedAt)
	}
	return channel
}

func publishedDataChannelState(channel *PublishedDataChannel) string {
	if channel == nil {
		return "unknown"
	}
	if strings.TrimSpace(channel.State) != "" {
		return strings.TrimSpace(channel.State)
	}
	return "detected"
}
