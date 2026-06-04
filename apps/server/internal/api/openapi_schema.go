package api

func openAPISchemas() map[string]any {
	return map[string]any{
		"HealthResponse":        openAPIHealthResponseSchema(),
		"SystemStatusResponse":  openAPISystemStatusResponseSchema(),
		"SystemComponentStatus": openAPISystemComponentStatusSchema(),
		"SystemConfig":          openAPISystemConfigSchema(),
		"ObjectStorageStatus":   openAPIObjectStorageStatusSchema(),
		"SystemSummary":         openAPISystemSummarySchema(),
		"ObjectStorageResponse": openAPIGenericObjectSchema("Object Storage 작업 결과입니다."),
		"ClearObjectStorageRequest": openAPIObjectSchema("Object Storage 초기화 요청입니다.", map[string]any{
			"confirmation": openAPIStringProperty("초기화 확인 문자열입니다.", "CLEAR_OBJECT_STORAGE"),
		}),
		"SensorDataClearResponse": openAPIObjectSchema("Sensor 데이터 초기화 결과입니다.", map[string]any{
			"sensorData": map[string]any{"$ref": "#/components/schemas/SensorDataClearResult"},
		}),
		"SensorDataClearResult": openAPIObjectSchema("삭제된 sensor 데이터 row 수입니다.", map[string]any{
			"sensorLatestSamplesDeleted": map[string]any{"type": "integer", "format": "int64", "example": 120000},
			"sensorSamplesDeleted":       map[string]any{"type": "integer", "format": "int64", "example": 120000},
			"sensorDescriptorsDeleted":   map[string]any{"type": "integer", "format": "int64", "example": 18},
		}),
		"ClearSensorDataRequest": openAPIObjectSchema("Sensor 데이터 초기화 요청입니다.", map[string]any{
			"confirmation": openAPIStringProperty("초기화 확인 문자열입니다.", "CLEAR_SENSOR_DATA"),
		}),
		"RTCConfigResponse":                openAPIRTCConfigResponseSchema(),
		"RecorderRecordingTargetsResponse": openAPIArrayEnvelopeSchema("targets", "#/components/schemas/RecorderRecordingTarget", "recorder-worker 내부 녹화 대상 임무 목록입니다."),
		"RecorderRecordingTarget":          openAPIRecorderRecordingTargetSchema(),
		"SensorDescriptorsResponse":        openAPIArrayEnvelopeSchema("sensorDescriptors", "#/components/schemas/SensorDescriptor", "센서 descriptor 목록입니다."),
		"SensorSamplesResponse":            openAPIArrayEnvelopeSchema("sensorSamples", "#/components/schemas/SensorSample", "센서 sample 목록입니다."),
		"SensorLatestEnvelope":             openAPISensorLatestEnvelopeSchema(),
		"SensorEnvelopeRequest":            openAPISensorEnvelopeRequestSchema(),
		"SensorDescriptor":                 openAPISensorDescriptorSchema(),
		"SensorSample":                     openAPISensorSampleSchema(),
		"SensorLatest":                     openAPISensorLatestSchema(),
		"SensorValueReading":               openAPISensorValueReadingSchema(),
		"OperatorRecordingsResponse":       openAPIArrayEnvelopeSchema("recordings", "#/components/schemas/OperatorRecordingChunk", "관제 UI가 조회하는 녹화 chunk 목록입니다."),
		"OperatorRecordingChunk":           openAPIOperatorRecordingChunkSchema(),
		"OperatorRecordingFile":            openAPIOperatorRecordingFileSchema(),
		"RecorderRecordingChunkEnvelope":   openAPIObjectSchema("recorder-worker 내부 녹화 chunk 응답입니다.", map[string]any{"chunk": map[string]any{"$ref": "#/components/schemas/RecorderRecordingChunk"}}),
		"RecorderRecordingChunk":           openAPIRecorderRecordingChunkSchema(),
		"RecorderTickRequest":              openAPIRecorderTickRequestSchema(),
		"RecorderRecordingTickResponse":    openAPIObjectSchema("recorder-worker 내부 녹화 tick 처리 결과입니다.", map[string]any{"chunk": map[string]any{"$ref": "#/components/schemas/RecorderRecordingChunk"}, "manifest": openAPIGenericObjectProperty("녹화 manifest")}),
		"RecorderUploadRequest":            openAPIRecorderUploadRequestSchema(),
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
		"RecorderFinalizationJobsResponse": openAPIObjectSchema("claim된 finalization job 목록입니다.", map[string]any{
			"jobs": map[string]any{
				"type":  "array",
				"items": map[string]any{"$ref": "#/components/schemas/RecorderFinalizationJob"},
			},
		}),
		"RecorderFinalizationJob":     openAPIRecorderFinalizationJobSchema(),
		"OKResponse":                  openAPIObjectSchema("처리 결과입니다.", map[string]any{"ok": map[string]any{"type": "boolean", "example": true}}),
		"RobotsResponse":              openAPIArrayEnvelopeSchema("robots", "#/components/schemas/Robot", "로봇 목록입니다."),
		"RobotEnvelope":               openAPIObjectSchema("로봇 응답입니다.", map[string]any{"robot": map[string]any{"$ref": "#/components/schemas/Robot"}}),
		"UpdateRobotRequest":          openAPICreateRobotRequestSchema(),
		"RobotConnectionInfoEnvelope": openAPIObjectSchema("로봇 연결 정보 응답입니다.", map[string]any{"connectionInfo": map[string]any{"$ref": "#/components/schemas/RobotConnectionInfo"}}),
		"MissionsResponse":            openAPIArrayEnvelopeSchema("missions", "#/components/schemas/Mission", "임무 목록입니다."),
		"MissionLiveStatusResponse":   openAPIMissionLiveStatusResponseSchema(),
		"RobotHeartbeatRequest":       openAPIRobotHeartbeatRequestSchema(),
		"RobotHeartbeatResponse":      openAPIRobotHeartbeatResponseSchema(),
		"RobotMissionResponse":        openAPIRobotMissionResponseSchema(),
		"RobotSFUConfig":              openAPIRobotSFUConfigSchema(),
		"TurnServer":                  openAPITurnServerSchema(),
		"CreateRobotRequest":          openAPICreateRobotRequestSchema(),
		"CreateRobotResponse":         openAPICreateRobotResponseSchema(),
		"Robot":                       openAPIRobotSchema(),
		"RobotConnectionInfo":         openAPIRobotConnectionInfoSchema(),
		"CreateMissionRequest":        openAPICreateMissionRequestSchema(),
		"MissionResponseEnvelope":     openAPIMissionResponseEnvelopeSchema(),
		"Mission":                     openAPIMissionSchema(),
		"ErrorResponse":               openAPIErrorResponseSchema(),
	}
}

