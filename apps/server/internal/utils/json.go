package utils

import "encoding/json"

func RawJSONOrEmpty(value any) json.RawMessage {
	payload, err := json.Marshal(value)
	if err != nil || len(payload) == 0 || string(payload) == "null" {
		return []byte("{}")
	}
	return payload
}

func RawJSONOrNil(value any) json.RawMessage {
	payload, err := json.Marshal(value)
	if err != nil || len(payload) == 0 || string(payload) == "null" {
		return nil
	}
	return payload
}
