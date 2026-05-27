package sfu

import (
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"
	"strings"
)

func (h *Hub) createPeerConnection() (*webrtc.PeerConnection, error) {
	mediaEngine := &webrtc.MediaEngine{}
	if err := mediaEngine.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}
	interceptorRegistry := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(mediaEngine, interceptorRegistry); err != nil {
		return nil, err
	}
	api := webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(interceptorRegistry),
	)

	configuration := webrtc.Configuration{}
	if strings.TrimSpace(h.config.TURNURL) != "" {
		configuration.ICEServers = []webrtc.ICEServer{
			{
				URLs:       []string{h.config.TURNURL},
				Username:   h.config.TURNUsername,
				Credential: h.config.TURNPassword,
			},
		}
		configuration.ICETransportPolicy = webrtc.ICETransportPolicyRelay
	}
	return api.NewPeerConnection(configuration)
}
