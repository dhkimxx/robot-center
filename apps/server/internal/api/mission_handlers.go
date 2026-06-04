package api

import (
	"fmt"
	"net/http"

	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/store"

	"strings"
)

// @Summary 임무 생성
// @Description 로봇을 배정한 임무를 생성합니다.
// @Tags Operator API
// @Accept json
// @Produce json
// @Param request body dto.CreateMissionRequest true "임무 생성 요청"
// @Success 201 {object} dto.MissionEnvelopeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/v1/operator/missions [post]
func (s *Server) handleCreateMission(w http.ResponseWriter, r *http.Request) {
	var request dto.CreateMissionRequest
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

	writeJSON(w, http.StatusCreated, dto.MissionPayload(mission))
}

// @Summary 임무 목록 조회
// @Description 관제 서버에 등록된 임무 목록을 반환합니다.
// @Tags Operator API
// @Produce json
// @Success 200 {object} dto.MissionsResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/operator/missions [get]
func (s *Server) handleListMissions(w http.ResponseWriter, r *http.Request) {
	missions, err := s.services.Missions.ListMissions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.MissionsPayload(missions))
}

// @Summary 임무 시작
// @Description ready 상태 임무를 active 상태로 전환합니다.
// @Tags Operator API
// @Produce json
// @Param missionCode path string true "임무 코드"
// @Success 200 {object} dto.MissionEnvelopeResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/v1/operator/missions/{missionCode}/start [post]
func (s *Server) handleStartMission(w http.ResponseWriter, r *http.Request) {
	mission, err := s.services.Missions.StartMission(r.Context(), r.PathValue("missionCode"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.MissionPayload(mission))
}

// @Summary 임무 종료
// @Description active 상태 임무를 ended 상태로 전환하고 SFU room을 닫습니다.
// @Tags Operator API
// @Produce json
// @Param missionCode path string true "임무 코드"
// @Success 200 {object} dto.MissionEnvelopeResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/v1/operator/missions/{missionCode}/end [post]
func (s *Server) handleEndMission(w http.ResponseWriter, r *http.Request) {
	mission, err := s.services.Missions.EndMission(r.Context(), r.PathValue("missionCode"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	s.sfuHub.CloseRoom(mission.MissionCode)
	writeJSON(w, http.StatusOK, dto.MissionPayload(mission))
}

func validMissionType(missionType string) bool {
	switch missionType {
	case "mountain_rescue", "collapse_site", "underground_facility":
		return true
	default:
		return false
	}
}
