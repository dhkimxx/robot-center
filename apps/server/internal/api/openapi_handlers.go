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
      url: "/api/docs/openapi.json",
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
			"description": "관제 서버 API 문서입니다. 로봇 런타임 API는 `/api/v1/robot/*` 하위 경로로 분리되며, robotToken으로 인증된 자기 로봇 범위의 정보만 반환합니다.",
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
				"description": "관제 서버 상태, WebRTC 설정, 운영성 작업 API입니다.",
			},
			{
				"name":        "로봇 런타임 API",
				"description": "로봇이 운용 중 호출하는 Bearer token 기반 API입니다. 서버는 token으로 로봇 신원을 판별합니다.",
			},
			{
				"name":        "로봇 관리 API",
				"description": "관제에서 로봇을 등록하고 연결 정보를 관리하는 API입니다.",
			},
			{
				"name":        "임무 관리 API",
				"description": "관제에서 임무를 생성하고 상태를 전환하는 API입니다.",
			},
			{
				"name":        "센서 API",
				"description": "센서 descriptor, sample, latest 조회와 저장 API입니다.",
			},
			{
				"name":        "녹화 API",
				"description": "녹화 chunk 조회와 recorder-worker callback API입니다.",
			},
			{
				"name":        "SFU WebSocket",
				"description": "관제 UI와 recorder-worker의 WebRTC signaling WebSocket endpoint입니다.",
			},
		},
		"paths": map[string]any{
			"/healthz":                              openAPIHealthPath(),
			"/api/system/status":                    openAPISystemStatusPath(),
			"/api/system/object-storage/clear":      openAPIObjectStorageClearPath(),
			"/api/rtc-config":                       openAPIRTCConfigPath(),
			"/api/recording-targets":                openAPIRecordingTargetsPath(),
			"/api/sensor-descriptors":               openAPISensorDescriptorsPath(),
			"/api/sensor-samples":                   openAPISensorSamplesPath(),
			"/api/sensor-latest":                    openAPISensorLatestPath(),
			"/api/recordings":                       openAPIRecordingsPath(),
			"/api/recorder/tick":                    openAPIRecorderTickPath(),
			"/api/recorder/finalization-jobs/claim": openAPIRecorderFinalizationJobsClaimPath(),
			"/api/recorder/finalization-jobs/{jobID}/completed": openAPIRecorderFinalizationJobStatusPath(
				"markRecorderFinalizationJobCompleted",
				"녹화 finalization job 완료 처리",
			),
			"/api/recorder/finalization-jobs/{jobID}/partial": openAPIRecorderFinalizationJobStatusPath(
				"markRecorderFinalizationJobPartial",
				"녹화 finalization job 부분 완료 처리",
			),
			"/api/recorder/finalization-jobs/{jobID}/failed": openAPIRecorderFinalizationJobStatusPath(
				"markRecorderFinalizationJobFailed",
				"녹화 finalization job 실패 처리",
			),
			"/api/recorder/chunks/{chunkID}/uploaded":                  openAPIRecorderChunkUploadedPath(),
			"/api/recorder/chunks/{chunkID}/files/{fileType}/uploaded": openAPIRecorderFileUploadedPath(),
			"/api/v1/robot/heartbeat":                                  openAPIRobotHeartbeatPath(),
			"/api/v1/robot/mission":                                    openAPIRobotMissionPath(),
			"/api/v1/robot/sfu/ws":                                     openAPIRobotSFUWebSocketPath(),
			"/api/robots":                                              openAPIRobotsPath(),
			"/api/robots/{robotCode}":                                  openAPIRobotItemPath(),
			"/api/robots/{robotCode}/connection-info":                  openAPIRobotConnectionInfoPath(),
			"/api/robots/{robotCode}/connection-token":                 openAPIRobotConnectionTokenPath(),
			"/api/missions":                                            openAPIMissionsPath(),
			"/api/missions/{missionCode}/live-status":                  openAPIMissionLiveStatusPath(),
			"/api/missions/{missionCode}/start": openAPIMissionStatePath(
				"startMission",
				"임무 시작",
				"ready 상태의 임무를 active 상태로 전환합니다. 로봇은 `/api/v1/robot/mission`에서 자기 token에 연결된 active 임무만 조회합니다.",
			),
			"/api/missions/{missionCode}/end": openAPIMissionStatePath(
				"endMission",
				"임무 종료",
				"active 임무를 ended 상태로 전환하고 해당 임무의 SFU room을 닫습니다.",
			),
			"/sfu/operator/ws": openAPISFUWebSocketPath(
				"openOperatorSFUWebSocket",
				"관제 UI SFU WebSocket 연결",
				"관제 UI operator peer가 mission room에 subscribe하기 위한 WebSocket endpoint입니다.",
			),
			"/sfu/recorder/ws": openAPISFUWebSocketPath(
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

func openAPIHealthPath() map[string]any {
	return map[string]any{
		"get": openAPIOperation("시스템 API", "getHealth", "서버 health 확인", "app-server health 상태를 반환합니다.", "HealthResponse", ""),
	}
}

func openAPISystemStatusPath() map[string]any {
	return map[string]any{
		"get": openAPIOperation("시스템 API", "getSystemStatus", "시스템 상태 조회", "app-server, recorder-worker, storage, SFU room 상태 요약을 반환합니다.", "SystemStatusResponse", ""),
	}
}

func openAPIObjectStorageClearPath() map[string]any {
	return map[string]any{
		"post": openAPIOperation("시스템 API", "clearObjectStorage", "Object Storage 초기화", "확인 문자열을 받은 뒤 object storage 데이터를 정리합니다.", "ObjectStorageResponse", "ClearObjectStorageRequest"),
	}
}

func openAPIRTCConfigPath() map[string]any {
	return map[string]any{
		"get": openAPIOperation("시스템 API", "getRTCConfig", "관제 WebRTC 설정 조회", "관제 UI operator peer가 사용할 signaling URL과 ICE 서버 설정을 반환합니다.", "RTCConfigResponse", ""),
	}
}

func openAPIRecordingTargetsPath() map[string]any {
	return map[string]any{
		"get": openAPIOperation("녹화 API", "listRecordingTargets", "녹화 대상 임무 조회", "recorder-worker가 구독해야 하는 active mission 목록을 반환합니다.", "RecordingTargetsResponse", ""),
	}
}

func openAPISensorDescriptorsPath() map[string]any {
	return map[string]any{
		"get":  openAPIOperationWithParameters("센서 API", "listSensorDescriptors", "센서 descriptor 조회", "missionId, robotCode 조건으로 센서 descriptor 목록을 조회합니다.", "SensorDescriptorsResponse", "", openAPISensorQueryParameters(false)),
		"post": openAPIOperation("센서 API", "createSensorDescriptors", "센서 descriptor/sample 저장", "DataChannel sensor envelope를 저장합니다. /api/sensor-samples POST와 같은 처리 경로를 사용합니다.", "SensorSamplesResponse", "SensorEnvelopeRequest"),
	}
}

func openAPISensorSamplesPath() map[string]any {
	return map[string]any{
		"get":  openAPIOperationWithParameters("센서 API", "listSensorSamples", "센서 sample 조회", "missionId, robotCode, sensorId 조건과 limit으로 센서 sample 목록을 조회합니다.", "SensorSamplesResponse", "", append(openAPISensorQueryParameters(false), openAPIQueryParameter("sensorId", "조회할 sensorId", false), openAPIIntegerQueryParameter("limit", "조회 개수 제한", false))),
		"post": openAPIOperation("센서 API", "createSensorSamples", "센서 sample 저장", "DataChannel sensor envelope를 저장합니다.", "SensorSamplesResponse", "SensorEnvelopeRequest"),
	}
}

func openAPISensorLatestPath() map[string]any {
	return map[string]any{
		"get": openAPIOperationWithParameters("센서 API", "listSensorLatest", "센서 latest 조회", "missionId, robotCode 조건으로 센서별 최신 sample을 조회합니다.", "SensorLatestEnvelope", "", openAPISensorQueryParameters(false)),
	}
}

func openAPIRecordingsPath() map[string]any {
	return map[string]any{
		"get": openAPIOperation("녹화 API", "listRecordings", "녹화 chunk 조회", "저장된 recording chunk와 파일 상태 목록을 반환합니다.", "RecordingsResponse", ""),
	}
}

func openAPIRecorderTickPath() map[string]any {
	return map[string]any{
		"post": openAPIOperation("녹화 API", "applyRecorderTick", "녹화 tick 반영", "recorder-worker가 mission/robot 기준 chunk 생성을 요청합니다.", "RecordingTickResponse", "RecorderTickRequest"),
	}
}

func openAPIRecorderFinalizationJobsClaimPath() map[string]any {
	return map[string]any{
		"post": openAPIOperation("녹화 API", "claimRecorderFinalizationJobs", "녹화 finalization job claim", "recorder-worker가 처리할 finalization job을 claim합니다.", "RecorderFinalizationJobsResponse", "RecorderFinalizationClaimRequest"),
	}
}

func openAPIRecorderFinalizationJobStatusPath(operationID string, summary string) map[string]any {
	return map[string]any{
		"post": openAPIOperationWithParameters("녹화 API", operationID, summary, "recorder-worker가 finalization job 처리 결과를 보고합니다.", "OKResponse", "RecorderFinalizationStatusRequest", []map[string]any{openAPIPathParameter("jobID", "finalization job ID")}),
	}
}

func openAPIRecorderChunkUploadedPath() map[string]any {
	return map[string]any{
		"post": openAPIOperationWithParameters("녹화 API", "markRecorderChunkUploaded", "녹화 chunk 업로드 완료", "recorder-worker가 chunk manifest 업로드 완료를 보고합니다.", "RecordingChunkEnvelope", "RecorderUploadRequest", []map[string]any{openAPIPathParameter("chunkID", "recording chunk ID")}),
	}
}

func openAPIRecorderFileUploadedPath() map[string]any {
	return map[string]any{
		"post": openAPIOperationWithParameters("녹화 API", "markRecorderFileUploaded", "녹화 파일 업로드 완료", "recorder-worker가 chunk의 개별 파일 업로드 완료를 보고합니다.", "RecordingChunkEnvelope", "RecorderUploadRequest", []map[string]any{openAPIPathParameter("chunkID", "recording chunk ID"), openAPIPathParameter("fileType", "업로드된 파일 타입")}),
	}
}

func openAPIRobotHeartbeatPath() map[string]any {
	return map[string]any{
		"post": map[string]any{
			"tags":        []string{"로봇 런타임 API"},
			"operationId": "sendRobotHeartbeat",
			"summary":     "로봇 heartbeat 전송",
			"description": "로봇이 자신의 현재 상태를 주기적으로 관제 서버에 보고합니다. Authorization Bearer token으로 로봇을 식별하므로 요청 body에 robotCode를 넣지 않습니다.",
			"security":    []map[string]any{{"robotBearerAuth": []string{}}},
			"requestBody": map[string]any{
				"required": true,
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{"$ref": "#/components/schemas/RobotHeartbeatRequest"},
					},
				},
			},
			"responses": map[string]any{
				"200": openAPIJSONResponse("heartbeat가 반영됐습니다.", "#/components/schemas/RobotHeartbeatResponse"),
				"400": openAPIErrorResponse("요청 body가 잘못됐습니다."),
				"401": openAPIErrorResponse("robotToken이 없거나 유효하지 않습니다."),
			},
		},
	}
}

