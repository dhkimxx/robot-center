package config

import "strings"

func (c AppServerConfig) SFURobotWebSocketURL() string {
	return joinWebSocketPath(c.SFUWebSocketBaseURL, "/sfu/robot/ws")
}

func (c AppServerConfig) SFUOperatorWebSocketURL() string {
	return joinWebSocketPath(c.SFUWebSocketBaseURL, "/sfu/operator/ws")
}

func (c AppServerConfig) SFURecorderWebSocketURL() string {
	return joinWebSocketPath(c.SFUWebSocketBaseURL, "/sfu/recorder/ws")
}

func (c RecorderWorkerConfig) SFURecorderWebSocketURL() string {
	return joinWebSocketPath(c.SFUWebSocketBaseURL, "/sfu/recorder/ws")
}

func joinWebSocketPath(baseURL string, path string) string {
	return strings.TrimRight(strings.TrimSpace(baseURL), "/") + path
}