func openAPIHealthResponseSchema() map[string]any {
	return openAPIObjectSchema("health 응답입니다.", map[string]any{
		"status":    openAPIStringProperty("상태입니다.", "ok"),
		"service":   openAPIStringProperty("서비스 이름입니다.", "app-server"),
		"startedAt": openAPIDateTimeProperty("서비스 시작 시각입니다."),
	})
}

func openAPISystemStatusResponseSchema() map[string]any {
	return openAPIObjectSchema("시스템 상태 응답입니다.", map[string]any{
		"service": openAPIStringProperty("서비스 이름입니다.", "app-server"),
		"status":  openAPIStringProperty("시스템 상태입니다.", "ok"),
		"components": map[string]any{
			"type":  "array",
			"items": map[string]any{"$ref": "#/components/schemas/SystemComponentStatus"},
		},
		"config":        map[string]any{"$ref": "#/components/schemas/SystemConfig"},
		"objectStorage": map[string]any{"$ref": "#/components/schemas/ObjectStorageStatus"},
		"summary":       map[string]any{"$ref": "#/components/schemas/SystemSummary"},
		"sfuRooms": map[string]any{
			"type":        "array",
			"description": "SFU room summary 목록입니다.",
			"items":       openAPIGenericObjectProperty("SFU room summary입니다."),
		},
	})
}

func openAPISystemComponentStatusSchema() map[string]any {
	return openAPIObjectSchema("시스템 구성요소 상태입니다.", map[string]any{
		"name":   openAPIStringProperty("구성요소 이름입니다.", "recorder-worker"),
		"status": openAPIStringProperty("구성요소 상태입니다.", "ok"),
	})
}