func openAPIRobotMissionPath() map[string]any {
	return map[string]any{
		"get": map[string]any{
			"tags":        []string{"로봇 런타임 API"},
			"operationId": "getRobotMission",
			"summary":     "내 active 임무 조회",
			"description": "robotToken에 연결된 로봇의 active 임무만 조회합니다. 다른 로봇이나 다른 임무 정보는 반환하지 않으며 robotCode 쿼리 파라미터는 허용하지 않습니다.",
			"security":    []map[string]any{{"robotBearerAuth": []string{}}},
			"responses": map[string]any{
				"200": openAPIJSONResponse("active 임무가 있으면 송출 설정을 반환하고, 없으면 missionStatus가 none입니다.", "#/components/schemas/RobotMissionResponse"),
				"400": openAPIErrorResponse("허용되지 않는 쿼리 파라미터가 포함됐습니다."),
				"401": openAPIErrorResponse("robotToken이 없거나 유효하지 않습니다."),
			},
		},
	}
}

func openAPIRobotSFUWebSocketPath() map[string]any {
	return map[string]any{
		"get": map[string]any{
			"tags":        []string{"로봇 런타임 API"},
			"operationId": "openRobotSFUWebSocket",
			"summary":     "로봇 SFU WebSocket 연결",
			"description": "WebRTC signaling용 WebSocket upgrade 엔드포인트입니다. `/api/v1/robot/mission` 응답의 `sfu.signalingUrl` 값을 그대로 사용합니다. robotCode, sessionId, roomId는 별도로 전달하지 않습니다.",
			"security":    []map[string]any{{"robotBearerAuth": []string{}}},
			"parameters": []map[string]any{
				{
					"name":        "room",
					"in":          "query",
					"required":    true,
					"description": "`/api/v1/robot/mission` 응답의 `missionCode`입니다. 이 값은 로봇이 임의로 선택하지 않고 mission 조회 응답에서 받은 값을 사용합니다.",
					"schema":      map[string]any{"type": "string", "example": "mission-002"},
				},
			},
			"responses": map[string]any{
				"101": map[string]any{"description": "WebSocket upgrade가 성공했습니다."},
				"400": openAPIErrorResponse("room 쿼리 파라미터가 없거나 잘못됐습니다."),
				"401": openAPIErrorResponse("robotToken이 없거나 유효하지 않습니다."),
				"403": openAPIErrorResponse("robotToken의 로봇이 해당 room에 배정되어 있지 않습니다."),
			},
		},
	}
}

