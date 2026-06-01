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
			"description": "관제 서버에서 외부 로봇이 사용하는 로봇 전용 API와 개발 서버 테스트 슬롯 생성에 필요한 최소 API를 정리한 문서입니다. 로봇 런타임은 `/api/v1/robot/*` 하위 경로만 호출합니다.",
		},
		"servers": []map[string]any{
			{
				"url":         openAPIServerURL(s.config.PublicURL),
				"description": "관제 API 서버",
			},
		},
		"tags": []map[string]any{
			{
				"name":        "로봇 런타임 API",
				"description": "실제 로봇 또는 Mock Robot이 런타임 중 호출하는 Bearer token 기반 self-scope API입니다.",
			},
			{
				"name":        "개발 서버 테스트 준비",
				"description": "개발 서버에서 로봇팀 테스트 슬롯을 만들고 임무를 시작하기 위한 관제팀 관리 API입니다. 로봇 런타임에서 호출하지 않습니다.",
			},
		},
		"paths": map[string]any{
			"/api/v1/robot/heartbeat": openAPIRobotHeartbeatPath(),
			"/api/v1/robot/mission":   openAPIRobotMissionPath(),
			"/api/v1/robot/sfu/ws":    openAPIRobotSFUWebSocketPath(),
			"/api/robots":             openAPIRobotsPath(),
			"/api/missions":           openAPIMissionsPath(),
			"/api/missions/{missionCode}/start": openAPIMissionStatePath(
				"startMission",
				"임무 시작",
				"생성된 임무를 active 상태로 전환합니다. 로봇은 이후 `/api/v1/robot/mission`에서 자기 토큰에 연결된 active 임무만 조회합니다.",
			),
			"/api/missions/{missionCode}/end": openAPIMissionStatePath(
				"endMission",
				"임무 종료",
				"active 임무를 ended 상태로 전환하고 해당 임무의 SFU room을 닫습니다.",
			),
		},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"robotBearerAuth": map[string]any{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "robotToken",
					"description":  "`POST /api/robots` 응답의 `connectionInfo.robotToken` 값입니다. 서버는 이 토큰으로 로봇 신원을 판별하므로 robotCode, robotId, roomId, sessionId를 별도 파라미터로 받지 않습니다.",
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
		"post": map[string]any{
			"tags":        []string{"개발 서버 테스트 준비"},
			"operationId": "createRobot",
			"summary":     "테스트용 로봇 생성",
			"description": "개발 서버에서 테스트에 사용할 로봇 슬롯을 생성하고 robotToken을 발급합니다. 이 API는 관제팀 관리 API이며 로봇 런타임이 호출하는 API가 아닙니다.",
			"requestBody": map[string]any{
				"required": true,
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{"$ref": "#/components/schemas/CreateRobotRequest"},
					},
				},
			},
			"responses": map[string]any{
				"201": openAPIJSONResponse("로봇 슬롯과 연결 정보가 생성됐습니다.", "#/components/schemas/CreateRobotResponse"),
				"400": openAPIErrorResponse("필수 값이 없거나 요청 body가 잘못됐습니다."),
			},
		},
	}
}

