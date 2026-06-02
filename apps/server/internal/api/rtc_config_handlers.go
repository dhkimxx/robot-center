package api

import (
	"net/http"

	"robot-center/apps/server/internal/api/dto"
)

func (s *Server) handleRTCConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, dto.RTCConfig(dto.RTCConfigInput{
		OperatorSignalingURL: s.config.SFUOperatorWebSocketURL(),
		TURNURL:              s.config.TURNPublicURL,
		TURNUsername:         s.config.TURNUsername,
		TURNPassword:         s.config.TURNPassword,
	}))
}