func openAPIRobotsPath() map[string]any {
	return map[string]any{
		"get": openAPIOperation("로봇 관리 API", "listRobots", "로봇 목록 조회", "관제 서버에 등록된 로봇 목록을 반환합니다.", "RobotsResponse", ""),
		"post": map[string]any{
			"tags":        []string{"로봇 관리 API"},
			"operationId": "createRobot",
			"summary":     "로봇 등록",
			"description": "관제 서버에 로봇을 등록하고 로봇 런타임이 사용할 연결 정보를 발급합니다. 이 API는 관리 API이며 로봇 런타임이 직접 호출하지 않습니다.",
			"requestBody": map[string]any{
				"required": true,
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{"$ref": "#/components/schemas/CreateRobotRequest"},
					},
				},
			},
			"responses": map[string]any{
				"201": openAPIJSONResponse("로봇과 연결 정보가 생성됐습니다.", "#/components/schemas/CreateRobotResponse"),
				"400": openAPIErrorResponse("필수 값이 없거나 요청 body가 잘못됐습니다."),
			},
		},
	}
}

func openAPIRobotItemPath() map[string]any {
	parameters := []map[string]any{openAPIPathParameter("robotCode", "로봇 코드")}
	return map[string]any{
		"patch":  openAPIOperationWithParameters("로봇 관리 API", "updateRobot", "로봇 수정", "로봇 표시 이름과 모델명을 수정합니다.", "RobotEnvelope", "UpdateRobotRequest", parameters),
		"delete": openAPIOperationWithParameters("로봇 관리 API", "archiveRobot", "로봇 보관 처리", "로봇을 active 목록에서 제외합니다.", "RobotEnvelope", "", parameters),
	}
}

func openAPIRobotConnectionInfoPath() map[string]any {
	return map[string]any{
		"get": openAPIOperationWithParameters("로봇 관리 API", "getRobotConnectionInfo", "로봇 연결 정보 조회", "로봇 런타임 접속에 필요한 serverUrl, robotCode, robotToken을 조회합니다.", "RobotConnectionInfoEnvelope", "", []map[string]any{openAPIPathParameter("robotCode", "로봇 코드")}),
	}
}

func openAPIRobotConnectionTokenPath() map[string]any {
	return map[string]any{
		"post": openAPIOperationWithParameters("로봇 관리 API", "rotateRobotConnectionToken", "로봇 token 재발급", "로봇 런타임 API용 robotToken을 재발급합니다.", "RobotConnectionInfoEnvelope", "", []map[string]any{openAPIPathParameter("robotCode", "로봇 코드")}),
	}
}

func openAPIMissionsPath() map[string]any {
	return map[string]any{
		"get": openAPIOperation("임무 관리 API", "listMissions", "임무 목록 조회", "관제 서버에 등록된 임무 목록을 반환합니다.", "MissionsResponse", ""),
		"post": map[string]any{
			"tags":        []string{"임무 관리 API"},
			"operationId": "createMission",
			"summary":     "임무 생성",
			"description": "임무를 생성하고 하나 이상의 로봇을 배정합니다. 여러 로봇을 배정할 때는 robotCodes 배열을 사용합니다.",
			"requestBody": map[string]any{
				"required": true,
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{"$ref": "#/components/schemas/CreateMissionRequest"},
					},
				},
			},
			"responses": map[string]any{
				"201": openAPIJSONResponse("임무가 생성됐습니다.", "#/components/schemas/MissionResponseEnvelope"),
				"400": openAPIErrorResponse("필수 값이 없거나 missionType 값이 허용 범위가 아닙니다."),
				"409": openAPIErrorResponse("배정하려는 로봇이 이미 active 임무에 포함되어 있습니다."),
			},
		},
	}
}

func openAPIMissionLiveStatusPath() map[string]any {
	return map[string]any{
		"get": openAPIOperationWithParameters("임무 관리 API", "getMissionLiveStatus", "임무 live status 조회", "임무에 배정된 로봇의 연결, 스트림, 녹화 상태를 반환합니다.", "MissionLiveStatusResponse", "", []map[string]any{openAPIPathParameter("missionCode", "임무 코드")}),
	}
}

func openAPIMissionStatePath(operationID string, summary string, description string) map[string]any {
	return map[string]any{
		"post": map[string]any{
			"tags":        []string{"임무 관리 API"},
			"operationId": operationID,
			"summary":     summary,
			"description": description,
			"parameters": []map[string]any{
				{
					"name":        "missionCode",
					"in":          "path",
					"required":    true,
					"description": "임무 생성 응답의 `mission.missionCode` 값입니다.",
					"schema":      map[string]any{"type": "string", "example": "mission-002"},
				},
			},
			"responses": map[string]any{
				"200": openAPIJSONResponse("임무 상태가 변경됐습니다.", "#/components/schemas/MissionResponseEnvelope"),
				"404": openAPIErrorResponse("missionCode에 해당하는 임무를 찾을 수 없습니다."),
				"409": openAPIErrorResponse("현재 상태에서 요청한 상태 전환을 할 수 없습니다."),
			},
		},
	}
}

