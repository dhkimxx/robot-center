package api

import (
	"errors"
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
	State          string    `json:"state"`
	BatteryPercent int       `json:"batteryPercent"`
	NetworkQuality string    `json:"networkQuality"`
	SentAt         time.Time `json:"sentAt"`
}

func (s *Server) handleRobotAPIHeartbeat(w http.ResponseWriter, r *http.Request) {
	var request heartbeatRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	robot, err := s.services.Robots.ApplyHeartbeat(r.Context(), store.HeartbeatInput{
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
		"robotCode":  robot.RobotCode,
		"status":     robot.DeviceState,
		"serverTime": time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func (s *Server) handleRobotAPIMission(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(r.URL.Query().Get("robotCode")) != "" {
		writeError(w, http.StatusBadRequest, errors.New("robotCode query is not allowed for robot API mission lookup"))
		return
	}
	mission, found, err := s.services.Missions.FindActiveMissionForRobot(r.Context(), bearerToken(r))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if !found {
		writeJSON(w, http.StatusOK, map[string]any{
			"missionStatus": "none",
			"serverTime":    time.Now().UTC().Format(time.RFC3339Nano),
		})
		return
	}

	writeJSON(w, http.StatusOK, s.robotMissionResponse(mission))
}

func (s *Server) robotMissionResponse(mission domain.Mission) map[string]any {
	roomID := mission.MissionCode
	return map[string]any{
		"missionCode":   mission.MissionCode,
		"missionStatus": mission.Status,
		"serverTime":    time.Now().UTC().Format(time.RFC3339Nano),
		"sfu": map[string]any{
			"signalingUrl":       s.config.SFURobotWebSocketURL() + "?room=" + url.QueryEscape(roomID),
			"iceTransportPolicy": "relay",
		},
		"turnServers": []map[string]any{
			{
				"urls":       []string{s.config.TURNPublicURL},
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
	}
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	return strings.TrimPrefix(header, "Bearer ")
}