func openAPISystemConfigSchema() map[string]any {
	return openAPIObjectSchema("시스템 설정 요약입니다.", map[string]any{
		"environment":               openAPIStringProperty("실행 환경입니다.", "development"),
		"appServerPublicUrl":        openAPIStringProperty("브라우저/로봇이 접근하는 app-server 외부 URL입니다.", "http://center.local:18080"),
		"recorderWorkerInternalUrl": openAPIStringProperty("app-server가 Docker 내부에서 recorder-worker에 접근하는 URL입니다.", "http://recorder-worker:8082"),
		"minioInternalUrl":          openAPIStringProperty("서버 컴포넌트가 Docker 내부에서 Object Storage에 접근하는 URL입니다.", "http://minio:9000"),
		"minioPublicUrl":            openAPIStringProperty("브라우저가 녹화 파일을 읽는 Object Storage 외부 URL입니다.", "http://center.local:19000"),
		"minioBucket":               openAPIStringProperty("Object Storage bucket입니다.", "robot-center-poc"),
	})
}

func openAPIObjectStorageStatusSchema() map[string]any {
	return openAPIObjectSchema("Object Storage 사용량 상태입니다.", map[string]any{
		"status":             openAPIStringProperty("Object Storage 상태입니다.", "ok"),
		"bucket":             openAPIStringProperty("bucket 이름입니다.", "robot-center-poc"),
		"objectCount":        map[string]any{"type": "integer", "example": 140},
		"bucketUsedBytes":    map[string]any{"type": "integer", "format": "int64", "example": 1303059484},
		"totalBytes":         map[string]any{"type": "integer", "format": "int64", "example": 741548665884},
		"usedBytes":          map[string]any{"type": "integer", "format": "int64", "example": 1303059484},
		"availableBytes":     map[string]any{"type": "integer", "format": "int64", "example": 740245606400},
		"usedPercent":        map[string]any{"type": "number", "example": 0.1757},
		"diskTotalBytes":     map[string]any{"type": "integer", "format": "int64", "example": 759793618944},
		"diskUsedBytes":      map[string]any{"type": "integer", "format": "int64", "example": 19548012544},
		"diskAvailableBytes": map[string]any{"type": "integer", "format": "int64", "example": 740245606400},
		"diskUsedPercent":    map[string]any{"type": "number", "example": 2.5728},
		"error":              openAPIStringProperty("Object Storage 조회 오류입니다.", "minio unavailable"),
	})
}

func openAPISystemSummarySchema() map[string]any {
	return openAPIObjectSchema("관제 서버 데이터 요약입니다.", map[string]any{
		"robots":     map[string]any{"type": "integer", "example": 5},
		"missions":   map[string]any{"type": "integer", "example": 15},
		"sfuRooms":   map[string]any{"type": "integer", "example": 1},
		"recordings": map[string]any{"type": "integer", "example": 265},
	})
}