func openAPISFUWebSocketPath(operationID string, summary string, description string) map[string]any {
	return map[string]any{
		"get": map[string]any{
			"tags":        []string{"SFU WebSocket"},
			"operationId": operationID,
			"summary":     summary,
			"description": description,
			"parameters": []map[string]any{
				openAPIQueryParameter("room", "접속할 missionCode room", true),
			},
			"responses": map[string]any{
				"101": map[string]any{"description": "WebSocket upgrade가 성공했습니다."},
				"400": openAPIErrorResponse("room 쿼리 파라미터가 없거나 허용되지 않는 query가 포함됐습니다."),
			},
		},
	}
}

func openAPIOperation(tag string, operationID string, summary string, description string, responseSchema string, requestSchema string) map[string]any {
	return openAPIOperationWithParameters(tag, operationID, summary, description, responseSchema, requestSchema, nil)
}

func openAPIOperationWithParameters(tag string, operationID string, summary string, description string, responseSchema string, requestSchema string, parameters []map[string]any) map[string]any {
	operation := map[string]any{
		"tags":        []string{tag},
		"operationId": operationID,
		"summary":     summary,
		"description": description,
		"responses": map[string]any{
			"200": openAPIJSONResponse("요청이 성공했습니다.", "#/components/schemas/"+responseSchema),
			"400": openAPIErrorResponse("요청 값이 잘못됐습니다."),
			"404": openAPIErrorResponse("대상을 찾을 수 없습니다."),
			"409": openAPIErrorResponse("현재 상태에서 요청을 처리할 수 없습니다."),
			"500": openAPIErrorResponse("서버 내부 오류입니다."),
		},
	}
	if len(parameters) > 0 {
		operation["parameters"] = parameters
	}
	if requestSchema != "" {
		operation["requestBody"] = map[string]any{
			"required": true,
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{"$ref": "#/components/schemas/" + requestSchema},
				},
			},
		}
	}
	return operation
}

func openAPIJSONResponse(description string, schemaReference string) map[string]any {
	return map[string]any{
		"description": description,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{"$ref": schemaReference},
			},
		},
	}
}

func openAPIErrorResponse(description string) map[string]any {
	return map[string]any{
		"description": description,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"},
			},
		},
	}
}

func openAPISchemas() map[string]any {
	return map[string]any{
		"HealthResponse":        openAPIHealthResponseSchema(),
		"SystemStatusResponse":  openAPIGenericObjectSchema("시스템 상태 응답입니다."),
		"ObjectStorageResponse": openAPIGenericObjectSchema("Object Storage 작업 결과입니다."),
		"ClearObjectStorageRequest": openAPIObjectSchema("Object Storage 초기화 요청입니다.", map[string]any{
			"confirmation": openAPIStringProperty("초기화 확인 문자열입니다.", "CLEAR_OBJECT_STORAGE"),
		}),
		"RTCConfigResponse":         openAPIRTCConfigResponseSchema(),
		"RecordingTargetsResponse":  openAPIArrayEnvelopeSchema("targets", "#/components/schemas/Mission", "녹화 대상 임무 목록입니다."),
		"SensorDescriptorsResponse": openAPIArrayEnvelopeSchema("sensorDescriptors", "#/components/schemas/SensorDescriptor", "센서 descriptor 목록입니다."),
		"SensorSamplesResponse":     openAPIArrayEnvelopeSchema("sensorSamples", "#/components/schemas/SensorSample", "센서 sample 목록입니다."),
		"SensorLatestEnvelope":      openAPISensorLatestEnvelopeSchema(),
		"SensorEnvelopeRequest":     openAPISensorEnvelopeRequestSchema(),
		"SensorDescriptor":          openAPISensorDescriptorSchema(),
		"SensorSample":              openAPISensorSampleSchema(),
		"SensorLatest":              openAPISensorLatestSchema(),
		"RecordingsResponse":        openAPIArrayEnvelopeSchema("recordings", "#/components/schemas/RecordingChunk", "녹화 chunk 목록입니다."),
		"RecordingChunkEnvelope":    openAPIObjectSchema("녹화 chunk 응답입니다.", map[string]any{"chunk": map[string]any{"$ref": "#/components/schemas/RecordingChunk"}}),
		"RecordingChunk":            openAPIRecordingChunkSchema(),
		"RecordingFile":             openAPIRecordingFileSchema(),
		"RecordingTickRequest":      openAPIRecorderTickRequestSchema(),
		"RecordingTickResponse":     openAPIObjectSchema("녹화 tick 처리 결과입니다.", map[string]any{"chunk": map[string]any{"$ref": "#/components/schemas/RecordingChunk"}, "manifest": openAPIGenericObjectProperty("녹화 manifest")}),
		"RecorderUploadRequest":     openAPIRecorderUploadRequestSchema(),
		"RecorderFinalizationClaimRequest": openAPIObjectSchema("녹화 finalization job claim 요청입니다.", map[string]any{
			"workerId":            openAPIStringProperty("recorder-worker ID입니다.", "recorder-1"),
			"limit":               map[string]any{"type": "integer", "example": 5},
			"lockDurationSeconds": map[string]any{"type": "integer", "example": 60},
		}),
		"RecorderFinalizationStatusRequest": openAPIObjectSchema("녹화 finalization job 상태 보고 요청입니다.", map[string]any{
			"workerId": openAPIStringProperty("recorder-worker ID입니다.", "recorder-1"),
			"attempt":  map[string]any{"type": "integer", "example": 1},
			"reason":   openAPIStringProperty("실패 또는 부분 완료 사유입니다.", "missing thermal file"),
		}),
		"RecorderFinalizationJobsResponse": openAPIObjectSchema("claim된 finalization job 목록입니다.", map[string]any{"jobs": map[string]any{"type": "array", "items": openAPIGenericObjectProperty("finalization job")}}),
		"OKResponse":                       openAPIObjectSchema("처리 결과입니다.", map[string]any{"ok": map[string]any{"type": "boolean", "example": true}}),
		"RobotsResponse":                   openAPIArrayEnvelopeSchema("robots", "#/components/schemas/Robot", "로봇 목록입니다."),
		"RobotEnvelope":                    openAPIObjectSchema("로봇 응답입니다.", map[string]any{"robot": map[string]any{"$ref": "#/components/schemas/Robot"}}),
		"UpdateRobotRequest":               openAPICreateRobotRequestSchema(),
		"RobotConnectionInfoEnvelope":      openAPIObjectSchema("로봇 연결 정보 응답입니다.", map[string]any{"connectionInfo": map[string]any{"$ref": "#/components/schemas/RobotConnectionInfo"}}),
		"MissionsResponse":                 openAPIArrayEnvelopeSchema("missions", "#/components/schemas/Mission", "임무 목록입니다."),
		"MissionLiveStatusResponse":        openAPIMissionLiveStatusResponseSchema(),
		"RobotHeartbeatRequest":            openAPIRobotHeartbeatRequestSchema(),
		"RobotHeartbeatResponse":           openAPIRobotHeartbeatResponseSchema(),
		"RobotMissionResponse":             openAPIRobotMissionResponseSchema(),
		"RobotSFUConfig":                   openAPIRobotSFUConfigSchema(),
		"TurnServer":                       openAPITurnServerSchema(),
		"CreateRobotRequest":               openAPICreateRobotRequestSchema(),
		"CreateRobotResponse":              openAPICreateRobotResponseSchema(),
		"Robot":                            openAPIRobotSchema(),
		"RobotConnectionInfo":              openAPIRobotConnectionInfoSchema(),
		"CreateMissionRequest":             openAPICreateMissionRequestSchema(),
		"MissionResponseEnvelope":          openAPIMissionResponseEnvelopeSchema(),
		"Mission":                          openAPIMissionSchema(),
		"ErrorResponse":                    openAPIErrorResponseSchema(),
	}
}

