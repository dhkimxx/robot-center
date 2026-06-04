package api

import (
	"errors"
	"net/http"
	"strings"

	"robot-center/apps/server/internal/sfu"
	"robot-center/apps/server/internal/store"
)

// @Summary Robot SFU WebSocket 연결
// @Description Robot publisher가 mission room에 WebRTC signaling용 WebSocket으로 접속합니다.
// @Tags Robot API
// @Security RobotBearerAuth
// @Param room query string true "접속할 missionCode room"
// @Success 101 {string} string "Switching Protocols"
// @Router /api/v1/robot/sfu/ws [get]
func (s *Server) handleRobotSFUWebSocket(w http.ResponseWriter, r *http.Request) {
	roomID, ok := s.sfuRoomID(w, r)
	if !ok {
		return
	}
	robot, err := s.services.Robots.ResolveRobotByBearerToken(r.Context(), bearerToken(r))
	if err != nil {
		if errors.Is(err, store.ErrUnauthorized) {
			writeError(w, http.StatusUnauthorized, err)
			return
		}
		writeStoreError(w, err)
		return
	}
	if err := s.services.Missions.ValidateActiveMissionRobot(r.Context(), roomID, robot.RobotCode); err != nil {
		if errors.Is(err, store.ErrUnauthorized) || errors.Is(err, store.ErrNotFound) || errors.Is(err, store.ErrInvalidState) {
			writeError(w, http.StatusForbidden, err)
			return
		}
		writeStoreError(w, err)
		return
	}
	s.sfuHub.ServePeer(w, r, sfu.PeerJoinRequest{
		RoomID:    roomID,
		Role:      "robot",
		RobotCode: robot.RobotCode,
	})
}

// @Summary Operator SFU WebSocket 연결
// @Description 관제 UI operator peer가 mission room에 WebRTC signaling용 WebSocket으로 접속합니다.
// @Tags Operator API
// @Param room query string true "접속할 missionCode room"
// @Success 101 {string} string "Switching Protocols"
// @Router /api/v1/operator/sfu/ws [get]
func (s *Server) handleOperatorSFUWebSocket(w http.ResponseWriter, r *http.Request) {
	roomID, ok := s.sfuRoomID(w, r)
	if !ok {
		return
	}
	if strings.TrimSpace(r.URL.Query().Get("robotCode")) != "" {
		writeError(w, http.StatusBadRequest, errors.New("robotCode query is not allowed for operator websocket"))
		return
	}
	s.sfuHub.ServePeer(w, r, sfu.PeerJoinRequest{
		RoomID: roomID,
		Role:   "operator",
	})
}

// @Summary Recorder SFU WebSocket 연결
// @Description recorder-worker peer가 mission room에 WebRTC signaling용 WebSocket으로 접속합니다.
// @Tags Recorder API
// @Param room query string true "접속할 missionCode room"
// @Success 101 {string} string "Switching Protocols"
// @Router /api/v1/recorder/sfu/ws [get]
func (s *Server) handleRecorderSFUWebSocket(w http.ResponseWriter, r *http.Request) {
	roomID, ok := s.sfuRoomID(w, r)
	if !ok {
		return
	}
	if strings.TrimSpace(r.URL.Query().Get("robotCode")) != "" {
		writeError(w, http.StatusBadRequest, errors.New("robotCode query is not allowed for recorder websocket"))
		return
	}
	s.sfuHub.ServePeer(w, r, sfu.PeerJoinRequest{
		RoomID: roomID,
		Role:   "recorder",
	})
}

func (s *Server) sfuRoomID(w http.ResponseWriter, r *http.Request) (string, bool) {
	roomID := strings.TrimSpace(r.URL.Query().Get("room"))
	if roomID == "" {
		writeError(w, http.StatusBadRequest, errors.New("room query parameter is required"))
		return "", false
	}
	if strings.TrimSpace(r.URL.Query().Get("role")) != "" {
		writeError(w, http.StatusBadRequest, errors.New("role query is not allowed"))
		return "", false
	}
	return roomID, true
}
