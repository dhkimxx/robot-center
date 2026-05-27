package domain

import (
	"testing"
	"time"
)

func TestRobotConnectionStateUsesHeartbeatFreshness(t *testing.T) {
	now := time.Date(2026, 5, 27, 11, 0, 0, 0, time.UTC)
	freshSeenAt := now.Add(-10 * time.Second)
	staleSeenAt := now.Add(-2 * time.Minute)

	tests := []struct {
		name  string
		robot Robot
		want  RobotConnectionState
	}{
		{
			name:  "fresh heartbeat is online",
			robot: Robot{DeviceState: RobotDeviceStateOnline, LastSeenAt: &freshSeenAt},
			want:  RobotConnectionStateOnline,
		},
		{
			name:  "stale persisted online status is offline",
			robot: Robot{DeviceState: RobotDeviceStateOnline, LastSeenAt: &staleSeenAt},
			want:  RobotConnectionStateOffline,
		},
		{
			name:  "missing heartbeat is offline",
			robot: Robot{DeviceState: RobotDeviceStateOnline},
			want:  RobotConnectionStateOffline,
		},
		{
			name:  "explicit offline device state wins over fresh heartbeat",
			robot: Robot{DeviceState: RobotDeviceStateOffline, LastSeenAt: &freshSeenAt},
			want:  RobotConnectionStateOffline,
		},
		{
			name:  "fault status wins over fresh heartbeat",
			robot: Robot{DeviceState: RobotDeviceStateFault, LastSeenAt: &freshSeenAt},
			want:  RobotConnectionStateFault,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.robot.ConnectionState(now, 30*time.Second); got != test.want {
				t.Fatalf("ConnectionState() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNormalizeRobotDeviceStateKeepsLegacyActiveStatesOnline(t *testing.T) {
	if got := NormalizeRobotDeviceState("streaming"); got != RobotDeviceStateOnline {
		t.Fatalf("NormalizeRobotDeviceState(streaming) = %q, want %q", got, RobotDeviceStateOnline)
	}
	if got := NormalizeRobotDeviceState(""); got != RobotDeviceStateOffline {
		t.Fatalf("NormalizeRobotDeviceState(empty) = %q, want %q", got, RobotDeviceStateOffline)
	}
}
