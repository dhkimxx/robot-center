package api

import (
	"fmt"
	"net/http"
	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/store"
	"strings"
)

type createMissionRequest struct {
	Name        string   `json:"name"`
	MissionType string   `json:"missionType"`
	SiteNote    string   `json:"siteNote"`
	RobotCode   string   `json:"robotCode"`
	RobotCodes  []string `json:"robotCodes"`
}

func (s *Server) handleCreateMission(w http.ResponseWriter, r *http.Request) {
	var request createMissionRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if !validMissionType(request.MissionType) {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid missionType %q", request.MissionType))
		return
	}

	mission, err := s.services.Missions.CreateMission(r.Context(), store.CreateMissionInput{
		Name:        strings.TrimSpace(request.Name),
		MissionType: strings.TrimSpace(request.MissionType),
		SiteNote:    strings.TrimSpace(request.SiteNote),
		RobotCode:   strings.TrimSpace(request.RobotCode),
		RobotCodes:  request.RobotCodes,
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"mission": dto.Mission(mission),
	})
}

func (s *Server) handleListMissions(w http.ResponseWriter, r *http.Request) {
	missions, err := s.services.Missions.ListMissions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"missions": dto.Missions(missions),
	})
}

func (s *Server) handleStartMission(w http.ResponseWriter, r *http.Request) {
	mission, err := s.services.Missions.StartMission(r.Context(), r.PathValue("missionCode"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"mission": dto.Mission(mission),
	})
}

func (s *Server) handleEndMission(w http.ResponseWriter, r *http.Request) {
	mission, err := s.services.Missions.EndMission(r.Context(), r.PathValue("missionCode"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	s.sfuHub.CloseRoom(mission.MissionCode)
	writeJSON(w, http.StatusOK, map[string]any{
		"mission": dto.Mission(mission),
	})
}

func validMissionType(missionType string) bool {
	switch missionType {
	case "mountain_rescue", "collapse_site", "underground_facility":
		return true
	default:
		return false
	}
}
