package api

import (
	"net/http"
	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
	"strings"
	"time"
)

// @Summary 로봇 생성
// @Description 관제 서버에 로봇을 등록하고 로봇 런타임 접속 정보를 함께 반환합니다.
// @Tags Operator API
// @Accept json
// @Produce json
// @Param request body dto.CreateRobotRequest true "로봇 생성 요청"
// @Success 201 {object} dto.CreateRobotResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/v1/operator/robots [post]
func (s *Server) handleCreateRobot(w http.ResponseWriter, r *http.Request) {
	var request dto.CreateRobotRequest
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
	writeJSON(w, http.StatusCreated, dto.CreateRobotPayload(robot, connectionInfo, now, domain.DefaultRobotHeartbeatTTL))
}

// @Summary 로봇 수정
// @Description 로봇 표시 이름과 모델명을 수정합니다.
// @Tags Operator API
// @Accept json
// @Produce json
// @Param robotCode path string true "로봇 코드"
// @Param request body dto.UpdateRobotRequest true "로봇 수정 요청"
// @Success 200 {object} dto.RobotEnvelopeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/operator/robots/{robotCode} [patch]
func (s *Server) handleUpdateRobot(w http.ResponseWriter, r *http.Request) {
	var request dto.UpdateRobotRequest
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
	writeJSON(w, http.StatusOK, dto.RobotEnvelope(robot, now, domain.DefaultRobotHeartbeatTTL))
}

// @Summary 로봇 보관 처리
// @Description 로봇을 active 목록에서 제외합니다.
// @Tags Operator API
// @Produce json
// @Param robotCode path string true "로봇 코드"
// @Success 200 {object} dto.RobotEnvelopeResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/operator/robots/{robotCode} [delete]
func (s *Server) handleArchiveRobot(w http.ResponseWriter, r *http.Request) {
	robot, err := s.services.Robots.ArchiveRobot(r.Context(), r.PathValue("robotCode"))
	if err != nil {
		writeStoreError(w, err)
		return
	}

	now := time.Now().UTC()
	writeJSON(w, http.StatusOK, dto.RobotEnvelope(robot, now, domain.DefaultRobotHeartbeatTTL))
}

// @Summary 로봇 목록 조회
// @Description 관제 서버에 등록된 로봇 목록을 반환합니다.
// @Tags Operator API
// @Produce json
// @Success 200 {object} dto.RobotsResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/operator/robots [get]
func (s *Server) handleListRobots(w http.ResponseWriter, r *http.Request) {
	robots, err := s.services.Robots.ListRobots(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	now := time.Now().UTC()
	writeJSON(w, http.StatusOK, dto.RobotsPayload(robots, now, domain.DefaultRobotHeartbeatTTL))
}

// @Summary 로봇 연결 정보 조회
// @Description 로봇 런타임 접속에 필요한 serverUrl, robotCode, robotToken을 조회합니다.
// @Tags Operator API
// @Produce json
// @Param robotCode path string true "로봇 코드"
// @Success 200 {object} dto.RobotConnectionInfoEnvelopeResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/operator/robots/{robotCode}/connection-info [get]
func (s *Server) handleGetRobotConnectionInfo(w http.ResponseWriter, r *http.Request) {
	connectionInfo, err := s.services.Robots.GetRobotConnectionInfo(r.Context(), r.PathValue("robotCode"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.RobotConnectionInfoPayload(connectionInfo))
}

// @Summary 로봇 token 재발급
// @Description 로봇 API용 robotToken을 재발급합니다.
// @Tags Operator API
// @Produce json
// @Param robotCode path string true "로봇 코드"
// @Success 200 {object} dto.RobotConnectionInfoEnvelopeResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/operator/robots/{robotCode}/connection-token [post]
func (s *Server) handleRotateRobotConnectionToken(w http.ResponseWriter, r *http.Request) {
	connectionInfo, err := s.services.Robots.RotateRobotConnectionToken(r.Context(), r.PathValue("robotCode"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.RobotConnectionInfoPayload(connectionInfo))
}
