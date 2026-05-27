package api

import (
	"errors"
	"net/http"
	"strings"

	"robot-center/apps/server/internal/sfu"
	"robot-center/apps/server/internal/store"
)

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