func openAPIRTCConfigResponseSchema() map[string]any {
	return openAPIObjectSchema("관제 WebRTC 설정 응답입니다.", map[string]any{
		"mode":                 openAPIStringProperty("WebRTC 모드입니다.", "sfu"),
		"signalingUrl":         openAPIStringProperty("operator signaling URL입니다.", "ws://center.local/api/v1/operator/sfu/ws"),
		"operatorSignalingUrl": openAPIStringProperty("operator signaling URL입니다.", "ws://center.local/api/v1/operator/sfu/ws"),
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
	schema["properties"].(map[string]any)["readings"] = map[string]any{
		"type":        "array",
		"description": "sensorType과 label 기준으로 해석한 표시용 readings입니다. 원본 values는 latestSample.values에 그대로 유지됩니다.",
		"items":       map[string]any{"$ref": "#/components/schemas/SensorValueReading"},
	}
	return schema
}

func openAPISensorValueReadingSchema() map[string]any {
	return openAPIObjectSchema("센서 sample values를 표시/해석용으로 풀어낸 reading입니다.", map[string]any{
		"fieldPath": openAPIStringProperty("values 내부 field path입니다.", "concentration"),
		"label":     openAPIStringProperty("표시 label입니다.", "CO"),
		"order":     openAPINumberProperty("관제 화면 표시 순서입니다.", 10.01),
		"unit":      openAPIStringProperty("표시 단위입니다.", "ppm"),
		"value":     openAPIGenericObjectProperty("표시 값입니다."),
	})
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

func openAPIRecorderRecordingTargetSchema() map[string]any {
	return openAPIObjectSchema("recorder-worker 내부 녹화 대상 임무입니다.", map[string]any{
		"id":          openAPIStringProperty("mission ID입니다.", "uuid"),
		"missionCode": openAPIStringProperty("mission code입니다.", "mission-001"),
		"name":        openAPIStringProperty("임무 이름입니다.", "산악 구조"),
		"missionType": openAPIStringProperty("임무 유형입니다.", "mountain_rescue"),
		"status":      openAPIStringProperty("임무 상태입니다.", "active"),
		"siteNote":    openAPIStringProperty("현장 메모입니다.", "북쪽 능선"),
		"robotCode":   openAPIStringProperty("대표 robot code입니다.", "robot-001"),
		"robotCodes":  openAPIStringArrayProperty("배정된 robot code 목록입니다."),
		"startedAt":   openAPINullableDateTimeProperty("시작 시각입니다."),
		"endedAt":     openAPINullableDateTimeProperty("종료 시각입니다."),
		"createdAt":   openAPIDateTimeProperty("생성 시각입니다."),
		"updatedAt":   openAPIDateTimeProperty("수정 시각입니다."),
	})
}

func openAPIOperatorRecordingChunkSchema() map[string]any {
	properties := openAPIRecordingChunkProperties()
	properties["files"] = map[string]any{
		"type":  "array",
		"items": map[string]any{"$ref": "#/components/schemas/OperatorRecordingFile"},
	}
	return openAPIObjectSchema("관제 UI가 조회하는 녹화 chunk입니다.", properties)
}

func openAPIRecorderRecordingChunkSchema() map[string]any {
	return openAPIObjectSchema("recorder-worker 내부 녹화 chunk입니다.", openAPIRecordingChunkProperties())
}

func openAPIRecordingChunkProperties() map[string]any {
	return map[string]any{
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
	}
}

func openAPIOperatorRecordingFileSchema() map[string]any {
	return openAPIObjectSchema("관제 UI가 조회하는 녹화 파일 상태입니다.", map[string]any{
		"type":        openAPIStringProperty("파일 타입입니다.", "rgb_audio_mp4"),
		"label":       openAPIStringProperty("표시 label입니다.", "RGB MP4"),
		"status":      openAPIStringProperty("파일 상태입니다.", "available"),
		"contentType": openAPIStringProperty("content type입니다.", "video/mp4"),
		"objectKey":   openAPIStringProperty("object storage key입니다.", "missions/mission-001/rgb.mp4"),
		"url":         openAPIStringProperty("브라우저가 접근하는 외부 다운로드 URL입니다.", "http://center.local:19000/bucket/object"),
	})
}

func openAPIRecorderFinalizationJobSchema() map[string]any {
	return openAPIObjectSchema("recorder-worker가 claim한 녹화 finalization job입니다.", map[string]any{
		"id":                 openAPIStringProperty("finalization job ID입니다.", "uuid"),
		"recordingChunkId":   openAPIStringProperty("recording chunk ID입니다.", "uuid"),
		"recordingSessionId": openAPIStringProperty("recording session ID입니다.", "uuid"),
		"missionId":          openAPIStringProperty("mission ID입니다.", "uuid"),
		"robotId":            openAPIStringProperty("robot ID입니다.", "uuid"),
		"status":             openAPIStringProperty("finalization job 상태입니다.", "claimed"),
		"reason":             openAPIStringProperty("부분 완료 또는 실패 사유입니다.", "missing media"),
		"attempts":           map[string]any{"type": "integer", "example": 1},
		"lockedBy":           openAPIStringProperty("claim한 recorder-worker ID입니다.", "recorder-1"),
		"lockedUntil":        openAPINullableDateTimeProperty("claim lock 만료 시각입니다."),
		"lastError":          openAPIStringProperty("마지막 오류 메시지입니다.", "upload failed"),
		"createdAt":          openAPIDateTimeProperty("생성 시각입니다."),
		"updatedAt":          openAPIDateTimeProperty("수정 시각입니다."),
		"completedAt":        openAPINullableDateTimeProperty("완료 시각입니다."),
		"chunk":              map[string]any{"$ref": "#/components/schemas/RecorderRecordingChunk"},
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
			"robotToken": openAPIStringProperty("로봇 API Authorization Bearer token 값입니다.", "rb_p0_example"),
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
