package dto

import (
	"time"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/sfu"
)

type RobotHeartbeatRequest struct {
	State          string    `json:"state"`
	BatteryPercent int       `json:"batteryPercent"`
	NetworkQuality string    `json:"networkQuality"`
	SentAt         time.Time `json:"sentAt"`
}

type RobotHeartbeatResponse struct {
	RobotCode  string                  `json:"robotCode"`
	Status     domain.RobotDeviceState `json:"status"`
	ServerTime string                  `json:"serverTime"`
}

type RobotMissionResponse struct {
	MissionCode   string                    `json:"missionCode,omitempty"`
	MissionStatus string                    `json:"missionStatus"`
	ServerTime    string                    `json:"serverTime"`
	SFU           *RobotSFUConfigResponse   `json:"sfu,omitempty"`
	TurnServers   []RobotTurnServerResponse `json:"turnServers,omitempty"`
	Tracks        []string                  `json:"tracks,omitempty"`
	DataChannels  []string                  `json:"dataChannels,omitempty"`
}

type RobotSFUConfigResponse struct {
	SignalingURL       string `json:"signalingUrl"`
	ICETransportPolicy string `json:"iceTransportPolicy"`
}

type RobotTurnServerResponse struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username"`
	Credential string   `json:"credential"`
}

type RobotMissionInput struct {
	Mission      domain.Mission
	SignalingURL string
	TURNURL      string
	TURNUsername string
	TURNPassword string
	Now          time.Time
}

func RobotHeartbeatPayload(robot domain.Robot, now time.Time) RobotHeartbeatResponse {
	return RobotHeartbeatResponse{
		RobotCode:  robot.RobotCode,
		Status:     robot.DeviceState,
		ServerTime: now.Format(time.RFC3339Nano),
	}
}

func RobotMissionNonePayload(now time.Time) RobotMissionResponse {
	return RobotMissionResponse{
		MissionStatus: "none",
		ServerTime:    now.Format(time.RFC3339Nano),
	}
}

func RobotMissionPayload(input RobotMissionInput) RobotMissionResponse {
	return RobotMissionResponse{
		MissionCode:   input.Mission.MissionCode,
		MissionStatus: input.Mission.Status,
		ServerTime:    input.Now.Format(time.RFC3339Nano),
		SFU: &RobotSFUConfigResponse{
			SignalingURL:       input.SignalingURL,
			ICETransportPolicy: "relay",
		},
		TurnServers: []RobotTurnServerResponse{
			{
				URLs:       []string{input.TURNURL},
				Username:   input.TURNUsername,
				Credential: input.TURNPassword,
			},
		},
		Tracks: []string{
			sfu.StreamRoleTrackVideo1,
			sfu.StreamRoleTrackVideo2,
			sfu.StreamRoleTrackAudio1,
			sfu.StreamRoleTrackAudio2,
		},
		DataChannels: []string{
			sfu.StreamRoleChannelTelemetry,
			sfu.StreamRoleChannelSpatial,
			sfu.StreamRoleChannelEvent,
			sfu.StreamRoleChannelControl,
		},
	}
}
