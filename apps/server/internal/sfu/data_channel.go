package sfu

import (
	"encoding/json"
	"strings"
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
