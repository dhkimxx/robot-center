package config

import "strings"

const RobotAPISFUWebSocketPath = "/api/v1/robot/sfu/ws"
const OperatorAPISFUWebSocketPath = "/api/v1/operator/sfu/ws"
const RecorderAPISFUWebSocketPath = "/api/v1/recorder/sfu/ws"

func (c AppServerConfig) SFURobotWebSocketURL() string {
	return joinWebSocketPath(c.SFUWebSocketPublicBaseURL, RobotAPISFUWebSocketPath)
}

func (c AppServerConfig) SFUOperatorWebSocketURL() string {
	return joinWebSocketPath(c.SFUWebSocketPublicBaseURL, OperatorAPISFUWebSocketPath)
}

func (c AppServerConfig) SFURecorderWebSocketURL() string {
	return joinWebSocketPath(c.SFUWebSocketPublicBaseURL, RecorderAPISFUWebSocketPath)
}

func (c RecorderWorkerConfig) SFURecorderWebSocketURL() string {
	return joinWebSocketPath(c.SFUWebSocketInternalBaseURL, RecorderAPISFUWebSocketPath)
}

func joinWebSocketPath(baseURL string, path string) string {
	return strings.TrimRight(strings.TrimSpace(baseURL), "/") + path
}