func openAPIMissionsPath() map[string]any {
	return map[string]any{
		"post": map[string]any{
			"tags":        []string{"개발 서버 테스트 준비"},
			"operationId": "createMission",
			"summary":     "테스트용 임무 생성",
			"description": "테스트할 로봇을 배정한 임무를 생성합니다. 여러 로봇 테스트는 robotCodes 배열을 사용합니다.",
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

func openAPIMissionStatePath(operationID string, summary string, description string) map[string]any {
	return map[string]any{
		"post": map[string]any{
			"tags":        []string{"개발 서버 테스트 준비"},
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
		"RobotHeartbeatRequest":   openAPIRobotHeartbeatRequestSchema(),
		"RobotHeartbeatResponse":  openAPIRobotHeartbeatResponseSchema(),
		"RobotMissionResponse":    openAPIRobotMissionResponseSchema(),
		"RobotSFUConfig":          openAPIRobotSFUConfigSchema(),
		"TurnServer":              openAPITurnServerSchema(),
		"CreateRobotRequest":      openAPICreateRobotRequestSchema(),
		"CreateRobotResponse":     openAPICreateRobotResponseSchema(),
		"Robot":                   openAPIRobotSchema(),
		"RobotConnectionInfo":     openAPIRobotConnectionInfoSchema(),
		"CreateMissionRequest":    openAPICreateMissionRequestSchema(),
		"MissionResponseEnvelope": openAPIMissionResponseEnvelopeSchema(),
		"Mission":                 openAPIMissionSchema(),
		"ErrorResponse":           openAPIErrorResponseSchema(),
	}
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
				"description": "ICE 후보 선택 정책입니다. 개발 서버 테스트는 TURN relay 경로 검증을 위해 relay를 사용합니다.",
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
		"description": "개발 서버에서 테스트용 로봇 슬롯을 만들 때 사용하는 요청입니다.",
		"required":    []string{"displayName"},
		"properties": map[string]any{
			"displayName": map[string]any{
				"type":        "string",
				"description": "관제 화면과 테스트 문서에서 구분할 로봇 표시 이름입니다.",
				"example":     "Robot Team Jetson 01",
			},
			"modelName": map[string]any{
				"type":        "string",
				"description": "로봇 모델 또는 테스트 장비 이름입니다. 선택 값입니다.",
				"example":     "Jetson Orin",
			},
		},
	}
}

func openAPICreateRobotResponseSchema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "생성된 로봇 슬롯과 로봇 런타임 접속 정보입니다.",
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
		"description": "관제 서버에 등록된 로봇 슬롯입니다.",
		"required":    []string{"id", "robotCode", "displayName", "status", "createdAt", "updatedAt"},
		"properties": map[string]any{
			"id":          openAPIStringProperty("관제 서버 내부 로봇 ID입니다.", "2f0af2f5-9f3b-4f02-a5d3-1f4c4a0e0001"),
			"robotCode":   openAPIStringProperty("관제 서버가 발급한 로봇 코드입니다.", "robot-004"),
			"displayName": openAPIStringProperty("로봇 표시 이름입니다.", "Robot Team Jetson 01"),
			"modelName":   openAPIStringProperty("로봇 모델 또는 테스트 장비 이름입니다.", "Jetson Orin"),
			"status":      openAPIStringProperty("현재 연결 상태입니다.", "offline"),
			"lastSeenAt": map[string]any{
				"type":        "string",
				"format":      "date-time",
				"description": "마지막 heartbeat 수신 시각입니다.",
				"nullable":    true,
			},
			"createdAt": openAPIDateTimeProperty("로봇 슬롯 생성 시각입니다."),
			"updatedAt": openAPIDateTimeProperty("로봇 슬롯 수정 시각입니다."),
		},
	}
}

func openAPIRobotConnectionInfoSchema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "로봇팀 테스트 클라이언트가 런타임 API에 접속할 때 필요한 정보입니다.",
		"required":    []string{"serverUrl", "robotCode", "robotToken"},
		"properties": map[string]any{
			"serverUrl":  openAPIStringProperty("관제 서버 Public URL입니다.", "http://192.168.20.12:18080"),
			"robotCode":  openAPIStringProperty("생성된 로봇 코드입니다. 런타임 API에서는 식별용 파라미터로 보내지 않습니다.", "robot-004"),
			"robotToken": openAPIStringProperty("로봇 런타임 API Authorization Bearer token 값입니다.", "rb_p0_example"),
		},
	}
}

func openAPICreateMissionRequestSchema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "테스트용 임무 생성 요청입니다.",
		"required":    []string{"name", "missionType", "robotCodes"},
		"properties": map[string]any{
			"name": openAPIStringProperty("임무 표시 이름입니다.", "Robot Team WebRTC Test"),
			"missionType": map[string]any{
				"type":        "string",
				"description": "임무 유형입니다. 현재 서버가 허용하는 값 중 하나를 사용합니다.",
				"enum":        []string{"mountain_rescue", "collapse_site", "underground_facility"},
				"example":     "mountain_rescue",
			},
			"siteNote": openAPIStringProperty("테스트 위치나 목적을 적는 메모입니다.", "WebRTC 영상/센서 송출 테스트"),
			"robotCode": map[string]any{
				"type":        "string",
				"description": "단일 로봇 테스트용 legacy 입력입니다. 새 테스트는 robotCodes 배열 사용을 권장합니다.",
				"example":     "robot-004",
			},
			"robotCodes": map[string]any{
				"type":        "array",
				"description": "임무에 배정할 로봇 코드 목록입니다. 여러 명이 테스트하면 각자 생성한 로봇 코드를 넣습니다.",
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
			"name":        openAPIStringProperty("임무 표시 이름입니다.", "Robot Team WebRTC Test"),
			"missionType": openAPIStringProperty("임무 유형입니다.", "mountain_rescue"),
			"status":      openAPIStringProperty("임무 상태입니다.", "active"),
			"siteNote":    openAPIStringProperty("임무 메모입니다.", "WebRTC 영상/센서 송출 테스트"),
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