func openAPIHealthResponseSchema() map[string]any {
	return openAPIObjectSchema("health 응답입니다.", map[string]any{
		"status":    openAPIStringProperty("상태입니다.", "ok"),
		"service":   openAPIStringProperty("서비스 이름입니다.", "app-server"),
		"startedAt": openAPIDateTimeProperty("서비스 시작 시각입니다."),
	})
}

func openAPIRTCConfigResponseSchema() map[string]any {
	return openAPIObjectSchema("관제 WebRTC 설정 응답입니다.", map[string]any{
		"mode":                 openAPIStringProperty("WebRTC 모드입니다.", "sfu"),
		"signalingUrl":         openAPIStringProperty("operator signaling URL입니다.", "ws://center.local/sfu/operator/ws"),
		"operatorSignalingUrl": openAPIStringProperty("operator signaling URL입니다.", "ws://center.local/sfu/operator/ws"),
		"iceTransportPolicy":   openAPIStringProperty("ICE transport policy입니다.", "relay"),
		"iceServers": map[string]any{
			"type":  "array",
			"items": map[string]any{"$ref": "#/components/schemas/TurnServer"},
		},
	})
}

func openAPISensorEnvelopeRequestSchema() map[string]any {
	return openAPIObjectSchema("DataChannel sensor envelope 요청입니다.", map[string]any{
		"messageId":   openAPIStringProperty("메시지 추적 ID입니다.", "telemetry-001"),
		"messageType": openAPIStringProperty("메시지 타입입니다.", "telemetry"),
		"robotCode":   openAPIStringProperty("로봇 코드입니다.", "robot-001"),
		"missionId":   openAPIStringProperty("임무 ID 또는 코드입니다.", "mission-001"),
		"channelRole": openAPIStringProperty("DataChannel role입니다.", "channel.telemetry"),
		"descriptors": map[string]any{
			"type":  "array",
			"items": map[string]any{"$ref": "#/components/schemas/SensorDescriptor"},
		},
		"samples": map[string]any{
			"type":  "array",
			"items": map[string]any{"$ref": "#/components/schemas/SensorSample"},
		},
	})
}

func openAPISensorDescriptorSchema() map[string]any {
	return openAPIObjectSchema("센서 descriptor입니다.", map[string]any{
		"id":          openAPIStringProperty("descriptor ID입니다.", "uuid"),
		"missionId":   openAPIStringProperty("mission ID입니다.", "mission-001"),
		"robotCode":   openAPIStringProperty("robot code입니다.", "robot-001"),
		"sensorId":    openAPIStringProperty("descriptor/sample 매칭용 sensor ID입니다.", "telemetry.gas.channel_1"),
		"channelRole": openAPIStringProperty("DataChannel role입니다.", "channel.telemetry"),
		"label":       openAPIStringProperty("표시 및 해석 보조 label입니다.", "CO"),
		"sensorType":  openAPIStringProperty("센서 타입입니다.", "gas"),
		"unit":        openAPIStringProperty("표시 단위입니다.", "ppm"),
		"enabled":     map[string]any{"type": "boolean", "example": true},
		"firstSeenAt": openAPIDateTimeProperty("최초 관측 시각입니다."),
		"lastSeenAt":  openAPIDateTimeProperty("마지막 관측 시각입니다."),
	})
}

func openAPISensorSampleSchema() map[string]any {
	return openAPIObjectSchema("센서 sample입니다.", map[string]any{
		"id":           openAPIStringProperty("sample ID입니다.", "uuid"),
		"descriptorId": openAPIStringProperty("descriptor ID입니다.", "uuid"),
		"missionId":    openAPIStringProperty("mission ID입니다.", "mission-001"),
		"robotCode":    openAPIStringProperty("robot code입니다.", "robot-001"),
		"sensorId":     openAPIStringProperty("descriptor와 매칭되는 sensor ID입니다.", "telemetry.gas.channel_1"),
		"channelRole":  openAPIStringProperty("DataChannel role입니다.", "channel.telemetry"),
		"messageId":    openAPIStringProperty("메시지 추적 ID입니다.", "telemetry-001"),
		"timestamp":    openAPINullableDateTimeProperty("sample 측정 시각입니다."),
		"receivedAt":   openAPIDateTimeProperty("서버 수신 시각입니다."),
		"values":       openAPIGenericObjectProperty("sensorType별 측정값 object입니다."),
		"objectKey":    openAPIStringProperty("object storage key입니다.", "missions/mission-001/sensor.jsonl"),
	})
}

func openAPISensorLatestSchema() map[string]any {
	schema := openAPISensorDescriptorSchema()
	schema["description"] = "센서 descriptor와 최신 sample입니다."
	schema["properties"].(map[string]any)["latestSample"] = map[string]any{"$ref": "#/components/schemas/SensorSample"}
	return schema
}

