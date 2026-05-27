package api

import (
	"net/http"
	"net/url"
	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/sfu"
	"robot-center/apps/server/internal/store"
	"robot-center/apps/server/internal/utils"
	"strings"
	"time"
)

type heartbeatRequest struct {
	RobotCode      string    `json:"robotCode"`
	State          string    `json:"state"`
	BatteryPercent int       `json:"batteryPercent"`
	NetworkQuality string    `json:"networkQuality"`
	SentAt         time.Time `json:"sentAt"`
}

func (s *Server) handleRobotGatewayHeartbeat(w http.ResponseWriter, r *http.Request) {
	var request heartbeatRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	robot, err := s.services.Robots.ApplyHeartbeat(r.Context(), store.HeartbeatInput{
		RobotCode:      strings.TrimSpace(request.RobotCode),
		State:          utils.FirstNonEmptyString(request.State, "online"),
		BatteryPercent: request.BatteryPercent,
		NetworkQuality: request.NetworkQuality,
		SentAt:         request.SentAt,
	}, bearerToken(r))
	if err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"robotId":    robot.ID,
		"robotCode":  robot.RobotCode,
		"status":     robot.DeviceState,
		"serverTime": time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func (s *Server) handleRobotGatewayMission(w http.ResponseWriter, r *http.Request) {
	robotCode := strings.TrimSpace(r.URL.Query().Get("robotCode"))
	mission, found, err := s.services.Missions.FindActiveMissionForRobot(r.Context(), robotCode, bearerToken(r))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if !found {
		writeJSON(w, http.StatusOK, map[string]any{
			"missionId":     nil,
			"missionStatus": "none",
		})
		return
	}

	roomID := mission.MissionCode
	legacyRoomID := domain.StreamRoomID(mission.MissionCode, robotCode)
	writeJSON(w, http.StatusOK, map[string]any{
		"missionId":     mission.ID,
		"missionCode":   mission.MissionCode,
		"missionStatus": mission.Status,
		"roomId":        roomID,
		"legacyRoomId":  legacyRoomID,
		"sfu": map[string]any{
			"signalingUrl":       s.config.SFUWebSocketURL + "?room=" + url.QueryEscape(roomID) + "&role=robot&robotCode=" + url.QueryEscape(robotCode),
			"iceTransportPolicy": "relay",
		},
		"turnServers": []map[string]any{
			{
				"urls":       []string{s.config.TURNURL},
				"username":   s.config.TURNUsername,
				"credential": s.config.TURNPassword,
			},
		},
		"tracks": []string{
			sfu.StreamRoleTrackVideo1,
			sfu.StreamRoleTrackVideo2,
			sfu.StreamRoleTrackAudio1,
			sfu.StreamRoleTrackAudio2,
		},
		"dataChannels": []string{
			sfu.StreamRoleChannelTelemetry,
			sfu.StreamRoleChannelSpatial,
			sfu.StreamRoleChannelEvent,
			sfu.StreamRoleChannelControl,
		},
		"legacyTracks":       []string{"rgb", "thermal", "audio"},
		"legacyDataChannels": []string{"sensor", "telemetry"},
		"videoPolicy": map[string]string{
			"mode": "robot_defined",
		},
	})
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	return strings.TrimPrefix(header, "Bearer ")
}
