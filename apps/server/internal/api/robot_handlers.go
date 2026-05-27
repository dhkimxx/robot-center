package api

import (
	"net/http"
	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
	"strings"
	"time"
)

type createRobotRequest struct {
	DisplayName string `json:"displayName"`
	ModelName   string `json:"modelName"`
}

type updateRobotRequest struct {
	DisplayName string `json:"displayName"`
	ModelName   string `json:"modelName"`
}

func (s *Server) handleCreateRobot(w http.ResponseWriter, r *http.Request) {
	var request createRobotRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	robot, connectionInfo, err := s.services.Robots.CreateRobot(r.Context(), store.CreateRobotInput{
		DisplayName: strings.TrimSpace(request.DisplayName),
		ModelName:   strings.TrimSpace(request.ModelName),
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	now := time.Now().UTC()
	writeJSON(w, http.StatusCreated, map[string]any{
		"robot":          dto.Robot(robot, now, domain.DefaultRobotHeartbeatTTL),
		"connectionInfo": dto.RobotConnectionInfo(connectionInfo),
	})
}

func (s *Server) handleUpdateRobot(w http.ResponseWriter, r *http.Request) {
	var request updateRobotRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	robot, err := s.services.Robots.UpdateRobot(r.Context(), r.PathValue("robotCode"), store.UpdateRobotInput{
		DisplayName: strings.TrimSpace(request.DisplayName),
		ModelName:   strings.TrimSpace(request.ModelName),
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}

	now := time.Now().UTC()
	writeJSON(w, http.StatusOK, map[string]any{
		"robot": dto.Robot(robot, now, domain.DefaultRobotHeartbeatTTL),
	})
}

func (s *Server) handleArchiveRobot(w http.ResponseWriter, r *http.Request) {
	robot, err := s.services.Robots.ArchiveRobot(r.Context(), r.PathValue("robotCode"))
	if err != nil {
		writeStoreError(w, err)
		return
	}

	now := time.Now().UTC()
	writeJSON(w, http.StatusOK, map[string]any{
		"robot": dto.Robot(robot, now, domain.DefaultRobotHeartbeatTTL),
	})
}

func (s *Server) handleListRobots(w http.ResponseWriter, r *http.Request) {
	robots, err := s.services.Robots.ListRobots(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	now := time.Now().UTC()
	writeJSON(w, http.StatusOK, map[string]any{
		"robots": dto.Robots(robots, now, domain.DefaultRobotHeartbeatTTL),
	})
}

func (s *Server) handleGetRobotConnectionInfo(w http.ResponseWriter, r *http.Request) {
	connectionInfo, err := s.services.Robots.GetRobotConnectionInfo(r.Context(), r.PathValue("robotCode"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"connectionInfo": dto.RobotConnectionInfo(connectionInfo),
	})
}

func (s *Server) handleRotateRobotConnectionToken(w http.ResponseWriter, r *http.Request) {
	connectionInfo, err := s.services.Robots.RotateRobotConnectionToken(r.Context(), r.PathValue("robotCode"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"connectionInfo": dto.RobotConnectionInfo(connectionInfo),
	})
}