func openAPISensorLatestEnvelopeSchema() map[string]any {
	return openAPIObjectSchema("센서 latest 응답입니다.", map[string]any{
		"missionId": openAPIStringProperty("mission ID 또는 코드입니다.", "mission-001"),
		"robotCode": openAPIStringProperty("robot code입니다.", "robot-001"),
		"sensors": map[string]any{
			"type":  "array",
			"items": map[string]any{"$ref": "#/components/schemas/SensorLatest"},
		},
	})
}

func openAPIMissionLiveStatusResponseSchema() map[string]any {
	return openAPIObjectSchema("임무 live status 응답입니다.", map[string]any{
		"missionCode":   openAPIStringProperty("mission code입니다.", "mission-001"),
		"missionStatus": openAPIStringProperty("mission 상태입니다.", "active"),
		"observedAt":    openAPIDateTimeProperty("관측 시각입니다."),
		"robots": map[string]any{
			"type": "array",
			"items": openAPIObjectSchema("로봇 live status입니다.", map[string]any{
				"robotCode":   openAPIStringProperty("robot code입니다.", "robot-001"),
				"displayName": openAPIStringProperty("로봇 표시 이름입니다.", "Jetson 01"),
				"connection":  openAPIGenericObjectProperty("연결 상태입니다."),
				"stream":      openAPIGenericObjectProperty("스트림 상태입니다."),
				"recording":   openAPIGenericObjectProperty("녹화 상태입니다."),
			}),
		},
	})
}

func openAPIRecordingChunkSchema() map[string]any {
	return openAPIObjectSchema("녹화 chunk입니다.", map[string]any{
		"id":                 openAPIStringProperty("chunk ID입니다.", "uuid"),
		"recordingSessionId": openAPIStringProperty("recording session ID입니다.", "uuid"),
		"missionId":          openAPIStringProperty("mission ID입니다.", "uuid"),
		"missionCode":        openAPIStringProperty("mission code입니다.", "mission-001"),
		"robotCode":          openAPIStringProperty("robot code입니다.", "robot-001"),
		"chunkIndex":         map[string]any{"type": "integer", "example": 1},
		"status":             openAPIStringProperty("chunk 상태입니다.", "uploaded"),
		"startedAt":          openAPIDateTimeProperty("chunk 시작 시각입니다."),
		"endedAt":            openAPIDateTimeProperty("chunk 종료 시각입니다."),
		"durationSeconds":    map[string]any{"type": "integer", "example": 60},
		"manifestObjectKey":  openAPIStringProperty("manifest object key입니다.", "missions/mission-001/manifest.json"),
		"mediaObjectKeys":    openAPIStringMapProperty("media object key map입니다."),
		"availableFileTypes": openAPIBoolMapProperty("사용 가능한 파일 타입 map입니다."),
		"createdAt":          openAPIDateTimeProperty("생성 시각입니다."),
		"updatedAt":          openAPIDateTimeProperty("수정 시각입니다."),
		"files": map[string]any{
			"type":  "array",
			"items": map[string]any{"$ref": "#/components/schemas/RecordingFile"},
		},
	})
}

func openAPIRecordingFileSchema() map[string]any {
	return openAPIObjectSchema("녹화 파일 상태입니다.", map[string]any{
		"type":        openAPIStringProperty("파일 타입입니다.", "rgb_audio_mp4"),
		"label":       openAPIStringProperty("표시 label입니다.", "RGB MP4"),
		"status":      openAPIStringProperty("파일 상태입니다.", "available"),
		"contentType": openAPIStringProperty("content type입니다.", "video/mp4"),
		"objectKey":   openAPIStringProperty("object storage key입니다.", "missions/mission-001/rgb.mp4"),
		"url":         openAPIStringProperty("다운로드 URL입니다.", "http://center.local:9000/bucket/object"),
	})
}

func openAPIRecorderTickRequestSchema() map[string]any {
	return openAPIObjectSchema("recorder tick 요청입니다.", map[string]any{
		"missionCode":          openAPIStringProperty("mission code입니다.", "mission-001"),
		"robotCode":            openAPIStringProperty("robot code입니다.", "robot-001"),
		"chunkDurationSeconds": map[string]any{"type": "integer", "example": 60},
		"tickAt":               openAPIDateTimeProperty("tick 시각입니다."),
	})
}

func openAPIRecorderUploadRequestSchema() map[string]any {
	return openAPIObjectSchema("recorder upload metadata입니다.", map[string]any{
		"sizeBytes": map[string]any{"type": "integer", "format": "int64", "example": 1048576},
		"checksum":  openAPIStringProperty("checksum입니다.", "sha256:..."),
		"workerId":  openAPIStringProperty("recorder-worker ID입니다.", "recorder-1"),
		"attempt":   map[string]any{"type": "integer", "example": 1},
	})
}

func openAPIRobotHeartbeatRequestSchema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "로봇이 주기적으로 전송하는 상태 보고입니다. 로봇 식별자는 body가 아니라 Authorization Bearer token에서 결정됩니다.",
		"properties": map[string]any{
			"state": map[string]any{
				"type":        "string",
				"description": "`online`, `offline`, `fault` 중 하나를 권장합니다. 빈 값이면 서버가 online으로 처리합니다.",
				"enum":        []string{"online", "offline", "fault"},
				"example":     "online",
			},
			"batteryPercent": map[string]any{
				"type":        "integer",
				"description": "로봇 배터리 잔량입니다. 0부터 100 사이 값을 권장합니다.",
				"minimum":     0,
				"maximum":     100,
				"example":     82,
			},
			"networkQuality": map[string]any{
				"type":        "string",
				"description": "로봇이 판단한 네트워크 상태 문자열입니다. 현재 서버는 값을 저장하되 enum으로 제한하지 않습니다.",
				"example":     "good",
			},
			"sentAt": map[string]any{
				"type":        "string",
				"format":      "date-time",
				"description": "로봇이 heartbeat를 보낸 시각입니다. RFC3339 형식을 사용합니다.",
				"example":     "2026-06-01T11:30:00Z",
			},
		},
	}
}

func openAPIRobotHeartbeatResponseSchema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "heartbeat 처리 결과입니다.",
		"required":    []string{"robotCode", "status", "serverTime"},
		"properties": map[string]any{
			"robotCode": map[string]any{
				"type":        "string",
				"description": "robotToken으로 식별된 로봇 코드입니다. 요청 body로 받은 값이 아닙니다.",
				"example":     "robot-004",
			},
			"status": map[string]any{
				"type":        "string",
				"description": "서버에 반영된 로봇 장치 상태입니다.",
				"example":     "online",
			},
			"serverTime": map[string]any{
				"type":        "string",
				"format":      "date-time",
				"description": "관제 서버가 응답을 생성한 시각입니다.",
				"example":     "2026-06-01T11:30:00.000000Z",
			},
		},
	}
}

