package api

import (
	"net/http"
	"strings"
)

const swaggerUIHTML = `<!doctype html>
<html lang="ko">
<head>
  <meta charset="utf-8">
  <title>Robot Center API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    body { margin: 0; background: #f7f7f7; }
    .swagger-ui .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: "/swagger/openapi.json",
      dom_id: "#swagger-ui",
      deepLinking: true,
      persistAuthorization: true
    });
  </script>
</body>
</html>`

func (s *Server) handleSwaggerUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swaggerUIHTML))
}

func (s *Server) handleOpenAPIJSON(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.openAPISpec())
}

func (s *Server) openAPISpec() map[string]any {
	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "Robot Center API",
			"version":     "0.1.0",
			"description": "관제 서버 API 문서입니다. API는 actor 책임 기준으로 `/api/v1/robot/*`, `/api/v1/recorder/*`, `/api/v1/operator/*`, `/api/v1/system/*` namespace로 분리됩니다.",
		},
		"servers": []map[string]any{
			{
				"url":         openAPIServerURL(s.config.PublicURL),
				"description": "관제 API 서버",
			},
		},
		"tags": []map[string]any{
			{
				"name":        "시스템 API",
				"description": "전체 시스템 상태, 운영성 작업, API 문서 endpoint입니다.",
			},
			{
				"name":        "로봇 API",
				"description": "로봇이 운용 중 호출하는 Bearer token 기반 API입니다. 서버는 token으로 로봇 신원을 판별합니다.",
			},
			{
				"name":        "Recorder API",
				"description": "recorder-worker가 녹화 대상 조회, chunk 상태 보고, 센서 payload 저장에 사용하는 내부 API입니다.",
			},
			{
				"name":        "Operator API",
				"description": "관제 브라우저가 로봇, 임무, 실시간 상태, 센서, 녹화 조회와 WebRTC subscribe에 사용하는 API입니다.",
			},
		},
		"paths": map[string]any{
			"/healthz":                                 openAPIHealthPath(),
			"/api/v1/system/status":                    openAPISystemStatusPath(),
			"/api/v1/system/object-storage/clear":      openAPIObjectStorageClearPath(),
			"/api/v1/system/sensors/clear":             openAPISensorDataClearPath(),
			"/api/v1/operator/rtc-config":              openAPIRTCConfigPath(),
			"/api/v1/operator/sensor-descriptors":      openAPISensorDescriptorsPath(),
			"/api/v1/operator/sensor-samples":          openAPISensorSamplesPath(),
			"/api/v1/operator/sensor-latest":           openAPISensorLatestPath(),
			"/api/v1/operator/recordings":              openAPIRecordingsPath(),
			"/api/v1/recorder/recording-targets":       openAPIRecordingTargetsPath(),
			"/api/v1/recorder/tick":                    openAPIRecorderTickPath(),
			"/api/v1/recorder/finalization-jobs/claim": openAPIRecorderFinalizationJobsClaimPath(),
			"/api/v1/recorder/finalization-jobs/{jobID}/completed": openAPIRecorderFinalizationJobStatusPath(
				"markRecorderFinalizationJobCompleted",
				"녹화 finalization job 완료 처리",
			),
			"/api/v1/recorder/finalization-jobs/{jobID}/partial": openAPIRecorderFinalizationJobStatusPath(
				"markRecorderFinalizationJobPartial",
				"녹화 finalization job 부분 완료 처리",
			),
			"/api/v1/recorder/finalization-jobs/{jobID}/failed": openAPIRecorderFinalizationJobStatusPath(
				"markRecorderFinalizationJobFailed",
				"녹화 finalization job 실패 처리",
			),
			"/api/v1/recorder/chunks/{chunkID}/uploaded":                  openAPIRecorderChunkUploadedPath(),
			"/api/v1/recorder/chunks/{chunkID}/files/{fileType}/uploaded": openAPIRecorderFileUploadedPath(),
			"/api/v1/recorder/sensor-samples":                             openAPIRecorderSensorSamplesPath(),
			"/api/v1/robot/heartbeat":                                     openAPIRobotHeartbeatPath(),
			"/api/v1/robot/mission":                                       openAPIRobotMissionPath(),
			"/api/v1/robot/sfu/ws":                                        openAPIRobotSFUWebSocketPath(),
			"/api/v1/operator/robots":                                     openAPIRobotsPath(),
			"/api/v1/operator/robots/{robotCode}":                         openAPIRobotItemPath(),
			"/api/v1/operator/robots/{robotCode}/connection-info":         openAPIRobotConnectionInfoPath(),
			"/api/v1/operator/robots/{robotCode}/connection-token":        openAPIRobotConnectionTokenPath(),
			"/api/v1/operator/missions":                                   openAPIMissionsPath(),
			"/api/v1/operator/missions/{missionCode}/live-status":         openAPIMissionLiveStatusPath(),
			"/api/v1/operator/missions/{missionCode}/start": openAPIMissionStatePath(
				"startMission",
				"임무 시작",
				"ready 상태의 임무를 active 상태로 전환합니다. 로봇은 `/api/v1/robot/mission`에서 자기 token에 연결된 active 임무만 조회합니다.",
			),
			"/api/v1/operator/missions/{missionCode}/end": openAPIMissionStatePath(
				"endMission",
				"임무 종료",
				"active 임무를 ended 상태로 전환하고 해당 임무의 SFU room을 닫습니다.",
			),
			"/api/v1/operator/sfu/ws": openAPISFUWebSocketPath(
				"Operator API",
				"openOperatorSFUWebSocket",
				"관제 UI SFU WebSocket 연결",
				"관제 UI operator peer가 mission room에 subscribe하기 위한 WebSocket endpoint입니다.",
			),
			"/api/v1/recorder/sfu/ws": openAPISFUWebSocketPath(
				"Recorder API",
				"openRecorderSFUWebSocket",
				"Recorder SFU WebSocket 연결",
				"recorder-worker peer가 mission room에 subscribe하기 위한 WebSocket endpoint입니다.",
			),
		},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"robotBearerAuth": map[string]any{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "robotToken",
					"description":  "로봇 연결 정보에 포함된 `robotToken` 값입니다. 서버는 이 token으로 로봇 신원을 판별하므로 robotCode, robotId, roomId, sessionId를 별도 파라미터로 받지 않습니다.",
				},
			},
			"schemas": openAPISchemas(),
		},
	}
}

func openAPIServerURL(publicURL string) string {
	value := strings.TrimRight(strings.TrimSpace(publicURL), "/")
	if value == "" {
		return "/"
	}
	return value
}
