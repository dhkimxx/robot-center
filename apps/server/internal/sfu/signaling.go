package sfu

import "time"

func serverPeerPresentMessage(roomID string) signalMessage {
	return signalMessage{
		Type: "peer-present",
		Payload: map[string]any{
			"room":     roomID,
			"role":     "sfu",
			"peerId":   serverPeerID,
			"joinedAt": time.Now().UTC().Format(time.RFC3339Nano),
		},
	}
}

func peerPresencePayload(peer *peer) map[string]any {
	payload := map[string]any{
		"room":     peer.roomID,
		"role":     peer.role,
		"peerId":   peer.id,
		"joinedAt": peer.joinedAt.Format(time.RFC3339Nano),
	}
	if peer.robotCode != "" {
		payload["robotCode"] = peer.robotCode
	}
	return payload
}

func isSubscriberRole(role string) bool {
	return role == "operator" || role == "recorder"
}

func isTargetingServer(payload map[string]any) bool {
	targetPeerID := payloadString(payload, "targetPeerId")
	return targetPeerID == "" || targetPeerID == serverPeerID
}

func payloadString(payload map[string]any, key string) string {
	value, _ := payload[key].(string)
	return value
}

func payloadStringPointer(payload map[string]any, key string) *string {
	value := payloadString(payload, key)
	if value == "" {
		return nil
	}
	return &value
}

func payloadUint16Pointer(payload map[string]any, key string) *uint16 {
	value, ok := payload[key]
	if !ok {
		return nil
	}
	var number uint16
	switch typed := value.(type) {
	case float64:
		number = uint16(typed)
	case int:
		number = uint16(typed)
	default:
		return nil
	}
	return &number
}

func dereferenceString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func dereferenceUint16(value *uint16) uint16 {
	if value == nil {
		return 0
	}
	return *value
}