func openAPIRobotMissionResponseSchema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "robotToken에 연결된 로봇의 active 임무와 송출 설정입니다. active 임무가 없으면 missionStatus와 serverTime만 반환됩니다.",
		"required":    []string{"missionStatus", "serverTime"},
		"properties": map[string]any{
			"missionCode": map[string]any{
				"type":        "string",
				"description": "현재 로봇이 배정된 active 임무 코드입니다. SFU room 값으로도 사용합니다.",
				"example":     "mission-002",
			},
			"missionStatus": map[string]any{
				"type":        "string",
				"description": "`active`이면 송출을 시작할 수 있고, `none`이면 현재 배정된 active 임무가 없습니다.",
				"enum":        []string{"active", "none"},
				"example":     "active",
			},
			"serverTime": map[string]any{
				"type":        "string",
				"format":      "date-time",
				"description": "관제 서버가 응답을 생성한 시각입니다.",
				"example":     "2026-06-01T11:30:00.000000Z",
			},
			"sfu": map[string]any{
				"$ref": "#/components/schemas/RobotSFUConfig",
			},
			"turnServers": map[string]any{
				"type":        "array",
				"description": "WebRTC PeerConnection 생성 시 사용하는 TURN 서버 목록입니다.",
				"items":       map[string]any{"$ref": "#/components/schemas/TurnServer"},
			},
			"tracks": map[string]any{
				"type":        "array",
				"description": "로봇이 offer 생성 전에 추가해야 하는 미디어 track role 목록입니다.",
				"items":       map[string]any{"type": "string"},
				"example":     []string{"track.video_1", "track.video_2", "track.audio_1", "track.audio_2"},
			},
			"dataChannels": map[string]any{
				"type":        "array",
				"description": "로봇이 offer 생성 전에 생성해야 하는 DataChannel label 목록입니다. 현재 payload 구조가 확정된 채널은 channel.telemetry입니다.",
				"items":       map[string]any{"type": "string"},
				"example":     []string{"channel.telemetry", "channel.spatial", "channel.event", "channel.control"},
			},
		},
	}
}

func openAPIRobotSFUConfigSchema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "로봇 publisher가 SFU에 연결하기 위한 signaling 설정입니다.",
		"required":    []string{"signalingUrl", "iceTransportPolicy"},
		"properties": map[string]any{
			"signalingUrl": map[string]any{
				"type":        "string",
				"description": "WebSocket signaling URL입니다. 로봇은 이 값을 그대로 사용하고 robotCode를 쿼리에 추가하지 않습니다.",
				"example":     "ws://192.168.20.12:18080/api/v1/robot/sfu/ws?room=mission-002",
			},
			"iceTransportPolicy": map[string]any{
				"type":        "string",
				"description": "ICE 후보 선택 정책입니다. `relay`이면 TURN relay candidate만 사용합니다.",
				"example":     "relay",
			},
		},
	}
}

func openAPITurnServerSchema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "WebRTC ICE 서버 설정입니다.",
		"required":    []string{"urls", "username", "credential"},
		"properties": map[string]any{
			"urls": map[string]any{
				"type":        "array",
				"description": "RTCIceServer.urls에 넣는 TURN URL 목록입니다.",
				"items":       map[string]any{"type": "string"},
				"example":     []string{"turn:192.168.20.12:3478?transport=udp"},
			},
			"username": map[string]any{
				"type":        "string",
				"description": "TURN 인증 사용자 이름입니다.",
				"example":     "robot",
			},
			"credential": map[string]any{
				"type":        "string",
				"description": "TURN 인증 비밀번호입니다.",
				"example":     "robot-pass",
			},
		},
	}
}

func openAPICreateRobotRequestSchema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "관제 서버에 로봇을 등록할 때 사용하는 요청입니다.",
		"required":    []string{"displayName"},
		"properties": map[string]any{
			"displayName": map[string]any{
				"type":        "string",
				"description": "관제 화면에서 표시할 로봇 이름입니다.",
				"example":     "Jetson 01",
			},
			"modelName": map[string]any{
				"type":        "string",
				"description": "로봇 모델명입니다. 선택 값입니다.",
				"example":     "Jetson Orin",
			},
		},
	}
}

func openAPICreateRobotResponseSchema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "생성된 로봇과 로봇 런타임 접속 정보입니다.",
		"required":    []string{"robot", "connectionInfo"},
		"properties": map[string]any{
			"robot":          map[string]any{"$ref": "#/components/schemas/Robot"},
			"connectionInfo": map[string]any{"$ref": "#/components/schemas/RobotConnectionInfo"},
		},
	}
}

func openAPIRobotSchema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "관제 서버에 등록된 로봇입니다.",
		"required":    []string{"id", "robotCode", "displayName", "status", "createdAt", "updatedAt"},
		"properties": map[string]any{
			"id":          openAPIStringProperty("관제 서버 내부 로봇 ID입니다.", "2f0af2f5-9f3b-4f02-a5d3-1f4c4a0e0001"),
			"robotCode":   openAPIStringProperty("관제 서버가 발급한 로봇 코드입니다.", "robot-004"),
			"displayName": openAPIStringProperty("로봇 표시 이름입니다.", "Jetson 01"),
			"modelName":   openAPIStringProperty("로봇 모델명입니다.", "Jetson Orin"),
			"status":      openAPIStringProperty("현재 연결 상태입니다.", "offline"),
			"lastSeenAt": map[string]any{
				"type":        "string",
				"format":      "date-time",
				"description": "마지막 heartbeat 수신 시각입니다.",
				"nullable":    true,
			},
			"createdAt": openAPIDateTimeProperty("로봇 등록 시각입니다."),
			"updatedAt": openAPIDateTimeProperty("로봇 정보 수정 시각입니다."),
		},
	}
}

func openAPIRobotConnectionInfoSchema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "로봇 런타임이 관제 서버에 접속할 때 사용하는 연결 정보입니다.",
		"required":    []string{"serverUrl", "robotCode", "robotToken"},
		"properties": map[string]any{
			"serverUrl":  openAPIStringProperty("관제 서버 public URL입니다.", "http://center.example.com"),
			"robotCode":  openAPIStringProperty("생성된 로봇 코드입니다. 런타임 API에서는 식별용 파라미터로 보내지 않습니다.", "robot-004"),
			"robotToken": openAPIStringProperty("로봇 런타임 API Authorization Bearer token 값입니다.", "rb_p0_example"),
		},
	}
}

