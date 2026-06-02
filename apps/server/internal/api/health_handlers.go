package api

import (
	"net/http"

	"robot-center/apps/server/internal/api/dto"
)

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, dto.Health(s.started))
}
