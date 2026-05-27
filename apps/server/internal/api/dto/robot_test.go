package dto

import (
	"testing"
	"time"

	"robot-center/apps/server/internal/domain"
)

func TestRobotResponseUsesDomainConnectionState(t *testing.T) {
	now := time.Date(2026, 5, 27, 11, 0, 0, 0, time.UTC)
	staleSeenAt := now.Add(-2 * time.Minute)

	response := Robot(domain.Robot{
		RobotCode:   "robot-003",
		DeviceState: domain.RobotDeviceStateOnline,
		LastSeenAt:  &staleSeenAt,
	}, now, 30*time.Second)

	if response.Status != domain.RobotConnectionStateOffline {
		t.Fatalf("status = %q, want %q", response.Status, domain.RobotConnectionStateOffline)
	}
}
