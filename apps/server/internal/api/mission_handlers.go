package api

import (
	"fmt"
	"net/http"
	"strings"

	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
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
// @Description 관제 서버에 등록된 임무 목록을 반환합니다. limit, offset, sort, order, filter query로 관제 화면용 페이지네이션과 정렬을 적용할 수 있습니다. limit을 생략하면 기존 호환을 위해 전체 목록을 반환합니다.
// @Tags Operator API
// @Produce json
// @Param limit query int false "반환할 최대 임무 개수. 생략하면 전체 반환, 최대 200."
// @Param offset query int false "건너뛸 임무 개수. 기본 0."
// @Param sort query string false "정렬 기준: missionCode, name, missionType, status, createdAt, startedAt, endedAt"
// @Param order query string false "정렬 방향: asc 또는 desc. 기본 asc."
// @Param filter query string false "임무명, 임무 코드, 시나리오, 상태, 현장 메모, 로봇 코드 검색어."
// @Success 200 {object} dto.MissionsResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/operator/missions [get]
func (s *Server) handleListMissions(w http.ResponseWriter, r *http.Request) {
	missions, err := s.services.Missions.ListMissions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	result := applyListQuery(
		missions,
		parseListQuery(r, operatorMissionListSorts()),
		missionMatchesOperatorListFilter,
		lessMissionByOperatorListSort,
	)
	writeJSON(w, http.StatusOK, dto.MissionsPayload(result.items, result.meta))
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

func operatorMissionListSorts() map[string]string {
	return map[string]string{
		"missionCode": "missionCode",
		"missioncode": "missionCode",
		"name":        "name",
		"missionType": "missionType",
		"missiontype": "missionType",
		"status":      "status",
		"createdAt":   "createdAt",
		"createdat":   "createdAt",
		"startedAt":   "startedAt",
		"startedat":   "startedAt",
		"endedAt":     "endedAt",
		"endedat":     "endedAt",
	}
}

func missionMatchesOperatorListFilter(mission domain.Mission, filter string) bool {
	values := []string{
		mission.MissionCode,
		mission.Name,
		mission.MissionType,
		mission.Status,
		mission.SiteNote,
		mission.RobotCode,
	}
	values = append(values, mission.RobotCodes...)
	return containsListFilterValue(filter, values...)
}

func lessMissionByOperatorListSort(left domain.Mission, right domain.Mission, sortKey string) bool {
	switch sortKey {
	case "missionCode":
		return lessListString(left.MissionCode, right.MissionCode)
	case "name":
		return lessListString(left.Name, right.Name)
	case "missionType":
		return lessListString(left.MissionType, right.MissionType)
	case "status":
		return lessListString(left.Status, right.Status)
	case "createdAt":
		return lessListTime(left.CreatedAt, right.CreatedAt)
	case "startedAt":
		return lessOptionalListTime(left.StartedAt, right.StartedAt)
	case "endedAt":
		return lessOptionalListTime(left.EndedAt, right.EndedAt)
	default:
		return false
	}
}
