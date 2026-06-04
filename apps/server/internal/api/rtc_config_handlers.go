package api

import (
	"net/http"

	"robot-center/apps/server/internal/api/dto"
)

// @Summary 관제 WebRTC 설정 조회
// @Description 관제 UI operator peer가 사용할 signaling URL과 ICE 서버 설정을 반환합니다.
// @Tags Operator API
// @Produce json
// @Success 200 {object} dto.RTCConfigResponse
// @Router /api/v1/operator/rtc-config [get]
func (s *Server) handleRTCConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, dto.RTCConfig(dto.RTCConfigInput{
		OperatorSignalingURL: s.config.SFUOperatorWebSocketURL(),
		TURNURL:              s.config.TURNPublicURL,
		TURNUsername:         s.config.TURNUsername,
		TURNPassword:         s.config.TURNPassword,
	}))
}
