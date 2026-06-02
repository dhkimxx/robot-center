package dto

type RTCConfigResponse struct {
	Mode                 string                 `json:"mode"`
	SignalingURL         string                 `json:"signalingUrl"`
	OperatorSignalingURL string                 `json:"operatorSignalingUrl"`
	ICETransportPolicy   string                 `json:"iceTransportPolicy"`
	ICEServers           []RTCICEServerResponse `json:"iceServers"`
}

type RTCConfigInput struct {
	OperatorSignalingURL string
	TURNURL              string
	TURNUsername         string
	TURNPassword         string
}

type RTCICEServerResponse struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username"`
	Credential string   `json:"credential"`
}

func RTCConfig(input RTCConfigInput) RTCConfigResponse {
	return RTCConfigResponse{
		Mode:                 "sfu",
		SignalingURL:         input.OperatorSignalingURL,
		OperatorSignalingURL: input.OperatorSignalingURL,
		ICETransportPolicy:   "relay",
		ICEServers: []RTCICEServerResponse{
			{
				URLs:       []string{input.TURNURL},
				Username:   input.TURNUsername,
				Credential: input.TURNPassword,
			},
		},
	}
}
