package sfu

import (
	"encoding/json"
	"strings"
)

func dataChannelPayloadWithRobotCode(robotCode string, payload []byte) []byte {
	if strings.TrimSpace(robotCode) == "" || !json.Valid(payload) {
		return payload
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil || object == nil {
		return payload
	}
	if _, ok := object["robotCode"]; ok {
		return payload
	}
	object["robotCode"] = strings.TrimSpace(robotCode)
	encoded, err := json.Marshal(object)
	if err != nil {
		return payload
	}
	return encoded
}
