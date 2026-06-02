package api

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

func openAPISensorDataClearPath() map[string]any {
	return map[string]any{
		"post": openAPIOperation("시스템 API", "clearSensorData", "Sensor 데이터 초기화", "확인 문자열을 받은 뒤 테스트용 sensor descriptor와 sample 데이터를 정리합니다. production 환경에서는 실행되지 않습니다.", "SensorDataClearResponse", "ClearSensorDataRequest"),
	}
}

func openAPIRTCConfigPath() map[string]any {
	return map[string]any{
		"get": openAPIOperation("Operator API", "getRTCConfig", "관제 WebRTC 설정 조회", "관제 UI operator peer가 사용할 signaling URL과 ICE 서버 설정을 반환합니다.", "RTCConfigResponse", ""),
	}
}

func openAPIRecordingTargetsPath() map[string]any {
	return map[string]any{
		"get": openAPIOperation("Recorder API", "listRecordingTargets", "녹화 대상 임무 조회", "recorder-worker가 구독해야 하는 active mission 목록을 반환합니다.", "RecordingTargetsResponse", ""),
	}
}

func openAPISensorDescriptorsPath() map[string]any {
	return map[string]any{
		"get": openAPIOperationWithParameters("Operator API", "listSensorDescriptors", "센서 descriptor 조회", "missionId, robotCode 조건으로 센서 descriptor 목록을 조회합니다.", "SensorDescriptorsResponse", "", openAPISensorQueryParameters(false)),
	}
}

func openAPISensorSamplesPath() map[string]any {
	return map[string]any{
		"get": openAPIOperationWithParameters("Operator API", "listSensorSamples", "센서 sample 조회", "missionId, robotCode, sensorId 조건과 limit으로 센서 sample 목록을 조회합니다.", "SensorSamplesResponse", "", append(openAPISensorQueryParameters(false), openAPIQueryParameter("sensorId", "조회할 sensorId", false), openAPIIntegerQueryParameter("limit", "조회 개수 제한", false))),
	}
}

func openAPIRecorderSensorSamplesPath() map[string]any {
	return map[string]any{
		"post": openAPIOperation("Recorder API", "createSensorSamples", "센서 sample 저장", "recorder-worker가 DataChannel sensor envelope를 저장합니다.", "SensorSamplesResponse", "SensorEnvelopeRequest"),
	}
}

func openAPISensorLatestPath() map[string]any {
	return map[string]any{
		"get": openAPIOperationWithParameters("Operator API", "listSensorLatest", "센서 latest 조회", "missionId, robotCode 조건으로 센서별 최신 sample을 조회합니다.", "SensorLatestEnvelope", "", openAPISensorQueryParameters(false)),
	}
}

func openAPIRecordingsPath() map[string]any {
	return map[string]any{
		"get": openAPIOperation("Operator API", "listRecordings", "녹화 chunk 조회", "저장된 recording chunk와 파일 상태 목록을 반환합니다.", "RecordingsResponse", ""),
	}
}

func openAPIRecorderTickPath() map[string]any {
	return map[string]any{
		"post": openAPIOperation("Recorder API", "applyRecorderTick", "녹화 tick 반영", "recorder-worker가 mission/robot 기준 chunk 생성을 요청합니다.", "RecordingTickResponse", "RecorderTickRequest"),
	}
}

func openAPIRecorderFinalizationJobsClaimPath() map[string]any {
	return map[string]any{
		"post": openAPIOperation("Recorder API", "claimRecorderFinalizationJobs", "녹화 finalization job claim", "recorder-worker가 처리할 finalization job을 claim합니다.", "RecorderFinalizationJobsResponse", "RecorderFinalizationClaimRequest"),
	}
}

func openAPIRecorderFinalizationJobStatusPath(operationID string, summary string) map[string]any {
	return map[string]any{
		"post": openAPIOperationWithParameters("Recorder API", operationID, summary, "recorder-worker가 finalization job 처리 결과를 보고합니다.", "OKResponse", "RecorderFinalizationStatusRequest", []map[string]any{openAPIPathParameter("jobID", "finalization job ID")}),
	}
}

