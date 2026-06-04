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

// @Summary Robot heartbeat 반영
// @Description Robot token으로 인증된 로봇의 heartbeat와 상태를 반영합니다.
// @Tags Robot API
// @Accept json
// @Produce json
// @Security RobotBearerAuth
// @Param request body dto.RobotHeartbeatRequest true "Robot heartbeat 요청"
// @Success 200 {object} dto.RobotHeartbeatResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/robot/heartbeat [post]
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

// @Summary Robot active mission 조회
// @Description Robot token으로 인증된 로봇의 현재 active mission과 SFU 연결 정보를 반환합니다.
// @Tags Robot API
// @Produce json
// @Security RobotBearerAuth
// @Success 200 {object} dto.RobotMissionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/robot/mission [get]
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