func openAPICreateMissionRequestSchema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "임무 생성 요청입니다.",
		"required":    []string{"name", "missionType", "robotCodes"},
		"properties": map[string]any{
			"name": openAPIStringProperty("임무 표시 이름입니다.", "Mountain Rescue A"),
			"missionType": map[string]any{
				"type":        "string",
				"description": "임무 유형입니다. 현재 서버가 허용하는 값 중 하나를 사용합니다.",
				"enum":        []string{"mountain_rescue", "collapse_site", "underground_facility"},
				"example":     "mountain_rescue",
			},
			"siteNote": openAPIStringProperty("임무 위치나 작업 내용을 적는 메모입니다.", "북측 진입로 수색"),
			"robotCode": map[string]any{
				"type":        "string",
				"description": "기존 단일 로봇 입력 필드입니다. 새 호출에서는 robotCodes 배열 사용을 권장합니다.",
				"example":     "robot-004",
			},
			"robotCodes": map[string]any{
				"type":        "array",
				"description": "임무에 배정할 로봇 코드 목록입니다.",
				"items":       map[string]any{"type": "string"},
				"example":     []string{"robot-004"},
			},
		},
	}
}

func openAPIMissionResponseEnvelopeSchema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "임무 생성 또는 상태 변경 결과입니다.",
		"required":    []string{"mission"},
		"properties": map[string]any{
			"mission": map[string]any{"$ref": "#/components/schemas/Mission"},
		},
	}
}

func openAPIMissionSchema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "관제 서버에 등록된 임무입니다.",
		"required":    []string{"id", "missionCode", "name", "missionType", "status", "createdAt", "updatedAt"},
		"properties": map[string]any{
			"id":          openAPIStringProperty("관제 서버 내부 임무 ID입니다.", "8c68200a-b18f-4f1d-94f6-6f409e000001"),
			"missionCode": openAPIStringProperty("로봇 런타임 mission 조회 응답과 SFU room에 사용되는 임무 코드입니다.", "mission-002"),
			"name":        openAPIStringProperty("임무 표시 이름입니다.", "Mountain Rescue A"),
			"missionType": openAPIStringProperty("임무 유형입니다.", "mountain_rescue"),
			"status":      openAPIStringProperty("임무 상태입니다.", "active"),
			"siteNote":    openAPIStringProperty("임무 메모입니다.", "북측 진입로 수색"),
			"robotCode":   openAPIStringProperty("첫 번째 배정 로봇 코드입니다. 다중 로봇 호환을 위해 robotCodes도 함께 확인합니다.", "robot-004"),
			"robotCodes": map[string]any{
				"type":        "array",
				"description": "임무에 배정된 로봇 코드 목록입니다.",
				"items":       map[string]any{"type": "string"},
				"example":     []string{"robot-004"},
			},
			"startedAt": openAPINullableDateTimeProperty("임무 시작 시각입니다."),
			"endedAt":   openAPINullableDateTimeProperty("임무 종료 시각입니다."),
			"createdAt": openAPIDateTimeProperty("임무 생성 시각입니다."),
			"updatedAt": openAPIDateTimeProperty("임무 수정 시각입니다."),
		},
	}
}

func openAPIErrorResponseSchema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "공통 오류 응답입니다.",
		"required":    []string{"error"},
		"properties": map[string]any{
			"error": map[string]any{
				"type":        "string",
				"description": "오류 원인 메시지입니다.",
				"example":     "unauthorized",
			},
		},
	}
}

func openAPIObjectSchema(description string, properties map[string]any) map[string]any {
	return map[string]any{
		"type":        "object",
		"description": description,
		"properties":  properties,
	}
}

func openAPIGenericObjectSchema(description string) map[string]any {
	return openAPIObjectSchema(description, map[string]any{
		"data": openAPIGenericObjectProperty("응답 payload입니다."),
	})
}

func openAPIGenericObjectProperty(description string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"description":          description,
		"additionalProperties": true,
	}
}

func openAPIArrayEnvelopeSchema(fieldName string, itemReference string, description string) map[string]any {
	return openAPIObjectSchema(description, map[string]any{
		fieldName: map[string]any{
			"type":  "array",
			"items": map[string]any{"$ref": itemReference},
		},
	})
}

func openAPIStringMapProperty(description string) map[string]any {
	return map[string]any{
		"type":        "object",
		"description": description,
		"additionalProperties": map[string]any{
			"type": "string",
		},
	}
}

func openAPIBoolMapProperty(description string) map[string]any {
	return map[string]any{
		"type":        "object",
		"description": description,
		"additionalProperties": map[string]any{
			"type": "boolean",
		},
	}
}

func openAPIPathParameter(name string, description string) map[string]any {
	return map[string]any{
		"name":        name,
		"in":          "path",
		"required":    true,
		"description": description,
		"schema":      map[string]any{"type": "string"},
	}
}

func openAPIQueryParameter(name string, description string, required bool) map[string]any {
	return map[string]any{
		"name":        name,
		"in":          "query",
		"required":    required,
		"description": description,
		"schema":      map[string]any{"type": "string"},
	}
}

func openAPIIntegerQueryParameter(name string, description string, required bool) map[string]any {
	return map[string]any{
		"name":        name,
		"in":          "query",
		"required":    required,
		"description": description,
		"schema":      map[string]any{"type": "integer"},
	}
}

func openAPISensorQueryParameters(required bool) []map[string]any {
	return []map[string]any{
		openAPIQueryParameter("missionId", "mission ID 또는 missionCode입니다.", required),
		openAPIQueryParameter("robotCode", "robot code입니다.", required),
	}
}

func openAPIStringProperty(description string, example string) map[string]any {
	return map[string]any{
		"type":        "string",
		"description": description,
		"example":     example,
	}
}

func openAPIDateTimeProperty(description string) map[string]any {
	return map[string]any{
		"type":        "string",
		"format":      "date-time",
		"description": description,
		"example":     "2026-06-01T11:30:00Z",
	}
}

func openAPINullableDateTimeProperty(description string) map[string]any {
	property := openAPIDateTimeProperty(description)
	property["nullable"] = true
	return property
}