func openAPIRecorderChunkUploadedPath() map[string]any {
	return map[string]any{
		"post": openAPIOperationWithParameters("Recorder API", "markRecorderChunkUploaded", "녹화 chunk 업로드 완료", "recorder-worker가 chunk manifest 업로드 완료를 보고합니다.", "RecordingChunkEnvelope", "RecorderUploadRequest", []map[string]any{openAPIPathParameter("chunkID", "recording chunk ID")}),
	}
}

func openAPIRecorderFileUploadedPath() map[string]any {
	return map[string]any{
		"post": openAPIOperationWithParameters("Recorder API", "markRecorderFileUploaded", "녹화 파일 업로드 완료", "recorder-worker가 chunk의 개별 파일 업로드 완료를 보고합니다.", "RecordingChunkEnvelope", "RecorderUploadRequest", []map[string]any{openAPIPathParameter("chunkID", "recording chunk ID"), openAPIPathParameter("fileType", "업로드된 파일 타입")}),
	}
}

func openAPIRobotHeartbeatPath() map[string]any {
	return map[string]any{
		"post": map[string]any{
			"tags":        []string{"로봇 API"},
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
			"tags":        []string{"로봇 API"},
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
			"tags":        []string{"로봇 API"},
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
		"get": openAPIOperation("Operator API", "listRobots", "로봇 목록 조회", "관제 서버에 등록된 로봇 목록을 반환합니다.", "RobotsResponse", ""),
		"post": map[string]any{
			"tags":        []string{"Operator API"},
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
		"patch":  openAPIOperationWithParameters("Operator API", "updateRobot", "로봇 수정", "로봇 표시 이름과 모델명을 수정합니다.", "RobotEnvelope", "UpdateRobotRequest", parameters),
		"delete": openAPIOperationWithParameters("Operator API", "archiveRobot", "로봇 보관 처리", "로봇을 active 목록에서 제외합니다.", "RobotEnvelope", "", parameters),
	}
}

func openAPIRobotConnectionInfoPath() map[string]any {
	return map[string]any{
		"get": openAPIOperationWithParameters("Operator API", "getRobotConnectionInfo", "로봇 연결 정보 조회", "로봇 런타임 접속에 필요한 serverUrl, robotCode, robotToken을 조회합니다.", "RobotConnectionInfoEnvelope", "", []map[string]any{openAPIPathParameter("robotCode", "로봇 코드")}),
	}
}

func openAPIRobotConnectionTokenPath() map[string]any {
	return map[string]any{
		"post": openAPIOperationWithParameters("Operator API", "rotateRobotConnectionToken", "로봇 token 재발급", "로봇 API용 robotToken을 재발급합니다.", "RobotConnectionInfoEnvelope", "", []map[string]any{openAPIPathParameter("robotCode", "로봇 코드")}),
	}
}

func openAPIMissionsPath() map[string]any {
	return map[string]any{
		"get": openAPIOperation("Operator API", "listMissions", "임무 목록 조회", "관제 서버에 등록된 임무 목록을 반환합니다.", "MissionsResponse", ""),
		"post": map[string]any{
			"tags":        []string{"Operator API"},
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
		"get": openAPIOperationWithParameters("Operator API", "getMissionLiveStatus", "임무 live status 조회", "임무에 배정된 로봇의 연결, 스트림, 녹화 상태를 반환합니다.", "MissionLiveStatusResponse", "", []map[string]any{openAPIPathParameter("missionCode", "임무 코드")}),
	}
}

func openAPIMissionStatePath(operationID string, summary string, description string) map[string]any {
	return map[string]any{
		"post": map[string]any{
			"tags":        []string{"Operator API"},
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

func openAPISFUWebSocketPath(tag string, operationID string, summary string, description string) map[string]any {
	return map[string]any{
		"get": map[string]any{
			"tags":        []string{tag},
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
