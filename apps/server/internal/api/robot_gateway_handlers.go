package api

import (
	"errors"
	"net/http"
	"net/url"

	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
	"robot-center/apps/server/internal/utils"
	"strings"
	"time"
)

func (s *Server) handleRobotAPIHeartbeat(w http.ResponseWriter, r *http.Request) {
	var request dto.RobotHeartbeatRequest
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

	writeJSON(w, http.StatusOK, dto.RobotHeartbeatPayload(robot, time.Now().UTC()))
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
		writeJSON(w, http.StatusOK, dto.RobotMissionNonePayload(time.Now().UTC()))
		return
	}

	writeJSON(w, http.StatusOK, s.robotMissionResponse(mission, time.Now().UTC()))
}

func (s *Server) robotMissionResponse(mission domain.Mission, now time.Time) dto.RobotMissionResponse {
	roomID := mission.MissionCode
	return dto.RobotMissionPayload(dto.RobotMissionInput{
		Mission:      mission,
		SignalingURL: s.config.SFURobotWebSocketURL() + "?room=" + url.QueryEscape(roomID),
		TURNURL:      s.config.TURNPublicURL,
		TURNUsername: s.config.TURNUsername,
		TURNPassword: s.config.TURNPassword,
		Now:          now,
	})
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	return strings.TrimPrefix(header, "Bearer ")
}
