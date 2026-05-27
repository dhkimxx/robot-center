package domain

import (
	"strings"
	"time"
)

type Robot struct {
	ID          string           `json:"id"`
	RobotCode   string           `json:"robotCode"`
	DisplayName string           `json:"displayName"`
	ModelName   string           `json:"modelName,omitempty"`
	DeviceState RobotDeviceState `json:"deviceState"`
	LastSeenAt  *time.Time       `json:"lastSeenAt,omitempty"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
}

type RobotConnectionInfo struct {
	ServerURL  string `json:"serverUrl"`
	RobotCode  string `json:"robotCode"`
	RobotToken string `json:"robotToken"`
}

type RobotDeviceState string
type RobotConnectionState string

const (
	RobotDeviceStateFault   RobotDeviceState = "fault"
	RobotDeviceStateOffline RobotDeviceState = "offline"
	RobotDeviceStateOnline  RobotDeviceState = "online"
)

const (
	RobotConnectionStateFault   RobotConnectionState = "fault"
	RobotConnectionStateOffline RobotConnectionState = "offline"
	RobotConnectionStateOnline  RobotConnectionState = "online"
)

const DefaultRobotHeartbeatTTL = 30 * time.Second

func NormalizeRobotDeviceState(state string) RobotDeviceState {
	switch strings.TrimSpace(state) {
	case "":
		return RobotDeviceStateOffline
	case string(RobotDeviceStateOffline):
		return RobotDeviceStateOffline
	case string(RobotDeviceStateFault):
		return RobotDeviceStateFault
	default:
		return RobotDeviceStateOnline
	}
}

func (r Robot) IsOnline(now time.Time, ttl time.Duration) bool {
	if r.DeviceState != RobotDeviceStateOnline || r.LastSeenAt == nil || r.LastSeenAt.IsZero() {
		return false
	}
	if ttl <= 0 {
		return false
	}
	delta := now.Sub(*r.LastSeenAt)
	if delta < 0 {
		delta = -delta
	}
	return delta <= ttl
}

func (r Robot) IsFault() bool {
	return r.DeviceState == RobotDeviceStateFault
}

func (r Robot) ConnectionState(now time.Time, ttl time.Duration) RobotConnectionState {
	if r.IsFault() {
		return RobotConnectionStateFault
	}
	if r.IsOnline(now, ttl) {
		return RobotConnectionStateOnline
	}
	return RobotConnectionStateOffline
}
