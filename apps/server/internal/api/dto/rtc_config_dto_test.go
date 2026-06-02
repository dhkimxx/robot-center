package dto

import (
	"encoding/json"
	"testing"
)

func TestRTCConfigResponseShape(t *testing.T) {
	response := RTCConfig(RTCConfigInput{
		OperatorSignalingURL: "ws://center.local/api/v1/operator/sfu/ws",
		TURNURL:              "turn:center.local:3478?transport=udp",
		TURNUsername:         "robot",
		TURNPassword:         "robot-pass",
	})

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var fields map[string]any
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if fields["mode"] != "sfu" {
		t.Fatalf("mode = %q, want sfu", fields["mode"])
	}
	if fields["signalingUrl"] != fields["operatorSignalingUrl"] {
		t.Fatalf("expected signalingUrl and operatorSignalingUrl to match, got %s", string(payload))
	}
	if fields["iceTransportPolicy"] != "relay" {
		t.Fatalf("iceTransportPolicy = %q, want relay", fields["iceTransportPolicy"])
	}
	servers, ok := fields["iceServers"].([]any)
	if !ok || len(servers) != 1 {
		t.Fatalf("expected one ice server, got %s", string(payload))
	}
}
