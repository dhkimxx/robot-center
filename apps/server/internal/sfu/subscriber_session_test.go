package sfu

import "testing"

func TestOperatorSubscriberReceivesOnlySelectedRobot(t *testing.T) {
	session := &subscriberSession{role: "operator"}
	currentRoom := &room{
		publishers: map[string]*publisherSession{
			"robot-002": {},
			"robot-001": {},
		},
	}

	session.ensureSelectedRobot(currentRoom)
	if session.selectedRobotCode != "robot-001" {
		t.Fatalf("selectedRobotCode = %q, want robot-001", session.selectedRobotCode)
	}
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
