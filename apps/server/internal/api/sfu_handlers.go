package api

import (
	"net/http"
)

func (s *Server) handleSFUWebSocket(w http.ResponseWriter, r *http.Request) {
	s.sfuHub.ServeHTTP(w, r)
}
