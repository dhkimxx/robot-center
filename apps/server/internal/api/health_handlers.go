package api

import (
	"net/http"

	"robot-center/apps/server/internal/api/dto"
)

// @Summary 서버 health 확인
// @Description app-server 기동 상태와 시작 시각을 반환합니다.
// @Tags 시스템 API
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Router /healthz [get]
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, dto.Health(s.started))
}
