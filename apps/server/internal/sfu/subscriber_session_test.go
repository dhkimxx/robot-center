package sfu

import "testing"

func TestOperatorSubscriberDoesNotAutoSelectRobot(t *testing.T) {
	session := &subscriberSession{role: "operator"}
	currentRoom := &room{
		publishers: map[string]*publisherSession{
			"robot-002": {},
			"robot-001": {},
		},
	}

	if session.beginOffer(currentRoom, nil, nil) {
		t.Fatalf("operator should not create an offer before selecting a robot")
	}
	if session.selectedRobotCode != "" {
		t.Fatalf("selectedRobotCode = %q, want empty before explicit selection", session.selectedRobotCode)
	}
	if session.shouldReceiveRobot("robot-001") || session.shouldReceiveRobot("robot-002") {
		t.Fatalf("operator should not receive robot streams before explicit selection")
	}
}

func TestOperatorSubscriberReceivesOnlySelectedRobot(t *testing.T) {
	session := &subscriberSession{role: "operator"}

	session.selectRobot("robot-001")
	if !session.shouldReceiveRobot("robot-001") {
		t.Fatalf("operator should receive the selected robot")
	}
	if session.shouldReceiveRobot("robot-002") {
		t.Fatalf("operator should not receive an unselected robot")
	}

	session.selectRobot("robot-002")
	if !session.shouldReceiveRobot("robot-002") {
		t.Fatalf("operator should receive the newly selected robot")
	}
	if session.shouldReceiveRobot("robot-001") {
		t.Fatalf("operator should stop receiving the previous robot")
	}
}

func TestRecorderSubscriberReceivesEveryRobot(t *testing.T) {
	session := &subscriberSession{role: "recorder"}
	session.selectRobot("robot-001")

	if !session.shouldReceiveRobot("robot-001") || !session.shouldReceiveRobot("robot-002") {
		t.Fatalf("recorder should receive all robot streams")
	}
}

func TestCanonicalDataChannelRoles(t *testing.T) {
	cases := map[string]string{
		"channel.telemetry": StreamRoleChannelTelemetry,
		"channel.event":     StreamRoleChannelEvent,
		"channel.spatial":   StreamRoleChannelSpatial,
		"channel.control":   StreamRoleChannelControl,
		"telemetry":         "telemetry",
		"sensor":            "sensor",
		"event":             "event",
		"spatial":           "spatial",
		"control":           "control",
	}
	for input, want := range cases {
		if got := normalizeDataChannelRole(input); got != want {
			t.Fatalf("normalizeDataChannelRole(%q) = %q, want %q", input, got, want)
		}
	}
}
